// SPDX-License-Identifier: MIT
package updates

import (
	"context"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// Launch copies the currently trusted executable as a short-lived helper. On
// Windows that helper can replace smolpng only after the invoking process exits.
// It is not the newly downloaded application or a persistent updater daemon.
func Launch(ctx context.Context, executable, stage, trustFile, stateFile string) (int, error) {
	j, e := ReadJournal(stage)
	if e != nil {
		return 0, e
	}
	if j.Phase != "staged" {
		return 0, fmt.Errorf("only a newly staged update can be scheduled")
	}
	b, e := ReadFile(executable, MaxFile)
	if e != nil {
		return 0, e
	}
	if SHA(b) != j.Previous[mainName(j.Platform)].SHA256 {
		return 0, fmt.Errorf("invoking executable changed since preparation; helper not started")
	}
	name := "worker"
	if filepath.Ext(executable) == ".exe" {
		name += ".exe"
	}
	worker := filepath.Join(stage, name)
	if e = fsutil.AtomicWrite(ctx, worker, b, 0700, false); e != nil {
		return 0, e
	}
	args := []string{"__update-worker", "--stage", stage, "--wait-pid", strconv.Itoa(os.Getpid())}
	if trustFile != "" {
		args = append(args, "--trust-key", trustFile)
	}
	log, e := os.OpenFile(filepath.Join(stage, "worker.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return 0, e
	}
	defer log.Close()
	c := exec.Command(worker, args...)
	configureWorker(c)
	c.Stdin = nil
	c.Stdout = log
	c.Stderr = log
	c.Env = cleanEnv()
	if e = c.Start(); e != nil {
		return 0, e
	}
	pid := c.Process.Pid
	_ = c.Process.Release()
	// This records scheduling, not completion. The stage journal is authoritative.
	if stateFile != "" {
		if e = fsutil.AtomicWrite(ctx, stateFile, []byte(stage+"\n"), 0600, true); e != nil {
			return pid, fmt.Errorf("worker scheduled but state pointer write failed; journal is %s", filepath.Join(stage, "journal.json"))
		}
	}
	return pid, nil
}
func Wait(ctx context.Context, pid int) error {
	if pid < 1 || pid == os.Getpid() {
		return fmt.Errorf("invalid parent process")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	return waitProcess(ctx, pid)
}
