// SPDX-License-Identifier: MIT
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/config"
	"github.com/Wolfsblvt/smolpng/internal/desktop"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/state"
	"github.com/Wolfsblvt/smolpng/internal/ui"
	"github.com/spf13/cobra"
)

func (a *app) copyOutputs(ctx context.Context, report domain.Report, mode string) (desktop.CopyReceipt, error) {
	if mode == "none" {
		return desktop.CopyReceipt{Modes: []string{}}, nil
	}
	if a.env.Copy == nil {
		return desktop.CopyReceipt{}, fmt.Errorf("clipboard writing is unavailable on this host")
	}
	return a.env.Copy.Copy(ctx, state.Outputs(report), mode)
}
func (a *app) finishHuman(ctx context.Context, report domain.Report, eff config.Effective, mode string) (desktop.CopyReceipt, []string) {
	if mode == "" {
		mode = eff.Clipboard
		if len(report.Results) > 1 {
			mode = eff.BatchClipboard
		}
	}
	warnings := []string{}
	receipt := desktop.CopyReceipt{}
	if len(state.Outputs(report)) > 0 {
		r, e := a.copyOutputs(ctx, report, mode)
		if e != nil {
			warnings = append(warnings, "Could not update clipboard: "+e.Error())
		} else {
			receipt = r
		}
	}
	p, e := a.dirs()
	if e != nil {
		warnings = append(warnings, e.Error())
		return receipt, warnings
	}
	last := state.Last{SchemaVersion: 1, AppVersion: Version, Completed: time.Now().UTC(), Report: report, Warnings: warnings}
	if e = state.Write(ctx, filepath.Join(p.State, "last-result.json"), last); e != nil {
		warnings = append(warnings, "Could not retain last result: "+e.Error())
	}
	return receipt, warnings
}
func (a *app) dropCommand() *cobra.Command {
	f := &compressFlags{}
	c := &cobra.Command{Use: "drop [INPUT...]", Short: "Configured background compression, clipboard result and last-result recovery"}
	f.bind(c)

	c.RunE = func(c *cobra.Command, args []string) error {
		copyMode, hold := f.CopyMode, f.ClipboardHold
		if !c.Flags().Changed("clipboard-hold") {
			hold = 5 * time.Minute
		}
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
			return domain.E(2, e.Error(), e)
		}
		if copyMode != "" && !desktop.ValidCopy(copyMode) {
			return domain.E(2, "unknown clipboard mode", nil)
		}
		if hold < 0 || hold > 30*time.Minute {
			return domain.E(2, "clipboard-hold must be 0..30m", nil)
		}
		if eff.DropMode == "interactive" && !a.json && !a.jsonl {
			return a.interactive(c, args, f)
		}
		if !c.Flags().Changed("preset") {
			if e = c.Flags().Set("preset", eff.DropPreset); e != nil {
				return e
			}
		}
		if len(args) == 0 && len(f.Input.Options.Globs) == 0 {
			c.Flags().Set("clipboard", "true")
		}
		if f.Stdout || f.Output.Path == "-" {
			return domain.E(2, "binary stdout belongs to compress, not the background drop workflow", nil)
		}
		var result domain.Report
		var effective config.Effective
		e = a.compression(c, args, f, func(r domain.Report, eff config.Effective) error { result = r; effective = eff; return nil })
		if e != nil {
			r := domain.NewResult("<drop>", f.Preset)
			r.Fail(e)
			result = domain.Summarize([]domain.Result{r})
			effective = eff
		}
		var receipt desktop.CopyReceipt
		warnings := []string{}
		if !f.DryRun {
			mode := copyMode
			if mode == "" {
				if p := w.User.Document.Drop.Copy; p != nil {
					mode = *p
				}
				if p := w.Project.Document.Drop.Copy; p != nil {
					mode = *p
				}
			}
			receipt, warnings = a.finishHuman(c.Context(), result, effective, mode)
			if effective.Notify && a.env.Desktop != nil {
				body := fmt.Sprintf("%d result(s), exit %d. Recover files with smolpng last.", len(result.Results), result.ExitCode)
				for _, item := range result.Results {
					if len(item.Warnings) > 0 {
						body = "Warning: " + item.Warnings[0] + " " + body
						break
					}
				}
				if er := a.env.Desktop.Notify(c.Context(), "smolpng", body); er != nil {
					warnings = append(warnings, "Desktop notification unavailable: "+er.Error())
				}
			}
		}
		if !f.DryRun && len(warnings) > 0 {
			last := state.Last{SchemaVersion: 1, AppVersion: Version, Completed: time.Now().UTC(), Report: result, Warnings: warnings}
			if er := state.Write(c.Context(), filepath.Join(dirs.State, "last-result.json"), last); er != nil {
				warnings = append(warnings, "Could not retain updated result warnings: "+er.Error())
			}
		}
		if len(result.Results) > 0 {
			result.Results[0].Warnings = append(result.Results[0].Warnings, warnings...)
		}
		if a.json {
			if e = a.emit(result); e != nil {
				return e
			}
		} else if a.jsonl {
			for _, r := range result.Results {
				if e = a.emit(r); e != nil {
					return e
				}
			}
		} else if !a.quiet {
			if e = human(a.env.Out, result); e != nil {
				return e
			}
		}
		if (a.json || a.jsonl) && !c.Flags().Changed("clipboard-hold") {
			hold = 0
		}
		desktop.Hold(c.Context(), receipt, hold)
		if result.ExitCode != 0 {
			return reported(result.ExitCode)
		}
		return nil
	}
	return c
}
func (a *app) lastCommand() *cobra.Command {
	var copyMode string
	var open, clear bool
	var hold time.Duration
	c := &cobra.Command{Use: "last", Short: "Inspect, copy, open or clear the last human/background result", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		dirs, e := a.dirs()
		if e != nil {
			return e
		}
		path := filepath.Join(dirs.State, "last-result.json")
		if clear {
			if e = os.Remove(path); e != nil && !os.IsNotExist(e) {
				return domain.E(5, "could not clear last result", e)
			}
			return a.emit(map[string]any{"schemaVersion": 1, "cleared": true, "artifactsDeleted": false})
		}
		if hold < 0 || hold > 30*time.Minute {
			return domain.E(2, "clipboard-hold must be 0..30m", nil)
		}
		last, e := state.ReadLast(path)
		if e != nil {
			return domain.E(3, "no readable last result: "+e.Error(), e)
		}
		files := state.Outputs(last.Report)
		exist := map[string]bool{}
		for _, p := range files {
			st, e := os.Stat(p)
			exist[p] = e == nil && st.Mode().IsRegular()
		}
		receipt := desktop.CopyReceipt{}
		if copyMode != "" {
			if !desktop.ValidCopy(copyMode) {
				return domain.E(2, "unknown clipboard mode", nil)
			}
			receipt, e = a.copyOutputs(c.Context(), last.Report, copyMode)
			if e != nil {
				return domain.E(5, e.Error(), e)
			}
		}
		if open {
			if len(files) == 0 {
				return domain.E(3, "last operation produced no file", nil)
			}
			if e = a.env.Desktop.OpenFolder(c.Context(), filepath.Dir(files[0])); e != nil {
				return domain.E(5, e.Error(), e)
			}
		}
		if a.json || a.jsonl {
			e = a.emit(map[string]any{"schemaVersion": 1, "last": last, "exists": exist, "clipboard": receipt})
		} else {
			e = human(a.env.Out, last.Report)
			for _, warning := range last.Warnings {
				fmt.Fprintln(a.env.Out, "Warning: "+ui.Safe(warning))
			}
			if receipt.NeedsOwner && hold == 0 {
				fmt.Fprintln(a.env.Err, "Clipboard ownership may end on exit; select --clipboard-hold to keep serving it.")
			}
		}
		if e != nil {
			return e
		}
		desktop.Hold(c.Context(), receipt, hold)
		return nil
	}}
	c.Flags().StringVar(&copyMode, "copy", "", "Explicit clipboard result: smart, path, image, files, directory or none")
	c.Flags().BoolVar(&open, "open", false, "Open the containing output folder")
	c.Flags().BoolVar(&clear, "clear", false, "Clear retained metadata without deleting image files")
	c.Flags().DurationVar(&hold, "clipboard-hold", 0, "On Linux keep serving a copied result until replaced or timeout")
	c.MarkFlagsMutuallyExclusive("clear", "copy")
	c.MarkFlagsMutuallyExclusive("clear", "open")
	return c
}
