// SPDX-License-Identifier: MIT
// Package domain carries the format-neutral request and report contracts.
package domain

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	Success           = 0
	InvalidArguments  = 2
	InvalidInput      = 3
	MissingDependency = 4
	OptimizerFailed   = 5
	TargetMissed      = 6
	NetworkFailed     = 7
	PartialFailure    = 8
	Cancelled         = 130
)

type Error struct {
	Code    int
	Message string
	Cause   error
}

func (e *Error) Error() string                      { return e.Message }
func (e *Error) Unwrap() error                      { return e.Cause }
func E(code int, message string, cause error) error { return &Error{code, message, cause} }
func Code(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, context.Canceled) {
		return Cancelled
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return OptimizerFailed
}

type Dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (d Dimensions) Area() int64    { return int64(d.Width) * int64(d.Height) }
func (d Dimensions) String() string { return fmt.Sprintf("%dx%d", d.Width, d.Height) }
func ParseDimensions(s string) (Dimensions, error) {
	p := strings.Split(strings.ToLower(strings.TrimSpace(s)), "x")
	if len(p) != 2 {
		return Dimensions{}, fmt.Errorf("dimensions must be WIDTHxHEIGHT")
	}
	w, e := strconv.Atoi(p[0])
	h, f := strconv.Atoi(p[1])
	if e != nil || f != nil || w < 1 || h < 1 || w > 1000000 || h > 1000000 {
		return Dimensions{}, fmt.Errorf("dimensions must be positive and at most 1000000 per side")
	}
	return Dimensions{w, h}, nil
}

var sizePattern = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)\s*(B|KB|KIB|MB|MIB|GB|GIB)?$`)

// ParseSize uses exact arithmetic, rejecting fractional bytes and overflow.
func ParseSize(s string) (int64, error) {
	m := sizePattern.FindStringSubmatch(strings.ToUpper(strings.TrimSpace(s)))
	if m == nil {
		return 0, fmt.Errorf("invalid size %q; use bytes, KB, KiB, MB, MiB, GB, or GiB", s)
	}
	units := map[string]int64{"": 1, "B": 1, "KB": 1000, "KIB": 1024, "MB": 1000000, "MIB": 1048576, "GB": 1000000000, "GIB": 1073741824}
	n, ok := new(big.Rat).SetString(m[1])
	if !ok {
		return 0, fmt.Errorf("invalid size")
	}
	n.Mul(n, new(big.Rat).SetInt64(units[m[2]]))
	if !n.IsInt() || !n.Num().IsInt64() || n.Sign() <= 0 {
		return 0, fmt.Errorf("size must be a positive whole number of bytes within int64")
	}
	return n.Num().Int64(), nil
}
func ParseQuality(s string) (int, int, error) {
	p := strings.Split(s, "-")
	if len(p) != 2 {
		return 0, 0, fmt.Errorf("quality must be MIN-MAX")
	}
	lo, e := strconv.Atoi(p[0])
	hi, f := strconv.Atoi(p[1])
	if e != nil || f != nil || lo < 0 || hi > 100 || lo > hi {
		return 0, 0, fmt.Errorf("quality must satisfy 0 <= MIN <= MAX <= 100")
	}
	return lo, hi, nil
}

// QualityRange contains encoder-specific quality bounds, not perceptual scores.
type QualityRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type Recipe struct {
	Background    string        `json:"background,omitempty"`
	Name          string        `json:"name"`
	Engine        string        `json:"engine"`
	RemotePolicy  string        `json:"remotePolicy"`
	Lossless      bool          `json:"lossless"`
	QualityMin    int           `json:"qualityMin"`
	QualityMax    int           `json:"qualityMax"`
	JPEGQuality   *QualityRange `json:"jpegQuality,omitempty"`
	Colors        int           `json:"colors"`
	Speed         int           `json:"speed"`
	Dither        string        `json:"dither"`
	OxiLevel      int           `json:"oxipngLevel"`
	Metadata      string        `json:"metadata"`
	OptimizeAlpha bool          `json:"optimizeAlpha"`
	TargetBytes   int64         `json:"targetBytes"`
	AcceptNearest bool          `json:"acceptNearest"`
	AllowResize   bool          `json:"allowResize"`
	Resize        Dimensions    `json:"resize"`
	MaxDimensions Dimensions    `json:"maxDimensions"`
	MinDimensions Dimensions    `json:"minDimensions"`
	ForceOutput   bool          `json:"forceOutput"`
	Timeout       time.Duration `json:"timeout"`
	PNGQuantArgs  []string      `json:"pngquantArgs,omitempty"`
	OxiPNGArgs    []string      `json:"oxipngArgs,omitempty"`
}

// JPEGQualityRange returns explicit JPEG bounds or the shared quality override.
func (r Recipe) JPEGQualityRange() QualityRange {
	if r.JPEGQuality != nil {
		return *r.JPEGQuality
	}
	return QualityRange{Min: r.QualityMin, Max: r.QualityMax}
}

// SetQuality applies a shared override to both supported formats.
func (r *Recipe) SetQuality(minimum, maximum int) {
	r.QualityMin, r.QualityMax = minimum, maximum
	r.JPEGQuality = &QualityRange{Min: minimum, Max: maximum}
}

// SetMinQuality preserves each format's upper bound while changing the floor.
func (r *Recipe) SetMinQuality(minimum int) {
	jpeg := r.JPEGQualityRange()
	jpeg.Min = minimum
	r.QualityMin, r.JPEGQuality = minimum, &jpeg
}

func Builtin(name string) (Recipe, error) {
	r := Recipe{Name: name, Engine: "local", RemotePolicy: "never", QualityMin: 70, QualityMax: 95, Colors: 256, Speed: 1, Dither: "auto", OxiLevel: 4, Metadata: "safe", MinDimensions: Dimensions{64, 64}, Timeout: 2 * time.Minute}
	switch name {
	case "auto":
		r.JPEGQuality = &QualityRange{Min: 70, Max: 85}
	case "web":
		r.JPEGQuality = &QualityRange{Min: 70, Max: 80}
	case "smallest":
		r.JPEGQuality = &QualityRange{Min: 55, Max: 70}
		r.QualityMin = 55
		r.QualityMax = 85
		r.Colors = 128
		r.OxiLevel = 6
	case "lossless":
		r.JPEGQuality = &QualityRange{Min: 70, Max: 95}
		r.Lossless = true
	default:
		return Recipe{}, fmt.Errorf("unknown built-in preset %q", name)
	}
	return r, nil
}
func IsBuiltin(name string) bool { _, e := Builtin(name); return e == nil }
func (r Recipe) Validate() error {
	if r.Background != "" {
		if _, err := ParseBackground(r.Background); err != nil {
			return err
		}
		if r.Lossless {
			return fmt.Errorf("strict lossless cannot flatten transparency")
		}
	}
	if r.QualityMin < 0 || r.QualityMax > 100 || r.QualityMin > r.QualityMax {
		return fmt.Errorf("quality must satisfy 0 <= MIN <= MAX <= 100")
	}
	jpeg := r.JPEGQualityRange()
	if jpeg.Min < 0 || jpeg.Max > 100 || jpeg.Min > jpeg.Max {
		return fmt.Errorf("JPEG quality must satisfy 0 <= MIN <= MAX <= 100")
	}
	if r.Colors < 2 || r.Colors > 256 {
		return fmt.Errorf("colors must be 2..256")
	}
	if r.Speed < 1 || r.Speed > 11 {
		return fmt.Errorf("speed must be 1..11")
	}
	if r.OxiLevel < 0 || r.OxiLevel > 6 {
		return fmt.Errorf("oxipng-level must be 0..6 or max")
	}
	if r.Dither != "auto" && r.Dither != "none" {
		n, e := strconv.ParseFloat(r.Dither, 64)
		if e != nil || !(n >= 0 && n <= 1) {
			return fmt.Errorf("dither must be auto, none, or 0..1")
		}
	}
	switch r.Metadata {
	case "keep", "safe", "all":
	default:
		return fmt.Errorf("metadata must be keep, safe, or all")
	}
	switch r.Engine {
	case "auto", "local", "tinypng":
	default:
		return fmt.Errorf("engine must be auto, local, or tinypng")
	}
	switch r.RemotePolicy {
	case "never", "ask", "fallback", "compare", "always":
	default:
		return fmt.Errorf("unknown remote policy")
	}
	if r.TargetBytes < 0 || r.Timeout < 0 {
		return fmt.Errorf("negative target or timeout")
	}
	if r.AcceptNearest && r.TargetBytes == 0 {
		return fmt.Errorf("accept-nearest needs a target size")
	}
	if r.AllowResize && r.TargetBytes == 0 {
		return fmt.Errorf("allow-resize needs a target size")
	}
	for _, d := range []Dimensions{r.Resize, r.MaxDimensions, r.MinDimensions} {
		if (d.Width == 0) != (d.Height == 0) || d.Width < 0 || d.Height < 0 || d.Width > 1000000 || d.Height > 1000000 {
			return fmt.Errorf("invalid dimensions")
		}
	}
	if r.Resize.Width > 0 && r.MaxDimensions.Width > 0 {
		return fmt.Errorf("resize and max-dimensions are mutually exclusive")
	}
	if r.AllowResize && (r.MinDimensions.Width <= 0 || r.MinDimensions.Height <= 0) {
		return fmt.Errorf("target resize floor must be positive")
	}
	if r.Lossless && (r.AllowResize || r.Resize.Width > 0 || r.MaxDimensions.Width > 0 || r.OptimizeAlpha || r.Engine == "tinypng" || r.RemotePolicy == "always" || r.RemotePolicy == "compare" || r.RemotePolicy == "fallback") {
		return fmt.Errorf("strict lossless cannot resize, normalize hidden RGB, or use a lossy remote provider")
	}
	if r.Engine == "tinypng" && (r.RemotePolicy == "never" || r.RemotePolicy == "ask") {
		return fmt.Errorf("TinyPNG needs explicit upload permission")
	}
	return nil
}

type Limits struct{ MaxBytes, MaxPixels int64 }

func DefaultLimits() Limits { return Limits{64 << 20, 64_000_000} }
func (l Limits) Validate() error {
	if l.MaxBytes < 1 || l.MaxBytes > 1<<30 || l.MaxPixels < 1 || l.MaxPixels > 256_000_000 {
		return fmt.Errorf("limits must be positive, at most 1GiB encoded and 256 million pixels")
	}
	return nil
}

type ImageInfo struct {
	AppendedBytes int64      `json:"appendedBytes,omitempty"`
	Orientation   int        `json:"orientation,omitempty"`
	Progressive   bool       `json:"progressive,omitempty"`
	Format        string     `json:"format"`
	Bytes         int64      `json:"bytes"`
	Dimensions    Dimensions `json:"dimensions"`
	BitDepth      int        `json:"bitDepth"`
	ColorType     int        `json:"colorType"`
	Alpha         bool       `json:"alpha"`
	Transparency  bool       `json:"transparency"`
	Interlaced    bool       `json:"interlaced"`
	Animated      bool       `json:"animated"`
	Chunks        []string   `json:"chunks"`
}
type Step struct {
	Tool      string   `json:"tool"`
	Version   string   `json:"version"`
	Arguments []string `json:"arguments,omitempty"`
}
type Target struct {
	RequestedBytes int64 `json:"requestedBytes"`
	Met            bool  `json:"met"`
	NearestBytes   int64 `json:"nearestBytes,omitempty"`
}
type DimensionChange struct {
	Before Dimensions `json:"before"`
	After  Dimensions `json:"after"`
}

// RemoteUsage describes an attempted upload even when a local candidate wins.
type RemoteUsage struct {
	Provider         string `json:"provider"`
	CredentialSource string `json:"credentialSource,omitempty"`
	Attempted        bool   `json:"attempted"`
	Selected         bool   `json:"selected"`
	CompressionCount *int   `json:"compressionCount,omitempty"`
}

// JPEGAttachment describes detected auxiliary data and the selected preservation
// policy. Output and Status establish whether a corresponding file was committed.
type JPEGAttachment struct {
	Bytes int64 `json:"bytes"`
	Keep  bool  `json:"keep"`
}

type Result struct {
	SchemaVersion  int             `json:"schemaVersion"`
	Status         string          `json:"status"`
	Input          string          `json:"input"`
	Output         string          `json:"output,omitempty"`
	OriginalBytes  int64           `json:"originalBytes"`
	OutputBytes    int64           `json:"outputBytes"`
	SavedBytes     int64           `json:"savedBytes"`
	SavingPercent  float64         `json:"savingPercent"`
	Format         string          `json:"format"`
	Recipe         string          `json:"recipe"`
	Dimensions     DimensionChange `json:"dimensions"`
	Pipeline       []Step          `json:"pipeline"`
	Target         *Target         `json:"target,omitempty"`
	ResizeUsed     bool            `json:"resizeUsed"`
	Remote         *RemoteUsage    `json:"remote,omitempty"`
	RemoteUsed     bool            `json:"remoteUsed"`
	Warnings       []string        `json:"warnings"`
	JPEGAttachment *JPEGAttachment `json:"jpegAttachment,omitempty"`
	Error          string          `json:"error,omitempty"`
	ExitCode       int             `json:"exitCode"`
	Bytes          []byte          `json:"-"`
}

func NewResult(input, recipe string) Result {
	return Result{SchemaVersion: 1, Status: "pending", Input: input, Recipe: recipe, Pipeline: []Step{}, Warnings: []string{}}
}
func (r *Result) Fail(e error) {
	r.ExitCode = Code(e)
	r.Error = e.Error()
	r.Status = "failed"
	if r.ExitCode == Cancelled {
		r.Status = "cancelled"
	}
	if r.ExitCode == TargetMissed {
		r.Status = "target-missed"
	}
}

type Summary struct {
	Total         int   `json:"total"`
	Succeeded     int   `json:"succeeded"`
	Failed        int   `json:"failed"`
	Skipped       int   `json:"skipped"`
	OriginalBytes int64 `json:"originalBytes"`
	OutputBytes   int64 `json:"outputBytes"`
}
type Report struct {
	SchemaVersion int      `json:"schemaVersion"`
	Results       []Result `json:"results"`
	Summary       Summary  `json:"summary"`
	ExitCode      int      `json:"exitCode"`
}

func Summarize(results []Result) Report {
	r := Report{SchemaVersion: 1, Results: results}
	cancelled := false
	for _, v := range results {
		r.Summary.Total++
		if v.Status == "skipped" {
			r.Summary.Skipped++
		} else if v.ExitCode != 0 {
			r.Summary.Failed++
			r.ExitCode = v.ExitCode
			cancelled = cancelled || v.ExitCode == Cancelled
		} else {
			r.Summary.Succeeded++
		}
		if v.Output != "" && v.Status != "planned" {
			r.Summary.OriginalBytes += v.OriginalBytes
			r.Summary.OutputBytes += v.OutputBytes
		}
	}
	if r.Summary.Failed > 0 && len(results) > 1 {
		r.ExitCode = PartialFailure
	}
	if cancelled {
		r.ExitCode = Cancelled
	}
	return r
}

// ParseBackground accepts an explicit opaque RGB matte. Empty means no change.
func ParseBackground(s string) ([3]uint16, error) {
	switch strings.ToLower(s) {
	case "white":
		s = "#ffffff"
	case "black":
		s = "#000000"
	}
	if len(s) != 7 || s[0] != '#' {
		return [3]uint16{}, fmt.Errorf("background must be #RRGGBB, white, or black (empty preserves transparency)")
	}
	n, e := strconv.ParseUint(s[1:], 16, 24)
	if e != nil {
		return [3]uint16{}, fmt.Errorf("background contains invalid hex digits")
	}
	return [3]uint16{uint16(n>>16) * 257, uint16((n>>8)&255) * 257, uint16(n&255) * 257}, nil
}
