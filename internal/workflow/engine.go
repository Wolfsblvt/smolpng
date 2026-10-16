// SPDX-License-Identifier: MIT
package workflow

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"github.com/Wolfsblvt/smolpng/internal/tinypng"
	"github.com/Wolfsblvt/smolpng/internal/toolchain"
	"os"
	"path/filepath"
)

type Engine struct {
	Remote   tinypng.Optimizer
	Runner   process.Runner
	Tools    toolchain.Set
	Limits   domain.Limits
	CacheDir string
	Verbose  bool
}
type Outcome struct {
	Remote     *domain.RemoteUsage
	RemoteUsed bool
	Original   domain.ImageInfo
	Candidate  Candidate
	Warnings   []string
}

func (e Engine) optimizePNG(ctx context.Context, data []byte, r domain.Recipe) (Outcome, error) {
	out := Outcome{Warnings: []string{}}
	if er := r.Validate(); er != nil {
		return out, domain.E(2, er.Error(), er)
	}
	if er := ValidateExtras(r); er != nil {
		return out, domain.E(2, er.Error(), er)
	}
	if er := e.Limits.Validate(); er != nil {
		return out, domain.E(2, er.Error(), er)
	}
	if r.Engine == "tinypng" || r.RemotePolicy == "always" || r.RemotePolicy == "compare" || r.RemotePolicy == "fallback" {
		return out, domain.E(7, "TinyPNG is not implemented; no upload was attempted", nil)
	}
	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Timeout)
		defer cancel()
	}
	if er := ctx.Err(); er != nil {
		return out, er
	}
	original, er := pngimage.Decode(data, e.Limits.MaxPixels)
	if er != nil {
		return out, domain.E(3, er.Error(), er)
	}
	out.Original = original.Info
	if original.Info.Animated {
		return out, domain.E(3, "animated PNG compression is unsupported; refusing to discard frames", nil)
	}
	normalized, er := pngimage.Filter(data, r.Metadata)
	if er != nil {
		return out, domain.E(3, er.Error(), er)
	}
	transform := []domain.Step{}
	flattened := r.Background != ""
	if flattened {
		normalized, er = pngimage.Flatten(original, r.Background, r.Metadata)
		if er != nil {
			return out, domain.E(2, er.Error(), er)
		}
		original, er = pngimage.Decode(normalized, e.Limits.MaxPixels)
		if er != nil {
			return out, er
		}
		transform = append(transform, domain.Step{Tool: "smolpng-background", Version: "1", Arguments: []string{r.Background}})
	}
	if r.Resize.Width == 0 && r.MaxDimensions.Width == 0 && r.TargetBytes > 0 && int64(len(normalized)) <= r.TargetBytes {
		out.Candidate = Candidate{Data: normalized, Info: original.Info, Rank: 100000000, Pipeline: transform}
		return out, nil
	}
	if !flattened && !e.Tools.OxiPNG.Available && (r.Lossless || !e.Tools.PNGQuant.Available) {
		out.Warnings = append(out.Warnings, "oxipng: "+e.Tools.OxiPNG.Error)
		if !r.Lossless {
			out.Warnings = append(out.Warnings, "pngquant: "+e.Tools.PNGQuant.Error)
		}
		return out, domain.E(4, "no suitable local optimizer; run deps status or configure executable paths", nil)
	}
	if !e.Tools.PNGQuant.Available && !r.Lossless {
		out.Warnings = append(out.Warnings, "pngquant unavailable: using lossless candidates only")
	}
	if !e.Tools.OxiPNG.Available {
		out.Warnings = append(out.Warnings, "OxiPNG unavailable: quantized candidates receive no lossless optimizer tail")
	}
	if e.Runner == nil {
		e.Runner = process.OSRunner{}
	}
	if e.CacheDir != "" {
		if er = os.MkdirAll(e.CacheDir, 0700); er != nil {
			return out, domain.E(5, "cannot create image cache", er)
		}
	}
	temp, er := os.MkdirTemp(e.CacheDir, "smolpng-job-*")
	if er != nil {
		return out, er
	}
	defer os.RemoveAll(temp)
	var best *Candidate
	validAttempts := 0
	if flattened {
		validAttempts = 1
	}
	serial := 0
	retain := func(c Candidate) {
		if !r.ForceOutput && len(c.Data) > len(data) {
			return
		}
		if Better(c, best, r.TargetBytes) {
			copy := c
			best = &copy
		}
	}
	warning := func(er error) {
		s := er.Error()
		for _, w := range out.Warnings {
			if w == s {
				return
			}
		}
		if len(out.Warnings) < 10 {
			out.Warnings = append(out.Warnings, s)
		}
	}
	for _, d := range Dimensions(original.Info.Dimensions, r) {
		if er := ctx.Err(); er != nil {
			return out, er
		}
		if d.Area() > e.Limits.MaxPixels {
			return out, domain.E(2, "resize exceeds pixel limit", nil)
		}
		basisBytes := normalized
		resized := d != original.Info.Dimensions
		if resized {
			basisBytes, er = pngimage.Resize(original, d, r.Metadata)
			if er != nil {
				return out, domain.E(5, "cannot resize image while preserving metadata", er)
			}
		}
		basis, er := pngimage.Decode(basisBytes, e.Limits.MaxPixels)
		if er != nil {
			return out, domain.E(5, "normalized source failed validation", er)
		}
		prefix := append([]domain.Step{}, transform...)
		if resized {
			prefix = append(prefix, domain.Step{Tool: "smolpng-resize", Version: "1"})
		} else if !flattened && !bytes.Equal(data, basisBytes) {
			prefix = append(prefix, domain.Step{Tool: "smolpng-metadata", Version: "1"})
		}
		retain(Candidate{Data: basisBytes, Info: basis.Info, Rank: 100000000, Pipeline: prefix})
		source := filepath.Join(temp, fmt.Sprintf("original-%d.png", serial))
		serial++
		if er = os.WriteFile(source, basisBytes, 0600); er != nil {
			return out, er
		}
		validate := func(b []byte, rank int, steps []domain.Step, strict, quantized bool) (Candidate, error) {
			asset, err := pngimage.Decode(b, e.Limits.MaxPixels)
			if err != nil {
				return Candidate{}, fmt.Errorf("optimizer returned an invalid PNG")
			}
			if asset.Info.Animated || asset.Info.Dimensions != d {
				return Candidate{}, fmt.Errorf("optimizer unexpectedly changed dimensions or animation")
			}
			if basis.Info.Transparency && !asset.Info.Transparency {
				return Candidate{}, fmt.Errorf("optimizer discarded transparency")
			}
			if strict && !pngimage.PixelEqual(basis.Image, asset.Image) {
				return Candidate{}, fmt.Errorf("optimizer violated strict pixel identity, including hidden RGB")
			}
			b, err = pngimage.Reconcile(basisBytes, b, r.Metadata, quantized)
			if err != nil {
				return Candidate{}, err
			}
			return Candidate{Data: b, Info: asset.Info, Rank: rank, Pipeline: append(append([]domain.Step{}, prefix...), steps...)}, nil
		}
		if e.Tools.OxiPNG.Available {
			dest := filepath.Join(temp, "lossless.png")
			b, step, _, err := e.invoke(ctx, e.Tools.OxiPNG, OxiArgs(source, dest, r, true), dest)
			if err != nil {
				warning(err)
			} else {
				c, err := validate(b, 100000000, []domain.Step{step}, true, false)
				if err != nil {
					warning(err)
				} else {
					validAttempts++
					retain(c)
				}
			}
		}
		fits := func() bool { return best != nil && r.TargetBytes > 0 && int64(len(best.Data)) <= r.TargetBytes }
		if fits() {
			validAttempts++
			break
		}
		if !r.Lossless && e.Tools.PNGQuant.Available {
			acceptedRank := -1
			for _, a := range Attempts(r) {
				if er := ctx.Err(); er != nil {
					return out, er
				}
				if acceptedRank >= 0 && a.Rank() < acceptedRank {
					break
				}
				dest := filepath.Join(temp, "quantized.png")
				b, step, rejected, err := e.invoke(ctx, e.Tools.PNGQuant, QuantArgs(source, dest, r, a), dest)
				if rejected {
					validAttempts++
					continue
				}
				if err != nil {
					warning(err)
					continue
				}
				c, err := validate(b, a.Rank(), []domain.Step{step}, false, true)
				if err != nil {
					warning(err)
					continue
				}
				validAttempts++
				retain(c)
				if e.Tools.OxiPNG.Available {
					if err = os.WriteFile(dest, c.Data, 0600); err != nil {
						return out, err
					}
					final := filepath.Join(temp, "tail.png")
					tail, tailStep, _, err := e.invoke(ctx, e.Tools.OxiPNG, OxiArgs(dest, final, r, false), final)
					if err != nil {
						warning(err)
					} else {
						tailAsset, derr := pngimage.Decode(tail, e.Limits.MaxPixels)
						quantAsset, qerr := pngimage.Decode(c.Data, e.Limits.MaxPixels)
						if derr != nil || qerr != nil {
							warning(fmt.Errorf("invalid optimizer tail"))
						} else if !r.OptimizeAlpha && !pngimage.PixelEqual(quantAsset.Image, tailAsset.Image) {
							warning(fmt.Errorf("OxiPNG tail changed quantized pixels"))
						} else {
							cleaned, err := validate(tail, a.Rank(), []domain.Step{step, tailStep}, false, true)
							if err != nil {
								warning(err)
							} else {
								retain(cleaned)
							}
						}
					}
				}
				if fits() {
					acceptedRank = a.Rank()
				}
			}
		}
		if fits() {
			break
		}
	}
	if er := ctx.Err(); er != nil {
		return out, er
	}
	if validAttempts == 0 {
		return out, domain.E(5, "all optimizer attempts failed validation; see warnings", nil)
	}
	if best == nil {
		return out, domain.E(5, "no candidate satisfies the size-growth policy; explicit enlarging transformations need --force-output", nil)
	}
	out.Candidate = *best
	if r.TargetBytes > 0 && int64(len(best.Data)) > r.TargetBytes {
		return out, domain.E(6, "target size was not reached within the permitted quality and dimensions", nil)
	}
	return out, nil
}
func (e Engine) invoke(ctx context.Context, t toolchain.Tool, args []string, dest string) ([]byte, domain.Step, bool, error) {
	step := domain.Step{Tool: t.Name, Version: t.Version}
	if e.Verbose {
		step.Arguments = append([]string{}, args...)
	}
	if er := os.Remove(dest); er != nil && !os.IsNotExist(er) {
		return nil, step, false, fmt.Errorf("cannot clear temporary candidate")
	}
	result, er := e.Runner.Run(ctx, process.Invocation{Executable: t.Path, Args: args})
	if ctx.Err() != nil {
		return nil, step, false, ctx.Err()
	}
	if t.Name == "pngquant" && (result.ExitCode == 98 || result.ExitCode == 99) {
		return nil, step, true, nil
	}
	if er != nil || result.ExitCode != 0 {
		return nil, step, false, fmt.Errorf("%s failed (exit %d)", t.Name, result.ExitCode)
	}
	st, er := os.Lstat(dest)
	if er != nil || !st.Mode().IsRegular() {
		return nil, step, false, fmt.Errorf("%s returned no regular output file", t.Name)
	}
	if st.Size() > e.Limits.MaxBytes {
		return nil, step, false, fmt.Errorf("optimizer output exceeds the byte limit")
	}
	f, er := os.Open(dest)
	if er != nil {
		return nil, step, false, er
	}
	defer f.Close()
	b, er := pngimage.ReadBounded(f, e.Limits.MaxBytes)
	return b, step, false, er
}
