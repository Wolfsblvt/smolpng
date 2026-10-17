package toolchain

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/process"
)

type transport func(*http.Request) (*http.Response, error)

func (t transport) RoundTrip(r *http.Request) (*http.Response, error) { return t(r) }

type managedRunner struct {
	calls int
	wrong bool
}

func (r *managedRunner) Run(_ context.Context, i process.Invocation) (process.Result, error) {
	r.calls++
	b, e := os.ReadFile(i.Executable)
	if e != nil {
		return process.Result{}, e
	}
	if len(i.Args) != 1 || i.Args[0] != "--version" {
		return process.Result{}, errors.New("unexpected process arguments")
	}
	v := strings.TrimPrefix(string(b), "fake version ")
	if r.wrong {
		v = "0.0.1"
	}
	return process.Result{Stdout: strings.TrimSuffix(filepath.Base(i.Executable), ".exe") + " " + v}, nil
}
func archiveZIP(t *testing.T, entries map[string][]byte) []byte {
	t.Helper()
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	for name, data := range entries {
		w, e := z.Create(name)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = w.Write(data); e != nil {
			t.Fatal(e)
		}
	}
	if e := z.Close(); e != nil {
		t.Fatal(e)
	}
	return b.Bytes()
}
func archiveTAR(t *testing.T, entries map[string][]byte, link bool) []byte {
	t.Helper()
	var b bytes.Buffer
	g := gzip.NewWriter(&b)
	z := tar.NewWriter(g)
	for name, data := range entries {
		h := &tar.Header{Name: name, Size: int64(len(data)), Mode: 0755, Typeflag: tar.TypeReg}
		if link {
			h.Typeflag = tar.TypeSymlink
			h.Linkname = "/outside"
			h.Size = 0
		}
		if e := z.WriteHeader(h); e != nil {
			t.Fatal(e)
		}
		if !link {
			if _, e := z.Write(data); e != nil {
				t.Fatal(e)
			}
		}
	}
	z.Close()
	g.Close()
	return b.Bytes()
}
func managedFixture(t *testing.T) (Managed, Artifact, []byte, *managedRunner, *atomic.Int32) {
	t.Helper()
	root, e := filepath.EvalSymlinks(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	a, e := ArtifactFor("oxipng", runtime.GOOS, runtime.GOARCH)
	if e != nil {
		a, _ = ArtifactFor("oxipng", "windows", "amd64")
		a.Arch = runtime.GOARCH
		a.OS = runtime.GOOS
	}
	a.Archive = "zip"
	payload := archiveZIP(t, map[string][]byte{"release/" + a.Executable: []byte("fake version " + a.Version)})
	a.SHA256 = digest(payload)
	calls := &atomic.Int32{}
	runner := &managedRunner{}
	client := &http.Client{Transport: transport(func(req *http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(payload)), ContentLength: int64(len(payload)), Header: make(http.Header)}, nil
	})}
	return Managed{Root: filepath.Join(root, "tools"), Runner: runner, Client: client}, a, payload, runner, calls
}
func TestManagedJPEGOptimRetainsLicenseAndUsesManagedDiscovery(t *testing.T) {
	m, _, _, runner, _ := managedFixture(t)
	a, e := ArtifactFor("jpegoptim", "windows", "amd64")
	if e != nil {
		t.Fatal(e)
	}
	a.OS, a.Arch = runtime.GOOS, runtime.GOARCH
	if runtime.GOOS != "windows" {
		a.Executable = "jpegoptim"
	}
	payload := archiveZIP(t, map[string][]byte{
		a.Executable: []byte("fake version v1.5.6"),
		"LICENSE":    []byte("GPL license fixture"),
		"COPYRIGHT":  []byte("jpegoptim copyright fixture"),
	})
	a.SHA256 = digest(payload)
	m.Client = &http.Client{Transport: transport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(payload)), ContentLength: int64(len(payload))}, nil
	})}
	installed, e := m.Install(context.Background(), a)
	if e != nil || installed.State != "current" {
		t.Fatal(installed, e)
	}
	license, e := os.ReadFile(filepath.Join(filepath.Dir(installed.Path), "LICENSE"))
	if e != nil || !bytes.Contains(license, []byte("GPL license fixture")) || !bytes.Contains(license, []byte("jpegoptim copyright fixture")) || bytes.Contains(license, []byte(oxipngLicense)) {
		t.Fatalf("wrong retained JPEG license: %q, %v", license, e)
	}
	tools := Discover(context.Background(), runner, Paths{}, Paths{ManagedRoot: m.Root, Preference: "managed"})
	if !tools.JPEGOptim.Available || tools.JPEGOptim.Source != "managed" || tools.JPEGOptim.Version != "1.5.6" || tools.JPEGOptim.Path != installed.Path {
		t.Fatalf("managed JPEG installation was not selected: %+v", tools.JPEGOptim)
	}
	if e := m.Remove(context.Background(), "jpegoptim"); e != nil {
		t.Fatal(e)
	}
	if got := m.Inspect("jpegoptim"); got.State != "missing" {
		t.Fatal(got)
	}
	for _, missing := range []string{"LICENSE", "COPYRIGHT"} {
		t.Run("missing-"+missing, func(t *testing.T) {
			entries := map[string][]byte{a.Executable: []byte("fake version v1.5.6"), "LICENSE": []byte("GPL fixture"), "COPYRIGHT": []byte("copyright fixture")}
			delete(entries, missing)
			payload := archiveZIP(t, entries)
			a.SHA256 = digest(payload)
			m.Client = &http.Client{Transport: transport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(payload)), ContentLength: int64(len(payload))}, nil
			})}
			before := runner.calls
			if _, err := m.Install(context.Background(), a); err == nil {
				t.Fatal("installation without upstream notices accepted")
			}
			if runner.calls != before || m.Inspect("jpegoptim").State != "missing" {
				t.Fatal("incomplete licensed archive was executed or activated")
			}
		})
	}
}

func TestManagedInstallVerifiesArchiveBeforeExecutionAndTracksBytes(t *testing.T) {
	m, a, _, runner, calls := managedFixture(t)
	s, e := m.Install(context.Background(), a)
	if e != nil || s.State != "current" || runner.calls != 1 || calls.Load() != 1 {
		t.Fatal(s, e, runner, calls.Load())
	}
	if _, e := os.Stat(s.Path); e != nil {
		t.Fatal(e)
	}
	s, e = m.Install(context.Background(), a)
	if e != nil || runner.calls != 1 || calls.Load() != 1 {
		t.Fatal("unchanged install should be local no-op", s, e)
	}
	if strings.Contains(s.Path, os.Getenv("PATH")) {
		t.Fatal("path mutation")
	}
	os.WriteFile(s.Path, []byte("changed"), 0700)
	if got := m.Inspect(a.Tool); got.State != "modified" {
		t.Fatal(got)
	}
	if _, e = m.Install(context.Background(), a); e == nil {
		t.Fatal("modified program overwritten")
	}
	if e = m.Remove(context.Background(), a.Tool); e == nil {
		t.Fatal("modified program removed")
	}
	b, _ := os.ReadFile(s.Path)
	if string(b) != "changed" {
		t.Fatal("modified bytes not preserved")
	}
}
func TestBadHashAndVersionNeverBecomeActive(t *testing.T) {
	for _, failure := range []string{"hash", "version"} {
		t.Run(failure, func(t *testing.T) {
			m, a, _, runner, _ := managedFixture(t)
			if failure == "hash" {
				a.SHA256 = strings.Repeat("0", 64)
			} else {
				runner.wrong = true
			}
			if _, e := m.Install(context.Background(), a); e == nil {
				t.Fatal("bad install succeeded")
			}
			if failure == "hash" && runner.calls != 0 {
				t.Fatal("unauthenticated bytes executed")
			}
			if got := m.Inspect(a.Tool); got.State != "missing" {
				t.Fatal("failed fresh install stranded active directory", got)
			}
			entries, _ := os.ReadDir(m.Root)
			for _, v := range entries {
				if v.Name() != "oxipng.lock" {
					t.Fatal("staging leaked", v.Name())
				}
			}
		})
	}
}
func TestUpdateKeepsOldImmutableVersionAndRemoveOnlyOwnedContents(t *testing.T) {
	m, a, _, _, _ := managedFixture(t)
	old, e := m.Install(context.Background(), a)
	if e != nil {
		t.Fatal(e)
	}
	next := a
	next.Version = "10.2.2"
	next.URL = strings.Replace(a.URL, "v10.2.1/", "v10.2.2/", 1)
	payload := archiveZIP(t, map[string][]byte{next.Executable: []byte("fake version " + next.Version)})
	next.SHA256 = digest(payload)
	m.Client = &http.Client{Transport: transport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(payload)), Header: make(http.Header)}, nil
	})}
	current, e := m.Install(context.Background(), next)
	if e != nil || current.Version != next.Version || current.Versions != 2 || current.Path == old.Path {
		t.Fatal(current, e)
	}
	b, _ := os.ReadFile(old.Path)
	if string(b) != "fake version "+a.Version {
		t.Fatal("old immutable bytes overwritten")
	}
	extra := filepath.Join(m.Root, a.Tool, "notes.txt")
	os.WriteFile(extra, []byte("user notes"), 0600)
	if e = m.Remove(context.Background(), a.Tool); e == nil {
		t.Fatal("untracked notes deleted")
	}
	if m.Inspect(a.Tool).State != "current" {
		t.Fatal("failed removal damaged active program")
	}
	os.Remove(extra)
	if e = m.Remove(context.Background(), a.Tool); e != nil {
		t.Fatal(e)
	}
	if m.Inspect(a.Tool).State != "missing" {
		t.Fatal("active record survives uninstall")
	}
}
func TestNoAdoptionOrCrossPlatformExecution(t *testing.T) {
	m, a, _, runner, calls := managedFixture(t)
	os.MkdirAll(filepath.Join(m.Root, a.Tool), 0700)
	os.WriteFile(filepath.Join(m.Root, a.Tool, "notes"), []byte("mine"), 0600)
	if _, e := m.Install(context.Background(), a); e == nil {
		t.Fatal("untracked directory adopted")
	}
	if calls.Load() != 0 || runner.calls != 0 {
		t.Fatal("untracked install caused network or execution")
	}
	a.Arch = "not-this-architecture"
	if _, e := m.Install(context.Background(), a); e == nil {
		t.Fatal("cross-platform managed binary executed")
	}
}
func TestArchivesRejectTraversalLinksDuplicateOrMissingExecutable(t *testing.T) {
	for _, bad := range []string{"../oxipng", "/oxipng", "C:/oxipng", "dir\\oxipng"} {
		if _, e := ArchiveExecutable(archiveZIP(t, map[string][]byte{bad: []byte("exe")}), "zip", "oxipng"); e == nil {
			t.Fatal("unsafe zip", bad)
		}
	}
	for _, bad := range []string{"../oxipng", "/oxipng"} {
		if _, e := ArchiveExecutable(archiveTAR(t, map[string][]byte{bad: []byte("exe")}, false), "tar.gz", "oxipng"); e == nil {
			t.Fatal("unsafe tar", bad)
		}
	}
	if _, e := ArchiveExecutable(archiveTAR(t, map[string][]byte{"oxipng": nil}, true), "tar.gz", "oxipng"); e == nil {
		t.Fatal("symlink accepted")
	}
	if _, e := ArchiveExecutable(archiveZIP(t, map[string][]byte{"a/oxipng": []byte("a"), "b/oxipng": []byte("b")}), "zip", "oxipng"); e == nil {
		t.Fatal("duplicate executable accepted")
	}
	if _, e := ArchiveExecutable(archiveZIP(t, map[string][]byte{"README": []byte("text")}), "zip", "oxipng"); e == nil {
		t.Fatal("missing executable accepted")
	}
	for _, kind := range []string{"zip", "tar.gz"} {
		var b []byte
		if kind == "zip" {
			b = archiveZIP(t, map[string][]byte{"dist/oxipng": []byte("good")})
		} else {
			b = archiveTAR(t, map[string][]byte{"dist/oxipng": []byte("good")}, false)
		}
		out, e := ArchiveExecutable(b, kind, "oxipng")
		if e != nil || string(out) != "good" {
			t.Fatal(kind, e)
		}
		b[len(b)-1] ^= 0xff
		if kind == "tar.gz" {
			if _, e := ArchiveExecutable(b, kind, "oxipng"); e == nil {
				t.Fatal("gzip trailer corruption accepted")
			}
		}
	}
}
func TestConcurrentInstallerRefusesSecondWriter(t *testing.T) {
	m, a, payload, _, _ := managedFixture(t)
	started, release := make(chan struct{}), make(chan struct{})
	m.Client = &http.Client{Transport: transport(func(*http.Request) (*http.Response, error) {
		close(started)
		<-release
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(payload)), Header: make(http.Header)}, nil
	})}
	done := make(chan error, 1)
	go func() { _, e := m.Install(context.Background(), a); done <- e }()
	<-started
	if _, e := m.Install(context.Background(), a); e == nil {
		t.Fatal("second writer bypassed management lock")
	}
	close(release)
	if e := <-done; e != nil {
		t.Fatal(e)
	}
}
func TestManagedCatalogAndPolicy(t *testing.T) {
	for _, platform := range []string{"windows", "linux", "darwin"} {
		a, err := ArtifactFor("jpegoptim", platform, "amd64")
		if err != nil || validateArtifact(a) != nil || a.License != "GPL-3.0-or-later" {
			t.Fatal(a, err)
		}
		a.URL = strings.Replace(a.URL, "/tjko/jpegoptim/", "/oxipng/oxipng/", 1)
		if validateArtifact(a) == nil {
			t.Fatal("wrong optimizer's upstream was accepted")
		}
		if _, err := ArtifactFor("jpegoptim", platform, "arm64"); err == nil {
			t.Fatal("invented native arm64 jpegoptim archive")
		}
	}
	for _, target := range []string{"linux/amd64", "linux/arm64", "windows/amd64", "darwin/amd64", "darwin/arm64"} {
		p := strings.Split(target, "/")
		a, e := ArtifactFor("oxipng", p[0], p[1])
		if e != nil || validateArtifact(a) != nil {
			t.Fatal(a, e)
		}
	}
	for _, tool := range []string{"pngquant", "unknown"} {
		if _, e := ArtifactFor(tool, "linux", "amd64"); e == nil {
			t.Fatal("unqualified tool advertised")
		}
	}
	if _, e := ArtifactFor("oxipng", "windows", "arm64"); e == nil {
		t.Fatal("invented native arm64 Windows archive")
	}
	m, a, _, _, calls := managedFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := m.Install(ctx, a); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if calls.Load() != 0 {
		t.Fatal("cancelled install downloaded")
	}
}

func TestOldCatalogDoesNotDowngradeAnInstalledTool(t *testing.T) {
	m, a, _, _, calls := managedFixture(t)
	newer := a
	newer.Version = "10.3.0"
	newer.URL = strings.Replace(a.URL, "v10.2.1/", "v10.3.0/", 1)
	payload := archiveZIP(t, map[string][]byte{newer.Executable: []byte("fake version " + newer.Version)})
	newer.SHA256 = digest(payload)
	m.Client = &http.Client{Transport: transport(func(*http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(payload)), Header: make(http.Header)}, nil
	})}
	current, e := m.Install(context.Background(), newer)
	if e != nil {
		t.Fatal(e)
	}
	before := calls.Load()
	if _, e = m.Install(context.Background(), a); e == nil || !strings.Contains(e.Error(), "downgrade") {
		t.Fatal(e)
	}
	if calls.Load() != before || m.Inspect(a.Tool).Path != current.Path {
		t.Fatal("old catalog modified the installed version")
	}
	for _, v := range []string{"10.2", "10.2.1/../bad", "10.02.1", "ten", "10.2.1-beta"} {
		bad := a
		bad.Version = v
		if validateArtifact(bad) == nil {
			t.Fatal("invalid version accepted", v)
		}
	}
}
func TestManagedDiscoveryPreferenceFallbackAndNoProbeOfUntrackedFiles(t *testing.T) {
	m, a, _, runner, _ := managedFixture(t)
	installed, e := m.Install(context.Background(), a)
	if e != nil {
		t.Fatal(e)
	}
	system := t.TempDir()
	exe := "oxipng"
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	path := filepath.Join(system, exe)
	if e = os.WriteFile(path, []byte("fake version 10.1.0"), 0700); e != nil {
		t.Fatal(e)
	}
	t.Setenv("PATH", system)
	config := Paths{ManagedRoot: m.Root}
	got := Discover(context.Background(), runner, Paths{}, config).OxiPNG
	if !got.Available || got.Source != "system" || got.Path != path {
		t.Fatal(got)
	}
	config.Preference = "managed"
	got = Discover(context.Background(), runner, Paths{}, config).OxiPNG
	if !got.Available || got.Source != "managed" || got.Path != installed.Path {
		t.Fatal(got)
	}
	got = Discover(context.Background(), runner, Paths{OxiPNG: path}, config).OxiPNG
	if !got.Available || got.Source != "invocation" || got.Path != path {
		t.Fatal(got)
	}
	config.Preference = "system"
	os.WriteFile(path, []byte("fake version 99.0.0"), 0700)
	got = Discover(context.Background(), runner, Paths{}, config).OxiPNG
	if !got.Available || got.Source != "managed" {
		t.Fatal("unsupported PATH entry shadowed valid managed copy", got)
	}
	config.Preference = "managed"
	os.WriteFile(installed.Path, []byte("modified"), 0700)
	os.WriteFile(path, []byte("fake version 10.1.0"), 0700)
	got = Discover(context.Background(), runner, Paths{}, config).OxiPNG
	if !got.Available || got.Source != "system" {
		t.Fatal("modified managed bytes should not be run", got)
	}
	// No executable-shaped sentinel is probed merely because managed status is missing.
	t.Setenv("PATH", t.TempDir())
	config.ManagedRoot = t.TempDir()
	os.WriteFile(filepath.Join(config.ManagedRoot, "__missing-managed-executable__"), []byte("fake version 10.2.1"), 0700)
	before := runner.calls
	got = Discover(context.Background(), runner, Paths{}, config).OxiPNG
	if got.Available || runner.calls != before {
		t.Fatal("untracked sentinel was executed", got)
	}
}
