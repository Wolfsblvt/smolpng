// SPDX-License-Identifier: MIT
package cli

import (
	"path/filepath"
	"runtime"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/toolchain"
	"github.com/spf13/cobra"
)

// Managed downloads use the app's pinned catalog. No user-supplied URL or checksum
// argument exists, and merely viewing setup/status never starts a network request.
func (a *app) managedTools() (toolchain.Managed, error) {
	p, e := a.dirs()
	return toolchain.Managed{Root: filepath.Join(p.Data, "tools"), Runner: a.env.Runner, Client: a.env.DependencyHTTP}, e
}
func (a *app) depsCommand() *cobra.Command {
	root := &cobra.Command{Use: "deps", Short: "Inspect, configure and explicitly install separate optimizer programs"}
	root.AddCommand(&cobra.Command{Use: "status", Args: cobra.NoArgs, RunE: a.doctor}, &cobra.Command{Use: "doctor", Args: cobra.NoArgs, RunE: a.doctor})
	root.AddCommand(&cobra.Command{Use: "licenses", Args: cobra.NoArgs, Short: "Show separate optimizer source and license boundaries", RunE: func(c *cobra.Command, _ []string) error {
		oxi, e := toolchain.ArtifactFor("oxipng", runtime.GOOS, runtime.GOARCH)
		jpeg, jpegErr := toolchain.ArtifactFor("jpegoptim", runtime.GOOS, runtime.GOARCH)
		return a.emit(map[string]any{"schemaVersion": 1, "bundledExecutables": false, "pngquant": map[string]any{"license": "GPL-3.0-or-later with upstream notices OR commercial", "source": "https://pngquant.org/licensing.html", "managedDownload": false, "reason": "no qualified artifact/checksum selected"}, "oxipng": map[string]any{"license": "MIT", "source": "https://github.com/oxipng/oxipng/tree/v10.2.1", "managedArtifact": oxi, "available": e == nil, "reason": errorText(e)}, "jpegoptim": map[string]any{"license": "GPL-3.0-or-later", "source": "https://github.com/tjko/jpegoptim/tree/v1.5.6", "managedArtifact": jpeg, "available": jpegErr == nil, "reason": errorText(jpegErr)}})
	}})
	for _, operation := range []string{"install", "update"} {
		var apply bool
		command := &cobra.Command{Use: operation + " [oxipng|jpegoptim]", Args: cobra.MaximumNArgs(1), Short: "Preview a pinned optimizer archive; --apply downloads, verifies and activates it"}
		command.Flags().BoolVar(&apply, "apply", false, "Accept the displayed upstream source/license and install the pinned artifact")
		command.RunE = func(c *cobra.Command, args []string) error {
			name := "oxipng"
			if len(args) > 0 {
				name = args[0]
			}
			artifact, e := toolchain.ArtifactFor(name, runtime.GOOS, runtime.GOARCH)
			if e != nil {
				return domain.E(4, e.Error(), e)
			}
			m, e := a.managedTools()
			if e != nil {
				return e
			}
			before := m.Inspect(name)
			if !apply {
				return a.emit(map[string]any{"schemaVersion": 1, "apply": false, "operation": operation, "artifact": artifact, "destination": filepath.Join(m.Root, name), "current": before, "notice": "No download or PATH change until --apply. Exact binary version is probed before activation; native corpus qualification is separate."})
			}
			result, e := m.Install(c.Context(), artifact)
			if e != nil {
				if c.Context().Err() != nil {
					return domain.E(130, "dependency installation cancelled", c.Context().Err())
				}
				return domain.E(5, "dependency installation refused: "+e.Error(), e)
			}
			return a.emit(map[string]any{"schemaVersion": 1, "apply": true, "operation": operation, "artifact": artifact, "managed": result, "pathModified": false, "selection": "configured paths and compatible system tools still take precedence unless managed preference was explicitly selected"})
		}
		root.AddCommand(command)
	}
	var removeApply bool
	remove := &cobra.Command{Use: "remove TOOL", Args: cobra.ExactArgs(1), Short: "Remove an unchanged tracked optimizer installation, never system tools", RunE: func(c *cobra.Command, args []string) error {
		m, e := a.managedTools()
		if e != nil {
			return e
		}
		s := m.Inspect(args[0])
		if !removeApply {
			return a.emit(map[string]any{"schemaVersion": 1, "apply": false, "managed": s})
		}
		if e = m.Remove(c.Context(), args[0]); e != nil {
			return domain.E(5, "managed removal refused", e)
		}
		return a.emit(map[string]any{"schemaVersion": 1, "removed": args[0], "systemToolsChanged": false})
	}}
	remove.Flags().BoolVar(&removeApply, "apply", false, "Remove only recorded, unmodified managed versions")
	root.AddCommand(remove)
	configure := &cobra.Command{Use: "configure TOOL PATH", Args: cobra.ExactArgs(2), Short: "Select a local executable after a successful version probe", RunE: func(c *cobra.Command, args []string) error {
		if args[0] != "pngquant" && args[0] != "oxipng" && args[0] != "jpegoptim" {
			return domain.E(2, "choose pngquant, oxipng, or jpegoptim", nil)
		}
		dirs, e := a.dirs()
		if e != nil {
			return e
		}
		cwd, e := a.cwd()
		if e != nil {
			return e
		}
		path := args[1]
		if !filepath.IsAbs(path) {
			path = filepath.Join(cwd, path)
		}
		paths := toolchain.Paths{}
		if args[0] == "jpegoptim" {
			paths.JPEGOptim = path
		} else if args[0] == "pngquant" {
			paths.PNGQuant = path
		} else {
			paths.OxiPNG = path
		}
		tools := toolchain.Discover(c.Context(), a.env.Runner, paths, toolchain.Paths{})
		chosen := tools.PNGQuant
		if args[0] == "oxipng" {
			chosen = tools.OxiPNG
		}
		if args[0] == "jpegoptim" {
			chosen = tools.JPEGOptim
		}
		if !chosen.Available {
			return domain.E(4, "selected executable failed qualification: "+chosen.Error, nil)
		}
		w, e := a.loadConfig(dirs)
		if e != nil {
			return e
		}
		if e = w.SaveUserSettings(c.Context(), map[string]any{"tools." + args[0]: path}); e != nil {
			return domain.E(2, e.Error(), e)
		}
		return a.emit(map[string]any{"schemaVersion": 1, "configured": chosen, "pathModified": false})
	}}
	root.AddCommand(configure)
	for _, preference := range []string{"system", "managed"} {
		var apply bool
		cmd := &cobra.Command{Use: "use-" + preference, Args: cobra.NoArgs, Short: "Preview switching discovery preference and clearing saved path overrides", RunE: func(c *cobra.Command, _ []string) error {
			dirs, e := a.dirs()
			if e != nil {
				return e
			}
			w, e := a.loadConfig(dirs)
			if e != nil {
				return e
			}
			changes := map[string]any{"tools.preference": preference, "tools.pngquant": "", "tools.oxipng": "", "tools.jpegoptim": ""}
			if apply {
				if e = w.SaveUserSettings(c.Context(), changes); e != nil {
					return domain.E(2, e.Error(), e)
				}
			}
			return a.emit(map[string]any{"schemaVersion": 1, "apply": apply, "settings": changes, "notice": "Explicit invocation paths still win. Unavailable preferred tools fall back to the other local source. No executables or PATH entries are changed."})
		}}
		cmd.Flags().BoolVar(&apply, "apply", false, "Apply this discovery preference and clear saved executable paths")
		root.AddCommand(cmd)
	}
	return root
}
func errorText(e error) string {
	if e == nil {
		return ""
	}
	return e.Error()
}
