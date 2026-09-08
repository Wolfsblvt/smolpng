// SPDX-License-Identifier: MIT
package cli

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/config"
	"github.com/Wolfsblvt/smolpng/internal/desktop"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/input"
	"github.com/Wolfsblvt/smolpng/internal/skill"
	"github.com/Wolfsblvt/smolpng/internal/state"
	"github.com/Wolfsblvt/smolpng/internal/toolchain"
	"github.com/Wolfsblvt/smolpng/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func (a *app) prompter() (ui.Prompter, error) {
	if a.json || a.jsonl {
		return nil, domain.E(2, "interactive workflows cannot write JSON; use compress, config, preset or skill commands", nil)
	}
	if a.env.Prompt != nil {
		return a.env.Prompt, nil
	}
	if !ui.IsTerminal(a.env.In, a.env.Out) && a.env.OpenTerminal != nil {
		c, e := a.env.OpenTerminal()
		if e != nil {
			return nil, domain.E(5, "cannot open selected interactive workflow", e)
		}
		a.env.In, a.env.Out, a.env.Err = c.In, c.Out, c.Err
		a.terminalRelease = c.Close
		a.env.OpenTerminal = nil
	}
	if !ui.IsTerminal(a.env.In, a.env.Out) {
		return nil, domain.E(2, "interactive mode needs a terminal; use smolpng compress for scripts and agents", nil)
	}
	return &ui.Terminal{In: a.env.In, Out: a.env.Out, Plain: a.plain, Accessible: a.accessible || os.Getenv("TERM") == "dumb"}, nil
}
func (a *app) notice(ctx context.Context, p ui.Prompter, title, description string) error {
	_, e := p.Choose(ctx, title, description, []ui.Option{{Label: "Back", Value: "back"}}, "back")
	return e
}
func (a *app) setupCommand() *cobra.Command {
	return &cobra.Command{Use: "setup", Short: "Inspect local tools and choose paths or optional Agent Skill installation", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		p, e := a.prompter()
		if e != nil {
			return e
		}
		e = a.onboard(c.Context(), p)
		if errors.Is(e, ui.ErrBack) {
			return nil
		}
		return e
	}}
}
func (a *app) settingsCommand() *cobra.Command {
	return &cobra.Command{Use: "settings", Short: "Edit ordinary user settings with keyboard forms", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		p, e := a.prompter()
		if e != nil {
			return e
		}
		e = a.editSettings(c.Context(), p)
		if errors.Is(e, ui.ErrBack) {
			return nil
		}
		return e
	}}
}
func toolDescription(t toolchain.Set) string {
	lines := []string{"Local optimizer setup (no startup network check)"}
	for _, v := range []toolchain.Tool{t.PNGQuant, t.OxiPNG, t.JPEGOptim} {
		if v.Available {
			lines = append(lines, fmt.Sprintf("%s %s: %s", v.Name, v.Version, strconv.Quote(v.Path)))
		} else {
			lines = append(lines, v.Name+": "+v.Error)
		}
	}
	return strings.Join(lines, "\n")
}
func (a *app) onboard(ctx context.Context, p ui.Prompter) error {
	for {
		dirs, e := a.dirs()
		if e != nil {
			return e
		}
		w, e := a.loadConfig(dirs)
		if e != nil {
			return e
		}
		tools := a.discover(ctx, dirs)
		desc := toolDescription(tools) + "\nExternal optimizer executables are not bundled.\npngquant remains a separate manual installation. OxiPNG can be installed from the pinned upstream archive on supported hosts."
		options := []ui.Option{{Label: "Continue with available local capabilities", Value: "continue"}}
		artifact, artifactErr := toolchain.ArtifactFor("oxipng", runtime.GOOS, runtime.GOARCH)
		if artifactErr == nil {
			options = append(options, ui.Option{Label: "Install/update a managed OxiPNG copy…", Value: "managed-oxipng"})
		}
		options = append(options, ui.Option{Label: "Configure optimizer executable paths", Value: "paths"}, ui.Option{Label: "Install the shared personal Agent Skill", Value: "skill"}, ui.Option{Label: "Install menu and drop shortcuts…", Value: "integrate"}, ui.Option{Label: "Open pngquant source/download page", Value: "pngquant"}, ui.Option{Label: "Open OxiPNG source/download page", Value: "oxipng"}, ui.Option{Label: "Cancel", Value: "quit", Key: "q"})
		initial := "continue"
		if !tools.OxiPNG.Available && artifactErr == nil {
			initial = "managed-oxipng"
		}
		action, e := p.Choose(ctx, "smolpng setup", desc, options, initial)
		if errors.Is(e, ui.ErrBack) {
			return e
		}
		if e != nil {
			return e
		}
		switch action {
		case "quit":
			return ui.ErrBack
		case "continue":
			return state.Write(ctx, filepath.Join(dirs.State, "setup.json"), state.Setup{SchemaVersion: 1, Acknowledged: true})
		case "managed-oxipng":
			confirm, er := p.Choose(ctx, "Install separate OxiPNG?", fmt.Sprintf("OxiPNG %s · %s\nSource: %s\nSHA-256: %s\nDestination: %s\nNo system PATH or existing configured/system tool is changed. The binary is verified and version-probed before activation.", artifact.Version, artifact.License, artifact.URL, artifact.SHA256, filepath.Join(dirs.Data, "tools", "oxipng")), []ui.Option{{Label: "Download, verify and install", Value: "install"}, {Label: "Back", Value: "back"}}, "back")
			if errors.Is(er, ui.ErrBack) || confirm == "back" {
				continue
			}
			if er != nil {
				return er
			}
			m, er := a.managedTools()
			if er != nil {
				return er
			}
			if _, er = m.Install(ctx, artifact); er != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				if er = a.notice(ctx, p, "Dependency installation refused", er.Error()); er != nil {
					return er
				}
			}
		case "integrate":
			if e = a.installIntegrationsUI(ctx, p); e != nil && !errors.Is(e, ui.ErrBack) {
				return e
			}
		case "paths":
			fields := []ui.Field{{Key: "tools.pngquant", Label: "pngquant executable (empty uses discovery)", Value: w.User.Document.Tools.PNGQuant}, {Key: "tools.oxipng", Label: "OxiPNG executable (empty uses discovery)", Value: w.User.Document.Tools.OxiPNG}, {Key: "tools.jpegoptim", Label: "jpegoptim executable (empty uses discovery)", Value: w.User.Document.Tools.JPEGOptim}}
			values, er := p.Edit(ctx, "Local executable paths; these run as your user", fields)
			if errors.Is(er, ui.ErrBack) {
				continue
			}
			if er != nil {
				return er
			}
			changes := map[string]any{}
			for _, f := range fields {
				if value := values[f.Key]; value != f.Value {
					changes[f.Key] = value
				}
			}
			if er = w.SaveUserSettings(ctx, changes); er != nil {
				if e = a.notice(ctx, p, "Settings were not saved", er.Error()); e != nil {
					return e
				}
			}
		case "skill":
			home, cwd, er := a.skillLocations()
			if er != nil {
				return er
			}
			target, er := skill.ResolveTarget("agents", "", home, cwd)
			if er != nil {
				return er
			}
			m, er := a.skillManager()
			if er != nil {
				return er
			}
			confirm, er := p.Choose(ctx, "Install Agent Skill?", target.Destination+"\nOnly smolpng's dedicated folder will be created. No existing content is overwritten.", []ui.Option{{Label: "Install", Value: "install"}, {Label: "Back", Value: "back"}}, "install")
			if errors.Is(er, ui.ErrBack) {
				continue
			}
			if er != nil {
				return er
			}
			if confirm == "install" {
				if er = m.Install(ctx, target.Destination, target.Name, target.Managed); er != nil {
					a.notice(ctx, p, "Skill installation refused", er.Error())
				}
			}
		case "pngquant":
			if e = a.env.Desktop.OpenURL(ctx, "https://pngquant.org/"); e != nil {
				a.notice(ctx, p, "Could not open the browser", e.Error())
			}
		case "oxipng":
			if e = a.env.Desktop.OpenURL(ctx, "https://github.com/oxipng/oxipng"); e != nil {
				a.notice(ctx, p, "Could not open the browser", e.Error())
			}
		}
	}
}
func (a *app) editSettings(ctx context.Context, p ui.Prompter) error {
	dirs, e := a.dirs()
	if e != nil {
		return e
	}
	w, e := a.loadConfig(dirs)
	if e != nil {
		return e
	}
	eff, e := w.Resolve("", config.Fields{})
	if e != nil {
		return e
	}
	presetOptions := []ui.Option{}
	for _, v := range w.List() {
		if v.Builtin || w.Project.Path == "" || v.Origin.File != w.Project.Path {
			presetOptions = append(presetOptions, ui.Option{Label: v.Label, Value: v.Name})
		}
	}
	defaultPreset := "auto"
	if w.User.Document.Defaults.Preset != nil {
		defaultPreset = *w.User.Document.Defaults.Preset
	}
	copyOptions := []ui.Option{}
	for _, v := range []string{"smart", "files", "path", "image", "none", "directory"} {
		copyOptions = append(copyOptions, ui.Option{Label: v, Value: v})
	}
	fields := []ui.Field{
		{Key: "defaults.preset", Label: "Default recipe", Value: defaultPreset, Options: presetOptions},
		{Key: "defaults.quality", Label: "Quality MIN-MAX (encoder-specific)", Value: fmt.Sprintf("%d-%d", eff.Recipe.QualityMin, eff.Recipe.QualityMax), Validate: quality},
		{Key: "output.template", Label: "Output filename template", Value: eff.Output.Template},
		{Key: "output.detached", Label: "Clipboard/URL output folder", Value: eff.Output.Detached},
		{Key: "clipboard.single", Label: "Copy a single result", Value: eff.Clipboard, Options: copyOptions},
		{Key: "drop.preset", Label: "Background/drop recipe", Value: eff.DropPreset, Options: presetOptions},
		{Key: "updates.check", Label: "Update checks (human startup only)", Value: eff.UpdateCheck, Options: []ui.Option{{Label: "Manual only", Value: "manual"}, {Label: "Once daily, when a trusted feed is configured", Value: "daily"}}},
		{Key: "ui.header", Label: "Small replaceable header", Value: eff.Header, Options: []ui.Option{{Label: "Cute placeholder", Value: "cute"}, {Label: "Compact", Value: "compact"}, {Label: "None", Value: "none"}}},
	}
	values, e := p.Edit(ctx, "User defaults (only changed fields are saved; project/preset overrides still apply)", fields)
	if e != nil {
		return e
	}
	changes := map[string]any{}
	for _, f := range fields {
		if value := values[f.Key]; value != f.Value {
			changes[f.Key] = value
		}
	}
	if len(changes) == 0 {
		return nil
	}
	if e = w.SaveUserSettings(ctx, changes); e != nil {
		return a.notice(ctx, p, "Settings were not saved", e.Error())
	}
	return nil
}
func quality(s string) error { _, _, e := domain.ParseQuality(s); return e }
func optionalSize(s string) error {
	if s == "" {
		return nil
	}
	_, e := domain.ParseSize(s)
	return e
}

type frozenClipboard struct{ snapshot input.Snapshot }

func (f frozenClipboard) Read(context.Context) (input.Snapshot, error) { return f.snapshot, nil }

// validClipboardSource does not mistake an unrelated secret copied as text for a
// filename to display. Only existing paths, explicit URLs, image/file data qualify.
func validClipboardSource(s input.Snapshot) bool {
	if len(s.Image) > 0 || len(s.Files) > 0 {
		return true
	}
	text := strings.Trim(strings.TrimSpace(s.Text), "\"")
	if strings.HasPrefix(text, "https://") || strings.HasPrefix(text, "http://") {
		return true
	}
	if text != "" {
		_, e := os.Stat(text)
		return e == nil
	}
	return false
}
func (a *app) importClipboardPreset(ctx context.Context, p ui.Prompter, s input.Snapshot) error {
	b, e := config.ParseBundle([]byte(s.Text))
	if e != nil {
		return a.notice(ctx, p, "Clipboard preset is invalid", e.Error())
	}
	dirs, e := a.dirs()
	if e != nil {
		return e
	}
	w, e := a.loadConfig(dirs)
	if e != nil {
		return e
	}
	plan, e := w.PlanImport(b, false)
	if e != nil {
		return a.notice(ctx, p, "Preset import unavailable", e.Error())
	}
	lines := []string{}
	for _, i := range plan.Items {
		line := i.SourceName + " -> " + i.Name
		f := i.After.Fields
		if f.Target != nil {
			line += " · target " + *f.Target
		}
		if f.AllowResize != nil && *f.AllowResize {
			line += " · resizing allowed"
		}
		if f.Remote != nil {
			line += " · remote " + *f.Remote
		}
		lines = append(lines, line)
		lines = append(lines, i.Warnings...)
	}
	action, e := p.Choose(ctx, "Preset found in clipboard", strings.Join(lines, "\n"), []ui.Option{{Label: "Import reviewed recipes", Value: "import"}, {Label: "Show full definition", Value: "review"}, {Label: "Ignore", Value: "ignore"}}, "import")
	if e != nil {
		return e
	}
	if action == "review" {
		data, _ := config.MarshalBundle(b)
		fmt.Fprintln(a.env.Out, ui.Safe(string(data)))
		return a.importClipboardPreset(ctx, p, s)
	}
	if action == "import" {
		if e = w.ApplyImport(ctx, &plan); e != nil {
			return a.notice(ctx, p, "Preset was not imported", e.Error())
		}
	}
	return nil
}
func (a *app) skillStartup(ctx context.Context, p ui.Prompter) error {
	m, e := a.skillManager()
	if e != nil {
		return e
	}
	prompt, e := m.Observe(ctx)
	if e != nil {
		return a.notice(ctx, p, "Agent Skill state needs attention", e.Error())
	}
	if !prompt.Prompt {
		return nil
	}
	action, e := p.Choose(ctx, "The bundled Agent Skill changed", "Only unmodified standard installations can be synchronized automatically. Custom and modified files are preserved.", []ui.Option{{Label: "Synchronize unmodified standard installations", Value: "sync"}, {Label: "Review destinations", Value: "review"}, {Label: "Ignore this Skill bundle", Value: "ignore"}, {Label: "Later", Value: "later"}}, "sync")
	if e != nil {
		return e
	}
	switch action {
	case "sync":
		results, e := m.SyncAfterUpdate(ctx)
		if e != nil {
			return e
		}
		for _, r := range results {
			fmt.Fprintf(a.env.Out, "%s: %s\n", ui.Safe(r.Action), ui.Safe(r.Destination))
		}
	case "review":
		s, e := m.Inspect()
		if e != nil {
			return e
		}
		for _, i := range s.Installations {
			fmt.Fprintf(a.env.Out, "%s: %s\n", ui.Safe(i.State), ui.Safe(i.Destination))
		}
		return a.skillStartup(ctx, p)
	case "ignore":
		return m.IgnoreCurrent(ctx)
	}
	return nil
}
func cloneInvocation(c *cobra.Command) (*cobra.Command, *compressFlags, error) {
	next := &cobra.Command{}
	next.SetContext(c.Context())
	f := &compressFlags{}
	f.bind(next)
	var failure error
	c.Flags().Visit(func(flag *pflag.Flag) {
		dst := next.Flags().Lookup(flag.Name)
		if dst == nil {
			return
		}
		if src, ok := flag.Value.(pflag.SliceValue); ok {
			if target, ok := dst.Value.(pflag.SliceValue); ok {
				failure = target.Replace(src.GetSlice())
				dst.Changed = true
				return
			}
		}
		if e := next.Flags().Set(flag.Name, flag.Value.String()); e != nil {
			failure = e
		}
	})
	return next, f, failure
}
func (a *app) interactive(c *cobra.Command, args []string, base *compressFlags) error {
	if base.Output.Path == "-" || base.Stdout {
		return domain.E(2, "binary stdout belongs to compress, not interactive mode", nil)
	}
	if len(args) > 0 && base.Input.Options.Clipboard {
		return domain.E(2, "choose explicit input or --clipboard, not both", nil)
	}
	for _, arg := range args {
		if arg == "-" {
			return domain.E(2, "stdin image streaming belongs to compress, not interactive mode", nil)
		}
	}
	p, e := a.prompter()
	if e != nil {
		return e
	}
	ctx := c.Context()
	dirs, e := a.dirs()
	if e != nil {
		return e
	}
	var setup state.Setup
	if e = state.Read(filepath.Join(dirs.State, "setup.json"), &setup); os.IsNotExist(e) || e == nil && setup.SchemaVersion == 1 && !setup.Acknowledged {
		if e = a.onboard(ctx, p); errors.Is(e, ui.ErrBack) {
			return nil
		}
		if e != nil {
			return e
		}
	} else if e != nil {
		return domain.E(3, "setup state is invalid; inspect the state directory", e)
	} else if setup.SchemaVersion != 1 {
		return domain.E(3, "unsupported setup state version; inspect the state directory", nil)
	}
	if e = a.skillStartup(ctx, p); e != nil && !errors.Is(e, ui.ErrBack) {
		return e
	}
	if scheduled, err := a.updateStartup(ctx, p, base.Input.Options.Offline); err != nil {
		return err
	} else if scheduled {
		return nil
	}
	originalClipboard := a.env.Clipboard
	defer func() { a.env.Clipboard = originalClipboard }()
	snapshot := input.Snapshot{}
	useClipboard := base.Input.Options.Clipboard
	if len(args) > 0 && useClipboard {
		return domain.E(2, "choose explicit input or --clipboard, not both", nil)
	}
	for _, arg := range args {
		if arg == "-" {
			return domain.E(2, "stdin image streaming belongs to compress, not interactive mode", nil)
		}
	}
	if len(args) == 0 || useClipboard {
		if originalClipboard != nil {
			snapshot, _ = originalClipboard.Read(ctx)
		}
		if validClipboardSource(snapshot) {
			useClipboard = true
			a.env.Clipboard = frozenClipboard{snapshot}
		}
	}
	hasPreset := strings.Contains(snapshot.Text, "[presets.")
	if len(args) == 0 && !useClipboard && hasPreset {
		if e = a.importClipboardPreset(ctx, p, snapshot); e != nil && !errors.Is(e, ui.ErrBack) {
			return e
		}
		hasPreset = false
	}
	for {
		if e = ctx.Err(); e != nil {
			return e
		}
		if len(args) == 0 && !useClipboard {
			values, er := p.Edit(ctx, "Choose an image source", []ui.Field{{Key: "source", Label: "Paste/drop a file, directory or HTTP(S) URL"}})
			if errors.Is(er, ui.ErrBack) {
				return nil
			}
			if er != nil {
				return er
			}
			path := strings.Trim(strings.TrimSpace(values["source"]), "\"")
			if path == "" {
				continue
			}
			args = []string{path}
		}
		w, er := a.loadConfig(dirs)
		if er != nil {
			return er
		}
		eff, er := w.Resolve("", config.Fields{})
		if er != nil {
			return er
		}
		source := "Clipboard image/file/path"
		if !useClipboard {
			labels := []string{}
			for _, v := range args[:min(3, len(args))] {
				labels = append(labels, strconv.Quote(sourceLabel(v)))
			}
			source = strings.Join(labels, "\n")
			if len(args) > 3 {
				source += fmt.Sprintf("\n... and %d more", len(args)-3)
			}
		}
		header := ""
		if !a.plain && !a.noHeader && !a.accessible && os.Getenv("TERM") != "dumb" {
			header = ui.Header(eff.Header, ui.Width(a.env.Out))
		}
		options := []ui.Option{}
		targetAdded := false
		initial := w.DefaultPreset()
		if c.Flags().Changed("preset") {
			initial = base.Preset
		}
		hasInitial := false
		for _, entry := range w.MenuEntries() {
			if entry.Builtin && entry.Name != "auto" && !targetAdded {
				options = append(options, ui.Option{Label: "Fit under a size…", Value: "@target"})
				targetAdded = true
			}
			options = append(options, ui.Option{Label: entry.Label, Value: entry.Name})
			hasInitial = hasInitial || entry.Name == initial
		}
		if !hasInitial {
			options = append(options, ui.Option{Label: "Configured default: " + initial, Value: initial})
		}
		options = append(options, ui.Option{Label: "TinyPNG (upload image)…", Value: "@tinypng"}, ui.Option{Label: "Saved recipes…", Value: "@saved"}, ui.Option{Label: "Custom…", Value: "@custom"}, ui.Option{Label: "Change source", Value: "@source", Key: "s"}, ui.Option{Label: "Settings", Value: "@settings", Key: ","}, ui.Option{Label: "Dependencies and setup", Value: "@setup", Key: "d"})
		if hasPreset {
			options = append(options, ui.Option{Label: "Import presets also found in clipboard", Value: "@import", Key: "i"})
		}
		options = append(options, ui.Option{Label: "Quit", Value: "@quit", Key: "q"})
		selected, er := p.Choose(ctx, header+"Choose a recipe", "Source: "+source+"\nLocal by default · originals preserved", options, initial)
		if errors.Is(er, ui.ErrBack) {
			return nil
		}
		if er != nil {
			return er
		}
		switch selected {
		case "@quit":
			return nil
		case "@source":
			args = nil
			useClipboard = false
			continue
		case "@settings":
			if er = a.editSettings(ctx, p); er != nil && !errors.Is(er, ui.ErrBack) {
				return er
			}
			continue
		case "@setup":
			if er = a.onboard(ctx, p); er != nil && !errors.Is(er, ui.ErrBack) {
				return er
			}
			continue
		case "@import":
			if er = a.importClipboardPreset(ctx, p, snapshot); er != nil && !errors.Is(er, ui.ErrBack) {
				return er
			}
			hasPreset = false
			continue
		case "@saved":
			saved := []ui.Option{}
			for _, entry := range w.List() {
				if !entry.Builtin {
					saved = append(saved, ui.Option{Label: entry.Label, Value: entry.Name})
				}
			}
			if len(saved) == 0 {
				a.notice(ctx, p, "No saved recipes yet", "Copy a preset or use smolpng preset import FILE --apply.")
				continue
			}
			selected, er = p.Choose(ctx, "Saved recipes", "", saved, saved[0].Value)
			if errors.Is(er, ui.ErrBack) {
				continue
			}
			if er != nil {
				return er
			}
		}
		cmd, f, er := cloneInvocation(c)
		if er != nil {
			return er
		}
		if useClipboard {
			cmd.Flags().Set("clipboard", "true")
		} else {
			cmd.Flags().Set("clipboard", "false")
		}
		if selected == "@tinypng" {
			choice, err := p.Choose(ctx, "Upload to TinyPNG?", "One original-derived image per input will leave this computer and may consume API quota. A saved key alone never authorizes uploads.", []ui.Option{{Label: "Upload this operation", Value: "upload"}, {Label: "Back", Value: "back"}}, "back")
			if err != nil {
				if errors.Is(err, ui.ErrBack) {
					continue
				}
				return err
			}
			if choice != "upload" {
				continue
			}
			cmd.Flags().Set("engine", "tinypng")
			cmd.Flags().Set("remote-policy", "always")
			selected = "auto"
		} else if selected == "@target" {
			value := base.Target
			if value == "" {
				value = "10KiB"
			}
			fields, er := p.Edit(ctx, "Strict maximum file size", []ui.Field{{Key: "target", Label: "Size (10KB = 10000 bytes; 10KiB = 10240)", Value: value, Validate: func(s string) error { _, e := domain.ParseSize(s); return e }}})
			if errors.Is(er, ui.ErrBack) {
				continue
			}
			if er != nil {
				return er
			}
			selected = "auto"
			cmd.Flags().Set("target-size", fields["target"])
		} else if selected == "@custom" {
			fields, er := p.Edit(ctx, "Custom local recipe", []ui.Field{
				{Key: "quality", Label: "Quality MIN-MAX", Value: fmt.Sprintf("%d-%d", eff.Recipe.QualityMin, eff.Recipe.QualityMax), Validate: quality},
				{Key: "colors", Label: "Palette colors (2..256)", Value: strconv.Itoa(eff.Recipe.Colors), Validate: func(s string) error {
					n, e := strconv.Atoi(s)
					if e != nil || n < 2 || n > 256 {
						return fmt.Errorf("use 2..256")
					}
					return nil
				}},
				{Key: "background", Label: "PNG background (#RRGGBB, white, black; empty keeps alpha)", Value: eff.Recipe.Background, Validate: func(s string) error {
					if s == "" {
						return nil
					}
					_, e := domain.ParseBackground(s)
					return e
				}},
				{Key: "dither", Label: "Dithering", Value: eff.Recipe.Dither, Options: []ui.Option{{Label: "Automatic", Value: "auto"}, {Label: "None", Value: "none"}}},
				{Key: "target-size", Label: "Optional target size (empty: smallest acceptable)", Value: base.Target, Validate: optionalSize},
				{Key: "allow-resize", Label: "Permit target resize fallback", Value: "false", Options: []ui.Option{{Label: "No dimension changes", Value: "false"}, {Label: "Allow fallback resizing", Value: "true"}}},
				{Key: "min-dimensions", Label: "Minimum fallback dimensions", Value: eff.Recipe.MinDimensions.String(), Validate: func(s string) error { _, e := domain.ParseDimensions(s); return e }},
				{Key: "strip", Label: "Metadata", Value: eff.Recipe.Metadata, Options: []ui.Option{{Label: "Safe", Value: "safe"}, {Label: "Keep", Value: "keep"}, {Label: "Strip all (may change color interpretation)", Value: "all"}}},
			})
			if errors.Is(er, ui.ErrBack) {
				continue
			}
			if er != nil {
				return er
			}
			selected = "auto"
			for k, v := range fields {
				if er = cmd.Flags().Set(k, v); er != nil {
					return er
				}
			}
		}
		if er = cmd.Flags().Set("preset", selected); er != nil {
			return er
		}
		if f.Output.Path == "-" || f.Stdout {
			return domain.E(2, "binary stdout belongs to compress, not interactive mode", nil)
		}
		fmt.Fprintln(a.env.Out, "Compressing with the selected recipe…")
		var result domain.Report
		var resolved config.Effective
		er = a.compression(cmd, args, f, func(report domain.Report, effective config.Effective) error {
			result = report
			resolved = effective
			return nil
		})
		if er != nil {
			if e = a.notice(ctx, p, "Image was not compressed", er.Error()); e != nil && !errors.Is(e, ui.ErrBack) {
				return e
			}
			continue
		}
		receipt := desktop.CopyReceipt{}
		warnings := []string{}
		if !f.DryRun {
			receipt, warnings = a.finishHuman(ctx, result, resolved, "")
		}
		_ = human(a.env.Out, result)
		for _, warning := range warnings {
			fmt.Fprintln(a.env.Out, "Warning: "+ui.Safe(warning))
		}
		if receipt.NeedsOwner {
			fmt.Fprintln(a.env.Out, "Clipboard is offered while this screen stays open. Paste before choosing Done; files remain saved afterwards.")
		}
		for {
			action, er := p.Choose(ctx, "Result", fmt.Sprintf("%d image result(s); exit %d", len(result.Results), result.ExitCode), []ui.Option{{Label: "Done", Value: "done"}, {Label: "Copy path(s)", Value: "path", Key: "p"}, {Label: "Copy image", Value: "image", Key: "c"}, {Label: "Copy files", Value: "files", Key: "f"}, {Label: "Open output folder", Value: "open", Key: "o"}, {Label: "Try another recipe", Value: "again", Key: "r"}, {Label: "New source", Value: "source", Key: "s"}}, "done")
			if errors.Is(er, ui.ErrBack) {
				return nil
			}
			if er != nil {
				return er
			}
			if action == "done" {
				return nil
			}
			if action == "again" {
				break
			}
			if action == "source" {
				args = nil
				useClipboard = false
				break
			}
			if action == "open" {
				files := state.Outputs(result)
				if len(files) > 0 {
					er = a.env.Desktop.OpenFolder(ctx, filepath.Dir(files[0]))
				} else {
					er = fmt.Errorf("no output exists to open")
				}
			} else {
				_, er = a.copyOutputs(ctx, result, action)
			}
			if er != nil {
				if e = a.notice(ctx, p, "Action unavailable", er.Error()); e != nil && !errors.Is(e, ui.ErrBack) {
					return e
				}
			}
		}
	}
}

func sourceLabel(s string) string {
	if u, e := url.Parse(s); e == nil && (u.Scheme == "https" || u.Scheme == "http") {
		u.RawQuery = ""
		u.Fragment = ""
		u.User = nil
		return u.String()
	}
	return s
}
