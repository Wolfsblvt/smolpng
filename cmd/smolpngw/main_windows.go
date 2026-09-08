// SPDX-License-Identifier: MIT
// smolpngw is the explicitly background Windows drop entry point. Build with
// -H=windowsgui to suppress console allocation; the ordinary CLI is unchanged.
package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/cli"
	"github.com/Wolfsblvt/smolpng/internal/desktop"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	args := os.Args[1:]
	if len(args) == 0 || args[0] != "_shell-run" {
		args = append([]string{"drop"}, args...)
	}
	code := cli.Run(ctx, args, cli.Environment{OpenTerminal: desktop.OpenConsole})
	if code != 0 {
		notice, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		_ = (desktop.NativeActions{}).Notify(notice, "smolpng could not finish", "Use smolpng last or smolpng doctor in a terminal for details.")
		cancel()
	}
	stop()
	os.Exit(code)
}
