// SPDX-License-Identifier: MIT
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/output"
	"github.com/Wolfsblvt/smolpng/internal/paths"
)

type Layer struct {
	Path     string   `json:"path"`
	Exists   bool     `json:"exists"`
	Document Document `json:"document"`
	Bytes    []byte   `json:"-"`
}
type Workspace struct {
	User       Layer       `json:"user"`
	Project    Layer       `json:"project"`
	StorePath  string      `json:"storePath"`
	Store      Bundle      `json:"store"`
	StoreBytes []byte      `json:"-"`
	CWD        string      `json:"cwd"`
	Paths      paths.Paths `json:"paths"`
	presets    map[string]Preset
	origins    map[string]Origin
}
type LoadOptions struct {
	Paths           paths.Paths
	CWD, ConfigPath string
	NoProject       bool
}

func ReadLayer(path string, required bool) (Layer, error) {
	l := Layer{Path: path}
	if path == "" {
		return l, nil
	}
	f, e := os.Open(path)
	if os.IsNotExist(e) && !required {
		return l, nil
	}
	if e != nil {
		return l, fmt.Errorf("open config %s: %w", path, e)
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil || !st.Mode().IsRegular() {
		return l, fmt.Errorf("config %s is not a regular file", path)
	}
	b, e := ReadBounded(f)
	if e != nil {
		return l, e
	}
	l.Document, e = Parse(b)
	if e != nil {
		return l, fmt.Errorf("%s: %w", path, e)
	}
	l.Bytes = b
	l.Exists = true
	return l, nil
}
func FindProject(cwd string) (string, error) {
	for dir := cwd; ; dir = filepath.Dir(dir) {
		p := filepath.Join(dir, ".smolpng.toml")
		if st, e := os.Stat(p); e == nil {
			if !st.Mode().IsRegular() {
				return "", fmt.Errorf("project config is not a file: %s", p)
			}
			return p, nil
		} else if !os.IsNotExist(e) {
			return "", e
		}
		if _, e := os.Stat(filepath.Join(dir, ".git")); e == nil {
			return "", nil
		}
		if filepath.Dir(dir) == dir {
			return "", nil
		}
	}
}
func Load(o LoadOptions) (*Workspace, error) {
	cwd, e := filepath.Abs(o.CWD)
	if e != nil {
		return nil, e
	}
	userPath := o.Paths.Config
	if o.ConfigPath != "" {
		userPath = o.ConfigPath
		if !filepath.IsAbs(userPath) {
			userPath = filepath.Join(cwd, userPath)
		}
	}
	user, e := ReadLayer(userPath, o.ConfigPath != "")
	if e != nil {
		return nil, e
	}
	projectPath := ""
	if !o.NoProject {
		projectPath, e = FindProject(cwd)
		if e != nil {
			return nil, e
		}
	}
	project, e := ReadLayer(projectPath, false)
	if e != nil {
		return nil, e
	}
	if project.Document.Tools.PNGQuant != "" || project.Document.Tools.OxiPNG != "" || project.Document.Tools.Preference != "" {
		return nil, fmt.Errorf("project configuration cannot choose executables; use user [tools] or explicit CLI paths")
	}
	w := &Workspace{User: user, Project: project, StorePath: filepath.Join(o.Paths.Data, "presets.toml"), Store: Bundle{Presets: map[string]Preset{}}, CWD: cwd, Paths: o.Paths}
	f, e := os.Open(w.StorePath)
	if e == nil {
		b, er := ReadBounded(f)
		f.Close()
		if er != nil {
			return nil, er
		}
		w.StoreBytes = b
		if !IsEmptyStore(b) {
			w.Store, e = ParseBundle(b)
			if e != nil {
				return nil, fmt.Errorf("installed presets: %w", e)
			}
		}
	} else if !os.IsNotExist(e) {
		return nil, e
	}
	if e = w.rebuild(); e != nil {
		return nil, e
	}
	// Check every stored definition, not just the selected one. Blocked remote
	// requests remain inspectable/importable; unknown fields and broken graphs do not.
	for name := range w.presets {
		if _, e = w.Resolve(name, Fields{}); e != nil {
			return nil, e
		}
	}
	return w, nil
}
func (w *Workspace) rebuild() error {
	w.presets = map[string]Preset{}
	w.origins = map[string]Origin{}
	for _, l := range []struct {
		path    string
		presets map[string]Preset
	}{{w.User.Path, w.User.Document.Presets}, {w.StorePath, w.Store.Presets}, {w.Project.Path, w.Project.Document.Presets}} {
		for n, p := range l.presets {
			w.presets[n] = p
			w.origins[n] = Origin{File: l.path, Key: "presets." + n}
		}
	}
	return nil
}
func (w *Workspace) DefaultPreset() string {
	name := "auto"
	for _, l := range []Layer{w.User, w.Project} {
		if p := l.Document.Defaults.Preset; p != nil {
			name = *p
		}
	}
	return name
}
func (w *Workspace) chain(name string) (string, []string, error) {
	seen := map[string]bool{}
	chain := []string{}
	for !domain.IsBuiltin(name) {
		if seen[name] {
			return "", nil, fmt.Errorf("preset inheritance cycle at %q", name)
		}
		seen[name] = true
		p, ok := w.presets[name]
		if !ok {
			return "", nil, fmt.Errorf("unknown preset %q", name)
		}
		chain = append(chain, name)
		if len(chain) > 64 {
			return "", nil, fmt.Errorf("preset inheritance exceeds 64 levels")
		}
		name = p.Extends
		if name == "" {
			name = "auto"
		}
	}
	slices.Reverse(chain)
	return name, chain, nil
}
func (w *Workspace) Resolve(name string, overrides Fields) (Effective, error) {
	if name == "" {
		name = w.DefaultPreset()
	}
	base, chain, e := w.chain(name)
	if e != nil {
		return Effective{}, e
	}
	r, _ := domain.Builtin(base)
	eff := Effective{SchemaVersion: 1, Recipe: r, Output: output.Options{Detached: w.Paths.Detached(), Template: "{stem}.smol{ext}", Collision: "increment"}, Jobs: 2, Clipboard: "smart", BatchClipboard: "files", DropMode: "background", DropPreset: "auto", Notify: true, Header: "cute", Animations: true, UpdateCheck: "manual", UpdateChannel: "stable", RemotePermission: "never", Formats: []string{"png"}, Provenance: map[string]Origin{}, Warnings: []string{}}
	for _, key := range []string{"quality_min", "quality_max", "colors", "speed", "dither", "oxipng_level", "metadata", "quantize", "optimize_alpha", "target_size", "allow_resize", "resize", "max_dimensions", "min_dimensions", "accept_nearest", "force_output", "engine", "remote", "timeout", "output.template", "output.directory", "output.detached", "output.collision", "output.overwrite", "recursive", "jobs", "clipboard", "clipboard.batch", "drop.mode", "drop.preset", "drop.notify", "ui.header", "ui.animations", "updates.check", "updates.channel", "remote_permission"} {
		eff.Provenance[key] = Origin{File: "builtin:" + base, Key: key}
	}
	for _, l := range []Layer{w.User, w.Project} {
		if !l.Exists {
			continue
		}
		if e = eff.apply(l.Document.Defaults.Fields, Origin{l.Path, "defaults"}, filepath.Dir(l.Path)); e != nil {
			return eff, e
		}
		if base == "auto" {
			if e = eff.apply(l.Document.PNG.Auto, Origin{l.Path, "png.auto"}, filepath.Dir(l.Path)); e != nil {
				return eff, e
			}
		}
		if e = eff.applyDocument(l, w.Paths.Downloads); e != nil {
			return eff, e
		}
	}
	// The user grant is the cap. A project can tighten or select a permitted mode,
	// but cannot silently turn a local-only personal installation into an uploader.
	permission := "never"
	if p := w.User.Document.Defaults.RemotePolicy; p != nil {
		permission = *p
	}
	if p := w.Project.Document.Defaults.RemotePolicy; p != nil {
		if !allows(permission, *p) {
			return eff, fmt.Errorf("project remote_policy %q exceeds the user permission %q", *p, permission)
		}
		permission = *p
	}
	eff.RemotePermission = permission
	eff.Recipe.RemotePolicy = permission
	for _, n := range chain {
		p := w.presets[n]
		if e = eff.apply(p.Fields, w.origins[n], w.CWD); e != nil {
			return eff, fmt.Errorf("preset %s: %w", n, e)
		}
		if len(p.Formats) > 0 {
			eff.Formats = slices.Clone(p.Formats)
		}
	}
	if e = validateFields(overrides, false); e != nil {
		return eff, e
	}
	if e = eff.apply(overrides, Origin{File: "CLI", Key: "flags"}, w.CWD); e != nil {
		return eff, e
	}
	if overrides.Remote != nil {
		eff.RemotePermission = *overrides.Remote
		eff.Provenance["remote_permission"] = Origin{File: "CLI", Key: "remote-policy"}
	}
	eff.Recipe.Name = name
	if base == "lossless" && !eff.Recipe.Lossless {
		return eff, fmt.Errorf("a lossless-derived preset cannot re-enable quantization; extend auto instead")
	}
	// Validate the recipe without turning a blocked provider request into a parse
	// failure. Invocation calls CheckUsable before any discovery/network/process.
	check := eff.Recipe
	blocked := !allows(eff.RemotePermission, check.RemotePolicy) || check.Engine == "tinypng" && (check.RemotePolicy == "never" || check.RemotePolicy == "ask")
	if blocked {
		eff.Warnings = append(eff.Warnings, "remote request is blocked by user policy; an explicit --remote-policy grant is required")
		if check.Engine == "tinypng" {
			check.RemotePolicy = "always"
		}
	}
	if e = check.Validate(); e != nil {
		return eff, fmt.Errorf("preset %s: %w", name, e)
	}
	if eff.Jobs < 1 || eff.Jobs > 32 {
		return eff, fmt.Errorf("jobs must be 1..32")
	}
	return eff, nil
}
func allows(permission, request string) bool {
	if request == "never" || request == "ask" {
		return true
	}
	switch permission {
	case "always", "compare":
		return true
	case "fallback":
		return request == "fallback"
	}
	return false
}
func (e Effective) CheckUsable() error {
	if !slices.Contains(e.Formats, "png") {
		return fmt.Errorf("selected preset does not admit PNG; JPEG compression is not implemented")
	}
	if !allows(e.RemotePermission, e.Recipe.RemotePolicy) || e.Recipe.Engine == "tinypng" && (e.Recipe.RemotePolicy == "never" || e.Recipe.RemotePolicy == "ask") {
		return fmt.Errorf("preset requests remote processing outside current permission %q; change user settings or explicitly select --remote-policy", e.RemotePermission)
	}
	return e.Recipe.Validate()
}
func (e *Effective) remember(key string, origin Origin, field string) {
	origin.Key += "." + field
	e.Provenance[key] = origin
}
func (e *Effective) apply(f Fields, o Origin, base string) error {
	mark := func(k, field string) { e.remember(k, o, field) }
	r := &e.Recipe
	if f.Quality != nil {
		lo, hi, err := domain.ParseQuality(*f.Quality)
		if err != nil {
			return err
		}
		r.QualityMin = lo
		r.QualityMax = hi
		mark("quality_min", "quality")
		mark("quality_max", "quality")
	}
	if f.MinQuality != nil {
		r.QualityMin = *f.MinQuality
		mark("quality_min", "min_quality")
	}
	if f.Colors != nil {
		r.Colors = *f.Colors
		mark("colors", "colors")
	}
	if f.Speed != nil {
		r.Speed = *f.Speed
		mark("speed", "speed")
	}
	if f.Dither != nil {
		r.Dither = *f.Dither
		mark("dither", "dither")
	}
	if f.OxiLevel != nil {
		r.OxiLevel = *f.OxiLevel
		mark("oxipng_level", "oxipng_level")
	}
	if f.Metadata != nil {
		r.Metadata = *f.Metadata
		mark("metadata", "metadata")
	}
	if f.Quantize != nil {
		r.Lossless = !*f.Quantize
		mark("quantize", "quantize")
	}
	if f.OptimizeAlpha != nil {
		r.OptimizeAlpha = *f.OptimizeAlpha
		mark("optimize_alpha", "optimize_alpha")
	}
	if f.Target != nil {
		r.TargetBytes = 0
		if *f.Target != "" {
			n, err := domain.ParseSize(*f.Target)
			if err != nil {
				return err
			}
			r.TargetBytes = n
		}
		mark("target_size", "target_size")
	}
	for _, d := range []struct {
		key  string
		s    *string
		dest *domain.Dimensions
	}{{"resize", f.Resize, &r.Resize}, {"max_dimensions", f.MaxDimensions, &r.MaxDimensions}, {"min_dimensions", f.MinDimensions, &r.MinDimensions}} {
		if d.s != nil {
			*d.dest = domain.Dimensions{}
			if *d.s != "" {
				v, err := domain.ParseDimensions(*d.s)
				if err != nil {
					return err
				}
				*d.dest = v
			}
			mark(d.key, d.key)
		}
	}
	if f.AllowResize != nil {
		r.AllowResize = *f.AllowResize
		mark("allow_resize", "allow_resize")
	}
	if f.AcceptNearest != nil {
		r.AcceptNearest = *f.AcceptNearest
		mark("accept_nearest", "accept_nearest")
	}
	if f.ForceOutput != nil {
		r.ForceOutput = *f.ForceOutput
		mark("force_output", "force_output")
	}
	if f.Engine != nil {
		r.Engine = *f.Engine
		mark("engine", "engine")
	}
	if f.Remote != nil {
		r.RemotePolicy = *f.Remote
		mark("remote", "remote")
	}
	if f.Timeout != nil {
		v, err := time.ParseDuration(*f.Timeout)
		if err != nil || v < 0 {
			return fmt.Errorf("timeout must be a nonnegative duration")
		}
		r.Timeout = v
		mark("timeout", "timeout")
	}
	if f.Template != nil {
		e.Output.Template = *f.Template
		mark("output.template", "template")
	}
	if f.Directory != nil {
		e.Output.Directory = relativeTo(*f.Directory, base)
		mark("output.directory", "output_directory")
	}
	if f.Jobs != nil {
		e.Jobs = *f.Jobs
		mark("jobs", "jobs")
	}
	if f.Recursive != nil {
		e.Recursive = *f.Recursive
		mark("recursive", "recursive")
	}
	if f.Clipboard != nil {
		e.Clipboard = *f.Clipboard
		mark("clipboard", "clipboard")
	}
	return nil
}
func relativeTo(p, base string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(base, filepath.FromSlash(p))
}
func (e *Effective) applyDocument(l Layer, downloads string) error {
	d := l.Document
	origin := Origin{File: l.Path}
	mark := func(k string) { o := origin; o.Key = k; e.Provenance[k] = o }
	if d.Defaults.Overwrite != nil {
		e.Output.Overwrite = *d.Defaults.Overwrite
		mark("output.overwrite")
	}
	if d.Defaults.RemotePolicy != nil {
		mark("remote_permission")
		e.Provenance["remote_permission"] = Origin{l.Path, "defaults.remote_policy"}
	}
	if d.Output.Directory != nil {
		e.Output.Directory = relativeTo(*d.Output.Directory, filepath.Dir(l.Path))
		mark("output.directory")
	}
	if d.Output.Template != nil {
		e.Output.Template = *d.Output.Template
		mark("output.template")
	}
	if d.Output.Detached != nil {
		v := *d.Output.Detached
		if strings.Contains(v, "{downloads}") && downloads == "" {
			return fmt.Errorf("Downloads could not be resolved")
		}
		v = strings.ReplaceAll(v, "{downloads}", downloads)
		if strings.ContainsAny(v, "{}\x00\r\n") {
			return fmt.Errorf("unknown detached output placeholder")
		}
		e.Output.Detached = relativeTo(v, filepath.Dir(l.Path))
		mark("output.detached")
	}
	if d.Output.Collision != nil {
		e.Output.Collision = *d.Output.Collision
		mark("output.collision")
	}
	if d.Clipboard.Single != nil {
		e.Clipboard = *d.Clipboard.Single
		mark("clipboard")
	}
	if d.Clipboard.Batch != nil {
		e.BatchClipboard = *d.Clipboard.Batch
		mark("clipboard.batch")
	}
	if d.Drop.Mode != nil {
		e.DropMode = *d.Drop.Mode
		mark("drop.mode")
	}
	if d.Drop.Preset != nil {
		e.DropPreset = *d.Drop.Preset
		mark("drop.preset")
	}
	if d.Drop.Notify != nil {
		e.Notify = *d.Drop.Notify
		mark("drop.notify")
	}
	if d.UI.Header != nil {
		e.Header = *d.UI.Header
		mark("ui.header")
	}
	if d.UI.Animations != nil {
		e.Animations = *d.UI.Animations
		mark("ui.animations")
	}
	if d.Updates.Check != nil {
		e.UpdateCheck = *d.Updates.Check
		mark("updates.check")
	}
	if d.Updates.Channel != nil {
		e.UpdateChannel = *d.Updates.Channel
		mark("updates.channel")
	}
	if d.Tools.Preference != "" {
		e.Tools.Preference = d.Tools.Preference
		mark("tools.preference")
	}
	if d.Tools.PNGQuant != "" {
		e.Tools.PNGQuant = relativeTo(d.Tools.PNGQuant, filepath.Dir(l.Path))
		mark("tools.pngquant")
	}
	if d.Tools.OxiPNG != "" {
		e.Tools.OxiPNG = relativeTo(d.Tools.OxiPNG, filepath.Dir(l.Path))
		mark("tools.oxipng")
	}
	return nil
}
