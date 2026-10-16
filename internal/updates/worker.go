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
	return startWorker(ctx, worker, stage, trustFile, stateFile, false)
}

// LaunchRollback schedules guarded recovery from a verified copy of the invoking
// executable, leaving its installed destination unlocked after the caller exits.
// The returned PID/log identify scheduling; the journal establishes completion.
func (m Manager) LaunchRollback(ctx context.Context, stage, trustFile, stateFile string) (int, string, error) {
	stage, e := filepath.Abs(stage)
	if e != nil {
		return 0, "", e
	}
	worker, e := m.prepareRollbackWorker(ctx, stage)
	if e != nil {
		return 0, "", e
	}
	pid, e := startWorker(ctx, worker, stage, trustFile, stateFile, true)
	if pid == 0 {
		_ = os.RemoveAll(filepath.Dir(worker))
	}
	return pid, filepath.Join(filepath.Dir(worker), "worker.log"), e
}

func (m Manager) prepareRollbackWorker(ctx context.Context, stage string) (string, error) {
	j, _, asset, e := m.load(stage, true)
	if e != nil {
		return "", e
	}
	if e = checkRollbackFiles(j, asset); e != nil {
		return "", e
	}
	b, e := ReadFile(m.Executable, MaxFile)
	if e != nil {
		return "", e
	}
	hash := SHA(b)
	trusted := hash == j.Previous[mainName(j.Platform)].SHA256
	for _, file := range asset.Files {
		if file.Name == mainName(j.Platform) && file.SHA256 == hash {
			trusted = true
		}
	}
	if !trusted {
		return "", fmt.Errorf("invoking executable does not match this update's verified old or new application")
	}
	// A prior apply/recovery worker and its evidence may still exist. Each explicit
	// recovery owns a fresh directory rather than overwriting that running image.
	directory, e := os.MkdirTemp(stage, "rollback-worker-")
	if e != nil {
		return "", e
	}
	worker := filepath.Join(directory, filepath.Base(m.Executable))
	if e = fsutil.AtomicWrite(ctx, worker, b, 0700, false); e != nil {
		_ = os.RemoveAll(directory)
		return "", e
	}
	return worker, nil
}

func startWorker(ctx context.Context, worker, stage, trustFile, stateFile string, rollback bool) (int, error) {
	args := []string{"__update-worker", "--stage", stage, "--wait-pid", strconv.Itoa(os.Getpid())}
	if rollback {
		args = append(args, "--rollback")
	}
	if trustFile != "" {
		args = append(args, "--trust-key", trustFile)
	}
	log, e := os.OpenFile(filepath.Join(filepath.Dir(worker), "worker.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
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
