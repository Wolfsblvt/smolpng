// SPDX-License-Identifier: MIT
// Package process invokes executables directly, never through a shell.
package process

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Invocation struct {
	Executable     string
	Args           []string
	Stdin          io.Reader
	Stdout, Stderr io.Writer
}
type Result struct {
	ExitCode       int
	Stdout, Stderr string
}
type Runner interface {
	Run(context.Context, Invocation) (Result, error)
}
type OSRunner struct{}
type capped struct{ bytes.Buffer }

// CommandContext creates a direct child process without opening a Windows console.
// Callers retain exec.Cmd's environment, working-directory and I/O controls.
func CommandContext(ctx context.Context, executable string, args ...string) *exec.Cmd {
	c := exec.CommandContext(ctx, executable, args...)
	configure(c)
	return c
}

func (b *capped) Write(p []byte) (int, error) {
	n := len(p)
	room := (64 << 10) - b.Len()
	if room > 0 {
		_, _ = b.Buffer.Write(p[:min(n, room)])
	}
	return n, nil
}
func (OSRunner) Run(ctx context.Context, i Invocation) (Result, error) {
	c := CommandContext(ctx, i.Executable, i.Args...)
	c.WaitDelay = 2 * time.Second
	c.Stdin = i.Stdin
	// A local optimizer has no reason to inherit the optional cloud API key.
	for _, e := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(e), "TINIFY_API_KEY=") {
			c.Env = append(c.Env, e)
		}
	}
	var out, er capped
	c.Stdout = &out
	c.Stderr = &er
	if i.Stdout != nil {
		c.Stdout = i.Stdout
	}
	if i.Stderr != nil {
		c.Stderr = i.Stderr
	}
	e := c.Run()
	r := Result{Stdout: out.String(), Stderr: er.String()}
	if e != nil {
		r.ExitCode = -1
		var ex *exec.ExitError
		if errors.As(e, &ex) {
			r.ExitCode = ex.ExitCode()
			e = nil
		}
	}
	if ctx.Err() != nil {
		return r, ctx.Err()
	}
	return r, e
}
