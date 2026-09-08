package cli

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/config"
	"github.com/Wolfsblvt/smolpng/internal/toolchain"
)

type dependencyTransport func(*http.Request) (*http.Response, error)

func (f dependencyTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestDependencyCommandsPreviewWithoutNetworkAndDoNotAdoptPNGQuant(t *testing.T) {
	h := setup(t)
	calls := 0
	h.DependencyHTTP = &http.Client{Transport: dependencyTransport(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("controlled network failure")
	})}
	if c := h.run("deps", "licenses", "--json"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	if !strings.Contains(h.out.String(), `"bundledExecutables": false`) {
		t.Fatal(h.out)
	}
	if _, e := toolchain.ArtifactFor("oxipng", runtime.GOOS, runtime.GOARCH); e == nil {
		if c := h.run("deps", "install", "oxipng", "--json"); c != 0 {
			t.Fatal(c, h.out, h.err)
		}
		if _, e = os.Stat(filepath.Join(h.Paths.Data, "tools")); !os.IsNotExist(e) {
			t.Fatal("preview created storage", e)
		}
		if calls != 0 {
			t.Fatal("preview downloaded")
		}
		if c := h.run("deps", "install", "oxipng", "--apply", "--json"); c != 5 || calls != 1 {
			t.Fatal(c, calls, h.out, h.err)
		}
	}
	before := calls
	if c := h.run("deps", "install", "pngquant", "--apply", "--json"); c != 4 || calls != before {
		t.Fatal("unqualified pngquant download attempted", c, calls)
	}
	if c := h.run("deps", "status", "--json"); c != 0 || !strings.Contains(h.out.String(), `"managedDependencies"`) {
		t.Fatal(c, h.out)
	}
}
func TestDependencyPreferenceChangesOnlyAfterExplicitApply(t *testing.T) {
	h := setup(t)
	if c := h.run("deps", "use-managed", "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	if _, e := os.Stat(h.Paths.Config); !os.IsNotExist(e) {
		t.Fatal("preference preview wrote config")
	}
	if c := h.run("deps", "use-managed", "--apply", "--json"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	w, e := config.Load(config.LoadOptions{Paths: *h.Paths, CWD: h.dir})
	if e != nil {
		t.Fatal(e)
	}
	eff, e := w.Resolve("", config.Fields{})
	if e != nil || eff.Tools.Preference != "managed" || eff.Provenance["tools.preference"].File != h.Paths.Config {
		t.Fatal(e, eff.Tools, eff.Provenance)
	}
	if c := h.run("deps", "configure", "oxipng", filepath.Join(h.dir, "oxipng"), "--json"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	if c := h.run("deps", "use-system", "--apply", "--json"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	w, e = config.Load(config.LoadOptions{Paths: *h.Paths, CWD: h.dir})
	if e != nil {
		t.Fatal(e)
	}
	if w.User.Document.Tools.Preference != "system" || w.User.Document.Tools.OxiPNG != "" {
		t.Fatal(w.User.Document.Tools)
	}
}
func TestOnboardingManagedDownloadRequiresSeparateConfirmation(t *testing.T) {
	if _, e := toolchain.ArtifactFor("oxipng", runtime.GOOS, runtime.GOARCH); e != nil {
		t.Skip("no pinned native artifact on this host")
	}
	h, p, _, _ := humanSetup(t, choose("smolpng setup", "managed-oxipng"), choose("Install separate OxiPNG?", "back"), choose("smolpng setup", "continue"))
	calls := 0
	h.DependencyHTTP = &http.Client{Transport: dependencyTransport(func(*http.Request) (*http.Response, error) { calls++; return nil, errors.New("must not download") })}
	if c := h.run("setup"); c != 0 || calls != 0 || len(p.steps) != 0 {
		t.Fatal(c, calls, h.out, h.err)
	}
	if _, e := os.Stat(filepath.Join(h.Paths.Data, "tools")); !os.IsNotExist(e) {
		t.Fatal("unconfirmed setup created tools")
	}
}
