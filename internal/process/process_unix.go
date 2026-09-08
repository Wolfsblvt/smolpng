//go:build !windows

// SPDX-License-Identifier: MIT
package process

import "os/exec"

func configure(*exec.Cmd) {}
