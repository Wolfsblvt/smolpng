// SPDX-License-Identifier: MIT
package fsutil

import "golang.org/x/sys/unix"

// MoveNoReplace atomically renames a file or directory only if the destination is
// absent. Unsupported filesystems fail rather than emulate a racy replacement.
func MoveNoReplace(from, to string) error {
	return unix.Renameat2(unix.AT_FDCWD, from, unix.AT_FDCWD, to, unix.RENAME_NOREPLACE)
}
