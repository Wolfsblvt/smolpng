// SPDX-License-Identifier: MIT
package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/integration"
	"github.com/Wolfsblvt/smolpng/internal/process"
)

// A background Linux entry can select the interactive workflow through config.
// Reuse the desktop's Terminal=true launch instead of guessing emulator flags or
// opening a terminal for ordinary background work. GIO success acknowledges the
// native launch; the child workflow owns its result and last-result recovery.
func (a *app) launchLinuxInteractiveDrop(ctx context.Context) error {
	gio, err := exec.LookPath("gio")
	if err != nil {
		return domain.E(5, "interactive Linux drop needs GLib's gio and a configured terminal", err)
	}
	executable, err := a.applicationExecutable()
	if err != nil {
		return err
	}
	directory, err := a.cwd()
	if err != nil {
		return err
	}
	data, err := integration.DesktopFile(integration.Entry{Name: "smolpng", Executable: executable, Arguments: a.invocation, WorkingDirectory: directory, Terminal: true})
	if err != nil {
		return domain.E(5, "cannot prepare interactive Linux drop", err)
	}
	f, err := os.CreateTemp("", "smolpng-interactive-*.desktop")
	if err != nil {
		return domain.E(5, "cannot prepare interactive Linux drop", err)
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			fmt.Fprintln(a.env.Err, "Warning: cannot remove temporary interactive entry:", err)
		}
	}()
	if _, err = f.Write(data); err != nil {
		_ = f.Close()
		return domain.E(5, "cannot write interactive Linux drop entry", err)
	}
	if err = f.Close(); err != nil {
		return domain.E(5, "cannot close interactive Linux drop entry", err)
	}
	// Keep the application's inherited environment, including an explicitly
	// selected provider credential. OSRunner intentionally strips that key for
	// optimizers, so this application handoff uses the direct process primitive.
	cmd := process.CommandContext(ctx, gio, "launch", f.Name())
	cmd.Stdout, cmd.Stderr = a.env.Out, a.env.Err
	if err = cmd.Run(); err != nil {
		return domain.E(5, "cannot open interactive Linux drop; check the desktop's terminal selection", err)
	}
	return nil
}
