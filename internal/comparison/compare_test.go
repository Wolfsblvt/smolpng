package comparison

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
	"github.com/Wolfsblvt/smolpng/internal/workflow"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestComparisonModesUseTheSameSourceAndReportActualPipeline(t *testing.T) {
	data := testutil.Corpus()["flat-icon.png"]
	r, _ := domain.Builtin("web")
	run := &testutil.Runner{}
	e := workflow.Engine{Runner: run, Tools: testutil.Tools(), Limits: domain.DefaultLimits()}
	for _, mode := range []string{"auto", "pipeline", "lossless", "quantizer"} {
		row := Evaluate(context.Background(), e, "source.png", data, Variant{ID: mode, Recipe: r, PNGMode: mode}, 1)
		if row.ExitCode != 0 || row.CandidateBytes == 0 || row.SourceSHA256 == "" || row.Metrics == nil {
			t.Fatalf("%s: %+v", mode, row)
		}
		if mode == "lossless" && !row.Metrics.PixelIdentical {
			t.Fatal("lossless comparison changed pixels")
		}
	}
	for _, c := range run.Calls() {
		if c.Invocation.Executable == "pngquant" && !bytes.Equal(c.Input, data) {
			t.Fatal("recompressed a previous result instead of original")
		}
	}
}
func TestHiddenRGBDoesNotInventVisibleError(t *testing.T) {
	a := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	b := image.NewNRGBA(a.Bounds())
	a.SetNRGBA(0, 0, color.NRGBA{R: 200, A: 0})
	b.SetNRGBA(0, 0, color.NRGBA{B: 100, A: 0})
	r, _ := domain.Builtin("web")
	m := measure(testutil.PNG(a, png.NoCompression), testutil.PNG(b, png.NoCompression), r, 100)
	if m.PixelIdentical || m.BlackRMSE == nil || *m.BlackRMSE != 0 || *m.WhiteRMSE != 0 {
		t.Fatalf("hidden-channel metrics: %+v", m)
	}
}
func TestComparisonPreservesTargetFailureAndRejectsRemote(t *testing.T) {
	r, _ := domain.Builtin("web")
	r.TargetBytes = 1
	e := workflow.Engine{Runner: &testutil.Runner{}, Tools: testutil.Tools(), Limits: domain.DefaultLimits()}
	row := Evaluate(context.Background(), e, "a.png", testutil.Corpus()["flat-icon.png"], Variant{ID: "one", Recipe: r, PNGMode: "auto"}, 1)
	if row.ExitCode != 6 || row.Status != "target-missed" || row.TargetMet == nil || *row.TargetMet {
		t.Fatalf("%+v", row)
	}
	r.RemotePolicy = "compare"
	if ValidateVariants([]Variant{{ID: "remote", Recipe: r, PNGMode: "auto"}}) == nil {
		t.Fatal("matrix could multiply billable uploads")
	}
}
func TestReportsEscapeNamesAndWithholdIncomparableMetrics(t *testing.T) {
	r := NewReport("dev", "abc", testutil.Tools(), nil)
	r.Rows = []Row{{Input: "=HYPERLINK(\"evil\")", Variant: "<script>evil</script>", Status: "failed"}}
	var b bytes.Buffer
	if err := CSV(&b, r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "'=HYPERLINK") {
		t.Fatal("CSV formula not guarded")
	}
	h, err := HTML(r)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(h, []byte("<script>evil")) {
		t.Fatal("HTML injection")
	}
	b.Reset()
	if err = JSON(&b, r); err != nil || !json.Valid(b.Bytes()) {
		t.Fatal(err)
	}
	q, _ := domain.Builtin("auto")
	a := testutil.Corpus()["alpha-gradient.png"]
	out := testutil.AddChunk(a, "sRGB", []byte{0})
	m := measure(a, out, q, 100000)
	if m.BlackRMSE != nil || m.OmittedReason == "" {
		t.Fatal("compared unaligned color spaces")
	}
}
