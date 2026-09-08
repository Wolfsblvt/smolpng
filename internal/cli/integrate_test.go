package cli

import (
	"bytes"
	"context"
	"github.com/Wolfsblvt/smolpng/internal/integration"
	"github.com/Wolfsblvt/smolpng/internal/paths"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type integrationRenderer struct{}

func (integrationRenderer) Render(_ context.Context, e integration.Entry) ([]byte, error) {
	return integration.DesktopFile(e)
}
func TestIntegrationCLIPlanAndInstall(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("native launchers are Windows/Linux only")
	}
	root, e := filepath.EvalSymlinks(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	exe := filepath.Join(root, "smolpng")
	os.WriteFile(exe, []byte("fake app"), 0755)
	os.WriteFile(filepath.Join(root, "smolpngw.exe"), []byte("fake background"), 0755)
	loc := integration.Locations{Menu: filepath.Join(root, "menu"), Desktop: filepath.Join(root, "desktop"), SendTo: filepath.Join(root, "sendto")}
	var out, er bytes.Buffer
	env := Environment{In: strings.NewReader(""), Out: &out, Err: &er, Paths: &paths.Paths{Config: filepath.Join(root, "config.toml"), State: filepath.Join(root, "state"), Data: root, Downloads: root}, CWD: root, Executable: func() (string, error) { return exe, nil }, IntegrationLocations: &loc, IntegrationRenderer: integrationRenderer{}}
	args := []string{"integrate", "install", "--desktop", "--json"}
	if c := Run(context.Background(), args, env); c != 0 {
		t.Fatal(c, out.String(), er.String())
	}
	if _, e := os.Stat(loc.Menu); !os.IsNotExist(e) {
		t.Fatal("preview wrote launchers")
	}
	out.Reset()
	if c := Run(context.Background(), append(args, "--apply"), env); c != 0 {
		t.Fatal(c, out.String(), er.String())
	}
	out.Reset()
	if c := Run(context.Background(), []string{"integrate", "status", "--json"}, env); c != 0 || !strings.Contains(out.String(), "current") {
		t.Fatal(c, out.String())
	}
}

func TestLaunchersUseDurableUserPresetAndExplicitConfiguration(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("native launchers are Windows/Linux only")
	}
	root, e := filepath.EvalSymlinks(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	project := filepath.Join(root, "project")
	if e = os.MkdirAll(project, 0700); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(project, ".smolpng.toml"), []byte("[presets.project-only]\nschema=1\nextends='auto'\n"), 0600); e != nil {
		t.Fatal(e)
	}
	user := filepath.Join(root, "special.toml")
	if e = os.WriteFile(user, []byte("[presets.user-copy]\nschema=1\nextends='auto'\n"), 0600); e != nil {
		t.Fatal(e)
	}
	exe := filepath.Join(root, "smolpng")
	_ = os.WriteFile(exe, []byte("app"), 0700)
	_ = os.WriteFile(filepath.Join(root, "smolpngw.exe"), []byte("wrapper"), 0700)
	var out, er bytes.Buffer
	loc := integration.Locations{Menu: filepath.Join(root, "menu"), Desktop: filepath.Join(root, "desktop")}
	env := Environment{In: strings.NewReader(""), Out: &out, Err: &er, Paths: &paths.Paths{Config: filepath.Join(root, "config.toml"), State: filepath.Join(root, "state"), Data: root, Downloads: root}, CWD: project, Executable: func() (string, error) { return exe, nil }, IntegrationLocations: &loc}
	if c := Run(context.Background(), []string{"integrate", "install", "--preset", "project-only", "--json"}, env); c != 2 {
		t.Fatal("project-only launcher should refuse", c, out.String(), er.String())
	}
	out.Reset()
	er.Reset()
	if c := Run(context.Background(), []string{"--config", user, "integrate", "install", "--preset", "user-copy", "--json"}, env); c != 0 {
		t.Fatal(c, out.String(), er.String())
	}
	if !strings.Contains(out.String(), "--no-project") || !strings.Contains(out.String(), "--config") {
		t.Fatal("lost durable launcher configuration", out.String())
	}
}
