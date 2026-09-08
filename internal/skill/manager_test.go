package skill

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	embedded "github.com/Wolfsblvt/smolpng/skills"
)

func testRoot(t *testing.T) string {
	t.Helper()
	root, e := filepath.EvalSymlinks(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	return root
}
func fixture(t *testing.T, content string) Manager {
	t.Helper()
	root := testRoot(t)
	b, e := ReadBundle(fstest.MapFS{"SKILL.md": {Data: []byte(content)}, "references/cli.md": {Data: []byte("commands")}})
	if e != nil {
		t.Fatal(e)
	}
	return Manager{StatePath: filepath.Join(root, "state", "skills.json"), Bundle: b, AppVersion: "0.1.0"}
}
func changed(t *testing.T, m Manager) Manager {
	t.Helper()
	m.AppVersion = "0.2.0"
	b, e := ReadBundle(fstest.MapFS{"SKILL.md": {Data: []byte("new instructions")}, "references/cli.md": {Data: []byte("new CLI")}, "references/presets.md": {Data: []byte("new preset guide")}})
	if e != nil {
		t.Fatal(e)
	}
	m.Bundle = b
	return m
}
func install(t *testing.T, m Manager, parent string, managed bool) string {
	t.Helper()
	dest := filepath.Join(filepath.Dir(filepath.Dir(m.StatePath)), parent, "smolpng")
	if e := m.Install(context.Background(), dest, "test", managed); e != nil {
		t.Fatal(e)
	}
	return dest
}
func TestStatusIsReadOnlyAndRecordsExactInstall(t *testing.T) {
	m := fixture(t, "original")
	s, e := m.Inspect()
	if e != nil || len(s.Installations) != 0 {
		t.Fatal(s, e)
	}
	if _, e = os.Stat(filepath.Dir(m.StatePath)); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("status created state directory")
	}
	dest := install(t, m, "shared", true)
	s, e = m.Inspect()
	if e != nil {
		t.Fatal(e)
	}
	got := s.Installations[0]
	if got.State != "current" || got.Destination != dest || !got.Managed || got.AppVersion != "0.1.0" || !reflect.DeepEqual(got.Files, m.Bundle.Hashes) {
		t.Fatalf("%+v", got)
	}
	for p, h := range got.Files {
		b, e := os.ReadFile(filepath.Join(dest, filepath.FromSlash(p)))
		if e != nil || digest(b) != h {
			t.Fatal(p, e)
		}
	}
	if e = m.Install(context.Background(), dest, "test", true); e == nil {
		t.Fatal("overwrote tracked install")
	}
}
func TestUnknownDestinationIsNeverAdoptedOrOverwritten(t *testing.T) {
	m := fixture(t, "original")
	dest := filepath.Join(testRoot(t), "smolpng")
	if e := os.MkdirAll(dest, 0755); e != nil {
		t.Fatal(e)
	}
	file := filepath.Join(dest, "SKILL.md")
	os.WriteFile(file, []byte("my instructions"), 0600)
	if e := m.Install(context.Background(), dest, "custom", false); e == nil {
		t.Fatal("unknown content adopted")
	}
	b, _ := os.ReadFile(file)
	if string(b) != "my instructions" {
		t.Fatal("unknown data changed")
	}
	if e := m.Install(context.Background(), filepath.Dir(dest), "custom", false); e == nil {
		t.Fatal("nondedicated directory accepted")
	}
}
func TestAutomaticUpdateOnlySynchronizesUnmodifiedManagedInstallations(t *testing.T) {
	m := fixture(t, "original")
	managed := install(t, m, "shared", true)
	custom := install(t, m, "custom", false)
	edited := install(t, m, "edited", true)
	os.WriteFile(filepath.Join(edited, "SKILL.md"), []byte("user custom instructions"), 0600)
	new := changed(t, m)
	results, e := new.SyncAfterUpdate(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	actions := map[string]string{}
	for _, r := range results {
		actions[r.Destination] = r.Action
	}
	if actions[managed] != "synced" || actions[custom] != "needs-consent" || actions[edited] != "preserved-modified" {
		t.Fatal(actions)
	}
	b, _ := os.ReadFile(filepath.Join(custom, "SKILL.md"))
	if string(b) != "original" {
		t.Fatal("custom auto-updated")
	}
	if e = new.Sync(context.Background(), custom); e != nil {
		t.Fatal(e)
	}
	b, _ = os.ReadFile(filepath.Join(edited, "SKILL.md"))
	if string(b) != "user custom instructions" {
		t.Fatal("modified Skill overwritten")
	}
	if e = new.Sync(context.Background(), edited); e == nil {
		t.Fatal("explicit sync overwrote modified Skill")
	}
}
func TestUninstallPreservesModifiedMissingAndUntrackedContent(t *testing.T) {
	for _, kind := range []string{"edited", "removed-resource", "extra", "symlink", "empty-directory"} {
		t.Run(kind, func(t *testing.T) {
			m := fixture(t, "original")
			dest := install(t, m, "shared", true)
			switch kind {
			case "edited":
				os.WriteFile(filepath.Join(dest, "SKILL.md"), []byte("mine"), 0600)
			case "removed-resource":
				os.Remove(filepath.Join(dest, "references", "cli.md"))
			case "extra":
				os.WriteFile(filepath.Join(dest, "personal.txt"), []byte("keep me"), 0600)
			case "empty-directory":
				os.Mkdir(filepath.Join(dest, "notes"), 0755)
			case "symlink":
				if e := os.Symlink(filepath.Join(dest, "SKILL.md"), filepath.Join(dest, "link")); e != nil {
					t.Skip(e)
				}
			}
			if e := m.Uninstall(context.Background(), dest); e == nil {
				t.Fatal("uninstall deleted user-modified content")
			}
			if _, e := os.Stat(dest); e != nil {
				t.Fatal("directory removed")
			}
			if e := m.Forget(context.Background(), dest); e != nil {
				t.Fatal(e)
			}
			if _, e := os.Stat(dest); e != nil {
				t.Fatal("forget removed bytes")
			}
		})
	}
	m := fixture(t, "original")
	dest := install(t, m, "shared", true)
	if e := m.Uninstall(context.Background(), dest); e != nil {
		t.Fatal(e)
	}
	if !missing(dest) {
		t.Fatal("unmodified installation not removed")
	}
	dest = install(t, m, "another", true)
	os.RemoveAll(dest)
	s, e := m.Inspect()
	if e != nil || s.Installations[0].State != "missing" {
		t.Fatal(s, e)
	}
	if e = m.Uninstall(context.Background(), dest); e != nil {
		t.Fatal(e)
	}
}
func TestReplacementPromptDependsOnBundleContentNotVersion(t *testing.T) {
	m := fixture(t, "original")
	install(t, m, "shared", true)
	versionOnly := m
	versionOnly.AppVersion = "0.2.0"
	before, e := versionOnly.Inspect()
	if e != nil || !before.Startup.ExecutableChanged || before.Startup.ContentChanged || before.Startup.Prompt {
		t.Fatal(before, e)
	}
	if v, e := versionOnly.Observe(context.Background()); e != nil || v.Prompt {
		t.Fatal(v, e)
	}
	new := changed(t, m)
	new.AppVersion = "0.3.0"
	s, e := new.Inspect()
	if e != nil || !s.Startup.Prompt {
		t.Fatal(s, e)
	}
	if _, e = new.Observe(context.Background()); e != nil {
		t.Fatal(e)
	}
	s, _ = new.Inspect()
	if !s.Startup.Prompt || !s.Startup.ContentChanged {
		t.Fatal("first startup consumed prompt before human response")
	}
	if e = new.IgnoreCurrent(context.Background()); e != nil {
		t.Fatal(e)
	}
	s, _ = new.Inspect()
	if s.Startup.Prompt || !s.Startup.Ignored {
		t.Fatal(s)
	}
	new2 := changed(t, new)
	b, e := ReadBundle(fstest.MapFS{"SKILL.md": {Data: []byte("third bundle")}})
	if e != nil {
		t.Fatal(e)
	}
	new2.Bundle = b
	s, _ = new2.Inspect()
	if !s.Startup.Prompt {
		t.Fatal("ignored previous version suppressed a different bundle")
	}
}
func TestBundleAndPackageHaveStableResourceHashesAndNoExecutable(t *testing.T) {
	b, e := ReadBundle(embedded.Files())
	if e != nil {
		t.Fatal(e)
	}
	first, e := b.Archive()
	if e != nil {
		t.Fatal(e)
	}
	second, e := b.Archive()
	if e != nil || !bytes.Equal(first, second) {
		t.Fatal("nondeterministic Skill package")
	}
	z, e := zip.NewReader(bytes.NewReader(first), int64(len(first)))
	if e != nil {
		t.Fatal(e)
	}
	if len(z.File) != len(b.Files) {
		t.Fatal("missing resources")
	}
	for _, f := range z.File {
		r, e := f.Open()
		if e != nil {
			t.Fatal(e)
		}
		data, e := io.ReadAll(r)
		r.Close()
		if e != nil {
			t.Fatal(e)
		}
		name := strings.TrimPrefix(f.Name, "smolpng/")
		if name == f.Name || digest(data) != b.Hashes[name] {
			t.Fatal(f.Name)
		}
	}
	if len(b.Files["SKILL.md"]) > 256<<10 || len(first) > 8<<20 {
		t.Fatal("package exceeds intended limits")
	}
}
func TestTargetsAreExplicitAndCustomPathsAreNotAutoManaged(t *testing.T) {
	home, cwd := testRoot(t), testRoot(t)
	a, e := ResolveTarget("auto", "", home, cwd)
	if e != nil || a.Destination != filepath.Join(home, ".agents", "skills", "smolpng") || !a.Managed {
		t.Fatal(a, e)
	}
	c, e := ResolveTarget("codex", "", home, cwd)
	if e != nil || c.Destination != a.Destination {
		t.Fatal(c, e)
	}
	c, e = ResolveTarget("auto", "custom", home, cwd)
	if e != nil || c.Managed || c.Destination != filepath.Join(cwd, "custom", "smolpng") {
		t.Fatal(c, e)
	}
	if _, e = ResolveTarget("claude", "custom", home, cwd); e == nil {
		t.Fatal("ambiguous target accepted")
	}
	for _, tgt := range Targets(home, cwd) {
		if _, e := os.Stat(filepath.Dir(tgt.Destination)); !os.IsNotExist(e) {
			t.Fatal("detection created a target")
		}
	}
}

// Simulate the on-disk journal at meaningful process-death boundaries; the next
// real operation reads and recovers it rather than relying on test-only hooks.
func journal(t *testing.T, m Manager, dest string, new Manager, landed bool) (backup string) {
	t.Helper()
	s, e := m.load()
	if e != nil {
		t.Fatal(e)
	}
	old := s.Installations[0]
	next := new.record(dest, old.Target, old.Managed)
	stage, e := os.MkdirTemp(filepath.Dir(dest), ".smolpng-skill-stage-")
	if e != nil {
		t.Fatal(e)
	}
	for name, data := range new.Bundle.Files {
		p := filepath.Join(stage, filepath.FromSlash(name))
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, data, 0600)
	}
	backup = filepath.Join(filepath.Dir(dest), ".smolpng-skill-backup-fixture")
	tx := &transaction{Destination: dest, Stage: stage, Backup: backup, Previous: &old, Next: &next}
	s.Pending = tx
	if e = m.save(context.Background(), s); e != nil {
		t.Fatal(e)
	}
	if e = os.Rename(dest, backup); e != nil {
		t.Fatal(e)
	}
	if landed {
		if e = os.Rename(stage, dest); e != nil {
			t.Fatal(e)
		}
	}
	return backup
}
func TestInterruptedReplacementRestoresOldOrCommitsExactNewBundle(t *testing.T) {
	for _, landed := range []bool{false, true} {
		t.Run(map[bool]string{false: "old-moved", true: "new-landed"}[landed], func(t *testing.T) {
			m := fixture(t, "original")
			dest := install(t, m, "shared", true)
			new := changed(t, m)
			backup := journal(t, m, dest, new, landed)
			before, e := new.Inspect()
			if e != nil || !before.RecoveryPending {
				t.Fatal(before, e)
			}
			if e = new.Recover(context.Background()); e != nil {
				t.Fatal(e)
			}
			s, e := new.Inspect()
			if e != nil || s.RecoveryPending {
				t.Fatal(s, e)
			}
			want := "stale"
			if landed {
				want = "current"
			}
			if s.Installations[0].State != want {
				t.Fatal(s.Installations[0])
			}
			if !missing(backup) {
				t.Fatal("owned backup leaked")
			}
		})
	}
}
func TestInterruptedReplacementWithUnexpectedContentPreservesBoth(t *testing.T) {
	m := fixture(t, "original")
	dest := install(t, m, "shared", true)
	new := changed(t, m)
	backup := journal(t, m, dest, new, false)
	os.Mkdir(dest, 0755)
	os.WriteFile(filepath.Join(dest, "personal.txt"), []byte("my data"), 0600)
	if e := new.Recover(context.Background()); e == nil {
		t.Fatal("ambiguous recovery silently overwrote")
	}
	if !exact(backup, m.Bundle.Hashes) {
		t.Fatal("backup lost")
	}
	b, _ := os.ReadFile(filepath.Join(dest, "personal.txt"))
	if string(b) != "my data" {
		t.Fatal("new user data changed")
	}
}
func TestRegistryCorruptionCannotBecomeADeletionCommand(t *testing.T) {
	m := fixture(t, "original")
	dest := install(t, m, "shared", true)
	b, _ := os.ReadFile(m.StatePath)
	var v map[string]any
	if e := json.Unmarshal(b, &v); e != nil {
		t.Fatal(e)
	}
	v["cleanup"] = []any{map[string]any{"path": filepath.Dir(dest), "files": m.Bundle.Hashes}}
	b, _ = json.Marshal(v)
	os.WriteFile(m.StatePath, b, 0600)
	if e := m.Recover(context.Background()); e == nil {
		t.Fatal("invalid cleanup path accepted")
	}
	if !exact(dest, m.Bundle.Hashes) {
		t.Fatal("data removed")
	}
}
