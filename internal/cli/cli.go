// SPDX-License-Identifier: MIT
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"github.com/Wolfsblvt/smolpng/internal/input"
	"github.com/Wolfsblvt/smolpng/internal/output"
	"github.com/Wolfsblvt/smolpng/internal/paths"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"github.com/Wolfsblvt/smolpng/internal/toolchain"
	"github.com/Wolfsblvt/smolpng/internal/workflow"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

var Version = "0.1.0-dev"
var Commit = "unknown"
var Distribution = "source"

type Environment struct {
	In        io.Reader
	Out, Err  io.Writer
	Runner    process.Runner
	Clipboard input.Clipboard
	Paths     *paths.Paths
	CWD       string
}

// Environment has only invocation and external-resource seams, not a second engine.
type app struct {
	env                                         Environment
	json, jsonl, quiet, verbose, portable, auto bool
	home, configPath                            string
	noProject                                   bool
	tools                                       toolchain.Paths
	configuredTools                             toolchain.Paths
}
type reported int

func (r reported) Error() string { return fmt.Sprintf("reported exit %d", r) }
func Run(ctx context.Context, args []string, env Environment) int {
	if env.In == nil {
		env.In = os.Stdin
	}
	if env.Out == nil {
		env.Out = os.Stdout
	}
	if env.Err == nil {
		env.Err = os.Stderr
	}
	if env.Runner == nil {
		env.Runner = process.OSRunner{}
	}
	a := &app{env: env}
	root := a.root()
	root.SetArgs(args)
	root.SetIn(env.In)
	root.SetOut(env.Out)
	root.SetErr(env.Err)
	er := root.ExecuteContext(ctx)
	if er == nil {
		return 0
	}
	var done reported
	if errors.As(er, &done) {
		return int(done)
	}
	code := domain.Code(er)
	var typed *domain.Error
	if !errors.As(er, &typed) && !errors.Is(er, context.Canceled) {
		code = 2
	}
	machine, l := a.json || a.jsonl, a.jsonl
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if arg == "--json" {
			machine = true
		}
		if arg == "--jsonl" {
			machine = true
			l = true
		}
	}
	if machine {
		r := domain.NewResult("<invocation>", "")
		r.Fail(domain.E(code, er.Error(), er))
		if l {
			_ = json.NewEncoder(env.Out).Encode(r)
		} else {
			_ = json.NewEncoder(env.Out).Encode(domain.Summarize([]domain.Result{r}))
		}
	} else {
		fmt.Fprintln(env.Err, "smolpng:", er)
	}
	return code
}
func (a *app) dirs() (paths.Paths, error) {
	if a.env.Paths != nil {
		return *a.env.Paths, nil
	}
	return paths.Resolve(a.portable, a.home)
}
func (a *app) cwd() (string, error) {
	if a.env.CWD != "" {
		return filepath.Abs(a.env.CWD)
	}
	return os.Getwd()
}
func (a *app) emit(v any) error {
	enc := json.NewEncoder(a.env.Out)
	if !a.jsonl {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(v)
}
func (a *app) resolver() input.Resolver {
	return input.Resolver{Stdin: a.env.In, Clipboard: a.env.Clipboard}
}
func (a *app) discover(ctx context.Context, p paths.Paths) toolchain.Set {
	return toolchain.Discover(ctx, a.env.Runner, a.tools, a.configuredTools)
}
func (a *app) root() *cobra.Command {
	c := &cobra.Command{Use: "smolpng [INPUT...]", Short: "Make images smol.", SilenceUsage: true, SilenceErrors: true}
	p := c.PersistentFlags()
	p.BoolVar(&a.json, "json", false, "Machine-readable JSON only")
	p.BoolVar(&a.jsonl, "jsonl", false, "One compact JSON result per image")
	p.BoolVarP(&a.quiet, "quiet", "q", false, "Suppress ordinary human reports")
	p.BoolVar(&a.verbose, "verbose", false, "Include exact optimizer argument arrays")
	p.BoolVar(&a.portable, "portable", false, "Use config/state beside the executable")
	p.StringVar(&a.home, "home", "", "Explicit portable data root")
	p.StringVar(&a.configPath, "config", "", "Explicit user config file (must exist)")
	p.BoolVar(&a.noProject, "no-project", false, "Do not load project .smolpng.toml")
	p.StringVar(&a.tools.PNGQuant, "pngquant-path", "", "Explicit pngquant executable")
	p.StringVar(&a.tools.OxiPNG, "oxipng-path", "", "Explicit OxiPNG executable")
	c.MarkFlagsMutuallyExclusive("json", "jsonl")
	rootFlags := &compressFlags{}
	rootFlags.bind(c)
	c.Flags().BoolVar(&a.auto, "auto", false, "One-shot deterministic compression")
	c.RunE = func(cmd *cobra.Command, args []string) error {
		if !a.auto {
			return domain.E(2, "interactive mode is not implemented in this cut; use compress or --auto", nil)
		}
		return a.compress(cmd, args, rootFlags)
	}
	cf := &compressFlags{}
	compress := &cobra.Command{Use: "compress INPUT...", Short: "Compress deterministically; never prompt"}
	cf.bind(compress)
	compress.RunE = func(c *cobra.Command, args []string) error { return a.compress(c, args, cf) }
	c.AddCommand(compress, a.inspectCommand(), a.configCommand(), a.presetCommand(), a.skillCommand())
	c.AddCommand(&cobra.Command{Use: "version", Args: cobra.NoArgs, Short: "Build identity", RunE: func(*cobra.Command, []string) error {
		return a.emit(map[string]any{"schemaVersion": 1, "version": Version, "commit": Commit, "distribution": Distribution, "platform": runtime.GOOS + "/" + runtime.GOARCH})
	}})
	c.AddCommand(&cobra.Command{Use: "capabilities", Args: cobra.NoArgs, Short: "Discover implemented features", RunE: func(*cobra.Command, []string) error {
		return a.emit(map[string]any{"schemaVersion": 1, "version": Version, "formats": []string{"png"}, "presets": []string{"auto", "web", "smallest", "lossless"}, "targetSize": true, "explicitResize": true, "remoteDefault": "never", "remoteImplemented": false, "config": true, "namedPresets": true, "presetImport": true, "agentSkill": true, "skillInstallationTracking": true, "interactive": false, "nativeClipboard": false, "exitCodes": map[string]int{"success": 0, "arguments": 2, "input": 3, "dependency": 4, "optimizer": 5, "target": 6, "network": 7, "batch": 8, "cancelled": 130}})
	}})
	c.AddCommand(&cobra.Command{Use: "doctor", Args: cobra.NoArgs, Short: "Inspect paths and optimizer status", RunE: a.doctor})
	deps := &cobra.Command{Use: "deps", Short: "Separately installed optimizer dependencies"}
	deps.AddCommand(&cobra.Command{Use: "status", Aliases: []string{"doctor"}, Args: cobra.NoArgs, RunE: a.doctor})
	deps.AddCommand(&cobra.Command{Use: "licenses", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		p, e := a.dirs()
		if e != nil {
			return e
		}
		if _, e = a.loadConfig(p); e != nil {
			return e
		}
		return a.emit(a.discover(c.Context(), p))
	}})
	c.AddCommand(deps)
	c.AddCommand(&cobra.Command{Use: "exec TOOL -- [ARG...]", Short: "Run a tool directly, relinquishing managed-output guarantees", DisableFlagParsing: true, RunE: func(c *cobra.Command, args []string) error {
		if len(args) == 0 || (args[0] != "pngquant" && args[0] != "oxipng") {
			return domain.E(2, "exec needs pngquant or oxipng", nil)
		}
		p, e := a.dirs()
		if e != nil {
			return e
		}
		if _, e = a.loadConfig(p); e != nil {
			return e
		}
		ts := a.discover(c.Context(), p)
		t := ts.PNGQuant
		if args[0] == "oxipng" {
			t = ts.OxiPNG
		}
		if !t.Available {
			return domain.E(4, t.Error, nil)
		}
		raw := args[1:]
		if len(raw) > 0 && raw[0] == "--" {
			raw = raw[1:]
		}
		r, e := a.env.Runner.Run(c.Context(), process.Invocation{Executable: t.Path, Args: raw, Stdin: a.env.In, Stdout: a.env.Out, Stderr: a.env.Err})
		if e != nil {
			return domain.E(domain.Code(e), "raw optimizer failed", e)
		}
		if r.ExitCode != 0 {
			return reported(r.ExitCode)
		}
		return nil
	}})
	c.AddCommand(&cobra.Command{Use: "completion SHELL", Short: "Generate shell completion", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return c.GenBashCompletionV2(a.env.Out, true)
		case "zsh":
			return c.GenZshCompletion(a.env.Out)
		case "fish":
			return c.GenFishCompletion(a.env.Out, true)
		case "powershell":
			return c.GenPowerShellCompletionWithDesc(a.env.Out)
		default:
			return domain.E(2, "choose bash, zsh, fish, or powershell", nil)
		}
	}})
	return c
}
func (a *app) compress(c *cobra.Command, args []string, f *compressFlags) error {
	p, e := a.dirs()
	if e != nil {
		return e
	}
	w, e := a.loadConfig(p)
	if e != nil {
		return e
	}
	name := ""
	if c.Flags().Changed("preset") {
		name = f.Preset
	}
	overrides, e := f.configOverrides(c)
	if e != nil {
		return domain.E(2, e.Error(), e)
	}
	effective, e := w.Resolve(name, overrides)
	if e != nil {
		return domain.E(2, e.Error(), e)
	}
	if e = effective.CheckUsable(); e != nil {
		return domain.E(2, e.Error(), e)
	}
	changed := c.Flags().Changed
	if !changed("jobs") {
		f.Jobs = effective.Jobs
	}
	if !changed("recursive") && !f.NoRecursive {
		f.Input.Options.Recursive = effective.Recursive
	}
	if !changed("template") {
		f.Output.Template = effective.Output.Template
	}
	if !changed("collision") {
		f.Output.Collision = effective.Output.Collision
	}
	if !changed("overwrite") {
		f.Output.Overwrite = effective.Output.Overwrite
	}
	if !changed("output-dir") && f.Output.Path == "" && !f.Stdout && !f.Output.InPlace {
		f.Output.Directory = effective.Output.Directory
	}
	f.Output.Detached = effective.Output.Detached
	r, e := f.apply(c, effective.Recipe)
	if e != nil {
		return domain.E(2, e.Error(), e)
	}
	if e = workflow.ValidateExtras(r); e != nil {
		return domain.E(2, e.Error(), e)
	}
	limits, e := f.Input.limits()
	if e != nil {
		return domain.E(2, e.Error(), e)
	}
	if f.Stdout {
		f.Output.Path = "-"
	}
	if f.Output.Path == "-" && (a.json || a.jsonl) {
		return domain.E(2, "binary stdout cannot share stdout with JSON; use --report-file", nil)
	}
	cwd, e := a.cwd()
	if e != nil {
		return e
	}
	absolute := func(p string) string {
		if p == "" || p == "-" || filepath.IsAbs(p) {
			return p
		}
		return filepath.Join(cwd, p)
	}
	f.Output.Path = absolute(f.Output.Path)
	f.Output.Directory = absolute(f.Output.Directory)
	f.ReportFile = absolute(f.ReportFile)
	// Detached output comes from effective configuration, never the working directory.
	f.Input.Options.CWD = cwd
	f.Input.Options.OutputDir = f.Output.Directory
	if f.NoRecursive {
		f.Input.Options.Recursive = false
	}
	sources, e := a.resolver().Resolve(c.Context(), args, f.Input.Options)
	if e != nil {
		return e
	}
	jobs, e := output.Plan(sources, f.Output, time.Now())
	if e != nil {
		return e
	}
	if f.ReportFile != "" {
		for _, j := range jobs {
			if j.Source.Path != "" && output.SameFile(f.ReportFile, j.Source.Path) || j.Output != "" && j.Output != "-" && output.SameFile(f.ReportFile, j.Output) {
				return domain.E(2, "report file cannot replace an input or output image", nil)
			}
		}
		if _, er := os.Stat(f.ReportFile); er == nil && !f.Output.Overwrite {
			return domain.E(2, "report file exists; choose a new name or --overwrite", nil)
		}
	}
	var tools toolchain.Set
	if !f.DryRun {
		tools = a.discover(c.Context(), p)
	}
	batch := workflow.Batch{Engine: workflow.Engine{Runner: a.env.Runner, Tools: tools, Limits: limits, CacheDir: p.Cache, Verbose: a.verbose}, Resolver: a.resolver(), Stdout: a.env.Out}
	result := batch.Run(c.Context(), jobs, workflow.Options{Recipe: r, Output: f.Output, Jobs: f.Jobs, FailFast: f.FailFast, DryRun: f.DryRun, Offline: f.Input.Options.Offline})
	if f.ReportFile != "" {
		b, e := json.MarshalIndent(result, "", "  ")
		if e != nil {
			return e
		}
		if e = fsutil.AtomicWrite(c.Context(), f.ReportFile, append(b, '\n'), 0600, f.Output.Overwrite); e != nil {
			return domain.E(5, "cannot write report file", e)
		}
	}
	if a.json {
		e = a.emit(result)
	} else if a.jsonl {
		enc := json.NewEncoder(a.env.Out)
		for _, r := range result.Results {
			if e = enc.Encode(r); e != nil {
				break
			}
		}
	} else if !a.quiet {
		w := a.env.Out
		if f.Output.Path == "-" {
			w = a.env.Err
		}
		e = human(w, result)
	}
	if e != nil {
		return domain.E(5, "cannot write result report", e)
	}
	if result.ExitCode != 0 {
		return reported(result.ExitCode)
	}
	return nil
}
func human(w io.Writer, report domain.Report) error {
	for _, r := range report.Results {
		if _, e := fmt.Fprintf(w, "%s: %s", r.Status, strconv.Quote(r.Input)); e != nil {
			return e
		}
		if r.Output != "" {
			if _, e := fmt.Fprintf(w, " -> %s", strconv.Quote(r.Output)); e != nil {
				return e
			}
		}
		if _, e := fmt.Fprintln(w); e != nil {
			return e
		}
		if r.OriginalBytes > 0 {
			fmt.Fprintf(w, "  %d B -> %d B | saved %.2f%%\n", r.OriginalBytes, r.OutputBytes, r.SavingPercent)
		}
		if r.Error != "" {
			fmt.Fprintf(w, "  %s (exit %d)\n", strconv.Quote(r.Error), r.ExitCode)
		}
		for _, s := range r.Warnings {
			fmt.Fprintf(w, "  warning: %s\n", strconv.Quote(s))
		}
	}
	return nil
}
func (a *app) doctor(c *cobra.Command, _ []string) error {
	p, e := a.dirs()
	if e != nil {
		return e
	}
	if _, e = a.loadConfig(p); e != nil {
		return e
	}
	ts := a.discover(c.Context(), p)
	var skillState any
	manager, skillErr := a.skillManager()
	if skillErr == nil {
		skillState, skillErr = manager.Inspect()
	}
	if skillErr != nil {
		skillState = map[string]any{"state": "unknown", "error": skillErr.Error()}
	}
	return a.emit(map[string]any{"schemaVersion": 1, "paths": p, "tools": ts, "skills": skillState, "networkCheckPerformed": false, "localAvailable": ts.PNGQuant.Available || ts.OxiPNG.Available})
}
func (a *app) inspectCommand() *cobra.Command {
	f := &inputFlags{}
	c := &cobra.Command{Use: "inspect INPUT...", Short: "Validate PNG structure and report dimensions, alpha, and metadata"}
	f.bind(c)
	c.RunE = func(c *cobra.Command, args []string) error {
		limits, e := f.limits()
		if e != nil {
			return domain.E(2, e.Error(), e)
		}
		f.Options.CWD, e = a.cwd()
		if e != nil {
			return e
		}
		sources, e := a.resolver().Resolve(c.Context(), args, f.Options)
		if e != nil {
			return e
		}
		type item struct {
			Input    string            `json:"input"`
			Info     *domain.ImageInfo `json:"info,omitempty"`
			Error    string            `json:"error,omitempty"`
			ExitCode int               `json:"exitCode"`
		}
		items := []item{}
		exit := 0
		for _, s := range sources {
			ctx, cancel := context.WithTimeout(c.Context(), 45*time.Second)
			b, er := a.resolver().Load(ctx, s, limits.MaxBytes, f.Options.Offline)
			it := item{Input: s.Display}
			if er == nil {
				var asset *pngimage.Asset
				asset, er = pngimage.Decode(b, limits.MaxPixels)
				if er == nil {
					it.Info = &asset.Info
				} else {
					er = domain.E(3, er.Error(), er)
				}
			}
			cancel()
			if er != nil {
				it.Error = er.Error()
				it.ExitCode = domain.Code(er)
				exit = it.ExitCode
			}
			items = append(items, it)
		}
		if exit != 0 && len(items) > 1 {
			exit = 8
		}
		if c.Context().Err() != nil {
			exit = 130
		}
		if a.jsonl {
			for _, it := range items {
				if e = json.NewEncoder(a.env.Out).Encode(it); e != nil {
					return e
				}
			}
		} else {
			e = a.emit(map[string]any{"schemaVersion": 1, "results": items, "exitCode": exit})
		}
		if e != nil {
			return e
		}
		if exit != 0 {
			return reported(exit)
		}
		return nil
	}
	return c
}
