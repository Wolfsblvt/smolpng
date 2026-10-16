// SPDX-License-Identifier: MIT
package cli

import (
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/input"
	"github.com/Wolfsblvt/smolpng/internal/output"
	"github.com/spf13/cobra"
	"strconv"
	"time"
)

type inputFlags struct {
	Options input.Options
	Size    string
	Pixels  int64
}

func (f *inputFlags) bind(c *cobra.Command) {
	p := c.Flags()
	p.BoolVar(&f.Options.Clipboard, "clipboard", false, "Read the clipboard explicitly")
	p.BoolVarP(&f.Options.Recursive, "recursive", "r", false, "Traverse directory descendants")
	p.BoolVar(&f.Options.Hidden, "hidden", false, "Include hidden discovered inputs")
	p.BoolVar(&f.Options.NoIgnore, "no-ignore", false, "Ignore .gitignore files")
	p.BoolVar(&f.Options.FollowSymlinks, "follow-symlinks", false, "Follow directory symlinks with cycle detection")
	p.BoolVar(&f.Options.Offline, "offline", false, "Reject URL input and remote optimization")
	p.StringArrayVar(&f.Options.Globs, "glob", nil, "Explicit quoted glob; repeatable")
	p.StringArrayVar(&f.Options.Include, "include", nil, "Include pattern; repeatable")
	p.StringArrayVar(&f.Options.Exclude, "exclude", nil, "Exclude pattern; repeatable")
	p.StringVar(&f.Options.StdinName, "stdin-name", "stdin.png", "Logical filename for stdin")
	p.StringVar(&f.Size, "max-input-size", "64MiB", "Maximum encoded image/download/candidate bytes")
	p.Int64Var(&f.Pixels, "max-pixels", 64_000_000, "Maximum decoded pixels")
}
func (f inputFlags) limits() (domain.Limits, error) {
	n, e := domain.ParseSize(f.Size)
	if e != nil {
		return domain.Limits{}, e
	}
	l := domain.Limits{MaxBytes: n, MaxPixels: f.Pixels}
	return l, l.Validate()
}

type compressFlags struct {
	Background                                                                                               string
	JPEGQuality                                                                                              string
	CopyMode                                                                                                 string
	ClipboardHold                                                                                            time.Duration
	Input                                                                                                    inputFlags
	Output                                                                                                   output.Options
	Preset, Quality, Dither, Level, Metadata, Target, Resize, MaxDimensions, MinDimensions, Engine, Remote   string
	Colors, Speed, MinQuality, Jobs                                                                          int
	Timeout                                                                                                  time.Duration
	Quantize, NoQuantize, Alpha, NoAlpha, AllowResize, Nearest, Force, FailFast, DryRun, Stdout, NoRecursive bool
	PNGArgs, OxiArgs                                                                                         []string
	ReportFile                                                                                               string
}

func (f *compressFlags) bind(c *cobra.Command) {
	f.Input.bind(c)
	p := c.Flags()
	p.StringVar(&f.Preset, "preset", "auto", "Built-in or saved recipe")
	p.StringVar(&f.Quality, "quality", "", "Shared MIN-MAX quality override; encoder-specific")
	p.StringVar(&f.JPEGQuality, "jpeg-quality", "", "JPEG-only MIN-MAX quality override")
	p.IntVar(&f.MinQuality, "min-quality", 70, "Minimum acceptable quantization quality")
	p.IntVar(&f.Colors, "colors", 256, "Palette limit, 2..256")
	p.IntVar(&f.Speed, "speed", 1, "pngquant speed, 1..11")
	p.StringVar(&f.Dither, "dither", "auto", "Dither: auto, none, or 0..1")
	p.StringVar(&f.Level, "oxipng-level", "4", "OxiPNG level, 0..6 or max")
	p.StringVar(&f.Background, "background", "", "Explicit PNG matte: #RRGGBB, white, black; removes transparency, not the subject background")
	p.StringVar(&f.Metadata, "strip", "safe", "Metadata: keep, safe, or all")
	p.BoolVar(&f.Quantize, "quantize", false, "Enable lossy palette candidates")
	p.BoolVar(&f.NoQuantize, "no-quantize", false, "Use only pixel-identical optimizer candidates")
	p.BoolVar(&f.Alpha, "optimize-alpha", false, "Allow hidden RGB cleanup on a quantized candidate")
	p.BoolVar(&f.NoAlpha, "no-optimize-alpha", false, "Preserve hidden RGB in the lossless tail")
	p.StringVar(&f.Target, "target-size", "", "Strict target, e.g. 10KiB")
	p.BoolVar(&f.AllowResize, "allow-resize", false, "Allow original-derived target resize fallback")
	p.StringVar(&f.Resize, "resize", "", "Explicit WIDTHxHEIGHT transformation")
	p.StringVar(&f.MaxDimensions, "max-dimensions", "", "Fit within WIDTHxHEIGHT without upscaling")
	p.StringVar(&f.MinDimensions, "min-dimensions", "64x64", "Minimum target-search dimensions")
	p.BoolVar(&f.Nearest, "accept-nearest", false, "Write nearest candidate on target miss; exit remains 6")
	p.BoolVar(&f.Force, "force-output", false, "Permit an output larger than its original")
	p.StringVar(&f.Engine, "engine", "local", "local, auto, or tinypng")
	p.StringVar(&f.Remote, "remote-policy", "never", "never, ask, fallback, compare, or always")
	p.StringVarP(&f.Output.Path, "output", "o", "", "Exact output for one input; - is binary stdout")
	p.StringVar(&f.Output.Directory, "output-dir", "", "Mirror input structure into this directory")
	p.StringVar(&f.Output.Template, "template", "{stem}.smol{ext}", "Filename template: {stem}, {ext}, {timestamp}")
	p.StringVar(&f.Output.Collision, "collision", "increment", "Existing output: increment or error")
	p.BoolVar(&f.Output.InPlace, "in-place", false, "Explicitly replace local originals atomically")
	p.BoolVar(&f.Output.Overwrite, "overwrite", false, "Replace existing outputs, never implicit originals")
	p.BoolVar(&f.Stdout, "stdout", false, "Stream a single PNG to stdout")
	p.IntVarP(&f.Jobs, "jobs", "j", 2, "Parallel image jobs, 1..32")
	p.DurationVar(&f.Timeout, "timeout", 2*time.Minute, "Per-image timeout including input read; 0 disables")
	p.BoolVar(&f.FailFast, "fail-fast", false, "Stop starting work after first failure; serial mode")
	p.BoolVar(&f.DryRun, "dry-run", false, "Plan only: no image read, optimizer, network, or image write")
	p.BoolVar(&f.NoRecursive, "no-recursive", false, "Override recursive traversal")
	p.StringArrayVar(&f.PNGArgs, "pngquant-arg", nil, "Safe extra --posterize=N; use exec for arbitrary arguments")
	p.StringArrayVar(&f.OxiArgs, "oxipng-arg", nil, "Safe extra --interlace=N or --zopfli")
	p.StringVar(&f.CopyMode, "copy", "", "Explicitly copy a file output: smart, path, image, files, directory or none")
	p.DurationVar(&f.ClipboardHold, "clipboard-hold", 0, "On Linux serve copied data until replaced or timeout (0..30m)")
	p.StringVar(&f.ReportFile, "report-file", "", "Also write a structured report to this file")
	c.MarkFlagsMutuallyExclusive("quantize", "no-quantize")
	c.MarkFlagsMutuallyExclusive("optimize-alpha", "no-optimize-alpha")
	c.MarkFlagsMutuallyExclusive("recursive", "no-recursive")
	c.MarkFlagsMutuallyExclusive("stdout", "output")
}
func (f *compressFlags) apply(c *cobra.Command, r domain.Recipe) (domain.Recipe, error) {
	changed := c.Flags().Changed
	var e error
	if changed("quality") {
		var lo, hi int
		lo, hi, e = domain.ParseQuality(f.Quality)
		if e != nil {
			return r, e
		}
		r.SetQuality(lo, hi)
	}
	if changed("min-quality") {
		r.SetMinQuality(f.MinQuality)
	}
	if changed("jpeg-quality") {
		lo, hi, err := domain.ParseQuality(f.JPEGQuality)
		if err != nil {
			return r, err
		}
		r.JPEGQuality = &domain.QualityRange{Min: lo, Max: hi}
	}
	if changed("colors") {
		r.Colors = f.Colors
	}
	if changed("speed") {
		r.Speed = f.Speed
	}
	if changed("dither") {
		r.Dither = f.Dither
	}
	if changed("oxipng-level") {
		if f.Level == "max" {
			r.OxiLevel = 6
		} else {
			r.OxiLevel, e = strconv.Atoi(f.Level)
			if e != nil {
				return r, e
			}
		}
	}
	if changed("background") {
		r.Background = f.Background
	}
	if changed("strip") {
		r.Metadata = f.Metadata
	}
	if changed("quantize") && f.Quantize {
		if r.Name == "lossless" {
			return r, fmt.Errorf("lossless preset cannot quantize")
		}
		r.Lossless = false
	}
	if f.NoQuantize {
		r.Lossless = true
	}
	if changed("optimize-alpha") {
		r.OptimizeAlpha = f.Alpha
	}
	if f.NoAlpha {
		r.OptimizeAlpha = false
	}
	if changed("target-size") {
		r.TargetBytes, e = domain.ParseSize(f.Target)
		if e != nil {
			return r, e
		}
	}
	for _, pair := range []struct {
		name, value string
		dest        *domain.Dimensions
	}{{"resize", f.Resize, &r.Resize}, {"max-dimensions", f.MaxDimensions, &r.MaxDimensions}, {"min-dimensions", f.MinDimensions, &r.MinDimensions}} {
		if changed(pair.name) {
			*pair.dest, e = domain.ParseDimensions(pair.value)
			if e != nil {
				return r, e
			}
		}
	}
	if changed("allow-resize") {
		r.AllowResize = f.AllowResize
	}
	if changed("accept-nearest") {
		r.AcceptNearest = f.Nearest
	}
	if changed("force-output") {
		r.ForceOutput = f.Force
	}
	if changed("engine") {
		r.Engine = f.Engine
	}
	if changed("remote-policy") {
		r.RemotePolicy = f.Remote
	}
	if changed("timeout") {
		r.Timeout = f.Timeout
	}
	r.PNGQuantArgs = f.PNGArgs
	r.OxiPNGArgs = f.OxiArgs
	if f.Input.Options.Offline && (r.Engine == "tinypng" || r.RemotePolicy != "never" && r.RemotePolicy != "ask") {
		return r, fmt.Errorf("offline mode prohibits the selected remote policy")
	}
	if f.Jobs < 1 || f.Jobs > 32 {
		return r, fmt.Errorf("jobs must be 1..32")
	}
	return r, r.Validate()
}
