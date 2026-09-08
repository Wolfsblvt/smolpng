// SPDX-License-Identifier: MIT
package fsutil

import "golang.org/x/sys/windows"

func publish(from, to string, overwrite bool) error {
	a, e := windows.UTF16PtrFromString(from)
	if e != nil {
		return e
	}
	b, e := windows.UTF16PtrFromString(to)
	if e != nil {
		return e
	}
	f := uint32(windows.MOVEFILE_WRITE_THROUGH)
	if overwrite {
		f |= windows.MOVEFILE_REPLACE_EXISTING
	}
	return windows.MoveFileEx(a, b, f)
}
