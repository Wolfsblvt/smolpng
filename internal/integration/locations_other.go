//go:build !windows

// SPDX-License-Identifier: MIT
package integration

import (
	"github.com/adrg/xdg"
	"path/filepath"
)

func NativeLocations() (Locations, error) {
	return Locations{Menu: filepath.Join(xdg.DataHome, "applications"), Desktop: xdg.UserDirs.Desktop}, nil
}
