package domain

import (
	"context"
	"fmt"
	"testing"
)

func TestExactSizeUnits(t *testing.T) {
	for s, n := range map[string]int64{"10000": 10000, "10KB": 10000, "10KiB": 10240, "1.5MiB": 1572864, "0.125KiB": 128, "9223372036854775807B": 9223372036854775807} {
		got, e := ParseSize(s)
		if e != nil || got != n {
			t.Fatalf("%s: %d %v", s, got, e)
		}
	}
	for _, s := range []string{"", "0", "-1", "1.1B", "NaN", "1e3", "10K", "9223372036854775808"} {
		if _, e := ParseSize(s); e == nil {
			t.Fatalf("accepted %q", s)
		}
	}
}
func TestRecipeValidation(t *testing.T) {
	for _, name := range []string{"auto", "web", "smallest", "lossless"} {
		r, e := Builtin(name)
		if e != nil || r.Validate() != nil {
			t.Fatal(name, e)
		}
	}
	r, _ := Builtin("lossless")
	for _, mutate := range []func(*Recipe){func(r *Recipe) { r.OptimizeAlpha = true }, func(r *Recipe) { r.Resize = Dimensions{10, 10} }, func(r *Recipe) { r.RemotePolicy = "fallback" }} {
		v := r
		mutate(&v)
		if v.Validate() == nil {
			t.Fatal("accepted destructive lossless override")
		}
	}
}
func TestStructuredErrorsAndPartialFailure(t *testing.T) {
	if Code(fmt.Errorf("wrapped: %w", E(6, "miss", nil))) != 6 || Code(context.Canceled) != 130 {
		t.Fatal("lost exit code")
	}
	a := NewResult("a", "auto")
	a.Status = "optimized"
	b := NewResult("b", "auto")
	b.Fail(E(3, "missing", nil))
	if Summarize([]Result{a, b}).ExitCode != 8 {
		t.Fatal("partial failure")
	}
	b.Fail(context.Canceled)
	if Summarize([]Result{a, b}).ExitCode != 130 {
		t.Fatal("lost cancellation")
	}
}
