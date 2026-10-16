package toolchain

import (
	"context"
	"errors"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeVersion struct{}

func (fakeVersion) Run(context.Context, process.Invocation) (process.Result, error) {
	return process.Result{Stdout: "pngquant 3.0.3"}, nil
}

type failedVersion struct {
	result process.Result
	err    error
}

func (f failedVersion) Run(context.Context, process.Invocation) (process.Result, error) {
	return f.result, f.err
}

func TestVersionProbePreservesFailureReason(t *testing.T) {
	p := filepath.Join(t.TempDir(), "optimizer")
	if err := os.WriteFile(p, nil, 0700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		runner failedVersion
		want   string
	}{
		{"deadline", failedVersion{err: context.DeadlineExceeded}, "context deadline exceeded"},
		{"cancelled", failedVersion{err: context.Canceled}, "context canceled"},
		{"launch", failedVersion{err: errors.New("access is denied")}, "access is denied"},
		{"exit", failedVersion{result: process.Result{ExitCode: 9}}, "exit code 9"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tool := probe(context.Background(), tc.runner, "oxipng", p, "")
			if tool.Available || !strings.Contains(tool.Error, tc.want) {
				t.Fatalf("probe lost failure reason: %+v", tool)
			}
		})
	}
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
