// SPDX-License-Identifier: MIT
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/config"
	"github.com/Wolfsblvt/smolpng/internal/desktop"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/integration"
	"github.com/Wolfsblvt/smolpng/internal/output"
	"github.com/Wolfsblvt/smolpng/internal/shellmenu"
	"github.com/Wolfsblvt/smolpng/internal/ui"
	"github.com/spf13/cobra"
)

type shellRenderer map[string][]byte

func (r shellRenderer) Render(_ context.Context, x integration.Entry) ([]byte, error) {
	b, ok := r[x.ID]
	if !ok {
		return nil, fmt.Errorf("unknown shell artifact")
	}
	return b, nil
}
func (a *app) shellManager() (integration.Manager, shellmenu.Locations, error) {
	l, e := shellmenu.NativeLocations(runtime.GOOS)
	if a.env.ShellLocations != nil {
		l = *a.env.ShellLocations
		e = nil
	}
	if e != nil {
		return integration.Manager{}, l, e
	}
	return integration.Manager{StatePath: filepath.Join(l.State, "ownership.json"), Version: Version, Allowed: integration.Locations{ShellState: l.State, ServiceMenus: l.ServiceMenus}}, l, nil
}
func (a *app) makeShellPlan(mode string, presets []string) (integration.Manager, shellmenu.Spec, []integration.Entry, error) {
	m, l, e := a.shellManager()
	if e != nil {
		return m, shellmenu.Spec{}, nil, e
	}
	p, e := a.dirs()
	if e != nil {
		return m, shellmenu.Spec{}, nil, e
	}
	cwd, e := a.cwd()
	if e != nil {
		return m, shellmenu.Spec{}, nil, e
	}
	configPath := a.configPath
	if configPath == "" {
		configPath = os.Getenv("SMOLPNG_CONFIG")
	}
	w, e := config.Load(config.LoadOptions{Paths: p, CWD: cwd, ConfigPath: configPath, NoProject: true})
	if e != nil {
		return m, shellmenu.Spec{}, nil, e
	}
	exe, e := a.applicationExecutable()
	if e != nil {
		return m, shellmenu.Spec{}, nil, e
	}
	exe, e = filepath.Abs(exe)
	if e != nil {
		return m, shellmenu.Spec{}, nil, e
	}
	if mode == "single" && len(presets) > 0 {
		return m, shellmenu.Spec{}, nil, fmt.Errorf("custom presets require the submenu mode")
	}
	custom := []shellmenu.Action{}
	seen := map[string]bool{}
	for _, name := range presets {
		if seen[name] {
			return m, shellmenu.Spec{}, nil, fmt.Errorf("duplicate shell preset")
		}
		seen[name] = true
		eff, e := w.Resolve(name, config.Fields{})
		if e != nil {
			return m, shellmenu.Spec{}, nil, fmt.Errorf("menu preset must be installed in user config, not a transient project: %w", e)
		}
		if eff.Recipe.Engine == "tinypng" || eff.Recipe.RemotePolicy != "never" && eff.Recipe.RemotePolicy != "ask" {
			return m, shellmenu.Spec{}, nil, fmt.Errorf("shell preset %s requests uploads; menu actions are local-only", name)
		}
		custom = append(custom, shellmenu.Action{ID: "preset-" + name, Label: name + " (keep original)", Preset: name})
	}
	portable := ""
	if p.Portable {
		portable = p.Data
	}
	spec, e := shellmenu.New(mode, exe, w.User.Path, portable, custom)
	if e != nil {
		return m, spec, nil, e
	}
	data, e := shellmenu.Encode(spec)
	if e != nil {
		return m, spec, nil, e
	}
	render := shellRenderer{"shell-menu": data}
	entries := []integration.Entry{{ID: "shell-menu", Path: filepath.Join(l.State, shellmenu.MenuFile), Name: "smolpng right-click settings", Executable: exe}}

	if runtime.GOOS == "linux" {
		files, err := shellmenu.DolphinFiles(spec, integration.DesktopExec)
		if err != nil {
			return m, spec, nil, err
		}
		ids := []string{}
		for id := range files {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			render[id] = files[id]
			entries = append(entries, integration.Entry{ID: id, Path: filepath.Join(l.ServiceMenus, "smolpng-"+id+".desktop"), Name: "smolpng Dolphin service menu", Executable: exe})
		}
	}

	m.Renderer = render
	return m, spec, entries, nil
}
func (a *app) readShell() (shellmenu.Spec, shellmenu.Locations, error) {
	m, l, e := a.shellManager()
	if e != nil {
		return shellmenu.Spec{}, l, e
	}
	status, pending, e := m.Status()
	if e != nil {
		return shellmenu.Spec{}, l, e
	}
	if pending {
		return shellmenu.Spec{}, l, fmt.Errorf("shell installation was interrupted; run integrate shell repair --apply")
	}
	owned := false
	for _, s := range status {
		if s.Entry.ID == "shell-menu" && s.State == "current" {
			owned = true
		}
	}
	if !owned {
		return shellmenu.Spec{}, l, fmt.Errorf("shell settings are missing or locally modified; not running an unrecognized menu")
	}
	b, e := os.ReadFile(filepath.Join(l.State, shellmenu.MenuFile))
	if e != nil {
		return shellmenu.Spec{}, l, e
	}
	spec, e := shellmenu.Decode(b)
	return spec, l, e
}
func (a *app) shellCommand() *cobra.Command {
	root := &cobra.Command{Use: "shell", Short: "Optional Windows 11 primary menu and KDE Dolphin service menu"}
	var apply bool
	var mode string
	var presets []string
	install := &cobra.Command{Use: "install", Short: "Preview menu settings; --apply installs them (Windows activation is separate)", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		m, s, entries, e := a.makeShellPlan(mode, presets)
		if e != nil {
			return domain.E(2, e.Error(), e)
		}
		if apply {
			if e = installShellPlan(c.Context(), m, entries); e != nil {
				return domain.E(5, e.Error(), e)
			}
		}
		note := "Dolphin service file; no Nautilus or other desktop claim"
		if runtime.GOOS == "windows" {
			note = "Settings only until the matching native DLL and signed sparse package are built and activated. This is not a legacy Show more options registry verb."
		}
		return a.emit(map[string]any{"schemaVersion": 1, "applied": apply, "menu": s, "entries": entries, "activation": note})
	}}
	install.Flags().BoolVar(&apply, "apply", false, "Write hash-tracked settings and Dolphin menu")
	install.Flags().StringVar(&mode, "mode", "single", "single: Make smol; submenu: labelled actions and presets")
	install.Flags().StringArrayVar(&presets, "preset", nil, "Installed local-only user preset for submenu; repeatable")
	root.AddCommand(install)
	root.AddCommand(&cobra.Command{Use: "status", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		m, l, e := a.shellManager()
		if e != nil {
			return e
		}
		s, p, e := m.Status()
		if e != nil {
			return e
		}
		native := any("Dolphin installed-file state, not proof the desktop loaded it")
		if runtime.GOOS == "windows" {
			native, e = a.shellPackageAction(c.Context(), "status", "")
			if e != nil {
				native = map[string]any{"state": "unknown", "error": e.Error()}
			}
		}
		return a.emit(map[string]any{"schemaVersion": 1, "entries": s, "recoveryPending": p, "locations": l, "nativePackage": native})
	}})
	for _, verb := range []string{"remove", "repair"} {
		var do bool
		c := &cobra.Command{Use: verb, Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
			m, _, e := a.shellManager()
			if e != nil {
				return e
			}
			if do {
				if verb == "repair" {
					e = m.Repair(c.Context())
				} else {
					e = m.Remove(c.Context(), "")
				}
				if e != nil {
					return domain.E(5, e.Error(), e)
				}
			}
			s, p, e := m.Status()
			if e != nil {
				return e
			}
			return a.emit(map[string]any{"schemaVersion": 1, "applied": do, "entries": s, "recoveryPending": p, "packageUnregistered": false})
		}}
		c.Flags().BoolVar(&do, "apply", false, "Apply the previewed owned-file operation; modified files are preserved")
		root.AddCommand(c)
	}
	for _, verb := range []string{"activate", "deactivate"} {
		var pkg string
		var do bool
		c := &cobra.Command{Use: verb, Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
			if runtime.GOOS != "windows" {
				return domain.E(2, "Windows package registration is Windows-only", nil)
			}
			if verb == "activate" {
				s, _, e := a.readShell()
				if e != nil {
					return e
				}
				if _, e = shellmenu.CheckPackage(pkg, filepath.Dir(s.Executable), Commit, runtime.GOARCH); e != nil {
					return domain.E(2, e.Error(), e)
				}
			}
			if !do {
				return a.emit(map[string]any{"schemaVersion": 1, "applied": false, "action": verb, "package": pkg, "packageName": shellmenu.PackageName, "trust": "Signature trust is checked by Windows; smolpng never installs a root certificate"})
			}
			result, e := a.shellPackageAction(c.Context(), verb, pkg)
			if e != nil {
				return domain.E(5, e.Error(), e)
			}
			return a.emit(result)
		}}
		c.Flags().BoolVar(&do, "apply", false, "Register/unregister the selected per-user Windows shell package")
		if verb == "activate" {
			c.Flags().StringVar(&pkg, "package", "", "Signed sparse MSIX produced by native/windows-shell/build.ps1")
		}
		root.AddCommand(c)
	}
	return root
}
func (a *app) shellUI(ctx context.Context, p ui.Prompter) error {
	choice, e := p.Choose(ctx, "Image right-click menu", "Single: Make smol preserves originals. Submenu: explicitly labelled replacement, interactive and custom user recipes. Windows 11 needs the signed native shell package; Linux support is Dolphin-specific.", []ui.Option{{Label: "Install one Make smol action", Value: "single"}, {Label: "Install smolpng submenu...", Value: "submenu"}, {Label: "Activate a signed Windows 11 menu package...", Value: "activate"}, {Label: "Remove menu entries/settings", Value: "remove"}, {Label: "Back", Value: "back"}}, "single")
	if e != nil || choice == "back" {
		return e
	}
	if choice == "activate" {
		if runtime.GOOS != "windows" {
			return a.notice(ctx, p, "Not Windows", "Dolphin needs no MSIX; choose Install above.")
		}
		values, e := p.Edit(ctx, "Windows 11 primary menu", []ui.Field{{Key: "package", Label: "Trusted signed sparse MSIX path (native/windows-shell/build.ps1)"}})
		if e != nil {
			return e
		}
		s, _, e := a.readShell()
		if e == nil {
			_, e = shellmenu.CheckPackage(values["package"], filepath.Dir(s.Executable), Commit, runtime.GOARCH)
		}
		if e != nil {
			return a.notice(ctx, p, "Package not ready", e.Error())
		}
		confirm, e := p.Choose(ctx, "Register per-user package?", values["package"]+"\nNo certificate trust, PATH or default associations will be changed.", []ui.Option{{Label: "Register", Value: "yes"}, {Label: "Back", Value: "no"}}, "no")
		if e != nil || confirm != "yes" {
			return e
		}
		result, e := a.shellPackageAction(ctx, "activate", values["package"])
		if e != nil {
			return a.notice(ctx, p, "Registration stopped", e.Error())
		}
		return a.notice(ctx, p, "Windows package registered", fmt.Sprint(result))
	}
	if choice == "remove" {
		m, _, e := a.shellManager()
		if e != nil {
			return e
		}
		confirm, e := p.Choose(ctx, "Remove app-owned menu settings?", "This hides native actions. Modified files are preserved; Windows package identity can separately be unregistered with integrate shell deactivate --apply.", []ui.Option{{Label: "Remove", Value: "yes"}, {Label: "Back", Value: "no"}}, "no")
		if e != nil || confirm != "yes" {
			return e
		}
		if e = m.Remove(ctx, ""); e != nil {
			return a.notice(ctx, p, "Removal stopped", e.Error())
		}
		return a.notice(ctx, p, "Menu disabled", "App-owned settings/service file removed.")
	}
	names := []string{}
	if choice == "submenu" {
		v, e := p.Edit(ctx, "Saved user presets", []ui.Field{{Key: "presets", Label: "Additional installed local-only presets, comma-separated (optional)"}})
		if e != nil {
			return e
		}
		for _, s := range strings.Split(v["presets"], ",") {
			if s = strings.TrimSpace(s); s != "" {
				names = append(names, s)
			}
		}
	}
	m, s, entries, e := a.makeShellPlan(choice, names)
	if e != nil {
		return a.notice(ctx, p, "Cannot prepare menu", e.Error())
	}
	data, _ := json.MarshalIndent(s, "", "  ")
	confirm, e := p.Choose(ctx, "Install this menu?", string(data), []ui.Option{{Label: "Install", Value: "yes"}, {Label: "Back", Value: "no"}}, "no")
	if e != nil || confirm != "yes" {
		return e
	}
	if e = installShellPlan(ctx, m, entries); e != nil {
		return a.notice(ctx, p, "Installation stopped", e.Error())
	}
	note := "Dolphin service file installed; desktop activation is still a native check."
	if runtime.GOOS == "windows" {
		note = "Menu settings saved. Activate the matching signed sparse package to put it in the Windows 11 primary menu. Settings alone do not register an extension."
	}
	return a.notice(ctx, p, "Right-click settings installed", note)
}
func (a *app) shellRunCommand() *cobra.Command {
	var request, id, generation string
	c := &cobra.Command{Use: "_shell-run", Hidden: true, RunE: func(c *cobra.Command, args []string) error {
		spec, l, e := a.readShell()
		if e != nil {
			return domain.E(3, e.Error(), e)
		}
		if request != "" {
			if len(args) > 0 || id != "" || generation != "" {
				return domain.E(2, "shell handoff modes conflict", nil)
			}
			q, e := shellmenu.Consume(request, l.Requests)
			if e != nil {
				return domain.E(3, e.Error(), e)
			}
			args = q.Files
			id = q.Action
			generation = q.Generation
		}
		if generation != spec.Generation {
			return domain.E(3, "stale menu invocation; reopen the file manager context menu", nil)
		}
		action, e := spec.Action(id)
		if e != nil {
			return domain.E(3, e.Error(), e)
		}
		if e = shellmenu.ValidateFiles(args); e != nil {
			return domain.E(3, e.Error(), e)
		}
		exe, e := a.applicationExecutable()
		if e != nil {
			return e
		}
		if !output.SameFile(exe, spec.Executable) {
			return domain.E(3, "menu belongs to another installation; install it again from this executable", nil)
		}
		// The native bridge conveys only a menu ID and files. Privacy/output policy is
		// reconstructed here, not accepted as arbitrary command arguments from a menu.
		a.noProject = true
		a.home = spec.Home
		a.configPath = spec.Config
		op := &cobra.Command{Use: "shell-operation"}
		op.SetContext(c.Context())
		f := &compressFlags{}
		f.bind(op)
		for k, v := range map[string]string{"preset": action.Preset, "engine": "local", "remote-policy": "never", "overwrite": "false", "output-dir": "", "template": "{stem}.smol{ext}", "in-place": fmt.Sprint(action.InPlace), "recursive": "false", "copy": "none"} {
			if e = op.Flags().Set(k, v); e != nil {
				return e
			}
		}
		if action.Interactive {
			return a.interactive(op, args, f)
		}
		var report domain.Report
		var eff config.Effective
		e = a.compression(op, args, f, func(r domain.Report, c config.Effective) error { report = r; eff = c; return nil })
		if e != nil {
			row := domain.NewResult("<shell selection>", action.Preset)
			row.Fail(e)
			report = domain.Summarize([]domain.Result{row})
		}
		receipt, _ := a.finishHuman(c.Context(), report, eff, "")
		if a.env.Desktop != nil {
			_ = a.env.Desktop.Notify(c.Context(), "smolpng", fmt.Sprintf("%d image result(s), exit %d. Recover with smolpng last.", len(report.Results), report.ExitCode))
		}
		desktop.Hold(c.Context(), receipt, 5*time.Minute)
		if report.ExitCode != 0 {
			return reported(report.ExitCode)
		}
		return nil
	}}
	c.Flags().StringVar(&request, "request", "", "Owned binary selection handoff")
	c.Flags().StringVar(&id, "action", "", "Installed menu action")
	c.Flags().StringVar(&generation, "generation", "", "Exact displayed menu generation")
	return c
}

func installShellPlan(ctx context.Context, m integration.Manager, entries []integration.Entry) error {
	keep := map[string]bool{}
	for _, x := range entries {
		keep[x.ID] = true
	}
	status, _, e := m.Status()
	if e != nil {
		return e
	}
	for _, x := range status {
		if !keep[x.Entry.ID] && x.State != "current" && x.State != "missing" {
			return fmt.Errorf("obsolete service file was modified; preserve/review it before changing menu mode: %s", x.Entry.Path)
		}
	}
	if e = m.Install(ctx, entries); e != nil {
		return e
	}
	for _, x := range status {
		if !keep[x.Entry.ID] {
			if e = m.Remove(ctx, x.Entry.ID); e != nil {
				return e
			}
		}
	}
	return nil
}
