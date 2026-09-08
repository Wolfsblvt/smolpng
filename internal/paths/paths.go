// SPDX-License-Identifier: MIT
package paths

import (
	"github.com/adrg/xdg"
	"os"
	"path/filepath"
	"runtime"
)

type Paths struct {
	Config    string `json:"config"`
	Data      string `json:"data"`
	State     string `json:"state"`
	Cache     string `json:"cache"`
	Downloads string `json:"downloads"`
	Portable  bool   `json:"portable"`
}

func Resolve(portable bool, explicit string) (Paths, error) {
	exe, e := os.Executable()
	if e != nil {
		return Paths{}, e
	}
	if _, e := os.Stat(filepath.Join(filepath.Dir(exe), "smolpng.portable")); e == nil {
		portable = true
	}
	root := explicit
	if root == "" {
		root = os.Getenv("SMOLPNG_HOME")
	}
	p := Paths{Downloads: xdg.UserDirs.Download}
	if root != "" || portable {
		if root == "" {
			root = filepath.Dir(exe)
		}
		root, e = filepath.Abs(root)
		if e != nil {
			return p, e
		}
		p.Config = filepath.Join(root, "config.toml")
		p.Data = root
		p.State = filepath.Join(root, "state")
		p.Cache = filepath.Join(root, "cache")
		p.Portable = true
	} else if runtime.GOOS == "windows" {
		root = filepath.Join(xdg.DataHome, "smolpng")
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			root = filepath.Join(local, "smolpng")
		}
		p.Config = filepath.Join(root, "config.toml")
		p.Data = root
		p.State = root
		p.Cache = filepath.Join(root, "cache")
	} else {
		p.Config = filepath.Join(xdg.ConfigHome, "smolpng", "config.toml")
		p.Data = filepath.Join(xdg.DataHome, "smolpng")
		p.State = filepath.Join(xdg.StateHome, "smolpng")
		p.Cache = filepath.Join(xdg.CacheHome, "smolpng")
	}
	if c := os.Getenv("SMOLPNG_CONFIG"); c != "" {
		p.Config, e = filepath.Abs(c)
	}
	return p, e
}
func (p Paths) Detached() string {
	if p.Downloads == "" {
		return ""
	}
	return filepath.Join(p.Downloads, "smolpng")
}
