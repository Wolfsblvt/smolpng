// SPDX-License-Identifier: MIT
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Wolfsblvt/smolpng/internal/comparison"
	"github.com/Wolfsblvt/smolpng/internal/config"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"github.com/Wolfsblvt/smolpng/internal/toolchain"
	"github.com/Wolfsblvt/smolpng/internal/workflow"
	"github.com/spf13/cobra"
)

func (a *app) compareCommand() *cobra.Command {
	var f inputFlags
	var presets, modes []string
	var repeat int
	var output, target, format string
	var dry, nearest bool
	c := &cobra.Command{Use: "compare INPUT...", Short: "Compare local recipes/tool pipelines from identical original bytes", Args: cobra.ArbitraryArgs, RunE: func(c *cobra.Command, args []string) error {
		if repeat < 1 || repeat > 5 {
			return domain.E(2, "repeat must be 1..5", nil)
		}
		if format != "human" && format != "csv" {
			return domain.E(2, "format must be human or csv; use --json or --jsonl for machine output", nil)
		}
		if format == "csv" && (a.json || a.jsonl) {
			return domain.E(2, "choose CSV or JSON stdout, not both", nil)
		}
		if len(presets) == 0 {
			presets = []string{"auto", "web", "smallest", "lossless"}
		}
		if len(modes) == 0 {
			modes = []string{"auto"}
		}
		limits, e := f.limits()
		if e != nil {
			return domain.E(2, e.Error(), e)
		}
		p, e := a.dirs()
		if e != nil {
			return e
		}
		w, e := a.loadConfig(p)
		if e != nil {
			return e
		}
		variants := []comparison.Variant{}
		override := config.Fields{}
		if target != "" {
			override.Target = config.Ptr(target)
		}
		if nearest {
			override.AcceptNearest = config.Ptr(true)
		}
		for _, name := range presets {
			eff, err := w.Resolve(name, override)
			if err != nil {
				return domain.E(2, err.Error(), err)
			}
			for _, mode := range modes {
				variants = append(variants, comparison.Variant{ID: name + "-" + mode, Recipe: eff.Recipe, PNGMode: mode, Formats: eff.Formats})
			}
		}
		if e = comparison.ValidateVariants(variants); e != nil {
			return domain.E(2, e.Error(), e)
		}
		outRoot := output
		if output != "" {
			if !filepath.IsAbs(output) {
				outRoot = filepath.Join(w.CWD, output)
			}
			outRoot, e = filepath.Abs(outRoot)
			if e != nil {
				return e
			}
			if _, e = os.Lstat(outRoot); !os.IsNotExist(e) {
				return domain.E(2, "comparison output must be a new directory; existing results are never replaced", e)
			}
		}
		f.Options.CWD = w.CWD
		f.Options.OutputDir = outRoot
		sources, e := a.resolver().Resolve(c.Context(), args, f.Options)
		if e != nil {
			return e
		}
		if len(sources) > 512 {
			return domain.E(2, "comparison is bounded to 512 input images; split the corpus", nil)
		}
		var tools toolchain.Set
		if !dry {
			tools = a.discover(c.Context(), p)
		}
		report := comparison.NewReport(Version, Commit, tools, variants)
		engine := workflow.Engine{Runner: a.env.Runner, Tools: tools, Limits: limits, CacheDir: p.Cache, Verbose: a.verbose}
		// Reserve the user-selected new artifact root only after all planning/argument checks.
		if outRoot != "" && !dry {
			if e = os.MkdirAll(filepath.Dir(outRoot), 0755); e != nil {
				return e
			}
			if e = os.Mkdir(outRoot, 0700); e != nil {
				return e
			}
		}
		for i, source := range sources {
			var data []byte
			var loadErr error
			if !dry {
				data, loadErr = a.resolver().Load(c.Context(), source, limits.MaxBytes, f.Options.Offline)
			}
			originalPath := ""
			if outRoot != "" && !dry && loadErr == nil {
				ext := ".png"
				if len(data) > 1 && data[0] == 0xff && data[1] == 0xd8 {
					ext = ".jpg"
				}
				originalPath = fmt.Sprintf("original/%04d%s", i+1, ext)
				// Original snapshots belong to the explicit comparison export, never its input tree.
				if e = fsutil.AtomicWrite(c.Context(), filepath.Join(outRoot, filepath.FromSlash(originalPath)), data, 0600, false); e != nil {
					return e
				}
			}
			for n, v := range variants {
				for pass := 1; pass <= repeat; pass++ {
					row := comparison.Row{Input: source.Display, Variant: v.ID, Repetition: pass, Recipe: v.Recipe, Pipeline: []domain.Step{}, Warnings: []string{}}
					if dry {
						row.Status = "planned"
						if source.Err != nil {
							row.Status = "failed"
							row.Error = source.Err.Error()
							row.ExitCode = domain.Code(source.Err)
						}
					} else if loadErr != nil {
						row.Status = "failed"
						row.Error = loadErr.Error()
						row.ExitCode = domain.Code(loadErr)
					} else {
						row = comparison.Evaluate(c.Context(), engine, source.Display, data, v, pass)
					}
					row.OriginalPath = originalPath
					if outRoot != "" && !dry && len(row.Data) > 0 && (row.ExitCode == 0 || row.ExitCode == 6 && nearest) {
						ext := ".png"
						if row.After.Format == "jpeg" {
							ext = ".jpg"
						}
						path := fmt.Sprintf("candidates/%04d-%02d-%02d%s", i+1, n+1, pass, ext)
						if e = fsutil.AtomicWrite(c.Context(), filepath.Join(outRoot, filepath.FromSlash(path)), row.Data, 0600, false); e != nil {
							return e
						}
						row.CandidatePath = path
					}
					row.Data = nil
					report.Rows = append(report.Rows, row)
				}
			}
			if c.Context().Err() != nil {
				break
			}
		}
		report.Finish()
		if outRoot != "" && !dry {
			files := map[string][]byte{}
			var b bytes.Buffer
			if e = comparison.JSON(&b, report); e != nil {
				return e
			}
			files["report.json"] = bytes.Clone(b.Bytes())
			b.Reset()
			if e = comparison.CSV(&b, report); e != nil {
				return e
			}
			files["report.csv"] = bytes.Clone(b.Bytes())
			html, e := comparison.HTML(report)
			if e != nil {
				return e
			}
			files["index.html"] = html
			for name, data := range files {
				if e = fsutil.AtomicWrite(c.Context(), filepath.Join(outRoot, name), data, 0600, false); e != nil {
					return e
				}
			}
		}
		if a.json {
			e = comparison.JSON(a.env.Out, report)
		} else if a.jsonl {
			enc := json.NewEncoder(a.env.Out)
			for _, row := range report.Rows {
				if e = enc.Encode(row); e != nil {
					break
				}
			}
		} else if format == "csv" {
			e = comparison.CSV(a.env.Out, report)
		} else if !a.quiet {
			e = comparison.Human(a.env.Out, report)
			if outRoot != "" && !dry {
				fmt.Fprintf(a.env.Out, "Saved comparison: %q\n", filepath.Join(outRoot, "index.html"))
			}
		}
		if e != nil {
			return domain.E(5, "write comparison report", e)
		}
		if report.ExitCode != 0 {
			return reported(report.ExitCode)
		}
		return nil
	}}
	f.bind(c)
	c.Flags().StringArrayVar(&presets, "preset", nil, "Compare selected built-in/user/project presets; repeatable (default: all four built-ins)")
	c.Flags().StringArrayVar(&modes, "png-mode", nil, "PNG tool set: auto, pipeline, quantizer or lossless; repeatable")
	c.Flags().StringVar(&target, "target-size", "", "Apply a common strict target to every comparison recipe")
	c.Flags().BoolVar(&nearest, "accept-nearest", false, "Include missed-target artifacts in an explicit comparison export; exit still indicates failure")
	c.Flags().IntVar(&repeat, "repeat", 1, "Serial measured runs per variant, 1..5")
	c.Flags().StringVar(&output, "output-dir", "", "New directory for original snapshots, candidates, JSON, CSV and offline HTML")
	c.Flags().StringVar(&format, "format", "human", "Human or CSV stdout (or use --json/--jsonl)")
	c.Flags().BoolVar(&dry, "dry-run", false, "Show the matrix without probes, downloads, compression or writes")
	return c
}
