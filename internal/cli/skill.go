// SPDX-License-Identifier: MIT
package cli

import (
	"os"
	"path/filepath"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"github.com/Wolfsblvt/smolpng/internal/skill"
	bundle "github.com/Wolfsblvt/smolpng/skills"
	"github.com/spf13/cobra"
)

func (a *app) skillManager() (skill.Manager, error) {
	p, e := a.dirs()
	if e != nil {
		return skill.Manager{}, e
	}
	b, e := skill.ReadBundle(bundle.Files())
	if e != nil {
		return skill.Manager{}, e
	}
	return skill.Manager{StatePath: filepath.Join(p.State, "skill-installations.json"), Bundle: b, AppVersion: Version}, nil
}
func (a *app) skillLocations() (string, string, error) {
	home, e := os.UserHomeDir()
	if e != nil {
		return "", "", e
	}
	cwd, e := a.cwd()
	return home, cwd, e
}
func skillError(e error) error {
	if e == nil {
		return nil
	}
	return domain.E(domain.OptimizerFailed, e.Error(), e)
}
func (a *app) skillCommand() *cobra.Command {
	root := &cobra.Command{Use: "skill", Short: "Install and maintain the version-bound Agent Skill"}
	root.AddCommand(&cobra.Command{Use: "status", Args: cobra.NoArgs, Short: "Read installation hashes and manual-replacement state without writes", RunE: func(c *cobra.Command, _ []string) error {
		m, e := a.skillManager()
		if e != nil {
			return e
		}
		s, e := m.Inspect()
		if e != nil {
			return skillError(e)
		}
		home, cwd, e := a.skillLocations()
		if e != nil {
			return e
		}
		return a.emit(map[string]any{"schemaVersion": 1, "status": s, "targets": skill.Targets(home, cwd), "registry": m.StatePath})
	}})
	root.AddCommand(&cobra.Command{Use: "print", Args: cobra.NoArgs, Short: "Print the embedded SKILL.md", RunE: func(c *cobra.Command, _ []string) error {
		m, e := a.skillManager()
		if e != nil {
			return e
		}
		content := m.Bundle.Files["SKILL.md"]
		if a.json || a.jsonl {
			return a.emit(map[string]any{"schemaVersion": 1, "bundle": m.Bundle.Version, "markdown": string(content)})
		}
		_, e = a.env.Out.Write(content)
		return e
	}})
	var target, parent string
	var installApply bool
	install := &cobra.Command{Use: "install", Short: "Preview one destination; --apply installs without overwriting existing content", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		home, cwd, e := a.skillLocations()
		if e != nil {
			return e
		}
		t, e := skill.ResolveTarget(target, parent, home, cwd)
		if e != nil {
			return domain.E(2, e.Error(), e)
		}
		m, e := a.skillManager()
		if e != nil {
			return e
		}
		state, e := m.Inspect()
		if e != nil {
			return skillError(e)
		}
		if installApply {
			if e = m.Install(c.Context(), t.Destination, t.Name, t.Managed); e != nil {
				return skillError(e)
			}
		}
		return a.emit(map[string]any{"schemaVersion": 1, "applied": installApply, "target": t, "bundle": m.Bundle, "registry": m.StatePath, "detected": skill.Targets(home, cwd), "recoveryPending": state.RecoveryPending})
	}}
	install.Flags().StringVar(&target, "target", "auto", "auto/agents/codex, claude, project, or path:<parent>; auto selects only shared personal Skills")
	install.Flags().StringVar(&parent, "path", "", "Custom parent directory; the dedicated smolpng child is created")
	install.Flags().BoolVar(&installApply, "apply", false, "Write the previewed Skill installation")
	root.AddCommand(install)
	var syncApply, custom bool
	var syncDest string
	sync := &cobra.Command{Use: "sync", Aliases: []string{"update"}, Args: cobra.NoArgs, Short: "Synchronize unmodified installations; custom paths need explicit consent", RunE: func(c *cobra.Command, _ []string) error {
		m, e := a.skillManager()
		if e != nil {
			return e
		}
		before, e := m.Inspect()
		if e != nil {
			return skillError(e)
		}
		if !syncApply {
			return a.emit(map[string]any{"schemaVersion": 1, "applied": false, "includeCustom": custom, "destination": syncDest, "status": before})
		}
		if syncDest != "" {
			dest, e := a.absoluteInvocationPath(syncDest)
			if e != nil {
				return e
			}
			if e = m.Sync(c.Context(), dest); e != nil {
				return skillError(e)
			}
			after, e := m.Inspect()
			if e != nil {
				return e
			}
			return a.emit(map[string]any{"schemaVersion": 1, "applied": true, "status": after})
		}
		results, e := m.SyncAll(c.Context(), custom)
		if e != nil {
			return skillError(e)
		}
		if e = a.emit(map[string]any{"schemaVersion": 1, "applied": true, "results": results}); e != nil {
			return e
		}
		for _, r := range results {
			if r.Action == "refused" {
				return reported(domain.PartialFailure)
			}
		}
		return nil
	}}
	sync.Flags().BoolVar(&syncApply, "apply", false, "Apply the safe synchronization plan")
	sync.Flags().BoolVar(&custom, "include-custom", false, "Consent to syncing unmodified custom installations too")
	sync.Flags().StringVar(&syncDest, "destination", "", "Synchronize exactly this tracked destination (explicit consent for custom paths)")
	root.AddCommand(sync)
	for _, action := range []string{"uninstall", "forget"} {
		var dest string
		var apply bool
		cmd := &cobra.Command{Use: action, Args: cobra.NoArgs, Short: map[string]string{"uninstall": "Remove only an unmodified tracked Skill", "forget": "Stop tracking a destination; do not delete its content"}[action], RunE: func(c *cobra.Command, _ []string) error {
			m, e := a.skillManager()
			if e != nil {
				return e
			}
			status, e := m.Inspect()
			if e != nil {
				return skillError(e)
			}
			if !apply {
				return a.emit(map[string]any{"schemaVersion": 1, "applied": false, "operation": action, "destination": dest, "status": status})
			}
			if dest == "" {
				return domain.E(2, "select one tracked --destination before applying", nil)
			}
			p, e := a.absoluteInvocationPath(dest)
			if e != nil {
				return e
			}
			if action == "uninstall" {
				e = m.Uninstall(c.Context(), p)
			} else {
				e = m.Forget(c.Context(), p)
			}
			if e != nil {
				return skillError(e)
			}
			return a.emit(map[string]any{"schemaVersion": 1, "applied": true, "operation": action, "destination": p})
		}}
		cmd.Flags().StringVar(&dest, "destination", "", "Exact destination from skill status")
		cmd.Flags().BoolVar(&apply, "apply", false, "Apply this management operation")
		root.AddCommand(cmd)
	}
	var repairApply bool
	repair := &cobra.Command{Use: "repair", Args: cobra.NoArgs, Short: "Recover a journaled interrupted directory update when hashes still agree", RunE: func(c *cobra.Command, _ []string) error {
		m, e := a.skillManager()
		if e != nil {
			return e
		}
		if repairApply {
			if e = m.Recover(c.Context()); e != nil {
				return skillError(e)
			}
		}
		s, e := m.Inspect()
		if e != nil {
			return skillError(e)
		}
		return a.emit(map[string]any{"schemaVersion": 1, "applied": repairApply, "status": s})
	}}
	repair.Flags().BoolVar(&repairApply, "apply", false, "Recover only the unambiguous journaled state")
	root.AddCommand(repair)
	root.AddCommand(&cobra.Command{Use: "ignore-update", Args: cobra.NoArgs, Short: "Remember this embedded bundle as ignored for interactive synchronization prompts", RunE: func(c *cobra.Command, _ []string) error {
		m, e := a.skillManager()
		if e != nil {
			return e
		}
		if e = m.IgnoreCurrent(c.Context()); e != nil {
			return skillError(e)
		}
		return a.emit(map[string]any{"schemaVersion": 1, "ignoredBundle": m.Bundle.Version})
	}})
	var packageTarget, packageOut string
	pack := &cobra.Command{Use: "package", Args: cobra.NoArgs, Short: "Create an instruction-only uploadable ChatGPT Skill ZIP", RunE: func(c *cobra.Command, _ []string) error {
		if packageTarget != "chatgpt" {
			return domain.E(2, "package target must be chatgpt", nil)
		}
		m, e := a.skillManager()
		if e != nil {
			return e
		}
		data, e := m.Bundle.Archive()
		if e != nil {
			return skillError(e)
		}
		dest := packageOut
		if dest == "" {
			p, e := a.dirs()
			if e != nil {
				return e
			}
			dest = filepath.Join(p.Detached(), "smolpng-skill-"+m.Bundle.Version+".zip")
		} else {
			dest, e = a.absoluteInvocationPath(dest)
			if e != nil {
				return e
			}
		}
		if e = fsutil.AtomicWrite(c.Context(), dest, data, 0644, false); e != nil {
			return skillError(e)
		}
		return a.emit(map[string]any{"schemaVersion": 1, "output": dest, "bundle": m.Bundle, "containsExecutable": false, "target": "chatgpt", "uploadAcceptance": "not tested"})
	}}
	pack.Flags().StringVar(&packageTarget, "target", "chatgpt", "Instruction-only Skill archive target")
	pack.Flags().StringVar(&packageOut, "output", "", "New ZIP destination; never replaces an existing archive")
	root.AddCommand(pack)
	return root
}
func (a *app) absoluteInvocationPath(path string) (string, error) {
	cwd, e := a.cwd()
	if e != nil {
		return "", e
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(cwd, path)
	}
	return filepath.Clean(path), nil
}
