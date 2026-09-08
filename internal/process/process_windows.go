// SPDX-License-Identifier: MIT
package process

import (
	"os/exec"
	"syscall"
)

func configure(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
