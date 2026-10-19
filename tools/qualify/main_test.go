package main

import (
	"context"
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/jpegimage"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
)

func TestQualifierValidatesArtifactNotJustSuccessfulExit(t *testing.T) {
	original := testutil.Corpus()["hidden-rgb.png"]
	a, e := pngimage.Decode(original, 1_000_000)
	if e != nil {
		t.Fatal(e)
	}
	output := testutil.PNG(a.Image, png.BestCompression)
	r := domain.Result{Format: "png", OriginalBytes: int64(len(original)), OutputBytes: int64(len(output)), Dimensions: domain.DimensionChange{Before: a.Info.Dimensions, After: a.Info.Dimensions}}
	if errors := validateImage(original, output, r, true); len(errors) != 0 {
		t.Fatal(errors)
	}
	if errors := validateImage(original, []byte("not a PNG"), r, true); len(errors) == 0 {
		t.Fatal("malformed result accepted")
	}
	r.RemoteUsed = true
	r.OutputBytes++
	if errors := validateImage(original, output, r, true); len(errors) != 2 {
		t.Fatal(errors)
	}
}

func TestQualifierValidatesJPEGArtifactPixelsAndOrientation(t *testing.T) {
	corpus := testutil.JPEGCorpus()
	original := corpus["jpeg-baseline.jpg"]
	a, err := jpegimage.Decode(original, 1_000_000)
	if err != nil {
		t.Fatal(err)
	}
	r := domain.Result{Format: "jpeg", OriginalBytes: int64(len(original)), OutputBytes: int64(len(original)), Dimensions: domain.DimensionChange{Before: a.Info.Dimensions, After: a.Info.Dimensions}}
	if failures := validateImage(original, original, r, true); len(failures) != 0 {
		t.Fatal(failures)
	}
	if failures := validateImage(original, []byte("not a JPEG"), r, false); len(failures) == 0 {
		t.Fatal("invalid JPEG output accepted")
	}
	changed := testutil.JPEG(a.Image, 70)
	r.OutputBytes = int64(len(changed))
	if failures := validateImage(original, changed, r, true); len(failures) == 0 {
		t.Fatal("lossless JPEG pixel change accepted")
	}
	if failures := validateImage(original, changed, r, false); len(failures) != 0 {
		t.Fatal("ordinary JPEG quality change rejected", failures)
	}
	oriented := corpus["jpeg-orientation-8.jpg"]
	r.OriginalBytes, r.OutputBytes = int64(len(oriented)), int64(len(original))
	if failures := validateImage(oriented, original, r, false); len(failures) == 0 {
		t.Fatal("orientation loss accepted")
	}
	r.OriginalBytes = int64(len(original))
	r.Format = "png"
	if failures := validateImage(original, original, r, false); len(failures) == 0 {
		t.Fatal("false format report accepted")
	}
}

func TestPartialEvidenceCannotClaimQualification(t *testing.T) {
	s := suite{root: t.TempDir(), data: evidence{Passed: true, Checks: []check{{Name: "preflight", Passed: true}}}}
	read := func() evidence {
		t.Helper()
		if err := s.save(); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(s.root, "report.json"))
		if err != nil {
			t.Fatal(err)
		}
		var got evidence
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		return got
	}
	if got := read(); got.Passed || got.Complete || !s.data.Passed {
		t.Fatal("partial evidence either claimed success or changed accumulated results", got)
	}
	s.data.Complete = true
	if got := read(); !got.Passed || !got.Complete {
		t.Fatal(got)
	}
	s.data.Checks = append(s.data.Checks, check{Name: "bad-output", Report: json.RawMessage("not JSON"), Passed: true})
	if got := read(); got.Passed || len(got.Checks) != 2 || len(got.Checks[0].Errors) != 0 || len(got.Checks[1].Errors) == 0 {
		t.Fatal("malformed output lost prior evidence or retained a passing verdict", got)
	}
}

func TestCancelledRunRetainsIncompleteEvidence(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "candidate")
	if err := os.WriteFile(exe, []byte("never executed"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	output := filepath.Join(dir, "qualification")
	if err := run(ctx, options{Executable: exe, Directory: output, Formats: "jpeg", Timeout: time.Second}); err == nil {
		t.Fatal("cancelled run succeeded")
	}
	data, err := os.ReadFile(filepath.Join(output, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Complete || got.Passed || len(got.Checks) == 0 || len(got.Formats) != 1 || got.Formats[0] != "jpeg" {
		t.Fatal(got)
	}
	if got.Host.OS == "" || got.Host.Architecture == "" || got.ExecutableSHA256 == "" {
		t.Fatal("interrupted run lost its captured host/application identities", got)
	}
}

type runnerFunc func(context.Context, process.Invocation) (process.Result, error)

func (run runnerFunc) Run(ctx context.Context, invocation process.Invocation) (process.Result, error) {
	return run(ctx, invocation)
}

func TestMixedBatchRejectsDuplicateResultInsteadOfMissingJPEG(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "mixed")
	if err := os.Mkdir(output, 0700); err != nil {
		t.Fatal(err)
	}
	name := "flat-icon.png"
	original := testutil.Corpus()[name]
	asset, err := pngimage.Decode(original, 1_000_000)
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(output, name)
	if err := os.WriteFile(dest, original, 0600); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(dir, "sources")
	result := domain.Result{Format: "png", Input: filepath.Join(source, name), Output: dest,
		OriginalBytes: int64(len(original)), OutputBytes: int64(len(original)),
		Dimensions: domain.DimensionChange{Before: asset.Info.Dimensions, After: asset.Info.Dimensions}}
	response, err := json.Marshal(domain.Report{Results: []domain.Result{result, result}})
	if err != nil {
		t.Fatal(err)
	}
	s := suite{root: dir, opt: options{Timeout: time.Second}, data: evidence{Passed: true},
		runner: runnerFunc(func(context.Context, process.Invocation) (process.Result, error) {
			return process.Result{Stdout: string(response)}, nil
		})}
	s.exerciseMixedBatch(context.Background(), source, []string{name, "jpeg-baseline.jpg"},
		map[string][]byte{name: original})
	if s.data.Passed {
		t.Fatal("two PNG results falsely qualified the requested PNG/JPEG batch")
	}
}
