package fsutil

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

func TestAtomicReplacementAndCancellation(t *testing.T) {
	p := filepath.Join(t.TempDir(), "out")
	if e := AtomicWrite(context.Background(), p, []byte("old"), 0600, false); e != nil {
		t.Fatal(e)
	}
	if e := AtomicWrite(context.Background(), p, []byte("new"), 0600, false); !errors.Is(e, os.ErrExist) {
		t.Fatal("expected wrapped exists", e)
	}
	b, _ := os.ReadFile(p)
	if string(b) != "old" {
		t.Fatal("clobbered existing output")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e := AtomicWrite(ctx, p, []byte("new"), 0600, true); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if e := AtomicWrite(context.Background(), p, []byte("new"), 0600, true); e != nil {
		t.Fatal(e)
	}
}
func TestConcurrentNoClobberHasOneWinner(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "out")
	var wins atomic.Int32
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			e := AtomicWrite(context.Background(), p, []byte("complete"), 0600, false)
			if e == nil {
				wins.Add(1)
			} else if !errors.Is(e, os.ErrExist) {
				t.Error(e)
			}
		})
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatal(wins.Load())
	}
	files, _ := os.ReadDir(d)
	if len(files) != 1 {
		t.Fatal("temporary file leaked")
	}
}
