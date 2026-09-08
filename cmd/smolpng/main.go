// SPDX-License-Identifier: MIT
package main

import (
	"context"
	"github.com/Wolfsblvt/smolpng/internal/cli"
	"os"
	"os/signal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := cli.Run(ctx, os.Args[1:], cli.Environment{})
	stop()
	os.Exit(code)
}
