package workflow

import (
	"context"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
	"image/png"
	"testing"
)

func TestRequestedBackgroundCannotSelectTransparentOriginal(t *testing.T) {
	r, _ := domain.Builtin("auto")
	r.Background = "#102030"
	r.TargetBytes = 1
	r.AcceptNearest = true
	data := testutil.PNG(testutil.Pattern(32, 32, "hidden"), png.NoCompression)
	e := Engine{Limits: domain.DefaultLimits(), Tools: testutil.Tools(), Runner: &testutil.Runner{}}
	out, err := e.Optimize(context.Background(), data, r)
	if domain.Code(err) != 6 {
		t.Fatalf("expected strict miss, got %v", err)
	}
	a, err := pngimage.Decode(out.Candidate.Data, 10000)
	if err != nil {
		t.Fatal(err)
	}
	if a.Info.Transparency {
		t.Fatal("unflattened original won")
	}
	if len(out.Candidate.Pipeline) == 0 || out.Candidate.Pipeline[0].Tool != "smolpng-background" {
		t.Fatal("missing transformation report")
	}
	if !out.Original.Transparency {
		t.Fatal("source history was overwritten by transformed metadata")
	}
}
func TestMatteNeverUploadsDifferentSourceOrWeakensLossless(t *testing.T) {
	r, _ := domain.Builtin("lossless")
	r.Background = "white"
	if r.Validate() == nil {
		t.Fatal("lossless matte allowed")
	}
	r, _ = domain.Builtin("auto")
	r.Background = "white"
	r.RemotePolicy = "compare"
	run := &testutil.Runner{}
	_, err := (Engine{Limits: domain.DefaultLimits(), Runner: run}).Optimize(context.Background(), nil, r)
	if domain.Code(err) != 2 || len(run.Calls()) != 0 {
		t.Fatal("remote/preservation validation came after effects")
	}
}
