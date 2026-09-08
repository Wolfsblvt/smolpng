//go:build !windows

// SPDX-License-Identifier: MIT
package fsutil

import "os"

func publish(from, to string, overwrite bool) error {
	if overwrite {
		return os.Rename(from, to)
	}
	return os.Link(from, to)
}
