package updates

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func signed(t *testing.T, m Manifest) ([]byte, Trust) {
	t.Helper()
	public, private, e := ed25519.GenerateKey(rand.Reader)
	if e != nil {
		t.Fatal(e)
	}
	payload, _ := json.Marshal(m)
	envelope, _ := json.Marshal(Envelope{Payload: base64.StdEncoding.EncodeToString(payload), Signatures: []Signature{{KeyID: "test", Value: base64.StdEncoding.EncodeToString(ed25519.Sign(private, payload))}}})
	return envelope, Trust{Keys: map[string]string{"test": base64.StdEncoding.EncodeToString(public)}}
}
func archive(t *testing.T, files map[string][]byte) ([]byte, Asset) {
	t.Helper()
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	a := Asset{URL: "https://releases.example.test/update.zip"}
	for name, data := range files {
		w, e := z.Create(name)
		if e != nil {
			t.Fatal(e)
		}
		w.Write(data)
		a.Files = append(a.Files, File{Name: name, SHA256: SHA(data), Size: int64(len(data))})
	}
	z.Close()
	a.Size = int64(b.Len())
	a.SHA256 = SHA(b.Bytes())
	return b.Bytes(), a
}
func manifest(a Asset) Manifest {
	now := time.Now().UTC()
	return Manifest{Schema: 1, Product: "smolpng", Version: "0.2.0", Commit: strings.Repeat("a", 40), Channel: "stable", Issued: now.Add(-time.Minute), Expires: now.Add(time.Hour), Assets: map[string]Asset{"linux/amd64": a}}
}
func TestSignedManifestBoundaries(t *testing.T) {
	_, a := archive(t, map[string][]byte{"smolpng": []byte("new"), "LICENSE": []byte("MIT")})
	m := manifest(a)
	data, trust := signed(t, m)
	if _, e := Verify(data, trust, time.Now()); e != nil {
		t.Fatal(e)
	}
	var envelope Envelope
	json.Unmarshal(data, &envelope)
	envelope.Payload = base64.StdEncoding.EncodeToString([]byte(`{"schema":1}`))
	bad, _ := json.Marshal(envelope)
	if _, e := Verify(bad, trust, time.Now()); e == nil {
		t.Fatal("tampered signature accepted")
	}
	for _, change := range []func(*Manifest){func(m *Manifest) { m.Product = "another" }, func(m *Manifest) { m.Expires = time.Now().Add(-time.Hour) }, func(m *Manifest) { m.Version = "0.2.0-beta" }, func(m *Manifest) { m.Commit = "unknown" }, func(m *Manifest) {
		a := m.Assets["linux/amd64"]
		a.URL = "http://insecure.test"
		m.Assets["linux/amd64"] = a
	}} {
		copy := manifest(a)
		change(&copy)
		b, k := signed(t, copy)
		if _, e := Verify(b, k, time.Now()); e == nil {
			t.Fatal("invalid release accepted", copy)
		}
	}
}
func TestUnsafeArchivesAndUnlistedFilesRefused(t *testing.T) {
	for _, name := range []string{"../escape", "/absolute", "LICENSES/../escape", "smolpng:evil", ".smolpng.toml"} {
		b, a := archive(t, map[string][]byte{"smolpng": []byte("new"), "LICENSE": []byte("MIT"), name: []byte("bad")})
		if e := ValidateAsset("linux/amd64", a); e == nil {
			t.Fatal("unsafe manifest accepted", name)
		}
		if e := Extract(b, a, "linux/amd64", t.TempDir()); e == nil {
			t.Fatal("unsafe archive accepted", name)
		}
	}
	b, a := archive(t, map[string][]byte{"smolpng": []byte("new"), "LICENSE": []byte("MIT")})
	b[len(b)/2] ^= 1
	if e := Extract(b, a, "linux/amd64", t.TempDir()); e == nil {
		t.Fatal("corrupt archive accepted")
	}
}

type fakeRunner struct {
	fail  bool
	calls []process.Invocation
}

func (f *fakeRunner) Run(_ context.Context, i process.Invocation) (process.Result, error) {
	f.calls = append(f.calls, i)
	if f.fail {
		return process.Result{ExitCode: 2}, nil
	}
	b, _ := json.Marshal(map[string]string{"version": "0.2.0", "commit": strings.Repeat("a", 40), "distribution": "standalone", "platform": "linux/amd64"})
	return process.Result{Stdout: string(b)}, nil
}
func fixture(t *testing.T) (Manager, Plan, *fakeRunner) {
	t.Helper()
	dir, e := filepath.EvalSymlinks(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	exe := filepath.Join(dir, "smolpng")
	os.WriteFile(exe, []byte("old-binary"), 0755)
	os.WriteFile(filepath.Join(dir, "config.toml"), []byte("preserve user config"), 0600)
	b, a := archive(t, map[string][]byte{"smolpng": []byte("new-binary"), "LICENSE": []byte("MIT")})
	data, keys := signed(t, manifest(a))
	archivePath := filepath.Join(t.TempDir(), "update.zip")
	os.WriteFile(archivePath, b, 0600)
	runner := &fakeRunner{}
	m := Manager{Executable: exe, Version: "0.1.0-dev", Distribution: "standalone", Platform: "linux/amd64", Trust: keys, Runner: runner, SkillState: filepath.Join(dir, "state", "skill-installations.json")}
	plan, e := m.Prepare(context.Background(), data, archivePath)
	if e != nil {
		t.Fatal(e)
	}
	return m, plan, runner
}
func TestUpdatePrepareApplyAndExplicitRollback(t *testing.T) {
	m, p, runner := fixture(t)
	b, _ := os.ReadFile(m.Executable)
	if string(b) != "old-binary" {
		t.Fatal("preparation mutated application")
	}
	if e := m.Apply(context.Background(), p.Stage); e != nil {
		t.Fatal(e)
	}
	b, _ = os.ReadFile(m.Executable)
	if string(b) != "new-binary" {
		t.Fatal("update not committed")
	}
	config, _ := os.ReadFile(filepath.Join(filepath.Dir(m.Executable), "config.toml"))
	if string(config) != "preserve user config" {
		t.Fatal("config overwritten")
	}
	j, _ := ReadJournal(p.Stage)
	if j.Phase != "complete" || len(runner.calls) != 2 || runner.calls[1].Args[0] != "__update-skill-sync" {
		t.Fatal(j, runner.calls)
	}
	if e := m.Rollback(context.Background(), p.Stage); e != nil {
		t.Fatal(e)
	}
	b, _ = os.ReadFile(m.Executable)
	if string(b) != "old-binary" {
		t.Fatal("rollback failed")
	}
	if _, e := os.Stat(filepath.Join(filepath.Dir(m.Executable), "LICENSE")); !os.IsNotExist(e) {
		t.Fatal("new file not rolled back")
	}
}
func TestFailedIdentityProbeRestoresOldApplication(t *testing.T) {
	m, p, r := fixture(t)
	r.fail = true
	if e := m.Apply(context.Background(), p.Stage); e == nil {
		t.Fatal("invalid executable identity accepted")
	}
	b, _ := os.ReadFile(m.Executable)
	if string(b) != "old-binary" {
		t.Fatal("failed update lost original")
	}
	j, _ := ReadJournal(p.Stage)
	if j.Phase != "rolled-back" {
		t.Fatal(j)
	}
}
func TestModifiedInstalledAndBackupFilesStopWithoutOverwrite(t *testing.T) {
	m, p, _ := fixture(t)
	os.WriteFile(m.Executable, []byte("user edit"), 0755)
	if e := m.Apply(context.Background(), p.Stage); e == nil {
		t.Fatal("modified installation overwritten")
	}
	b, _ := os.ReadFile(m.Executable)
	if string(b) != "user edit" {
		t.Fatal("lost user bytes")
	}
	m, p, _ = fixture(t)
	os.WriteFile(filepath.Join(p.Stage, "old", "smolpng"), []byte("corrupt"), 0600)
	if e := m.Apply(context.Background(), p.Stage); e == nil {
		t.Fatal("corrupt backup accepted")
	}
}
func TestInterruptedApplyIsResumable(t *testing.T) {
	m, p, _ := fixture(t)
	j, _ := ReadJournal(p.Stage)
	j.Phase = "applying"
	writeJournal(context.Background(), p.Stage, j)
	b, _ := os.ReadFile(filepath.Join(p.Stage, "new", "smolpng"))
	os.WriteFile(m.Executable, b, 0755)
	if e := m.Apply(context.Background(), p.Stage); e != nil {
		t.Fatal(e)
	}
	j, _ = ReadJournal(p.Stage)
	if j.Phase != "complete" {
		t.Fatal(j)
	}
}
func TestChannelGuardsAndMonotonicVersions(t *testing.T) {
	for _, p := range []string{`C:\Users\x\scoop\apps\smolpng\current\smolpng.exe`, "/opt/homebrew/Cellar/smolpng/1.0/bin/smolpng", "/usr/bin/smolpng", `C:\Program Files\WindowsApps\smolpng.exe`} {
		if Installation(p, "standalone").CanApply {
			t.Fatal("package owner bypass", p)
		}
	}
	if Installation("/tmp/smolpng", "source").CanApply {
		t.Fatal("source self-update allowed")
	}
	for _, p := range [][2]string{{"1.0.0", "0.9.9"}, {"1.0.0", "1.0.0-dev"}, {"1.10.0", "1.9.0"}} {
		if !Newer(p[0], p[1]) {
			t.Fatal(p)
		}
	}
	if Newer("0.1.0", "1.0.0") || Newer("1.0.0", "1.0.0") {
		t.Fatal("downgrade allowed")
	}
}

func TestUntrustedKeyAndCorruptedStageNeverReachProcess(t *testing.T) {
	m, p, r := fixture(t)
	m.Trust = Trust{}
	if e := m.Apply(context.Background(), p.Stage); e == nil {
		t.Fatal("untrusted key applied")
	}
	if len(r.calls) != 0 {
		t.Fatal("executed without trust")
	}
	m, p, r = fixture(t)
	os.WriteFile(filepath.Join(p.Stage, "new", "smolpng"), []byte("corrupt"), 0600)
	if e := m.Apply(context.Background(), p.Stage); e == nil {
		t.Fatal("staged mutation applied")
	}
	if len(r.calls) != 0 {
		t.Fatal("corrupt stage executed")
	}
}
func TestRollbackNeverOverwritesPostUpdateEdits(t *testing.T) {
	m, p, _ := fixture(t)
	if e := m.Apply(context.Background(), p.Stage); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(m.Executable, []byte("later version"), 0755)
	if e := m.Rollback(context.Background(), p.Stage); e == nil {
		t.Fatal("rollback discarded later bytes")
	}
	b, _ := os.ReadFile(m.Executable)
	if string(b) != "later version" {
		t.Fatal("lost new bytes")
	}
}
