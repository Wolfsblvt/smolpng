package workflow

import (
	"bytes"
	"context"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/input"
	"github.com/Wolfsblvt/smolpng/internal/output"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestUnavailableOptimizerReasonsReachFailedImageReport(t *testing.T) {
	for _, preset := range []string{"auto", "lossless"} {
		t.Run(preset, func(t *testing.T) {
			tools := testutil.Tools()
			tools.OxiPNG.Available = false
			tools.OxiPNG.Error = "version probe failed: context deadline exceeded"
			tools.PNGQuant.Available = false
			tools.PNGQuant.Error = "version probe failed: exit code 9"
			recipe, _ := domain.Builtin(preset)
			dir := t.TempDir()
			path := filepath.Join(dir, "image.png")
			if err := os.WriteFile(path, testutil.Corpus()["flat-icon.png"], 0600); err != nil {
				t.Fatal(err)
			}
			jobs, err := output.Plan([]input.Source{{Kind: "file", Path: path, Display: path, Name: "image.png"}}, output.Options{}, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			batch := Batch{Engine: Engine{Tools: tools, Limits: domain.DefaultLimits()}}
			report := batch.Run(context.Background(), jobs, Options{Recipe: recipe})
			if report.ExitCode != 4 || len(report.Results) != 1 {
				t.Fatal(report)
			}
			result := report.Results[0]
			warnings := strings.Join(result.Warnings, "; ")
			if result.Output != "" || !strings.Contains(warnings, "oxipng: "+tools.OxiPNG.Error) {
				t.Fatal(result)
			}
			if preset == "auto" && !strings.Contains(warnings, "pngquant: "+tools.PNGQuant.Error) {
				t.Fatal(result)
			}
		})
	}
}

func TestAutoUsesOriginalForEveryQuantization(t *testing.T) {
	r, _ := domain.Builtin("auto")
	data := testutil.Corpus()["flat-icon.png"]
	runner := &testutil.Runner{}
	cache := t.TempDir()
	e := Engine{Runner: runner, Tools: testutil.Tools(), Limits: domain.DefaultLimits(), CacheDir: cache}
	out, er := e.Optimize(context.Background(), data, r)
	if er != nil || len(out.Candidate.Data) >= len(data) {
		t.Fatal(out, er)
	}
	calls := 0
	for _, c := range runner.Calls() {
		if c.Invocation.Executable == "pngquant" {
			calls++
			if !bytes.Equal(data, c.Input) {
				t.Fatal("lossy generation chaining")
			}
		}
	}
	if calls < 2 {
		t.Fatal("Auto alternatives not exercised")
	}
	files, _ := os.ReadDir(cache)
	if len(files) > 0 {
		t.Fatal("temporary output leak")
	}
}
func TestStrictLosslessAcrossCorpus(t *testing.T) {
	r, _ := domain.Builtin("lossless")
	e := Engine{Runner: &testutil.Runner{}, Tools: testutil.Tools(), Limits: domain.DefaultLimits()}
	for name, b := range testutil.Corpus() {
		t.Run(name, func(t *testing.T) {
			out, err := e.Optimize(context.Background(), b, r)
			if err != nil {
				t.Fatal(err)
			}
			a, _ := pngimage.Decode(b, 1000000)
			c, err := pngimage.Decode(out.Candidate.Data, 1000000)
			if err != nil || !pngimage.PixelEqual(a.Image, c.Image) {
				t.Fatal("lost pixels", err)
			}
		})
	}
}
func TestTargetMetMissAndNearestOutput(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "a.png")
	data := testutil.Corpus()["flat-icon.png"]
	_ = os.WriteFile(p, data, 0600)
	src := input.Source{Kind: "file", Path: p, Display: p, Name: "a.png"}
	r, _ := domain.Builtin("auto")
	e := Engine{Runner: &testutil.Runner{}, Tools: testutil.Tools(), Limits: domain.DefaultLimits()}
	r.TargetBytes = 512
	out, err := e.Optimize(context.Background(), data, r)
	if err != nil || len(out.Candidate.Data) > 512 {
		t.Fatal(err)
	}
	r.TargetBytes = 1
	js, _ := output.Plan([]input.Source{src}, output.Options{}, time.Now())
	b := Batch{Engine: e}
	report := b.Run(context.Background(), js, Options{Recipe: r})
	if report.ExitCode != 6 || report.Results[0].Output != "" {
		t.Fatal(report)
	}
	r.AcceptNearest = true
	report = b.Run(context.Background(), js, Options{Recipe: r})
	if report.ExitCode != 6 || report.Results[0].Output == "" {
		t.Fatal(report)
	}
}
func TestQualityRejectionNotHardFailure(t *testing.T) {
	runner := &testutil.Runner{Handler: func(context.Context, process.Invocation) (process.Result, error, bool) {
		return process.Result{ExitCode: 99}, nil, true
	}}
	ts := testutil.Tools()
	ts.OxiPNG.Available = false
	r, _ := domain.Builtin("auto")
	data := testutil.Corpus()["flat-icon.png"]
	out, e := (Engine{Runner: runner, Tools: ts, Limits: domain.DefaultLimits()}).Optimize(context.Background(), data, r)
	if e != nil || !bytes.Equal(out.Candidate.Data, data) {
		t.Fatal(e)
	}
}
func TestCorruptOptimizerAndCancellationDoNotCommit(t *testing.T) {
	runner := &testutil.Runner{Handler: func(ctx context.Context, i process.Invocation) (process.Result, error, bool) {
		dest := ""
		for n, a := range i.Args {
			if a == "--out" {
				dest = i.Args[n+1]
			}
		}
		_ = os.WriteFile(dest, []byte("not PNG"), 0600)
		return process.Result{}, nil, true
	}}
	r, _ := domain.Builtin("lossless")
	e := Engine{Runner: runner, Tools: testutil.Tools(), Limits: domain.DefaultLimits()}
	if _, er := e.Optimize(context.Background(), testutil.Corpus()["flat-icon.png"], r); domain.Code(er) != 5 {
		t.Fatal(er)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, er := e.Optimize(ctx, nil, r); domain.Code(er) != 130 {
		t.Fatal(er)
	}
}
func TestTargetResizeNeedsPermission(t *testing.T) {
	r, _ := domain.Builtin("auto")
	r.TargetBytes = 300
	r.MinDimensions = domain.Dimensions{Width: 4, Height: 4}
	ts := testutil.Tools()
	ts.PNGQuant.Available = false
	e := Engine{Runner: &testutil.Runner{}, Tools: ts, Limits: domain.DefaultLimits()}
	data := testutil.Corpus()["synthetic-noise.png"]
	if _, err := e.Optimize(context.Background(), data, r); domain.Code(err) != 6 {
		t.Fatal(err)
	}
	r.AllowResize = true
	out, err := e.Optimize(context.Background(), data, r)
	if err != nil || out.Candidate.Info.Dimensions.Width < 4 || len(out.Candidate.Data) > 300 {
		t.Fatal(out.Candidate.Info, err)
	}
}
func TestRemotePermissionAndSafeArgumentBoundary(t *testing.T) {
	r, _ := domain.Builtin("auto")
	r.Engine = "tinypng"
	if _, e := (Engine{Limits: domain.DefaultLimits()}).Optimize(context.Background(), nil, r); domain.Code(e) != 2 {
		t.Fatal(e)
	}
	r, _ = domain.Builtin("auto")
	for _, s := range []string{"--output=evil", "--force", "--quality=0-1", "file.png"} {
		r.PNGQuantArgs = []string{s}
		if ValidateExtras(r) == nil {
			t.Fatal(s)
		}
	}
}
