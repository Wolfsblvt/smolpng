package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type textRenderer struct{}

func (textRenderer) Render(_ context.Context, x Entry) ([]byte, error) { return DesktopFile(x) }
func fixture(t *testing.T) (Manager, []Entry) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(root, "smolpng")
	if e := os.WriteFile(exe, []byte("fixture executable"), 0755); e != nil {
		t.Fatal(e)
	}
	loc := Locations{Menu: filepath.Join(root, "apps"), Desktop: filepath.Join(root, "desktop")}
	plan, e := Plan(loc, Request{OS: "linux", Executable: exe, WorkingDirectory: root, Desktop: true})
	if e != nil {
		t.Fatal(e)
	}
	return Manager{StatePath: filepath.Join(root, "state", "integrations.json"), Version: "test", Allowed: loc, Renderer: textRenderer{}}, plan
}
func TestPlanDoesNotInstallAndSeparatesExplicitModes(t *testing.T) {
	m, p := fixture(t)
	if len(p) != 4 {
		t.Fatal(len(p))
	}
	if _, e := os.Stat(m.Allowed.Menu); !os.IsNotExist(e) {
		t.Fatal("planning wrote files")
	}
	if !p[0].Terminal || p[1].Terminal || p[1].Arguments[0] != "drop" {
		t.Fatal(p)
	}
	if p[0].Arguments[len(p[0].Arguments)-1] != "--" {
		t.Fatal("no positional barrier")
	}
}
func TestTrackedInstallPreservesModifiedAndUnownedFiles(t *testing.T) {
	m, p := fixture(t)
	ctx := context.Background()
	if e := m.Install(ctx, p); e != nil {
		t.Fatal(e)
	}
	s, pending, e := m.Status()
	if e != nil || pending || len(s) != 4 || s[0].State != "current" {
		t.Fatal(s, e)
	}
	if e = m.Install(ctx, p); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(p[0].Path, []byte("user change"), 0600)
	if e = m.Remove(ctx, ""); e == nil {
		t.Fatal("modified removal accepted")
	}
	if e = m.Install(ctx, p); e == nil {
		t.Fatal("modified replacement accepted")
	}
	b, _ := os.ReadFile(p[0].Path)
	if string(b) != "user change" {
		t.Fatal("lost edits")
	}
	m2, p2 := fixture(t)
	os.MkdirAll(filepath.Dir(p2[0].Path), 0755)
	os.WriteFile(p2[0].Path, []byte("unowned"), 0600)
	if e = m2.Install(ctx, p2); e == nil {
		t.Fatal("unowned overwrite")
	}
}
func TestInterruptedInstallRecoveryRequiresExactHashes(t *testing.T) {
	m, p := fixture(t)
	ctx := context.Background()
	b, _ := DesktopFile(p[0])
	os.MkdirAll(filepath.Dir(p[0].Path), 0755)
	os.WriteFile(p[0].Path, b, 0755)
	s := Registry{Schema: 1, Records: []Record{}, Pending: &Pending{Next: Record{Entry: p[0], SHA256: digest(b), Version: "test"}}}
	os.MkdirAll(filepath.Dir(m.StatePath), 0700)
	raw, _ := json.Marshal(s)
	os.WriteFile(m.StatePath, raw, 0600)
	if e := m.Repair(ctx); e != nil {
		t.Fatal(e)
	}
	st, pending, e := m.Status()
	if e != nil || pending || len(st) != 1 {
		t.Fatal(st, pending, e)
	}
	if e = m.Remove(ctx, ""); e != nil {
		t.Fatal(e)
	}
	if _, e = os.Stat(p[0].Path); !os.IsNotExist(e) {
		t.Fatal("remove did not delete")
	}
}
func TestNativeRenderContractsAndEscaping(t *testing.T) {
	m, p := fixture(t)
	_ = m
	p[0].Executable = "/home/a space/$USER`x`%/smolpng"
	p[0].Arguments = []string{"--preset", "name", "--"}
	b, e := DesktopFile(p[0])
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Contains(b, []byte(" %F\n")) || bytes.Contains(b, []byte(`"%F"`)) || !bytes.Contains(b, []byte("%%")) || !bytes.Contains(b, []byte(`\\$`)) {
		t.Fatal(string(b))
	}
	args := WindowsArguments([]string{`a b`, `ends\`, `q"uote`})
	if args != `"a b" "ends\\" "q\"uote"` {
		t.Fatal(args)
	}
	script := PowerShellScript(Entry{Executable: `C:\a'b\smolpng.exe`, Arguments: []string{`hello'; exit; '`}, WorkingDirectory: `C:\`}, `C:\out.lnk`)
	if !strings.Contains(script, "a''b") || strings.Contains(script, "hello'; exit") {
		t.Fatal(script)
	}
}
func TestRegistryCannotRemoveOutsideNativeDirectories(t *testing.T) {
	m, p := fixture(t)
	p[0].Path = filepath.Join(t.TempDir(), "smolpng-interactive.desktop")
	if e := m.Install(context.Background(), p); e == nil {
		t.Fatal("outside install accepted")
	}
}
