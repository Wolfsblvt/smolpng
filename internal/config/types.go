// SPDX-License-Identifier: MIT
// Package config owns data-only configuration, preset inheritance and provenance.
package config

import (
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/output"
)

const MaxDocumentBytes = 1 << 20

// Fields uses pointers so an explicit false, zero, empty string or empty list is
// distinguishable from inheritance. No execution, environment or secret fields
// are available to a shared preset.
type Fields struct {
	Background    *string `toml:"background,omitempty" json:"background,omitempty"`
	Quality       *string `toml:"quality,omitempty" json:"quality,omitempty"`
	MinQuality    *int    `toml:"min_quality,omitempty" json:"min_quality,omitempty"`
	Colors        *int    `toml:"colors,omitempty" json:"colors,omitempty"`
	Speed         *int    `toml:"speed,omitempty" json:"speed,omitempty"`
	Dither        *string `toml:"dither,omitempty" json:"dither,omitempty"`
	OxiLevel      *int    `toml:"oxipng_level,omitempty" json:"oxipng_level,omitempty"`
	Metadata      *string `toml:"metadata,omitempty" json:"metadata,omitempty"`
	Quantize      *bool   `toml:"quantize,omitempty" json:"quantize,omitempty"`
	OptimizeAlpha *bool   `toml:"optimize_alpha,omitempty" json:"optimize_alpha,omitempty"`
	Target        *string `toml:"target_size,omitempty" json:"target_size,omitempty"`
	AllowResize   *bool   `toml:"allow_resize,omitempty" json:"allow_resize,omitempty"`
	Resize        *string `toml:"resize,omitempty" json:"resize,omitempty"`
	MaxDimensions *string `toml:"max_dimensions,omitempty" json:"max_dimensions,omitempty"`
	MinDimensions *string `toml:"min_dimensions,omitempty" json:"min_dimensions,omitempty"`
	AcceptNearest *bool   `toml:"accept_nearest,omitempty" json:"accept_nearest,omitempty"`
	ForceOutput   *bool   `toml:"force_output,omitempty" json:"force_output,omitempty"`
	Engine        *string `toml:"engine,omitempty" json:"engine,omitempty"`
	Remote        *string `toml:"remote,omitempty" json:"remote,omitempty"`
	Template      *string `toml:"template,omitempty" json:"template,omitempty"`
	Directory     *string `toml:"output_directory,omitempty" json:"output_directory,omitempty"`
	Clipboard     *string `toml:"clipboard,omitempty" json:"clipboard,omitempty"`
	Recursive     *bool   `toml:"recursive,omitempty" json:"recursive,omitempty"`
	Jobs          *int    `toml:"jobs,omitempty" json:"jobs,omitempty"`
	Timeout       *string `toml:"timeout,omitempty" json:"timeout,omitempty"`
}

type Menu struct {
	Pinned   *bool `toml:"pinned,omitempty" json:"pinned,omitempty"`
	Position *int  `toml:"position,omitempty" json:"position,omitempty"`
	Hidden   *bool `toml:"hidden,omitempty" json:"hidden,omitempty"`
}
type Preset struct {
	Schema      int      `toml:"schema" json:"schema"`
	Label       string   `toml:"label,omitempty" json:"label,omitempty"`
	Description string   `toml:"description,omitempty" json:"description,omitempty"`
	Extends     string   `toml:"extends,omitempty" json:"extends,omitempty"`
	Formats     []string `toml:"formats,omitempty" json:"formats,omitempty"`
	Author      string   `toml:"author,omitempty" json:"author,omitempty"`
	Source      string   `toml:"source,omitempty" json:"source,omitempty"`
	Fields
	Menu Menu `toml:"menu,omitempty" json:"menu,omitempty"`
}
type Bundle struct {
	Presets map[string]Preset `toml:"presets" json:"presets"`
}
type Defaults struct {
	Preset       *string `toml:"preset,omitempty" json:"preset,omitempty"`
	RemotePolicy *string `toml:"remote_policy,omitempty" json:"remote_policy,omitempty"`
	Overwrite    *bool   `toml:"overwrite,omitempty" json:"overwrite,omitempty"`
	Fields
}
type Output struct {
	LocalFiles *string `toml:"local_files,omitempty" json:"local_files,omitempty"`
	Directory  *string `toml:"directory,omitempty" json:"directory,omitempty"`
	Detached   *string `toml:"detached,omitempty" json:"detached,omitempty"`
	Template   *string `toml:"template,omitempty" json:"template,omitempty"`
	Collision  *string `toml:"collision,omitempty" json:"collision,omitempty"`
}
type Clipboard struct {
	Single *string `toml:"single,omitempty" json:"single,omitempty"`
	Batch  *string `toml:"batch,omitempty" json:"batch,omitempty"`
}
type Drop struct {
	Mode   *string `toml:"mode,omitempty" json:"mode,omitempty"`
	Preset *string `toml:"preset,omitempty" json:"preset,omitempty"`
	Notify *bool   `toml:"notify,omitempty" json:"notify,omitempty"`
	Copy   *string `toml:"copy,omitempty" json:"copy,omitempty"`
}
type UI struct {
	Header     *string `toml:"header,omitempty" json:"header,omitempty"`
	Animations *bool   `toml:"animations,omitempty" json:"animations,omitempty"`
}
type Updates struct {
	Check   *string `toml:"check,omitempty" json:"check,omitempty"`
	Channel *string `toml:"channel,omitempty" json:"channel,omitempty"`
}
type Tools struct {
	JPEGOptim  string `toml:"jpegoptim,omitempty" json:"jpegoptim,omitempty"`
	Preference string `toml:"preference,omitempty" json:"preference,omitempty"`
	PNGQuant   string `toml:"pngquant,omitempty" json:"pngquant,omitempty"`
	OxiPNG     string `toml:"oxipng,omitempty" json:"oxipng,omitempty"`
}
type PNG struct {
	Auto Fields `toml:"auto,omitempty" json:"auto,omitempty"`
}
type Document struct {
	Version   int               `toml:"config_version,omitempty" json:"config_version"`
	Defaults  Defaults          `toml:"defaults,omitempty" json:"defaults"`
	Output    Output            `toml:"output,omitempty" json:"output"`
	Clipboard Clipboard         `toml:"clipboard,omitempty" json:"clipboard"`
	Drop      Drop              `toml:"drop,omitempty" json:"drop"`
	UI        UI                `toml:"ui,omitempty" json:"ui"`
	Updates   Updates           `toml:"updates,omitempty" json:"updates"`
	Tools     Tools             `toml:"tools,omitempty" json:"tools"`
	PNG       PNG               `toml:"png,omitempty" json:"png"`
	Presets   map[string]Preset `toml:"presets,omitempty" json:"presets,omitempty"`
}

type Origin struct {
	File string `json:"file"`
	Key  string `json:"key"`
}
type Effective struct {
	SchemaVersion    int               `json:"schemaVersion"`
	Recipe           domain.Recipe     `json:"recipe"`
	Output           output.Options    `json:"output"`
	Jobs             int               `json:"jobs"`
	Recursive        bool              `json:"recursive"`
	Clipboard        string            `json:"clipboard"`
	BatchClipboard   string            `json:"batchClipboard"`
	DropMode         string            `json:"dropMode"`
	DropPreset       string            `json:"dropPreset"`
	Notify           bool              `json:"notify"`
	Header           string            `json:"header"`
	Animations       bool              `json:"animations"`
	UpdateCheck      string            `json:"updateCheck"`
	UpdateChannel    string            `json:"updateChannel"`
	RemotePermission string            `json:"remotePermission"`
	Tools            Tools             `json:"tools"`
	Formats          []string          `json:"formats"`
	Provenance       map[string]Origin `json:"provenance"`
	Warnings         []string          `json:"warnings"`
}

func Ptr[T any](v T) *T { return &v }
