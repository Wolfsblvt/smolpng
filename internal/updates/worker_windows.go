// SPDX-License-Identifier: MIT
package updates

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"os/exec"
	"syscall"
)

func configureWorker(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000008}
}
func waitProcess(ctx context.Context, pid int) error {
	h, e := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if errors.Is(e, windows.ERROR_INVALID_PARAMETER) {
		return nil
	}
	if e != nil {
		return e
	}
	defer windows.CloseHandle(h)
	for {
		if e := ctx.Err(); e != nil {
			return e
		}
		r, e := windows.WaitForSingleObject(h, 100)
		if e != nil {
			return e
		}
		if r == windows.WAIT_OBJECT_0 {
			return nil
		}
		if r != uint32(windows.WAIT_TIMEOUT) {
			return fmt.Errorf("unexpected process wait result")
		}
	}
}
