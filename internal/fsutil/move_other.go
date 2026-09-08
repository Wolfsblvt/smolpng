//go:build !windows && !linux && !darwin

// SPDX-License-Identifier: MIT
package fsutil

import "fmt"

func MoveNoReplace(from, to string) error {
	return fmt.Errorf("atomic no-replace directory move is not supported on this platform")
}
