// SPDX-License-Identifier: MIT
package cli

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/Wolfsblvt/smolpng/internal/config"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"github.com/Wolfsblvt/smolpng/internal/paths"
	"github.com/Wolfsblvt/smolpng/internal/toolchain"
	"github.com/spf13/cobra"
)

func (a *app) loadConfig(p paths.Paths) (*config.Workspace, error) {
	cwd, e := a.cwd()
	if e != nil {
		return nil, e
	}
	configPath := a.configPath
	if configPath == "" {
		configPath = os.Getenv("SMOLPNG_CONFIG")
	}
	w, e := config.Load(config.LoadOptions{Paths: p, CWD: cwd, ConfigPath: configPath, NoProject: a.noProject})
	if e != nil {
		return nil, domain.E(2, e.Error(), e)
	}
	a.configuredTools = toolchain.Paths{JPEGOptim: w.User.Document.Tools.JPEGOptim, PNGQuant: w.User.Document.Tools.PNGQuant, OxiPNG: w.User.Document.Tools.OxiPNG, ManagedRoot: filepath.Join(p.Data, "tools"), Preference: w.User.Document.Tools.Preference}
	base := filepath.Dir(w.User.Path)
	if v := a.configuredTools.JPEGOptim; v != "" && !filepath.IsAbs(v) {
		a.configuredTools.JPEGOptim = filepath.Join(base, v)
	}
	if v := a.configuredTools.PNGQuant; v != "" && !filepath.IsAbs(v) {
		a.configuredTools.PNGQuant = filepath.Join(base, v)
	}
	if v := a.configuredTools.OxiPNG; v != "" && !filepath.IsAbs(v) {
		a.configuredTools.OxiPNG = filepath.Join(base, v)
	}
	return w, nil
}
func (a *app) configCommand() *cobra.Command {
	root := &cobra.Command{Use: "config", Short: "User/project settings and effective-value provenance"}
	root.AddCommand(&cobra.Command{Use: "path", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		p, e := a.dirs()
		if e != nil {
			return e
		}
		cwd, e := a.cwd()
		if e != nil {
			return e
		}
		user := p.Config
		if a.configPath != "" {
			user = a.configPath
			if !filepath.IsAbs(user) {
				user = filepath.Join(cwd, user)
			}
		}
		project := ""
		if !a.noProject {
			project, e = config.FindProject(cwd)
			if e != nil {
				return e
			}
		}
		return a.emit(map[string]any{"schemaVersion": 1, "user": user, "project": project, "presets": filepath.Join(p.Data, "presets.toml"), "paths": p})
	}})
	var effective, provenance bool
	var preset string
	show := &cobra.Command{Use: "show", Args: cobra.NoArgs, Short: "Show loaded sources or effective settings", RunE: func(c *cobra.Command, _ []string) error {
		p, e := a.dirs()
		if e != nil {
			return e
		}
		w, e := a.loadConfig(p)
		if e != nil {
			return e
		}
		if !effective && !provenance {
			return a.emit(w)
		}
		eff, e := w.Resolve(preset, config.Fields{})
		if e != nil {
			return domain.E(2, e.Error(), e)
		}
		if !provenance {
			eff.Provenance = nil
		}
		return a.emit(eff)
	}}
	show.Flags().BoolVar(&effective, "effective", false, "Resolve defaults and selected preset")
	show.Flags().BoolVar(&provenance, "provenance", false, "Include each effective field's file and exact TOML key")
	show.Flags().StringVar(&preset, "preset", "", "Inspect a particular preset")
	root.AddCommand(show)
	root.AddCommand(&cobra.Command{Use: "validate", Args: cobra.NoArgs, Short: "Validate config, inherited recipes and selected default", RunE: func(c *cobra.Command, _ []string) error {
		p, e := a.dirs()
		if e != nil {
			return e
		}
		w, e := a.loadConfig(p)
		if e != nil {
			return e
		}
		if _, e = w.Resolve("", config.Fields{}); e != nil {
			return domain.E(2, e.Error(), e)
		}
		return a.emit(map[string]any{"schemaVersion": 1, "valid": true, "user": w.User.Path, "project": w.Project.Path, "presets": w.StorePath})
	}})
	var project bool
	init := &cobra.Command{Use: "init", Args: cobra.NoArgs, Short: "Create a minimal config without replacing an existing file", RunE: func(c *cobra.Command, _ []string) error {
		p, e := a.dirs()
		if e != nil {
			return e
		}
		cwd, e := a.cwd()
		if e != nil {
			return e
		}
		dest := p.Config
		if a.configPath != "" {
			dest = a.configPath
			if !filepath.IsAbs(dest) {
				dest = filepath.Join(cwd, dest)
			}
		}
		data := []byte("# smolpng user settings. Presets imported by the app use presets.toml.\nconfig_version = 1\n\n[defaults]\npreset = 'auto'\nremote_policy = 'never'\n\n[output]\ntemplate = '{stem}.smol{ext}'\ncollision = 'increment'\n\n[ui]\nheader = 'cute'\n")
		if project {
			dest = filepath.Join(cwd, ".smolpng.toml")
			data = []byte("# Project settings, inherited by humans, scripts, and agents.\nconfig_version = 1\n\n[defaults]\npreset = 'auto'\n\n[output]\ntemplate = '{stem}.smol{ext}'\n")
		}
		if e = fsutil.AtomicWrite(c.Context(), dest, data, 0600, false); e != nil {
			return domain.E(2, "config already exists or cannot be created: "+dest, e)
		}
		return a.emit(map[string]any{"schemaVersion": 1, "created": dest})
	}}
	init.Flags().BoolVar(&project, "project", false, "Create .smolpng.toml in the current directory")
	root.AddCommand(init)
	root.AddCommand(&cobra.Command{Use: "set KEY VALUE", Args: cobra.ExactArgs(2), Short: "Set a supported user preference while preserving comments", RunE: func(c *cobra.Command, args []string) error {
		p, e := a.dirs()
		if e != nil {
			return e
		}
		w, e := a.loadConfig(p)
		if e != nil {
			return e
		}
		value, e := config.SettingValue(args[0], args[1])
		if e != nil {
			return domain.E(2, e.Error(), e)
		}
		if e = w.SaveUserSettings(c.Context(), map[string]any{args[0]: value}); e != nil {
			return domain.E(2, e.Error(), e)
		}
		return a.emit(map[string]any{"schemaVersion": 1, "path": w.User.Path, "key": args[0], "value": value})
	}})
	return root
}
func (a *app) presetCommand() *cobra.Command {
	root := &cobra.Command{Use: "preset", Short: "Inspect, share and manage data-only named recipes"}
	load := func() (*config.Workspace, error) {
		p, e := a.dirs()
		if e != nil {
			return nil, e
		}
		return a.loadConfig(p)
	}
	root.AddCommand(&cobra.Command{Use: "list", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		w, e := load()
		if e != nil {
			return e
		}
		return a.emit(map[string]any{"schemaVersion": 1, "presets": w.List(), "menu": w.MenuEntries()})
	}})
	root.AddCommand(&cobra.Command{Use: "show NAME", Args: cobra.ExactArgs(1), RunE: func(c *cobra.Command, args []string) error {
		w, e := load()
		if e != nil {
			return e
		}
		eff, e := w.Resolve(args[0], config.Fields{})
		if e != nil {
			return domain.E(2, e.Error(), e)
		}
		definition, _ := w.Definition(args[0])
		return a.emit(map[string]any{"schemaVersion": 1, "definition": definition, "effective": eff})
	}})
	var clip, apply, replace bool
	imp := &cobra.Command{Use: "import [FILE|-]", Short: "Preview preset import; --apply commits after validation", Args: cobra.MaximumNArgs(1), RunE: func(c *cobra.Command, args []string) error {
		w, e := load()
		if e != nil {
			return e
		}
		var data []byte
		if clip {
			if len(args) != 0 {
				return domain.E(2, "choose a file or --clipboard", nil)
			}
			if a.env.Clipboard == nil {
				return domain.E(3, "clipboard is unavailable", nil)
			}
			snapshot, er := a.env.Clipboard.Read(c.Context())
			if er != nil {
				return domain.E(3, "cannot read clipboard", er)
			}
			data = []byte(snapshot.Text)
		} else {
			if len(args) == 0 {
				return domain.E(2, "provide a preset file, -, or --clipboard", nil)
			}
			if args[0] == "-" {
				data, e = config.ReadBounded(a.env.In)
			} else {
				path := args[0]
				if !filepath.IsAbs(path) {
					path = filepath.Join(w.CWD, path)
				}
				f, er := os.Open(path)
				if er != nil {
					return domain.E(3, "cannot read preset file", er)
				}
				data, e = config.ReadBounded(f)
				f.Close()
			}
			if e != nil {
				return domain.E(3, e.Error(), e)
			}
		}
		bundle, e := config.ParseBundle(data)
		if e != nil {
			return domain.E(2, e.Error(), e)
		}
		plan, e := w.PlanImport(bundle, replace)
		if e != nil {
			return domain.E(2, e.Error(), e)
		}
		if apply {
			if e = w.ApplyImport(c.Context(), &plan); e != nil {
				return domain.E(5, e.Error(), e)
			}
		}
		return a.emit(plan)
	}}
	imp.Flags().BoolVar(&clip, "clipboard", false, "Read textual clipboard without replacing its image")
	imp.Flags().BoolVar(&apply, "apply", false, "Commit the import; otherwise only preview")
	imp.Flags().BoolVar(&replace, "replace", false, "Replace app-managed name collisions instead of renaming")
	root.AddCommand(imp)
	var resolved bool
	var dest, newName string
	export := &cobra.Command{Use: "export NAME", Short: "Export inheritance plus parents, or freeze effective values", Args: cobra.ExactArgs(1), RunE: func(c *cobra.Command, args []string) error {
		w, e := load()
		if e != nil {
			return e
		}
		data, e := w.Export(args[0], resolved)
		if e != nil {
			return domain.E(2, e.Error(), e)
		}
		if newName != "" {
			old := args[0]
			if domain.IsBuiltin(old) && resolved {
				old += "-copy"
			}
			data, e = config.RenameExport(data, old, newName)
			if e != nil {
				return domain.E(2, e.Error(), e)
			}
		}
		if dest != "" {
			out := dest
			if !filepath.IsAbs(out) {
				out = filepath.Join(w.CWD, out)
			}
			if e = fsutil.AtomicWrite(c.Context(), out, data, 0600, false); e != nil {
				return domain.E(5, "cannot create export without replacing existing content", e)
			}
			return a.emit(map[string]any{"schemaVersion": 1, "output": out})
		}
		if a.json || a.jsonl {
			return a.emit(map[string]any{"schemaVersion": 1, "toml": string(data)})
		}
		_, e = a.env.Out.Write(data)
		return e
	}}
	export.Flags().BoolVar(&resolved, "resolved", false, "Freeze resolved recipe values without embedding machine-specific paths")
	export.Flags().StringVar(&dest, "output", "", "Write to a new file instead of stdout")
	export.Flags().StringVar(&newName, "name", "", "Name the exported leaf (useful for copying built-ins)")
	root.AddCommand(export)
	for _, action := range []string{"pin", "unpin"} {
		pinned := action == "pin"
		root.AddCommand(&cobra.Command{Use: action + " NAME", Args: cobra.ExactArgs(1), RunE: func(c *cobra.Command, args []string) error {
			w, e := load()
			if e != nil {
				return e
			}
			if e = w.ChangeMenu(c.Context(), args[0], &pinned, nil); e != nil {
				return domain.E(2, e.Error(), e)
			}
			return a.emit(map[string]any{"schemaVersion": 1, "name": args[0], "pinned": pinned})
		}})
	}
	var position int
	move := &cobra.Command{Use: "move NAME", Args: cobra.ExactArgs(1), RunE: func(c *cobra.Command, args []string) error {
		if !c.Flags().Changed("position") {
			return domain.E(2, "provide --position", nil)
		}
		w, e := load()
		if e != nil {
			return e
		}
		if e = w.ChangeMenu(c.Context(), args[0], nil, &position); e != nil {
			return domain.E(2, e.Error(), e)
		}
		return a.emit(map[string]any{"schemaVersion": 1, "name": args[0], "position": position})
	}}
	move.Flags().IntVar(&position, "position", 2, "1-based menu position; Auto always remains first")
	root.AddCommand(move)
	root.AddCommand(&cobra.Command{Use: "remove NAME", Args: cobra.ExactArgs(1), Short: "Remove an app-managed recipe unless another recipe depends on it", RunE: func(c *cobra.Command, args []string) error {
		w, e := load()
		if e != nil {
			return e
		}
		if e = w.Remove(c.Context(), args[0]); e != nil {
			return domain.E(2, e.Error(), e)
		}
		return a.emit(map[string]any{"schemaVersion": 1, "removed": args[0]})
	}})
	return root
}

func (f *compressFlags) configOverrides(c *cobra.Command) (config.Fields, error) {
	p := config.Fields{}
	changed := c.Flags().Changed
	if changed("copy") && f.CopyMode != "" {
		p.Clipboard = &f.CopyMode
	}
	if changed("quality") {
		p.Quality = &f.Quality
	}
	if changed("min-quality") {
		p.MinQuality = &f.MinQuality
	}
	if changed("colors") {
		p.Colors = &f.Colors
	}
	if changed("speed") {
		p.Speed = &f.Speed
	}
	if changed("dither") {
		p.Dither = &f.Dither
	}
	if changed("background") {
		p.Background = &f.Background
	}
	if changed("strip") {
		p.Metadata = &f.Metadata
	}
	if changed("target-size") {
		p.Target = &f.Target
	}
	if changed("resize") {
		p.Resize = &f.Resize
	}
	if changed("max-dimensions") {
		p.MaxDimensions = &f.MaxDimensions
	}
	if changed("min-dimensions") {
		p.MinDimensions = &f.MinDimensions
	}
	if changed("allow-resize") {
		p.AllowResize = &f.AllowResize
	}
	if changed("accept-nearest") {
		p.AcceptNearest = &f.Nearest
	}
	if changed("force-output") {
		p.ForceOutput = &f.Force
	}
	if changed("engine") {
		p.Engine = &f.Engine
	}
	if changed("remote-policy") {
		p.Remote = &f.Remote
	}
	if changed("jobs") {
		p.Jobs = &f.Jobs
	}
	if changed("recursive") {
		p.Recursive = &f.Input.Options.Recursive
	}
	if f.NoRecursive {
		p.Recursive = config.Ptr(false)
	}
	if changed("quantize") {
		p.Quantize = &f.Quantize
	}
	if f.NoQuantize {
		p.Quantize = config.Ptr(false)
	}
	if changed("optimize-alpha") {
		p.OptimizeAlpha = &f.Alpha
	}
	if f.NoAlpha {
		p.OptimizeAlpha = config.Ptr(false)
	}
	if changed("template") {
		p.Template = &f.Output.Template
	}
	if changed("output-dir") {
		p.Directory = &f.Output.Directory
	}
	if changed("timeout") {
		p.Timeout = config.Ptr(f.Timeout.String())
	}
	if changed("oxipng-level") {
		n := 6
		if f.Level != "max" {
			var e error
			n, e = strconv.Atoi(f.Level)
			if e != nil {
				return p, e
			}
		}
		p.OxiLevel = &n
	}
	return p, nil
}
