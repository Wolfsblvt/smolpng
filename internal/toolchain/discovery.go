// SPDX-License-Identifier: MIT
package toolchain

import (
	"context"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

type Tool struct {
	Name      string `json:"name"`
	Path      string `json:"path,omitempty"`
	Version   string `json:"version,omitempty"`
	Source    string `json:"source"`
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
	License   string `json:"license"`
	Upstream  string `json:"upstream"`
}
type Set struct {
	JPEGOptim Tool `json:"jpegoptim"`
	PNGQuant  Tool `json:"pngquant"`
	OxiPNG    Tool `json:"oxipng"`
}
type Paths struct{ PNGQuant, OxiPNG, JPEGOptim, ManagedRoot, Preference string }

var version = regexp.MustCompile(`\b([0-9]+)\.([0-9]+)(?:\.([0-9]+))?\b`)

func Discover(ctx context.Context, r process.Runner, explicit, config Paths) Set {
	if r == nil {
		r = process.OSRunner{}
	}
	selectTool := func(name, override, saved string) Tool {
		if override != "" || saved != "" {
			return probe(ctx, r, name, override, saved)
		}
		managed := Tool{}
		managedResult := func() Tool {
			state := (Managed{Root: config.ManagedRoot}).Inspect(name)
			if state.State == "current" {
				t := probe(ctx, r, name, state.Path, "")
				t.Source = "managed"
				return t
			}
			t := Tool{Name: name, Source: "managed", License: "MIT", Upstream: "https://github.com/oxipng/oxipng"}
			if name == "pngquant" {
				t.License = "GPL-3.0-or-later with upstream notices OR commercial"
				t.Upstream = "https://pngquant.org/"
			}
			t.Error = "managed dependency " + state.State
			if state.Reason != "" {
				t.Error += ": " + state.Reason
			}
			return t
		}
		if config.Preference == "managed" && config.ManagedRoot != "" {
			managed = managedResult()
			if managed.Available {
				return managed
			}
		}
		system := probe(ctx, r, name, "", "")
		if system.Available {
			return system
		}
		if config.ManagedRoot != "" {
			if managed.Name == "" {
				managed = managedResult()
			}
			if managed.Available {
				return managed
			}
			if system.Source == "none" && managed.Error != "managed dependency missing" {
				return managed
			}
		}
		return system
	}
	return Set{PNGQuant: selectTool("pngquant", explicit.PNGQuant, config.PNGQuant), OxiPNG: selectTool("oxipng", explicit.OxiPNG, config.OxiPNG), JPEGOptim: probe(ctx, r, "jpegoptim", explicit.JPEGOptim, config.JPEGOptim)}
}
func probe(ctx context.Context, r process.Runner, name, explicit, configured string) Tool {
	t := Tool{Name: name, Source: "none", License: "MIT", Upstream: "https://github.com/oxipng/oxipng"}
	if name == "pngquant" {
		t.License = "GPL-3.0-or-later with upstream notices OR commercial"
		t.Upstream = "https://pngquant.org/"
	}
	if name == "jpegoptim" {
		t.License = "GPL (exact terms depend on the installed jpegoptim release)"
		t.Upstream = "https://github.com/tjko/jpegoptim"
	}
	switch {
	case explicit != "":
		t.Path = explicit
		t.Source = "invocation"
	case configured != "":
		t.Path = configured
		t.Source = "configured"
	default:
		if p, e := exec.LookPath(name); e == nil {
			t.Path = p
			t.Source = "system"
		}
	}
	if t.Path == "" {
		t.Error = "not found; install separately or configure an executable path"
		return t
	}
	p, e := filepath.Abs(t.Path)
	if e != nil {
		t.Error = "invalid executable path"
		return t
	}
	t.Path = p
	s, e := os.Stat(p)
	if e != nil || !s.Mode().IsRegular() {
		t.Error = "executable path is unavailable or not a regular file"
		return t
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	v, e := r.Run(ctx, process.Invocation{Executable: p, Args: []string{"--version"}})
	if e != nil {
		t.Error = fmt.Sprintf("version probe failed: %v", e)
		return t
	}
	if v.ExitCode != 0 {
		t.Error = fmt.Sprintf("version probe failed: exit code %d", v.ExitCode)
		return t
	}
	m := version.FindStringSubmatch(v.Stdout + " " + v.Stderr)
	if m == nil {
		t.Error = "version probe returned no recognized version"
		return t
	}
	t.Version = m[0]
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	if !(name == "pngquant" && (major == 3 || major == 2 && minor >= 12) || name == "oxipng" && (major == 9 || major == 10) || name == "jpegoptim" && major == 1 && (minor == 4 || minor == 5)) {
		t.Error = fmt.Sprintf("unsupported argument family for %s %s", name, t.Version)
		return t
	}
	t.Available = true
	return t
}
