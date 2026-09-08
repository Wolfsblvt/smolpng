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
