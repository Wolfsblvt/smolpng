// SPDX-License-Identifier: MIT
package workflow

import (
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"math"
	"sort"
	"strconv"
	"strings"
)

type Candidate struct {
	Data     []byte
	Info     domain.ImageInfo
	Rank     int
	Pipeline []domain.Step
}
type Attempt struct {
	QualityMax, Colors int
	Dither             string
}

func (a Attempt) Rank() int { return a.QualityMax*10000 + a.Colors }
func Attempts(r domain.Recipe) []Attempt {
	if r.Lossless {
		return nil
	}
	quality := []int{r.QualityMax}
	colors := []int{r.Colors}
	dither := []string{r.Dither}
	if r.TargetBytes > 0 {
		quality = append(quality, (r.QualityMax+r.QualityMin)/2, r.QualityMin)
		for _, n := range []int{128, 64, 32, 16, 8, 4, 2} {
			if n < r.Colors {
				colors = append(colors, n)
			}
		}
		if r.Dither == "auto" {
			dither = append(dither, "none")
		}
	} else if r.Name == "auto" {
		quality = append(quality, max(r.QualityMin, r.QualityMax-10))
	}
	seen := map[Attempt]bool{}
	out := []Attempt{}
	for _, q := range quality {
		for _, c := range colors {
			for _, d := range dither {
				a := Attempt{q, c, d}
				if !seen[a] {
					out = append(out, a)
					seen[a] = true
				}
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Rank() > out[j].Rank() })
	return out
}

// Recipe rank is declared preference, not a claimed visual-quality measurement.
func Better(a Candidate, b *Candidate, target int64) bool {
	if b == nil {
		return true
	}
	if target > 0 {
		am, bm := int64(len(a.Data)) <= target, int64(len(b.Data)) <= target
		if am != bm {
			return am
		}
		if am {
			if a.Info.Dimensions.Area() != b.Info.Dimensions.Area() {
				return a.Info.Dimensions.Area() > b.Info.Dimensions.Area()
			}
			if a.Rank != b.Rank {
				return a.Rank > b.Rank
			}
		}
	}
	if len(a.Data) != len(b.Data) {
		return len(a.Data) < len(b.Data)
	}
	return a.Rank > b.Rank
}
func Dimensions(original domain.Dimensions, r domain.Recipe) []domain.Dimensions {
	base := original
	if r.Resize.Width > 0 {
		base = r.Resize
	} else if r.MaxDimensions.Width > 0 {
		base = fit(original, r.MaxDimensions)
	}
	out := []domain.Dimensions{base}
	if !r.AllowResize || r.TargetBytes == 0 {
		return out
	}
	floor := math.Max(float64(r.MinDimensions.Width)/float64(base.Width), float64(r.MinDimensions.Height)/float64(base.Height))
	if floor >= 1 {
		return out
	}
	for s := 0.8; ; s *= 0.8 {
		s = math.Max(s, floor)
		d := domain.Dimensions{Width: int(math.Ceil(float64(base.Width) * s)), Height: int(math.Ceil(float64(base.Height) * s))}
		if d != out[len(out)-1] {
			out = append(out, d)
		}
		if s == floor {
			return out
		}
	}
}
func fit(d, box domain.Dimensions) domain.Dimensions {
	if d.Width <= box.Width && d.Height <= box.Height {
		return d
	}
	s := math.Min(float64(box.Width)/float64(d.Width), float64(box.Height)/float64(d.Height))
	return domain.Dimensions{Width: max(1, int(float64(d.Width)*s)), Height: max(1, int(float64(d.Height)*s))}
}
func QuantArgs(source, dest string, r domain.Recipe, a Attempt) []string {
	args := []string{strconv.Itoa(a.Colors), fmt.Sprintf("--quality=%d-%d", r.QualityMin, a.QualityMax), fmt.Sprintf("--speed=%d", r.Speed), "--force", "--output", dest}
	if r.Metadata == "all" {
		args = append(args, "--strip")
	}
	if a.Dither == "none" {
		args = append(args, "--nofs")
	} else if a.Dither != "auto" {
		args = append(args, "--floyd="+a.Dither)
	}
	args = append(args, r.PNGQuantArgs...)
	return append(args, "--", source)
}
func OxiArgs(source, dest string, r domain.Recipe, strict bool) []string {
	args := []string{"-o", strconv.Itoa(r.OxiLevel), "--threads", "1", "--force", "--out", dest}
	if r.OptimizeAlpha && !strict {
		args = append(args, "--alpha")
	}
	args = append(args, r.OxiPNGArgs...)
	return append(args, "--", source)
}
func ValidateExtras(r domain.Recipe) error {
	seen := map[string]bool{}
	for _, arg := range r.PNGQuantArgs {
		k, v, ok := strings.Cut(arg, "=")
		n, e := strconv.Atoi(v)
		if !ok || k != "--posterize" || e != nil || n < 0 || n > 4 || seen[k] || r.Lossless {
			return fmt.Errorf("unsafe pngquant extra %q; use typed flags or exec", arg)
		}
		seen[k] = true
	}
	for _, arg := range r.OxiPNGArgs {
		k, v, has := strings.Cut(arg, "=")
		if seen[k] {
			return fmt.Errorf("duplicate optimizer argument")
		}
		seen[k] = true
		switch k {
		case "--interlace":
			if !has || (v != "0" && v != "1" && v != "keep") {
				return fmt.Errorf("interlace extra must be 0, 1, or keep")
			}
		case "--zopfli":
			if has {
				return fmt.Errorf("zopfli takes no value")
			}
		default:
			return fmt.Errorf("unsafe OxiPNG extra %q; use typed flags or exec", arg)
		}
	}
	return nil
}
