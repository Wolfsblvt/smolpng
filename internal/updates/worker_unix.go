//go:build !windows

// SPDX-License-Identifier: MIT
package updates

import (
	"context"
	"errors"
	"os/exec"
	"syscall"
	"time"
)

func configureWorker(c *exec.Cmd) { c.SysProcAttr = &syscall.SysProcAttr{Setsid: true} }
func waitProcess(ctx context.Context, pid int) error {
	timer := time.NewTicker(100 * time.Millisecond)
	defer timer.Stop()
	for {
		e := syscall.Kill(pid, 0)
		if errors.Is(e, syscall.ESRCH) {
			return nil
		}
		if e != nil {
			return e
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
}
