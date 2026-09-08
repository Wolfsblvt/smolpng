// SPDX-License-Identifier: MIT
// Package comparison evaluates named local recipes through the production engine.
// It never selects a universal quality winner, uploads, writes originals or changes settings.
package comparison

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"math"
	"runtime"
	"slices"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/imagefile"
	"github.com/Wolfsblvt/smolpng/internal/jpegimage"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/toolchain"
	"github.com/Wolfsblvt/smolpng/internal/workflow"
)

type Variant struct {
	ID      string        `json:"id"`
	Recipe  domain.Recipe `json:"recipe"`
	PNGMode string        `json:"pngMode"`
	Formats []string      `json:"formats"`
}
type Metrics struct {
	PixelIdentical bool     `json:"pixelIdentical"`
	Samples        int      `json:"samples"`
	BlackRMSE      *float64 `json:"blackCompositeRMSE,omitempty"`
	WhiteRMSE      *float64 `json:"whiteCompositeRMSE,omitempty"`
	AlphaMAE       *float64 `json:"alphaMAE,omitempty"`
	OmittedReason  string   `json:"omittedReason,omitempty"`
}
type Row struct {
	Input           string           `json:"input"`
	SourceSHA256    string           `json:"sourceSHA256"`
	Variant         string           `json:"variant"`
	Repetition      int              `json:"repetition"`
	Recipe          domain.Recipe    `json:"recipe"`
	Status          string           `json:"status"`
	ExitCode        int              `json:"exitCode"`
	Error           string           `json:"error,omitempty"`
	ElapsedMS       float64          `json:"elapsedMilliseconds"`
	OriginalBytes   int              `json:"originalBytes"`
	CandidateBytes  int              `json:"candidateBytes"`
	CandidateSHA256 string           `json:"candidateSHA256,omitempty"`
	Before          domain.ImageInfo `json:"before"`
	After           domain.ImageInfo `json:"after"`
	TargetMet       *bool            `json:"targetMet,omitempty"`
	Pipeline        []domain.Step    `json:"pipeline"`
	Metrics         *Metrics         `json:"metrics,omitempty"`
	Warnings        []string         `json:"warnings"`
	OriginalPath    string           `json:"originalPath,omitempty"`
	CandidatePath   string           `json:"candidatePath,omitempty"`
	Data            []byte           `json:"-"`
}
type Report struct {
	SchemaVersion int           `json:"schemaVersion"`
	Version       string        `json:"version"`
	Commit        string        `json:"commit"`
	Platform      string        `json:"platform"`
	Tools         toolchain.Set `json:"tools"`
	Variants      []Variant     `json:"variants"`
	Rows          []Row         `json:"rows"`
	Notes         []string      `json:"notes"`
	ExitCode      int           `json:"exitCode"`
}

func NewReport(version, commit string, tools toolchain.Set, variants []Variant) Report {
	return Report{SchemaVersion: 1, Version: version, Commit: commit, Platform: runtime.GOOS + "/" + runtime.GOARCH, Tools: tools, Variants: variants, Rows: []Row{}, Notes: []string{
		"Local only. Every variant and repetition consumes the same original bytes; no setting, clipboard, credential or input file is changed.",
		"Timings include validation and process startup. Repetitions are serial measurements, not a controlled CPU benchmark.",
		"Recipe quality numbers are backend-specific. No universal winner is selected from unlike quality or dimension constraints.",
		"RMSE is sampled encoded-channel error over black/white composites, normalized to 0..1; it is not perceptual or color-managed quality.",
	}}
}
func ValidateVariants(vs []Variant) error {
	if len(vs) == 0 || len(vs) > 64 {
		return fmt.Errorf("comparison needs 1..64 variants")
	}
	seen := map[string]bool{}
	for _, v := range vs {
		if seen[v.ID] || v.ID == "" {
			return fmt.Errorf("comparison IDs must be distinct and nonempty")
		}
		seen[v.ID] = true
		if e := v.Recipe.Validate(); e != nil {
			return e
		}
		if v.Recipe.Engine == "tinypng" || v.Recipe.RemotePolicy != "never" && v.Recipe.RemotePolicy != "ask" {
			return fmt.Errorf("comparison is local-only; select a local preset (remote compare remains an explicit compression policy)")
		}
		if !slices.Contains([]string{"auto", "pipeline", "quantizer", "lossless"}, v.PNGMode) {
			return fmt.Errorf("png-mode must be auto, pipeline, quantizer, or lossless")
		}
	}
	return nil
}
func Evaluate(ctx context.Context, engine workflow.Engine, source string, data []byte, v Variant, repetition int) Row {
	row := Row{Input: source, SourceSHA256: fmt.Sprintf("%x", sha256.Sum256(data)), Variant: v.ID, Repetition: repetition, Recipe: v.Recipe, OriginalBytes: len(data), Pipeline: []domain.Step{}, Warnings: []string{}}
	fail := func(e error) Row {
		row.ExitCode = domain.Code(e)
		row.Status = "failed"
		row.Error = e.Error()
		return row
	}
	if err := ValidateVariants([]Variant{v}); err != nil {
		return fail(domain.E(2, err.Error(), err))
	}
	before, err := imagefile.Inspect(data, engine.Limits.MaxPixels)
	if err != nil {
		return fail(domain.E(3, err.Error(), err))
	}
	row.Before = before
	if len(v.Formats) > 0 && !slices.Contains(v.Formats, before.Format) {
		row.Status = "skipped"
		row.Warnings = append(row.Warnings, "preset does not admit this format")
		return row
	}
	recipe := v.Recipe
	recipe.Engine = "local"
	recipe.RemotePolicy = "never"
	engine.Remote = nil
	if v.PNGMode != "auto" && before.Format != "png" {
		row.Status = "skipped"
		row.Warnings = append(row.Warnings, "explicit PNG pipeline is not a JPEG backend")
		return row
	}
	switch v.PNGMode {
	case "pipeline":
		if !engine.Tools.PNGQuant.Available || !engine.Tools.OxiPNG.Available {
			return fail(domain.E(4, "requested full PNG pipeline needs both pngquant and OxiPNG", nil))
		}
		if recipe.Lossless {
			return fail(domain.E(2, "quantizing pipeline conflicts with a lossless recipe", nil))
		}
	case "quantizer":
		if !engine.Tools.PNGQuant.Available {
			return fail(domain.E(4, "requested quantizer requires pngquant", nil))
		}
		if recipe.Lossless {
			return fail(domain.E(2, "quantizer conflicts with a lossless recipe", nil))
		}
		engine.Tools.OxiPNG.Available = false
	case "lossless":
		if !engine.Tools.OxiPNG.Available {
			return fail(domain.E(4, "requested lossless pipeline requires OxiPNG", nil))
		}
		recipe.Lossless = true
		recipe.OptimizeAlpha = false
	}
	row.Recipe = recipe
	started := time.Now()
	out, err := engine.Optimize(ctx, bytes.Clone(data), recipe)
	row.ElapsedMS = float64(time.Since(started).Nanoseconds()) / 1e6
	row.Warnings = append(row.Warnings, out.Warnings...)
	row.Pipeline = out.Candidate.Pipeline
	if row.Pipeline == nil {
		row.Pipeline = []domain.Step{}
	}
	row.ExitCode = domain.Code(err)
	if err != nil {
		row.Error = err.Error()
	}
	if len(out.Candidate.Data) == 0 {
		row.Status = "failed"
		return row
	}
	candidate := out.Candidate.Data
	after, e := imagefile.Inspect(candidate, engine.Limits.MaxPixels)
	if e != nil {
		return fail(domain.E(5, "comparison candidate failed independent decoding", e))
	}
	row.After = after
	row.Data = candidate
	row.CandidateBytes = len(candidate)
	row.CandidateSHA256 = fmt.Sprintf("%x", sha256.Sum256(candidate))
	row.Status = "candidate"
	if bytes.Equal(candidate, data) {
		row.Status = "unchanged"
	}
	if row.ExitCode == 6 {
		row.Status = "target-missed"
	}
	if recipe.TargetBytes > 0 {
		met := int64(len(candidate)) <= recipe.TargetBytes
		row.TargetMet = &met
	}
	row.Metrics = measure(data, candidate, recipe, engine.Limits.MaxPixels)
	return row
}
func (r *Report) Finish() {
	r.ExitCode = 0
	failures := 0
	for _, row := range r.Rows {
		if row.ExitCode != 0 {
			failures++
			r.ExitCode = row.ExitCode
		}
		if row.ExitCode == 130 {
			r.ExitCode = 130
			return
		}
	}
	if failures > 0 && len(r.Rows) > 1 {
		r.ExitCode = 8
	}
}
func decode(b []byte, maxPixels int64) (image.Image, []byte, error) {
	var colorData bytes.Buffer
	if jpegimage.Is(b) {
		a, e := jpegimage.Decode(b, maxPixels)
		if e != nil {
			return nil, nil, e
		}
		for _, s := range a.Segments {
			if s.Marker == 0xe2 && bytes.HasPrefix(s.Data, []byte("ICC_PROFILE")) {
				colorData.Write(s.Data)
			}
		}
		return jpegimage.Upright(a), colorData.Bytes(), nil
	}
	a, e := pngimage.Decode(b, maxPixels)
	if e != nil {
		return nil, nil, e
	}
	for _, s := range a.Chunks {
		if slices.Contains([]string{"iCCP", "sRGB", "gAMA", "cHRM", "cICP"}, s.Type) {
			colorData.WriteString(s.Type)
			colorData.Write(s.Data)
		}
	}
	return a.Image, colorData.Bytes(), nil
}
func channels(c color.Color) [4]float64 {
	var v color.NRGBA64
	switch x := c.(type) {
	case color.NRGBA:
		v = color.NRGBA64{R: uint16(x.R) * 257, G: uint16(x.G) * 257, B: uint16(x.B) * 257, A: uint16(x.A) * 257}
	case color.NRGBA64:
		v = x
	default:
		v = color.NRGBA64Model.Convert(c).(color.NRGBA64)
	}
	return [4]float64{float64(v.R) / 65535, float64(v.G) / 65535, float64(v.B) / 65535, float64(v.A) / 65535}
}
func measure(source, candidate []byte, r domain.Recipe, maxPixels int64) *Metrics {
	m := &Metrics{}
	a, ac, e := decode(source, maxPixels)
	b, bc, f := decode(candidate, maxPixels)
	if e != nil || f != nil {
		m.OmittedReason = "independent decode unavailable"
		return m
	}
	if a.Bounds().Size() != b.Bounds().Size() {
		m.OmittedReason = "dimensions differ; no resize-normalized score is implied"
		return m
	}
	m.PixelIdentical = pngimage.PixelEqual(a, b)
	if r.Background != "" {
		m.OmittedReason = "requested alpha flattening changes the reference; compare visually against the selected matte"
		return m
	}
	if !bytes.Equal(ac, bc) {
		m.OmittedReason = "color metadata differs; a color-managed evaluator is required"
		return m
	}
	w, h := a.Bounds().Dx(), a.Bounds().Dy()
	sx, sy := max(1, (w+511)/512), max(1, (h+511)/512)
	black, white, alpha := 0.0, 0.0, 0.0
	for y := 0; y < h; y += sy {
		for x := 0; x < w; x += sx {
			p := channels(a.At(a.Bounds().Min.X+x, a.Bounds().Min.Y+y))
			q := channels(b.At(b.Bounds().Min.X+x, b.Bounds().Min.Y+y))
			for k := 0; k < 3; k++ {
				d := p[k]*p[3] - q[k]*q[3]
				black += d * d
				d += (1 - p[3]) - (1 - q[3])
				white += d * d
			}
			alpha += math.Abs(p[3] - q[3])
			m.Samples++
		}
	}
	black = math.Sqrt(black / float64(m.Samples*3))
	white = math.Sqrt(white / float64(m.Samples*3))
	alpha /= float64(m.Samples)
	m.BlackRMSE = &black
	m.WhiteRMSE = &white
	m.AlphaMAE = &alpha
	return m
}
