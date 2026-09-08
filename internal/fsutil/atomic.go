// SPDX-License-Identifier: MIT
package fsutil

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// AtomicWrite never deletes the old target before a replacement is complete.
func AtomicWrite(ctx context.Context, path string, data []byte, mode os.FileMode, overwrite bool) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(path), ".smolpng-*")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if e = f.Chmod(mode.Perm()); e == nil {
		var n int
		n, e = f.Write(data)
		if e == nil && n != len(data) {
			e = io.ErrShortWrite
		}
	}
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		return e
	}
	if e = ctx.Err(); e != nil {
		return e
	}
	if e = publish(f.Name(), path, overwrite); e != nil {
		return fmt.Errorf("publish output: %w", e)
	}
	return nil
}
