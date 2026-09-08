// SPDX-License-Identifier: MIT
package updates

import (
	"path/filepath"
	"strings"
)

type Channel struct {
	Owner       string `json:"owner"`
	CanApply    bool   `json:"canApply"`
	Instruction string `json:"instruction"`
}

func Installation(executable, distribution string) Channel {
	p := strings.ToLower(filepath.ToSlash(executable))
	p = strings.ReplaceAll(p, `\`, "/")
	owner := distribution
	// Recognizable package-manager paths override an accidentally copied standalone
	// build marker. Unknown/source builds never infer permission from writable files.
	switch {
	case strings.Contains(p, "/windowsapps/"):
		owner = "msix"
	case strings.Contains(p, "/scoop/"):
		owner = "scoop"
	case strings.Contains(p, "/chocolatey/"):
		owner = "chocolatey"
	case strings.Contains(p, "/microsoft/winget/"):
		owner = "winget"
	case strings.Contains(p, "/cellar/") || strings.Contains(p, "/caskroom/"):
		owner = "homebrew"
	case strings.HasPrefix(p, "/usr/bin/") || strings.HasPrefix(p, "/usr/sbin/") || strings.HasPrefix(p, "/bin/"):
		if owner == "standalone" || owner == "portable" {
			owner = "system-package"
		}
	}
	instructions := map[string]string{"winget": "Use winget upgrade with the installed smolpng package ID.", "scoop": "Use scoop update smolpng.", "chocolatey": "Use choco upgrade smolpng.", "homebrew": "Use brew upgrade with the installed smolpng formula/cask.", "deb": "Use the package's owning apt repository.", "rpm": "Use the package's owning dnf/yum repository.", "apk": "Use apk upgrade smolpng.", "arch": "Use the package's owning pacman/AUR route.", "msix": "Use Microsoft Store or the package's App Installer route.", "system-package": "Use the operating-system package manager; smolpng will not replace system files.", "source": "Update source and rebuild; self-update is disabled for source builds."}
	if owner == "standalone" || owner == "portable" {
		return Channel{Owner: owner, CanApply: true, Instruction: "Select a trusted signed manifest and run update apply; settings, dependencies and custom Skills are preserved."}
	}
	text := instructions[owner]
	if text == "" {
		owner = "unknown"
		text = "Installation owner unknown. Use the original installation route; self-update is disabled."
	}
	return Channel{Owner: owner, Instruction: text}
}
