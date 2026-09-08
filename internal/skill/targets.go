// SPDX-License-Identifier: MIT
package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Target struct {
	Name        string `json:"name"`
	Destination string `json:"destination"`
	Managed     bool   `json:"managed"`
	Detected    bool   `json:"detected"`
	Reason      string `json:"reason"`
}

func Targets(home, cwd string) []Target {
	exists := func(p string) bool { s, e := os.Stat(p); return e == nil && s.IsDir() }
	return []Target{
		{"agents", filepath.Join(home, ".agents", "skills", "smolpng"), true, exists(filepath.Join(home, ".agents")) || exists(filepath.Join(home, ".codex")) || exists(filepath.Join(home, ".copilot")), "Shared personal Skills; current Codex and supported Copilot surfaces use ~/.agents/skills"},
		{"claude", filepath.Join(home, ".claude", "skills", "smolpng"), true, exists(filepath.Join(home, ".claude")), "Claude Code personal Skills (custom CLAUDE_CONFIG_DIR: use --path explicitly)"},
		{"project", filepath.Join(cwd, ".agents", "skills", "smolpng"), true, exists(filepath.Join(cwd, ".agents")) || exists(filepath.Join(cwd, ".git")), "Project-local shared Skill; does not mutate other project instructions"},
	}
}
func ResolveTarget(name, customParent, home, cwd string) (Target, error) {
	if customParent != "" {
		if name != "" && name != "auto" {
			return Target{}, fmt.Errorf("choose --target or --path, not both")
		}
		p := customParent
		if !filepath.IsAbs(p) {
			p = filepath.Join(cwd, p)
		}
		p = filepath.Join(p, "smolpng")
		return Target{Name: "custom", Destination: p, Managed: false, Reason: "Explicit custom parent; updates require synchronization consent"}, nil
	}
	if strings.HasPrefix(name, "path:") {
		return ResolveTarget("", strings.TrimPrefix(name, "path:"), home, cwd)
	}
	if name == "auto" || name == "" || name == "codex" {
		name = "agents"
	}
	for _, t := range Targets(home, cwd) {
		if t.Name == name {
			return t, nil
		}
	}
	return Target{}, fmt.Errorf("target must be auto, agents, codex, claude, project, or path:<parent>")
}
