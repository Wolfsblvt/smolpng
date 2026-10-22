// SPDX-License-Identifier: MIT
package integration

import (
	"os"
	"path/filepath"
)

func NativeLocations() (Locations, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Locations{}, err
	}
	return Locations{Menu: filepath.Join(home, "Applications"), Desktop: filepath.Join(home, "Desktop")}, nil
}
