// SPDX-License-Identifier: MIT
package fsutil

import "golang.org/x/sys/unix"

func MoveNoReplace(from, to string) error { return unix.RenamexNp(from, to, unix.RENAME_EXCL) }
