// SPDX-License-Identifier: MIT
package shellmenu

import (
	"fmt"
	"strings"
)

// DolphinFile emits a KDE service menu, not a universal Linux file-manager plug-in.
// quote is the shared freedesktop Exec encoder and appends %F independently.
func DolphinFile(s Spec, quote func([]string) string) ([]byte, error) {
	if e := s.Validate(); e != nil {
		return nil, e
	}
	if strings.Contains(s.Executable, "=") {
		return nil, fmt.Errorf("Dolphin executable path cannot contain '='")
	}
	var b strings.Builder
	b.WriteString("[Desktop Entry]\nType=Service\nMimeType=image/png;image/jpeg;\nIcon=image-x-generic\nX-KDE-Protocols=file\nX-KDE-MinNumberOfUrls=1\nX-KDE-MaxNumberOfUrls=512\nX-Smolpng-Managed=true\n")
	terminal := s.Actions[0].Interactive
	for _, a := range s.Actions {
		if a.Interactive != terminal {
			return nil, fmt.Errorf("Dolphin Terminal applies to the whole service, not individual actions; split the menu")
		}
	}
	fmt.Fprintf(&b, "Terminal=%t\n", terminal)
	if s.Mode == "submenu" {
		b.WriteString("X-KDE-Submenu=smolpng\nX-KDE-Priority=TopLevel\n")
	} else {
		b.WriteString("X-KDE-Priority=TopLevel\n")
	}
	b.WriteString("Actions=")
	for _, a := range s.Actions {
		b.WriteString(a.ID + ";")
	}
	b.WriteString("\n")
	for _, a := range s.Actions {
		args := []string{s.Executable, "_shell-run", "--action", a.ID, "--generation", s.Generation, "--"}
		fmt.Fprintf(&b, "\n[Desktop Action %s]\nName=%s\nIcon=image-x-generic\nExec=%s\n", a.ID, strings.ReplaceAll(a.Label, `\`, `\\`), quote(args))
	}
	return []byte(b.String()), nil
}

// Dolphin inherits Terminal from the service's main section, not each action.
// Two owned service files share the submenu so only the interactive action opens
// the desktop's selected terminal; background compression remains windowless.
func DolphinFiles(s Spec, quote func([]string) string) (map[string][]byte, error) {
	if e := s.Validate(); e != nil {
		return nil, e
	}
	out := map[string][]byte{}
	background, interactive := s, s
	background.Actions = nil
	interactive.Actions = nil
	for _, a := range s.Actions {
		if a.Interactive {
			interactive.Actions = append(interactive.Actions, a)
		} else {
			background.Actions = append(background.Actions, a)
		}
	}
	for id, spec := range map[string]Spec{"dolphin-menu": background, "dolphin-interactive": interactive} {
		if len(spec.Actions) == 0 {
			continue
		}
		b, e := DolphinFile(spec, quote)
		if e != nil {
			return nil, e
		}
		out[id] = b
	}
	return out, nil
}
