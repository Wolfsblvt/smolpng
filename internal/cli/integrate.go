// SPDX-License-Identifier: MIT
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/config"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/integration"
	"github.com/Wolfsblvt/smolpng/internal/ui"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func (a *app) applicationExecutable() (string, error) {
	p := ""
	var e error
	if a.env.Executable != nil {
		p, e = a.env.Executable()
	} else {
		p, e = os.Executable()
	}
	if e != nil {
		return "", e
	}
	if strings.EqualFold(filepath.Base(p), "smolpngw.exe") {
		p = filepath.Join(filepath.Dir(p), "smolpng.exe")
	}
	return p, nil
}
func (a *app) integrations() (integration.Manager, error) {
	p, e := a.dirs()
	if e != nil {
		return integration.Manager{}, e
	}
	l, e := integration.NativeLocations()
	if a.env.IntegrationLocations != nil {
		l = *a.env.IntegrationLocations
		e = nil
	}
	if e != nil {
		return integration.Manager{}, e
	}
	renderer := a.env.IntegrationRenderer
	if renderer == nil {
		renderer = integration.Native{Runner: a.env.Runner}
	}
	return integration.Manager{StatePath: integration.RegistryName(p.State), Version: Version, Allowed: l, Renderer: renderer}, nil
}
func (a *app) integrationPlan(m integration.Manager, desktop, sendto bool, preset, name string, interactive bool) ([]integration.Entry, error) {
	p, e := a.dirs()
	if e != nil {
		return nil, e
	}
	exe, e := a.applicationExecutable()
	if e != nil {
		return nil, e
	}
	home, e := os.UserHomeDir()
	if e != nil {
		return nil, e
	}
	if preset == "" {
		preset = "auto"
	}
	// A durable launcher starts from the user's home. Resolve its named recipe
	// without transient project definitions rather than generating a broken link.
	cwd, e := a.cwd()
	if e != nil {
		return nil, e
	}
	configPath := a.configPath
	if configPath == "" {
		configPath = os.Getenv("SMOLPNG_CONFIG")
	}
	w, e := config.Load(config.LoadOptions{Paths: p, CWD: cwd, ConfigPath: configPath, NoProject: true})
	if e != nil {
		return nil, e
	}
	if _, e = w.Resolve(preset, config.Fields{}); e != nil {
		return nil, fmt.Errorf("launcher needs a built-in or installed user preset; import a project-only preset into user storage first: %w", e)
	}
	if configPath != "" {
		configPath = w.User.Path
	}
	portable := ""
	if p.Portable {
		portable = p.Data
	}
	return integration.Plan(m.Allowed, integration.Request{OS: runtime.GOOS, Executable: exe, ConfigPath: configPath, NoProject: true, WorkingDirectory: home, Home: portable, Preset: preset, Desktop: desktop, SendTo: sendto, Custom: name != "", Name: name, Interactive: interactive})
}
func (a *app) integrationCommand() *cobra.Command {
	root := &cobra.Command{Use: "integrate", Short: "Opt-in per-user launchers with hash-protected removal"}
	root.AddCommand(a.shellCommand())
	root.AddCommand(&cobra.Command{Use: "status", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		m, e := a.integrations()
		if e != nil {
			return e
		}
		s, pending, e := m.Status()
		if e != nil {
			return skillError(e)
		}
		return a.emit(map[string]any{"schemaVersion": 1, "entries": s, "recoveryPending": pending, "locations": m.Allowed})
	}})
	var apply, desktop, sendto bool
	var preset string
	install := &cobra.Command{Use: "install", Short: "Preview menu launchers; --apply writes them", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		m, e := a.integrations()
		if e != nil {
			return e
		}
		p, e := a.integrationPlan(m, desktop, sendto, preset, "", false)
		if e != nil {
			return domain.E(2, e.Error(), e)
		}
		if apply {
			if e = m.Install(c.Context(), p); e != nil {
				return skillError(e)
			}
		}
		return a.emit(map[string]any{"schemaVersion": 1, "applied": apply, "entries": p})
	}}
	install.Flags().BoolVar(&apply, "apply", false, "Create only previewed app-owned launchers")
	install.Flags().BoolVar(&desktop, "desktop", false, "Also create background and interactive desktop targets")
	install.Flags().BoolVar(&sendto, "send-to", false, "Also create Windows Send To target")
	install.Flags().StringVar(&preset, "preset", "auto", "Preset used by launchers")
	root.AddCommand(install)
	var removeApply bool
	var removeID string
	remove := &cobra.Command{Use: "remove", Short: "Preview removal; modified files are always preserved", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		m, e := a.integrations()
		if e != nil {
			return e
		}
		s, _, e := m.Status()
		if e != nil {
			return e
		}
		if removeApply {
			if e = m.Remove(c.Context(), removeID); e != nil {
				return skillError(e)
			}
		}
		return a.emit(map[string]any{"schemaVersion": 1, "applied": removeApply, "id": removeID, "entries": s})
	}}
	remove.Flags().BoolVar(&removeApply, "apply", false, "Remove unmodified app-created launchers")
	remove.Flags().StringVar(&removeID, "id", "", "Remove one ID; empty removes all app-managed entries")
	root.AddCommand(remove)
	var repairApply bool
	repair := &cobra.Command{Use: "repair", Short: "Reconcile an interrupted journal only when file hashes agree", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		m, e := a.integrations()
		if e != nil {
			return e
		}
		if repairApply {
			if e = m.Repair(c.Context()); e != nil {
				return skillError(e)
			}
		}
		s, p, e := m.Status()
		if e != nil {
			return e
		}
		return a.emit(map[string]any{"schemaVersion": 1, "applied": repairApply, "entries": s, "recoveryPending": p})
	}}
	repair.Flags().BoolVar(&repairApply, "apply", false, "Repair recorded interrupted operation")
	root.AddCommand(repair)
	shortcut := &cobra.Command{Use: "shortcut", Short: "Create a preset-specific desktop drop target"}
	var shortcutApply, interactive bool
	var shortcutPreset string
	create := &cobra.Command{Use: "create NAME", Args: cobra.ExactArgs(1), RunE: func(c *cobra.Command, args []string) error {
		m, e := a.integrations()
		if e != nil {
			return e
		}
		p, e := a.integrationPlan(m, true, false, shortcutPreset, args[0], interactive)
		if e != nil {
			return domain.E(2, e.Error(), e)
		}
		if shortcutApply {
			if e = m.Install(c.Context(), p); e != nil {
				return skillError(e)
			}
		}
		return a.emit(map[string]any{"schemaVersion": 1, "applied": shortcutApply, "entries": p})
	}}
	create.Flags().BoolVar(&shortcutApply, "apply", false, "Create the previewed shortcut")
	create.Flags().BoolVar(&interactive, "interactive", false, "Open the recipe picker instead of background processing")
	create.Flags().StringVar(&shortcutPreset, "preset", "auto", "Installed preset name")
	shortcut.AddCommand(create)
	root.AddCommand(shortcut)
	return root
}
func (a *app) installIntegrationsUI(ctx context.Context, p ui.Prompter) error {
	m, e := a.integrations()
	if e != nil {
		return e
	}
	action, e := p.Choose(ctx, "Desktop integration", "Per-user only. Existing unrelated or modified files are never replaced.", []ui.Option{{Label: "Right-click menu (Windows 11 / Dolphin)...", Value: "shell"}, {Label: "Menu launchers", Value: "menu"}, {Label: "Menu plus desktop drop targets", Value: "desktop"}, {Label: "Menu, desktop and Windows Send To", Value: "all"}, {Label: "Back", Value: "back"}}, "menu")
	if errors.Is(e, ui.ErrBack) || action == "back" {
		return nil
	}
	if e != nil {
		return e
	}
	if action == "shell" {
		return a.shellUI(ctx, p)
	}
	plan, e := a.integrationPlan(m, action != "menu", action == "all", "auto", "", false)
	if e != nil {
		return a.notice(ctx, p, "Integration unavailable", e.Error())
	}
	b, _ := json.MarshalIndent(plan, "", "  ")
	confirm, e := p.Choose(ctx, "Install these exact launchers?", string(b), []ui.Option{{Label: "Install", Value: "yes"}, {Label: "Back", Value: "no"}}, "no")
	if e != nil {
		return e
	}
	if confirm != "yes" {
		return nil
	}
	if e = m.Install(ctx, plan); e != nil {
		return a.notice(ctx, p, "Installation stopped", e.Error())
	}
	return a.notice(ctx, p, "Launchers installed", fmt.Sprintf("%d per-user launchers recorded. Native drop/clipboard behavior still depends on your desktop.", len(plan)))
}
