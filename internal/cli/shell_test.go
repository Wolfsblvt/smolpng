package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/shellmenu"
	"github.com/Wolfsblvt/smolpng/internal/state"
)

func shellSetup(t *testing.T) *harness {
	h := setup(t)
	h.ShellLocations = &shellmenu.Locations{State: filepath.Join(h.dir, "shell-menu"), Requests: filepath.Join(h.dir, "shell-requests"), ServiceMenus: filepath.Join(h.dir, "dolphin")}
	exe := filepath.Join(h.dir, "smolpng")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	_ = os.WriteFile(exe, []byte("fake executable"), 0700)
	h.Executable = func() (string, error) { return exe, nil }
	return h
}
func installedShell(t *testing.T, h *harness) shellmenu.Spec {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(h.ShellLocations.State, shellmenu.MenuFile))
	if e != nil {
		t.Fatal(e)
	}
	s, e := shellmenu.Decode(b)
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func TestShellPreviewAndOwnedInstallRemoval(t *testing.T) {
	h := shellSetup(t)
	if c := h.run("integrate", "shell", "install", "--json"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	if _, e := os.Stat(h.ShellLocations.State); !os.IsNotExist(e) {
		t.Fatal("preview wrote settings")
	}
	if c := h.run("integrate", "shell", "install", "--apply", "--json"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	s := installedShell(t, h)
	if s.Mode != "single" || s.Actions[0].InPlace {
		t.Fatal(s)
	}
	if runtime.GOOS == "linux" {
		p := filepath.Join(h.ShellLocations.ServiceMenus, "smolpng-dolphin-menu.desktop")
		info, e := os.Stat(p)
		if e != nil || info.Mode().Perm()&0111 == 0 {
			t.Fatal("Dolphin menu not executable", e)
		}
	}
	p := filepath.Join(h.ShellLocations.State, shellmenu.MenuFile)
	_ = os.WriteFile(p, []byte("user edit"), 0600)
	if c := h.run("integrate", "shell", "remove", "--apply"); c != 5 {
		t.Fatal("modified settings removed", c, h.out, h.err)
	}
	b, _ := os.ReadFile(p)
	if string(b) != "user edit" {
		t.Fatal("edit lost")
	}
}
func TestShellSelectionUsesExactSnapshotKeepsOriginalAndLastResult(t *testing.T) {
	h := shellSetup(t)
	before, _ := os.ReadFile(h.source)
	// Current project/global privacy defaults cannot turn a shell click into upload.
	_ = os.WriteFile(h.Paths.Config, []byte("config_version=1\n[defaults]\nremote_policy=\"always\"\noverwrite=true\n[clipboard]\nsingle=\"none\"\nbatch=\"none\"\n"), 0600)
	if c := h.run("integrate", "shell", "install", "--apply"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	s := installedShell(t, h)
	if c := h.run("_shell-run", "--action", "auto", "--generation", s.Generation, "--", h.source); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	b, _ := os.ReadFile(h.source)
	if !bytes.Equal(before, b) {
		t.Fatal("default context action overwrote original")
	}
	if _, e := os.Stat(filepath.Join(h.dir, "a.smol.png")); e != nil {
		t.Fatal(e)
	}
	last, e := state.ReadLast(filepath.Join(h.Paths.State, "last-result.json"))
	if e != nil || len(last.Report.Results) != 1 || last.Report.Results[0].RemoteUsed {
		t.Fatal("missing/locality result", e)
	}
	if c := h.run("_shell-run", "--action", "auto", "--generation", s.Generation, "--"); c != 3 {
		t.Fatal("empty fell back to clipboard", c)
	}
	if c := h.run("_shell-run", "--action", "auto", "--generation", "stale", "--", h.source); c != 3 {
		t.Fatal("stale menu executed", c)
	}
}
func TestShellSubmenuCustomPresetsAreDurableAndLocal(t *testing.T) {
	h := shellSetup(t)
	_ = os.WriteFile(h.Paths.Config, []byte("config_version=1\n[presets.icon]\nschema=1\nextends=\"auto\"\ntarget_size=\"10KiB\"\n"), 0600)
	if c := h.run("integrate", "shell", "install", "--mode", "submenu", "--preset", "icon", "--apply"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	s := installedShell(t, h)
	if len(s.Actions) != 5 || s.Actions[0].ID != "replace" || s.Actions[4].Preset != "icon" {
		t.Fatal(s)
	}
	if c := h.run("integrate", "shell", "install", "--mode", "single", "--preset", "icon"); c != 2 {
		t.Fatal("single custom accepted", c)
	}
	_ = os.WriteFile(filepath.Join(h.dir, ".smolpng.toml"), []byte("[presets.project-only]\nschema=1\nextends=\"auto\"\n"), 0600)
	if c := h.run("integrate", "shell", "install", "--mode", "submenu", "--preset", "project-only"); c != 2 {
		t.Fatal("transient project preset accepted", c, h.out, h.err)
	}
	if c := h.run("integrate", "shell", "status", "--json"); c != 0 || !json.Valid(h.out.Bytes()) {
		t.Fatal(c, h.out, h.err)
	}
}
func TestShellBinaryRequestIsOneShotAndBackgroundWrapperDispatch(t *testing.T) {
	h := shellSetup(t)
	_ = os.WriteFile(h.Paths.Config, []byte("[clipboard]\nsingle=\"none\"\n"), 0600)
	if c := h.run("integrate", "shell", "install", "--apply"); c != 0 {
		t.Fatal(c, h.err)
	}
	s := installedShell(t, h)
	b, e := shellmenu.EncodeRequest(shellmenu.Request{Generation: s.Generation, Action: "auto", Files: []string{h.source}})
	if e != nil {
		t.Fatal(e)
	}
	_ = os.MkdirAll(h.ShellLocations.Requests, 0700)
	p := filepath.Join(h.ShellLocations.Requests, "smolpng-00112233445566778899aabbccddeeff.request")
	_ = os.WriteFile(p, b, 0600)
	if c := h.run("_shell-run", "--request", p); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	if _, e = os.Stat(p); !os.IsNotExist(e) {
		t.Fatal("handoff not consumed")
	}
}

func TestShellSwitchToSingleRemovesOnlyOwnedInteractiveService(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Dolphin artifact transition")
	}
	h := shellSetup(t)
	if c := h.run("integrate", "shell", "install", "--mode", "submenu", "--apply"); c != 0 {
		t.Fatal(c, h.err)
	}
	p := filepath.Join(h.ShellLocations.ServiceMenus, "smolpng-dolphin-interactive.desktop")
	b, e := os.ReadFile(p)
	if e != nil || !bytes.Contains(b, []byte("Terminal=true")) {
		t.Fatal("interactive service missing", e)
	}
	if c := h.run("integrate", "shell", "install", "--mode", "single", "--apply"); c != 0 {
		t.Fatal(c, h.err)
	}
	if _, e = os.Stat(p); !os.IsNotExist(e) {
		t.Fatal("obsolete interactive service retained")
	}
	if c := h.run("integrate", "shell", "install", "--mode", "submenu", "--apply"); c != 0 {
		t.Fatal(c, h.err)
	}
	_ = os.WriteFile(p, []byte("edited"), 0600)
	old := installedShell(t, h)
	if c := h.run("integrate", "shell", "install", "--mode", "single", "--apply"); c != 5 {
		t.Fatal("obsolete local edit ignored", c)
	}
	if installedShell(t, h).Generation != old.Generation {
		t.Fatal("changed settings before edited obsolete-file preflight")
	}
}
