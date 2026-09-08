// SPDX-License-Identifier: MIT
package fsutil

import (
	"golang.org/x/sys/windows"
	"os"
)

func TryLock(path string) (*os.File, error) {
	f, e := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return nil, e
	}
	var overlap windows.Overlapped
	e = windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlap)
	if e != nil {
		f.Close()
		return nil, e
	}
	return f, nil
}
