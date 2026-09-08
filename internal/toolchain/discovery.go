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
	"runtime"
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
	PNGQuant Tool `json:"pngquant"`
	OxiPNG   Tool `json:"oxipng"`
}
type Paths struct{ PNGQuant, OxiPNG, ManagedRoot string }

var version = regexp.MustCompile(`\b([0-9]+)\.([0-9]+)(?:\.([0-9]+))?\b`)

func Discover(ctx context.Context, r process.Runner, explicit, config Paths) Set {
	return Set{probe(ctx, r, "pngquant", explicit.PNGQuant, config.PNGQuant, config.ManagedRoot), probe(ctx, r, "oxipng", explicit.OxiPNG, config.OxiPNG, config.ManagedRoot)}
}
func probe(ctx context.Context, r process.Runner, name, explicit, configured, managed string) Tool {
	t := Tool{Name: name, Source: "none", License: "MIT", Upstream: "https://github.com/oxipng/oxipng"}
	if name == "pngquant" {
		t.License = "GPL-3.0-or-later with upstream notices OR commercial"
		t.Upstream = "https://pngquant.org/"
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
		} else if managed != "" {
			exe := name
			if runtime.GOOS == "windows" {
				exe += ".exe"
			}
			p := filepath.Join(managed, name, exe)
			if s, e := os.Stat(p); e == nil && s.Mode().IsRegular() {
				t.Path = p
				t.Source = "managed"
			}
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
	if e != nil || v.ExitCode != 0 {
		t.Error = "version probe failed"
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
	if !(name == "pngquant" && (major == 3 || major == 2 && minor >= 12) || name == "oxipng" && (major == 9 || major == 10)) {
		t.Error = fmt.Sprintf("unsupported argument family for %s %s", name, t.Version)
		return t
	}
	t.Available = true
	return t
}
