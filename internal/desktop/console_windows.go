// SPDX-License-Identifier: MIT
package desktop

import (
	"fmt"
	"golang.org/x/sys/windows"
	"os"
)

func OpenConsole() (Console, error) {
	kernel := windows.NewLazySystemDLL("kernel32.dll")
	window, _, _ := kernel.NewProc("GetConsoleWindow").Call()
	allocated := false
	if window == 0 {
		ok, _, err := kernel.NewProc("AllocConsole").Call()
		if ok == 0 {
			return Console{}, fmt.Errorf("cannot open an interactive console: %w", err)
		}
		allocated = true
	}
	release := func() {
		if allocated {
			kernel.NewProc("FreeConsole").Call()
		}
	}
	in, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
	if err != nil {
		release()
		return Console{}, err
	}
	out, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	if err != nil {
		in.Close()
		release()
		return Console{}, err
	}
	return Console{In: in, Out: out, Err: out, Close: func() { in.Close(); out.Close(); release() }}, nil
}
