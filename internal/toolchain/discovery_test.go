package toolchain

import (
	"context"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"os"
	"path/filepath"
	"testing"
)

type fakeVersion struct{}

func (fakeVersion) Run(context.Context, process.Invocation) (process.Result, error) {
	return process.Result{Stdout: "pngquant 3.0.3"}, nil
}
func TestExplicitPathAndBrokenOverrideNeverFallsBack(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "pngquant")
	_ = os.WriteFile(p, nil, 0700)
	s := Discover(context.Background(), fakeVersion{}, Paths{PNGQuant: p}, Paths{})
	if !s.PNGQuant.Available || s.PNGQuant.Source != "invocation" {
		t.Fatal(s)
	}
	s = Discover(context.Background(), fakeVersion{}, Paths{PNGQuant: filepath.Join(d, "missing")}, Paths{PNGQuant: p})
	if s.PNGQuant.Available || s.PNGQuant.Source != "invocation" {
		t.Fatal("fallback ignored explicit broken path", s)
	}
}
