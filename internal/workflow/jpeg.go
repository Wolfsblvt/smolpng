// SPDX-License-Identifier: MIT
package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/jpegimage"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/process"
)

// optimizeJPEG uses jpegoptim for entropy-only and ordinary JPEG re-encode
// candidates. Explicit geometry transformations are encoded directly from the
// original decode with Go's JPEG encoder, with only lossless cleanup afterward.
func (e Engine) optimizeJPEG(ctx context.Context, data []byte, r domain.Recipe) (Outcome, error) {
	out := Outcome{Warnings: []string{}}
	original, err := jpegimage.Decode(data, e.Limits.MaxPixels)
	if err != nil {
		return out, domain.E(3, err.Error(), err)
	}
	out.Original = original.Info
	transform := r.Resize.Width > 0 || r.MaxDimensions.Width > 0 || r.Metadata == "all" && original.Orientation != 1
	if r.Lossless && transform {
		return out, domain.E(2, "strict JPEG lossless cannot normalize orientation or resize; use safe metadata", nil)
	}
	if len(original.Trailer) > 0 {
		if r.Metadata == "keep" {
			if transform {
				return out, domain.E(2, "JPEG geometry changes with keep metadata are unsupported; choose safe or all", nil)
			}
			out.Candidate = Candidate{Data: data, Info: original.Info, Rank: 100000000, Pipeline: []domain.Step{}}
			out.Warnings = append(out.Warnings, fmt.Sprintf("Kept complete JPEG unchanged to preserve %d attached bytes; compression and resizing cannot preserve opaque camera attachments", len(original.Trailer)))
			if r.TargetBytes > 0 && int64(len(data)) > r.TargetBytes {
				return out, domain.E(6, "JPEG target was not met; keep metadata preserves the complete file unchanged", nil)
			}
			return out, nil
		}
		out.Warnings = append(out.Warnings, jpegAttachmentWarning(original.Info.AppendedBytes))
	}
	normalized := data
	if !transform {
		normalized, err = jpegimage.Filter(original, r.Metadata)
		if err != nil {
			return out, domain.E(3, err.Error(), err)
		}
	}
	var best *Candidate
	retain := func(c Candidate) {
		if !r.ForceOutput && len(c.Data) > len(data) {
			return
		}
		if Better(c, best, r.TargetBytes) {
			copy := c
			best = &copy
		}
	}
	if !transform {
		info := original.Info
		info.Bytes = int64(len(normalized))
		info.AppendedBytes = 0
		retain(Candidate{Data: normalized, Info: info, Rank: 100000000, Pipeline: []domain.Step{{Tool: "smolpng-jpeg-metadata", Version: "1"}}})
	}
	fits := func() bool { return best != nil && r.TargetBytes > 0 && int64(len(best.Data)) <= r.TargetBytes }
	if fits() {
		out.Candidate = *best
		return out, nil
	}
	if e.Runner == nil {
		e.Runner = process.OSRunner{}
	}
	if e.CacheDir != "" {
		if err = os.MkdirAll(e.CacheDir, 0700); err != nil {
			return out, err
		}
	}
	temp, err := os.MkdirTemp(e.CacheDir, "smolpng-jpeg-*")
	if err != nil {
		return out, err
	}
	defer os.RemoveAll(temp)
	valid := 0
	serial := 0
	warn := func(err error) {
		for _, s := range out.Warnings {
			if s == err.Error() {
				return
			}
		}
		if len(out.Warnings) < 8 {
			out.Warnings = append(out.Warnings, err.Error())
		}
	}
	run := func(source []byte, quality int, strict bool, prefix []domain.Step) {
		if ctx.Err() != nil {
			return
		}
		dir := filepath.Join(temp, strconv.Itoa(serial))
		serial++
		if err := os.Mkdir(dir, 0700); err != nil {
			warn(err)
			return
		}
		defer os.RemoveAll(dir)
		src := filepath.Join(dir, "original.jpg")
		destDir := filepath.Join(dir, "out")
		if err := os.Mkdir(destDir, 0700); err != nil {
			warn(err)
			return
		}
		if err := os.WriteFile(src, source, 0600); err != nil {
			warn(err)
			return
		}
		args := []string{"--force", "--overwrite", "--keep-all", "--dest=" + destDir, "--all-progressive"}
		if !strict {
			args = append(args, fmt.Sprintf("--max=%d", quality))
		}
		args = append(args, "--", src)
		b, step, _, err := e.invoke(ctx, e.Tools.JPEGOptim, args, filepath.Join(destDir, "original.jpg"))
		if err != nil {
			warn(err)
			return
		}
		before, err := jpegimage.Decode(source, e.Limits.MaxPixels)
		if err != nil {
			warn(err)
			return
		}
		after, err := jpegimage.Decode(b, e.Limits.MaxPixels)
		if err != nil {
			warn(err)
			return
		}
		if before.Info.Dimensions != after.Info.Dimensions || before.Orientation != after.Orientation {
			warn(fmt.Errorf("jpegoptim changed dimensions/orientation"))
			return
		}
		if strict && !pngimage.PixelEqual(before.Image, after.Image) {
			warn(fmt.Errorf("jpegoptim violated strict decoded pixel identity"))
			return
		}
		b, err = jpegimage.Reconcile(before, after, r.Metadata, false)
		if err != nil {
			warn(err)
			return
		}
		valid++
		after.Info.Bytes = int64(len(b))
		after.Info.AppendedBytes = 0
		rank := quality * 100000
		if strict && len(prefix) == 0 {
			rank = 100000000
		}
		retain(Candidate{Data: b, Info: after.Info, Rank: rank, Pipeline: append(append([]domain.Step{}, prefix...), step)})
	}
	if !transform && e.Tools.JPEGOptim.Available {
		run(normalized, 100, true, nil)
	}
	if fits() {
		out.Candidate = *best
		return out, nil
	}
	if !transform && !e.Tools.JPEGOptim.Available {
		return out, domain.E(4, "jpegoptim 1.4+/1.5.x is required for local JPEG optimization; configure it separately", nil)
	}
	dims := []domain.Dimensions{original.Info.Dimensions}
	if transform || r.AllowResize {
		base := jpegimage.DisplayDimensions(original)
		dims = Dimensions(base, r)
		if !transform && len(dims) > 0 {
			dims[0] = original.Info.Dimensions
		}
	}
	qualities := []int{r.QualityMax}
	if r.TargetBytes > 0 {
		for q := r.QualityMax - 5; q > r.QualityMin; q -= 5 {
			qualities = append(qualities, q)
		}
		if r.QualityMin != r.QualityMax {
			qualities = append(qualities, r.QualityMin)
		}
	}
	if !r.Lossless {
		for i, d := range dims {
			changed := transform || i > 0
			for _, q := range qualities {
				if err := ctx.Err(); err != nil {
					return out, err
				}
				if !changed {
					run(normalized, q, false, nil)
				} else {
					if d.Area() > e.Limits.MaxPixels {
						return out, domain.E(2, "JPEG resize exceeds pixel limit", nil)
					}
					b, err := jpegimage.EncodeOriginal(original, d, q, r.Metadata)
					if err != nil {
						return out, domain.E(2, err.Error(), err)
					}
					a, err := jpegimage.Decode(b, e.Limits.MaxPixels)
					if err != nil {
						return out, err
					}
					valid++
					steps := []domain.Step{{Tool: "smolpng-jpeg-encode", Version: "Go", Arguments: []string{fmt.Sprintf("quality=%d", q), "orientation=normalized"}}}
					retain(Candidate{Data: b, Info: a.Info, Rank: q * 100000, Pipeline: steps})
					if e.Tools.JPEGOptim.Available {
						run(b, q, true, steps)
					}
				}
				if fits() {
					break
				}
			}
			if fits() {
				break
			}
		}
	}
	if ctx.Err() != nil {
		return out, ctx.Err()
	}
	if valid == 0 {
		return out, domain.E(5, "all JPEG optimizer attempts failed; see warnings", nil)
	}
	if best == nil {
		return out, domain.E(5, "no JPEG candidate satisfies no-growth; explicit transformations may need --force-output", nil)
	}
	out.Candidate = *best
	if r.TargetBytes > 0 && int64(len(best.Data)) > r.TargetBytes {
		return out, domain.E(6, "JPEG target was not met within the selected quality and dimension bounds", nil)
	}
	return out, nil
}

func jpegAttachmentWarning(n int64) string {
	return fmt.Sprintf("Removed %d attached bytes after the primary JPEG; output is a still image without attached depth, refocus, motion or other auxiliary content; use keep metadata to preserve the complete file unchanged", n)
}
