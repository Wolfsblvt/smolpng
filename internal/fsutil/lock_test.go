package fsutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOSLockCannotBeHeldTwiceAndCloseReleases(t *testing.T) {
	p := filepath.Join(t.TempDir(), "operation.lock")
	a, e := TryLock(p)
	if e != nil {
		t.Fatal(e)
	}
	if b, e := TryLock(p); e == nil {
		b.Close()
		a.Close()
		t.Fatal("two independent writers held the registry lock")
	}
	a.Close()
	b, e := TryLock(p)
	if e != nil {
		t.Fatal("closed lock did not release", e)
	}
	b.Close()
}
func TestDirectoryMoveNeverReplacesExistingDirectory(t *testing.T) {
	root := t.TempDir()
	from, to := filepath.Join(root, "from"), filepath.Join(root, "to")
	os.Mkdir(from, 0700)
	os.Mkdir(to, 0700)
	os.WriteFile(filepath.Join(from, "source"), []byte("complete"), 0600)
	if e := MoveNoReplace(from, to); e == nil {
		t.Fatal("existing destination replaced")
	}
	if _, e := os.Stat(filepath.Join(from, "source")); e != nil {
		t.Fatal(e)
	}
	os.Remove(to)
	if e := MoveNoReplace(from, to); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(from); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("source still exists after commit")
	}
}
