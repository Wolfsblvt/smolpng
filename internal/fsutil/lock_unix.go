//go:build !windows

// SPDX-License-Identifier: MIT
package fsutil

import (
	"golang.org/x/sys/unix"
	"os"
)

// TryLock holds an advisory OS lock until the returned file is closed. The lock
// file deliberately remains: unlinking it can create two independent locks.
func TryLock(path string) (*os.File, error) {
	f, e := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return nil, e
	}
	if e = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); e != nil {
		f.Close()
		return nil, e
	}
	return f, nil
}
