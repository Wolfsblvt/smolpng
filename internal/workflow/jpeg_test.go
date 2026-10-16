package workflow

import (
	"bytes"
	"context"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/jpegimage"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
	"github.com/Wolfsblvt/smolpng/internal/toolchain"
	"image/jpeg"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type jpegRunner struct{ sources [][]byte }

func (r *jpegRunner) Run(ctx context.Context, i process.Invocation) (process.Result, error) {
	b, e := os.ReadFile(i.Args[len(i.Args)-1])
	if e != nil {
		return process.Result{}, e
	}
	r.sources = append(r.sources, bytes.Clone(b))
	dest := ""
	quality := -1
	for _, a := range i.Args {
		if strings.HasPrefix(a, "--dest=") {
			dest = strings.TrimPrefix(a, "--dest=")
		}
		if strings.HasPrefix(a, "--max=") {
			quality, _ = strconv.Atoi(strings.TrimPrefix(a, "--max="))
		}
	}
	if quality >= 0 {
		a, e := jpegimage.Decode(b, 1000000)
		if e != nil {
			return process.Result{}, e
		}
		var out bytes.Buffer
		_ = jpeg.Encode(&out, a.Image, &jpeg.Options{Quality: quality})
		b = out.Bytes()
	}
	return process.Result{}, os.WriteFile(filepath.Join(dest, "original.jpg"), b, 0600)
}
func jpegData() []byte {
	var b bytes.Buffer
	_ = jpeg.Encode(&b, testutil.Pattern(64, 64, "noise"), &jpeg.Options{Quality: 98})
	return b.Bytes()
}
func TestJPEGCandidatesAlwaysUseOriginalAndStrictTarget(t *testing.T) {
	data := jpegData()
	runner := &jpegRunner{}
	engine := Engine{Runner: runner, Tools: toolchain.Set{JPEGOptim: toolchain.Tool{Name: "jpegoptim", Path: "jpegoptim", Version: "1.5.6", Available: true}}, Limits: domain.DefaultLimits()}
	r, _ := domain.Builtin("auto")
	r.TargetBytes = 1
	r.QualityMin = 50
	r.QualityMax = 85
	out, e := engine.Optimize(context.Background(), data, r)
	if domain.Code(e) != 6 || len(out.Candidate.Data) == 0 || out.Candidate.Info.Format != "jpeg" {
		t.Fatal(out, e)
	}
	for _, source := range runner.sources {
		if !bytes.Equal(source, data) {
			t.Fatal("lossy JPEG generation used a previous candidate")
		}
	}
	r.TargetBytes = 5000
	out, e = engine.Optimize(context.Background(), data, r)
	if e != nil || len(out.Candidate.Data) > 5000 {
		t.Fatal(len(out.Candidate.Data), e)
	}
}
func TestJPEGExplicitResizeWorksWithoutAnExternalEncoder(t *testing.T) {
	data := jpegData()
	r, _ := domain.Builtin("auto")
	r.Resize = domain.Dimensions{Width: 16, Height: 16}
	out, e := (Engine{Limits: domain.DefaultLimits()}).Optimize(context.Background(), data, r)
	if e != nil || out.Candidate.Info.Dimensions != r.Resize {
		t.Fatal(out, e)
	}
}

func TestJPEGAttachedDataKeepPreservesWholeFileWithoutOptimizer(t *testing.T) {
	data := append(jpegData(), []byte("DualShot_Meta_Info\x00SEFT")...)
	for _, target := range []int64{0, int64(len(data)), 1} {
		runner := &jpegRunner{}
		engine := Engine{Runner: runner, Tools: toolchain.Set{JPEGOptim: toolchain.Tool{Name: "jpegoptim", Path: "jpegoptim", Available: true}}, Limits: domain.DefaultLimits()}
		r, _ := domain.Builtin("auto")
		r.Metadata, r.TargetBytes, r.AllowResize = "keep", target, target > 0
		out, err := engine.Optimize(context.Background(), data, r)
		wantCode := 0
		if target == 1 {
			wantCode = 6
		}
		if domain.Code(err) != wantCode || !bytes.Equal(out.Candidate.Data, data) || len(runner.sources) != 0 || len(out.Warnings) == 0 {
			t.Fatalf("target %d: code %d, calls %d, warnings %v", target, domain.Code(err), len(runner.sources), out.Warnings)
		}
		r.Resize = domain.Dimensions{Width: 16, Height: 16}
		if _, err := engine.Optimize(context.Background(), data, r); domain.Code(err) != 2 {
			t.Fatal("keep accepted geometry change", err)
		}
	}
}

func TestJPEGAttachedDataSafeAndAllProduceWarnedStillImage(t *testing.T) {
	primary := jpegData()
	trailer := []byte("DualShot_Meta_Info\x00SEFT")
	data := append(bytes.Clone(primary), trailer...)
	snapshot := bytes.Clone(data)
	for _, policy := range []string{"safe", "all"} {
		runner := &jpegRunner{}
		engine := Engine{Runner: runner, Tools: toolchain.Set{JPEGOptim: toolchain.Tool{Name: "jpegoptim", Path: "jpegoptim", Available: true}}, Limits: domain.DefaultLimits()}
		r, _ := domain.Builtin("lossless")
		r.Metadata = policy
		out, err := engine.Optimize(context.Background(), data, r)
		if err != nil || !bytes.Equal(out.Candidate.Data, primary) || len(out.Warnings) == 0 {
			t.Fatal(policy, out.Warnings, err)
		}
		for _, source := range runner.sources {
			if !bytes.Equal(source, primary) {
				t.Fatal("attachment sent to optimizer")
			}
		}
	}
	if !bytes.Equal(data, snapshot) {
		t.Fatal("source bytes changed")
	}
}
