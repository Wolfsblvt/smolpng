// SPDX-License-Identifier: MIT
package config

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/output"
	"github.com/pelletier/go-toml/v2"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)

func validName(name string) bool { return namePattern.MatchString(name) }
func decode(data []byte, dest any) error {
	if len(data) > MaxDocumentBytes {
		return fmt.Errorf("TOML document exceeds 1MiB")
	}
	decoder := toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields()
	if e := decoder.Decode(dest); e != nil {
		return fmt.Errorf("invalid TOML: %w", e)
	}
	return nil
}
func Parse(data []byte) (Document, error) {
	var d Document
	if e := decode(data, &d); e != nil {
		return d, e
	}
	if !slices.Contains([]string{"", "system", "managed"}, d.Tools.Preference) {
		return d, fmt.Errorf("tools.preference must be system or managed")
	}
	if d.Version != 0 && d.Version != 1 {
		return d, fmt.Errorf("unsupported config_version %d", d.Version)
	}
	if e := validatePresets(d.Presets); e != nil {
		return d, e
	}
	if e := validateFields(d.Defaults.Fields, false); e != nil {
		return d, fmt.Errorf("defaults: %w", e)
	}
	if e := validateFields(d.PNG.Auto, false); e != nil {
		return d, fmt.Errorf("png.auto: %w", e)
	}
	for _, s := range []*string{d.Output.Template} {
		if s != nil {
			if e := output.ValidateTemplate(*s); e != nil {
				return d, e
			}
		}
	}
	if s := d.Output.LocalFiles; s != nil && *s != "source" {
		return d, fmt.Errorf("output.local_files must be source; use output.directory for a destination")
	}
	if s := d.Output.Collision; s != nil && !slices.Contains([]string{"increment", "error"}, *s) {
		return d, fmt.Errorf("output.collision must be increment or error")
	}
	for _, s := range []*string{d.Clipboard.Single, d.Clipboard.Batch, d.Drop.Copy} {
		if s != nil && !validCopy(*s) {
			return d, fmt.Errorf("unknown clipboard mode %q", *s)
		}
	}
	if s := d.Defaults.RemotePolicy; s != nil && !validRemote(*s) {
		return d, fmt.Errorf("unknown remote_policy %q", *s)
	}
	if s := d.Drop.Mode; s != nil && !slices.Contains([]string{"background", "interactive"}, *s) {
		return d, fmt.Errorf("drop.mode must be background or interactive")
	}
	if s := d.UI.Header; s != nil && !slices.Contains([]string{"cute", "compact", "none"}, *s) {
		return d, fmt.Errorf("ui.header must be cute, compact, or none")
	}
	if s := d.Updates.Check; s != nil && !slices.Contains([]string{"manual", "daily"}, *s) {
		return d, fmt.Errorf("updates.check must be manual or daily")
	}
	if s := d.Updates.Channel; s != nil && !slices.Contains([]string{"stable"}, *s) {
		return d, fmt.Errorf("updates.channel must be stable; prerelease feeds are not enabled")
	}
	return d, nil
}
func validRemote(s string) bool {
	return slices.Contains([]string{"never", "ask", "fallback", "compare", "always"}, s)
}
func validCopy(s string) bool {
	return slices.Contains([]string{"none", "smart", "path", "image", "file", "files", "directory"}, s)
}
func validateFields(f Fields, shared bool) error {
	if f.Quality != nil {
		if _, _, e := domain.ParseQuality(*f.Quality); e != nil {
			return e
		}
	}
	if f.MinQuality != nil && (*f.MinQuality < 0 || *f.MinQuality > 100) {
		return fmt.Errorf("min_quality must be 0..100")
	}
	if f.Colors != nil && (*f.Colors < 2 || *f.Colors > 256) {
		return fmt.Errorf("colors must be 2..256")
	}
	if f.Speed != nil && (*f.Speed < 1 || *f.Speed > 11) {
		return fmt.Errorf("speed must be 1..11")
	}
	if f.OxiLevel != nil && (*f.OxiLevel < 0 || *f.OxiLevel > 6) {
		return fmt.Errorf("oxipng_level must be 0..6")
	}
	if f.Jobs != nil && (*f.Jobs < 1 || *f.Jobs > 32) {
		return fmt.Errorf("jobs must be 1..32")
	}
	if f.Remote != nil && !validRemote(*f.Remote) {
		return fmt.Errorf("unknown remote request")
	}
	if f.Engine != nil && !slices.Contains([]string{"local", "auto", "tinypng"}, *f.Engine) {
		return fmt.Errorf("unknown engine")
	}
	if f.Metadata != nil && !slices.Contains([]string{"safe", "keep", "all"}, *f.Metadata) {
		return fmt.Errorf("metadata must be safe, keep, or all")
	}
	if f.Clipboard != nil && !validCopy(*f.Clipboard) {
		return fmt.Errorf("unknown clipboard mode")
	}
	if f.Template != nil {
		if e := output.ValidateTemplate(*f.Template); e != nil {
			return e
		}
	}
	if f.Target != nil && *f.Target != "" {
		if _, e := domain.ParseSize(*f.Target); e != nil {
			return e
		}
	}
	for _, d := range []*string{f.Resize, f.MaxDimensions, f.MinDimensions} {
		if d != nil && *d != "" {
			if _, e := domain.ParseDimensions(*d); e != nil {
				return e
			}
		}
	}
	if shared && f.Directory != nil && *f.Directory != "" && !portableRelative(*f.Directory) {
		return fmt.Errorf("shared output_directory must be a relative path without traversal or machine-specific syntax")
	}
	return nil
}
func portableRelative(s string) bool {
	s = strings.ReplaceAll(s, "\\", "/")
	if strings.ContainsAny(s, ":\x00\r\n") || strings.HasPrefix(s, "/") {
		return false
	}
	for _, part := range strings.Split(s, "/") {
		if part == ".." {
			return false
		}
	}
	return filepath.IsLocal(filepath.FromSlash(s))
}
func validatePresets(presets map[string]Preset) error {
	if len(presets) > 256 {
		return fmt.Errorf("at most 256 presets per document")
	}
	for name, p := range presets {
		if !validName(name) || domain.IsBuiltin(name) {
			return fmt.Errorf("preset %q is invalid or reserved by an immutable built-in", name)
		}
		if p.Schema != 1 {
			return fmt.Errorf("preset %q requires schema = 1", name)
		}
		if p.Extends != "" && !validName(p.Extends) {
			return fmt.Errorf("preset %q has an invalid parent", name)
		}
		if p.Menu.Position != nil && (*p.Menu.Position < 1 || *p.Menu.Position > 1000) {
			return fmt.Errorf("menu.position must be 1..1000")
		}
		for _, f := range p.Formats {
			if f != "png" && f != "jpeg" {
				return fmt.Errorf("unknown preset format %q", f)
			}
		}
		if e := validateFields(p.Fields, true); e != nil {
			return fmt.Errorf("preset %q: %w", name, e)
		}
	}
	return nil
}

// ParseBundle accepts plain TOML or a single fenced block, never global settings.
func ParseBundle(data []byte) (Bundle, error) {
	if len(data) > MaxDocumentBytes {
		return Bundle{}, fmt.Errorf("preset input exceeds 1MiB")
	}
	s := strings.TrimSpace(string(data))
	if strings.HasPrefix(s, "```") {
		first, rest, ok := strings.Cut(s, "\n")
		if !ok || first != "```toml" && first != "```" {
			return Bundle{}, fmt.Errorf("use one TOML fenced block")
		}
		if !strings.HasSuffix(rest, "```") {
			return Bundle{}, fmt.Errorf("unclosed preset fence")
		}
		s = strings.TrimSpace(strings.TrimSuffix(rest, "```"))
	}
	var b Bundle
	if e := decode([]byte(s), &b); e != nil {
		return b, e
	}
	if len(b.Presets) == 0 {
		return b, fmt.Errorf("no [presets.NAME] definitions found")
	}
	return b, validatePresets(b.Presets)
}
func LooksLikePreset(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "[presets.") || strings.HasPrefix(s, "```toml\n") && strings.Contains(s, "[presets.")
}
func ReadBounded(r io.Reader) ([]byte, error) {
	b, e := io.ReadAll(io.LimitReader(r, MaxDocumentBytes+1))
	if e != nil {
		return nil, e
	}
	if len(b) > MaxDocumentBytes {
		return nil, fmt.Errorf("configuration input exceeds 1MiB")
	}
	return b, nil
}
func MarshalBundle(b Bundle) ([]byte, error) { return toml.Marshal(b) }
