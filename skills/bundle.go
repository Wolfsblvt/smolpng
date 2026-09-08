// SPDX-License-Identifier: MIT
// Package skills embeds the user-facing product Skill, independently of repository AGENTS.md.
package skills

import (
	"embed"
	"io/fs"
)

//go:embed smolpng
var content embed.FS

func Files() fs.FS { f, _ := fs.Sub(content, "smolpng"); return f }
