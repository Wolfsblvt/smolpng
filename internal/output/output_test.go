package output

import (
	"github.com/Wolfsblvt/smolpng/internal/input"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNamesMirroringAndCollisions(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "icon.png")
	_ = os.WriteFile(p, []byte("source"), 0600)
	_ = os.WriteFile(filepath.Join(d, "icon.smol.png"), []byte("prior"), 0600)
	s := input.Source{Kind: "file", Name: "icon.png", Path: p, Display: p, Relative: "nested/icon.png"}
	js, e := Plan([]input.Source{s}, Options{}, time.Time{})
	if e != nil || js[0].Output != filepath.Join(d, "icon.smol-2.png") {
		t.Fatal(js, e)
	}
	js, e = Plan([]input.Source{s}, Options{Directory: filepath.Join(d, "out")}, time.Time{})
	if e != nil || js[0].Output != filepath.Join(d, "out", "nested", "icon.smol.png") {
		t.Fatal(js, e)
	}
	if _, e = Plan([]input.Source{s}, Options{Path: p, Overwrite: true}, time.Time{}); e == nil {
		t.Fatal("original alias overwritten")
	}
}
func TestNameSafety(t *testing.T) {
	for _, s := range []string{"../{stem}.png", "{wrong}.png", "{stem}.jpg"} {
		if _, e := Name("icon.png", s, time.Time{}); e == nil {
			t.Fatal(s)
		}
	}
	n, e := Name("icon.smol-2.png", "", time.Time{})
	if e != nil || n != "icon.smol.png" {
		t.Fatal(n, e)
	}
}
