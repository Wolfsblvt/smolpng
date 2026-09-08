// SPDX-License-Identifier: MIT
// Package integration owns opt-in per-user launchers, not MIME defaults or PATH.
package integration

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type Locations struct {
	Menu    string `json:"menu"`
	Desktop string `json:"desktop"`
	SendTo  string `json:"sendTo,omitempty"`
}
type Entry struct {
	ID               string   `json:"id"`
	Path             string   `json:"path"`
	Name             string   `json:"name"`
	Executable       string   `json:"executable"`
	Arguments        []string `json:"arguments"`
	WorkingDirectory string   `json:"workingDirectory"`
	Terminal         bool     `json:"terminal"`
}
type Request struct {
	OS, Executable, Home, WorkingDirectory, Preset, Name string
	ConfigPath                                           string
	NoProject                                            bool
	Desktop, SendTo, Interactive, Custom                 bool
}

var identifier = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

func Plan(loc Locations, q Request) ([]Entry, error) {
	if q.OS != "windows" && q.OS != "linux" {
		return nil, fmt.Errorf("launcher installation is implemented for Windows and Linux; macOS uses the CLI/TUI")
	}
	if q.Preset == "" {
		q.Preset = "auto"
	}
	if !identifier.MatchString(q.Preset) {
		return nil, fmt.Errorf("preset must be a simple preset identifier")
	}
	exe, e := filepath.Abs(q.Executable)
	if e != nil {
		return nil, e
	}
	for _, s := range []string{exe, q.Home, q.WorkingDirectory, q.Name, q.ConfigPath} {
		if strings.ContainsAny(s, "\r\n\x00") {
			return nil, fmt.Errorf("launcher values cannot contain control characters")
		}
	}
	if q.OS == "linux" && strings.Contains(exe, "=") {
		return nil, fmt.Errorf("desktop-entry executable names cannot contain '='")
	}
	create := func(id, name, dir string, interactive bool) (Entry, error) {
		if dir == "" || !filepath.IsAbs(dir) {
			return Entry{}, fmt.Errorf("native user directory unavailable for %s", id)
		}
		args := []string{}
		program := exe
		if !interactive {
			if q.OS == "windows" {
				program = filepath.Join(filepath.Dir(exe), "smolpngw.exe")
			} else {
				args = append(args, "drop")
			}
		}
		if q.Home != "" {
			args = append(args, "--home", q.Home)
		}
		if q.ConfigPath != "" {
			args = append(args, "--config", q.ConfigPath)
		}
		if q.NoProject {
			args = append(args, "--no-project")
		}
		args = append(args, "--preset", q.Preset, "--")
		extension := ".desktop"
		if q.OS == "windows" {
			extension = ".lnk"
		}
		return Entry{ID: id, Path: filepath.Join(dir, "smolpng-"+id+extension), Name: name, Executable: program, Arguments: args, WorkingDirectory: q.WorkingDirectory, Terminal: interactive}, nil
	}
	entries := []Entry{}
	add := func(id, name, dir string, interactive bool) error {
		x, e := create(id, name, dir, interactive)
		if e == nil {
			entries = append(entries, x)
		}
		return e
	}
	if q.Custom {
		if !identifier.MatchString(q.Name) || len(q.Name) > 48 {
			return nil, fmt.Errorf("shortcut name must be 1..48 letters/digits/dashes/underscores")
		}
		if e = add("preset-"+q.Name, "smolpng · "+q.Name, loc.Desktop, q.Interactive); e != nil {
			return nil, e
		}
	} else {
		if e = add("interactive", "smolpng", loc.Menu, true); e != nil {
			return nil, e
		}
		if e = add("auto", "smolpng Auto", loc.Menu, false); e != nil {
			return nil, e
		}
		if q.Desktop {
			if e = add("desktop-auto", "smolpng Auto", loc.Desktop, false); e != nil {
				return nil, e
			}
			if e = add("desktop-interactive", "smolpng · Choose recipe", loc.Desktop, true); e != nil {
				return nil, e
			}
		}
		if q.SendTo {
			if q.OS != "windows" {
				return nil, fmt.Errorf("Send To integration is Windows-specific")
			}
			if e = add("send-to", "smolpng Auto", loc.SendTo, false); e != nil {
				return nil, e
			}
		}
	}
	return entries, nil
}

// DesktopExec quotes twice: command argument escaping then desktop string-value
// escaping. %F is appended outside quotes and receives independent file arguments.
func DesktopExec(args []string) string {
	out := make([]string, 0, len(args))
	for _, s := range args {
		s = strings.ReplaceAll(s, "%", "%%")
		s = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\\`", "$", "\\$").Replace(s)
		s = "\"" + s + "\""
		s = strings.ReplaceAll(s, `\`, `\\`)
		out = append(out, s)
	}
	return strings.Join(out, " ") + " %F"
}
func DesktopFile(x Entry) ([]byte, error) {
	for _, s := range append([]string{x.Name, x.Executable, x.WorkingDirectory}, x.Arguments...) {
		if strings.ContainsAny(s, "\n\r\x00") {
			return nil, fmt.Errorf("invalid desktop entry control character")
		}
	}
	escape := func(s string) string { return strings.ReplaceAll(s, `\`, `\\`) }
	return []byte(fmt.Sprintf("[Desktop Entry]\nType=Application\nName=%s\nComment=Make images smol\nExec=%s\nPath=%s\nTerminal=%t\nIcon=image-x-generic\nCategories=Graphics;Utility;\nStartupNotify=false\nX-Smolpng-Managed=true\n", escape(x.Name), DesktopExec(append([]string{x.Executable}, x.Arguments...)), escape(x.WorkingDirectory), x.Terminal)), nil
}

// WindowsArguments follows the CommandLineToArgvW quoting convention. It is a
// command line for a shortcut, never an interpolated shell command.
func WindowsArguments(args []string) string {
	out := []string{}
	for _, s := range args {
		var b strings.Builder
		b.WriteByte('"')
		slashes := 0
		for _, c := range s {
			if c == '\\' {
				slashes++
				continue
			}
			if c == '"' {
				b.WriteString(strings.Repeat(`\`, slashes*2+1))
				b.WriteRune(c)
			} else {
				b.WriteString(strings.Repeat(`\`, slashes))
				b.WriteRune(c)
			}
			slashes = 0
		}
		b.WriteString(strings.Repeat(`\`, slashes*2))
		b.WriteByte('"')
		out = append(out, b.String())
	}
	return strings.Join(out, " ")
}
