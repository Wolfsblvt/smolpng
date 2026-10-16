package toolchain

import (
	"context"
	"errors"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

type fakeVersion struct{}

func (fakeVersion) Run(context.Context, process.Invocation) (process.Result, error) {
	return process.Result{Stdout: "pngquant 3.0.3"}, nil
}

type failedVersion struct {
	result process.Result
	err    error
}

type delayedVersion struct{ delay time.Duration }

func (f delayedVersion) Run(ctx context.Context, _ process.Invocation) (process.Result, error) {
	select {
	case <-time.After(f.delay):
		return process.Result{Stdout: "oxipng 10.2.1"}, nil
	case <-ctx.Done():
		return process.Result{}, ctx.Err()
	}
}

func TestSlowCompatibleVersionProbeRemainsAvailable(t *testing.T) {
	p := filepath.Join(t.TempDir(), "optimizer")
	if err := os.WriteFile(p, nil, 0700); err != nil {
		t.Fatal(err)
	}
	synctest.Test(t, func(t *testing.T) {
		tool := probe(context.Background(), delayedVersion{5 * time.Second}, "oxipng", p, "")
		if !tool.Available || tool.Version != "10.2.1" {
			t.Fatalf("slow compatible optimizer was refused: %+v", tool)
		}
	})
}

func TestVersionProbeHonorsCallerBudgetAndBoundsHungTools(t *testing.T) {
	p := filepath.Join(t.TempDir(), "optimizer")
	if err := os.WriteFile(p, nil, 0700); err != nil {
		t.Fatal(err)
	}
	for _, callerBudget := range []time.Duration{time.Second, time.Hour} {
		t.Run(callerBudget.String(), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), callerBudget)
				defer cancel()
				started := time.Now()
				tool := probe(ctx, delayedVersion{time.Hour}, "oxipng", p, "")
				if tool.Available || !strings.Contains(tool.Error, "context deadline exceeded") {
					t.Fatalf("hung probe did not preserve its deadline: %+v", tool)
				}
				if elapsed := time.Since(started); elapsed > min(callerBudget, 30*time.Second) {
					t.Fatalf("probe exceeded its bounded wait: %s", elapsed)
				}
			})
		})
	}
}

func (f failedVersion) Run(context.Context, process.Invocation) (process.Result, error) {
	return f.result, f.err
}

func TestJPEGOptimVersionBanner(t *testing.T) {
	p := filepath.Join(t.TempDir(), "jpegoptim")
	if err := os.WriteFile(p, nil, 0700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		banner    string
		version   string
		available bool
	}{
		{"jpegoptim v1.5.6  Win64 (Sep 15 2025)\nlibjpeg version: 6b  27-Mar-1998", "1.5.6", true},
		{"jpegoptim 1.4.7", "1.4.7", true},
		{"jpegoptim v2.0.1", "2.0.1", false},
	} {
		t.Run(tc.version, func(t *testing.T) {
			runner := failedVersion{result: process.Result{Stdout: tc.banner}}
			tool := probe(context.Background(), runner, "jpegoptim", p, "")
			if tool.Version != tc.version || tool.Available != tc.available {
				t.Fatalf("wrong version or compatibility from banner %q: %+v", tc.banner, tool)
			}
		})
	}
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
