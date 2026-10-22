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

type jpegRunner struct {
	sources   [][]byte
	qualities []int
	legacy14  bool
}

func (r *jpegRunner) Run(ctx context.Context, i process.Invocation) (process.Result, error) {
	if r.legacy14 {
		for _, argument := range i.Args {
			if argument == "--keep-all" {
				return process.Result{ExitCode: 1, Stderr: "unrecognized option '--keep-all'"}, nil
			}
		}
	}
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
	r.qualities = append(r.qualities, quality)
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

func TestJPEGMetadataPreservationSupportsAdmitted14Backend(t *testing.T) {
	runner := &jpegRunner{legacy14: true}
	engine := Engine{Runner: runner, Tools: toolchain.Set{JPEGOptim: toolchain.Tool{Name: "jpegoptim", Path: "jpegoptim", Version: "1.4.7", Available: true}}, Limits: domain.DefaultLimits()}
	recipe, _ := domain.Builtin("auto")
	result, err := engine.Optimize(context.Background(), jpegData(), recipe)
	if err != nil || result.Candidate.Info.Format != "jpeg" || len(runner.qualities) != 2 {
		t.Fatalf("admitted jpegoptim 1.4 must support both original-derived attempts: %v; qualities %v", err, runner.qualities)
	}
}

func TestJPEGBuiltinsUseSelectedFormatSpecificQuality(t *testing.T) {
	for _, test := range []struct {
		name        string
		jpegQuality int
		pngQuality  int
	}{{"auto", 85, 95}, {"web", 80, 95}, {"smallest", 70, 85}, {"lossless", -1, 95}} {
		t.Run(test.name, func(t *testing.T) {
			runner := &jpegRunner{}
			engine := Engine{Runner: runner, Tools: toolchain.Set{JPEGOptim: toolchain.Tool{Name: "jpegoptim", Path: "jpegoptim", Available: true}}, Limits: domain.DefaultLimits()}
			recipe, err := domain.Builtin(test.name)
			if err != nil {
				t.Fatal(err)
			}
			if recipe.QualityMax != test.pngQuality {
				t.Fatal("PNG quality changed", recipe.QualityMax)
			}
			if _, err = engine.Optimize(context.Background(), jpegData(), recipe); err != nil {
				t.Fatal(err)
			}
			wantCalls := 2
			if test.name == "lossless" {
				wantCalls = 1
			}
			if len(runner.qualities) != wantCalls || runner.qualities[0] != -1 || runner.qualities[len(runner.qualities)-1] != test.jpegQuality {
				t.Fatalf("JPEG encoder qualities: got %v, want final %d and an initial lossless attempt", runner.qualities, test.jpegQuality)
			}
		})
	}
}
func TestJPEGCandidatesAlwaysUseOriginalAndStrictTarget(t *testing.T) {
	data := jpegData()
	runner := &jpegRunner{}
	engine := Engine{Runner: runner, Tools: toolchain.Set{JPEGOptim: toolchain.Tool{Name: "jpegoptim", Path: "jpegoptim", Version: "1.5.6", Available: true}}, Limits: domain.DefaultLimits()}
	r, _ := domain.Builtin("auto")
	r.TargetBytes = 1
	r.SetQuality(50, 85)
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
