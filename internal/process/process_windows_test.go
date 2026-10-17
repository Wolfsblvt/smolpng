// SPDX-License-Identifier: MIT
package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func TestConsoleHandleChild(t *testing.T) {
	mode := os.Getenv("SMOLPNG_TEST_CONSOLE_HANDLE")
	if mode == "" {
		return
	}
	if mode == "descendant" {
		child := exec.Command(os.Args[0], "-test.run=^TestConsoleHandleChild$")
		child.Env = append(os.Environ(), "SMOLPNG_TEST_CONSOLE_HANDLE=direct")
		// Hide a newly allocated console if inheritance regresses: the test must
		// never flash a visible window merely to discover the wrong allocation.
		child.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		output, err := child.CombinedOutput()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Fprint(os.Stdout, string(output))
		os.Exit(0)
	}
	handle, _, _ := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetConsoleWindow").Call()
	fmt.Fprint(os.Stdout, handle)
	os.Exit(0)
}

func TestDirectChildHasNoWindowsConsole(t *testing.T) {
	for _, mode := range []string{"direct", "descendant"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cmd := CommandContext(ctx, os.Args[0], "-test.run=^TestConsoleHandleChild$")
			cmd.Env = append(os.Environ(), "SMOLPNG_TEST_CONSOLE_HANDLE="+mode)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("console probe: %v; output: %s", err, output)
			}
			if handle := strings.TrimSpace(string(output)); handle != "0" {
				t.Fatalf("child owns console window %q", handle)
			}
		})
	}
}
