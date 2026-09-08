// SPDX-License-Identifier: MIT
package integration

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf16"
)

type Renderer interface {
	Render(context.Context, Entry) ([]byte, error)
}
type Native struct{ Runner process.Runner }

func PowerShellScript(x Entry, dest string) string {
	q := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	return "$ErrorActionPreference='Stop'; $w=New-Object -ComObject WScript.Shell; $s=$w.CreateShortcut(" + q(dest) + "); $s.TargetPath=" + q(x.Executable) + "; $s.Arguments=" + q(WindowsArguments(x.Arguments)) + "; $s.WorkingDirectory=" + q(x.WorkingDirectory) + "; $s.Description=" + q(x.Name) + "; $s.WindowStyle=1; $s.Save(); if(-not (Test-Path -LiteralPath " + q(dest) + ")){throw 'Shortcut not created'}"
}
func (n Native) Render(ctx context.Context, x Entry) ([]byte, error) {
	if runtime.GOOS == "linux" {
		return DesktopFile(x)
	}
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("native shortcut rendering unavailable on %s", runtime.GOOS)
	}
	dir, e := os.MkdirTemp("", "smolpng-link-*")
	if e != nil {
		return nil, e
	}
	defer os.RemoveAll(dir)
	dest := filepath.Join(dir, "candidate.lnk")
	script := PowerShellScript(x, dest)
	words := utf16.Encode([]rune(script))
	b := make([]byte, len(words)*2)
	for i, w := range words {
		binary.LittleEndian.PutUint16(b[i*2:], w)
	}
	runner := n.Runner
	if runner == nil {
		runner = process.OSRunner{}
	}
	system := os.Getenv("SystemRoot")
	if system == "" {
		return nil, fmt.Errorf("SystemRoot unavailable")
	}
	result, e := runner.Run(ctx, process.Invocation{Executable: filepath.Join(system, "System32", "WindowsPowerShell", "v1.0", "powershell.exe"), Args: []string{"-NoProfile", "-NonInteractive", "-EncodedCommand", base64.StdEncoding.EncodeToString(b)}})
	if e != nil || result.ExitCode != 0 {
		return nil, fmt.Errorf("Windows shortcut creation failed (exit %d); no integration target changed", result.ExitCode)
	}
	b, e = os.ReadFile(dest)
	if e != nil || len(b) < 76 || len(b) > 1<<20 {
		return nil, fmt.Errorf("Windows shortcut was not created as expected")
	}
	return b, nil
}
