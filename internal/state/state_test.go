package state

import (
	"context"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLastRoundTripRetainsMetadataNotImageBytes(t *testing.T) {
	p := filepath.Join(t.TempDir(), "last.json")
	r := domain.NewResult("source", "auto")
	r.Status = "optimized"
	r.Output = filepath.Join(t.TempDir(), "image.png")
	r.Bytes = []byte("private image payload")
	l := Last{SchemaVersion: 1, AppVersion: "test", Completed: time.Now().UTC(), Report: domain.Summarize([]domain.Result{r}), Warnings: []string{"clipboard unavailable"}}
	if e := Write(context.Background(), p, l); e != nil {
		t.Fatal(e)
	}
	data, _ := os.ReadFile(p)
	if strings.Contains(string(data), "payload") {
		t.Fatal("image data retained in state")
	}
	got, e := ReadLast(p)
	if e != nil || got.AppVersion != "test" || got.Report.Results[0].Output != r.Output || len(got.Warnings) != 1 {
		t.Fatal(got, e)
	}
}
func TestStateRejectsMalformedExcessiveUnknownAndSymlink(t *testing.T) {
	p := filepath.Join(t.TempDir(), "state.json")
	for _, text := range []string{`{"schemaVersion":1,"acknowledged":true} {}`, `{"schemaVersion":1,"secret":"hidden"}`, strings.Repeat(" ", 4<<20) + "{}"} {
		os.WriteFile(p, []byte(text), 0600)
		var s Setup
		if e := Read(p, &s); e == nil {
			t.Fatal("invalid state accepted")
		}
	}
	os.WriteFile(p, []byte(`{"schemaVersion":1,"acknowledged":true}`), 0600)
	link := p + ".link"
	if e := os.Symlink(p, link); e == nil {
		if e := Write(context.Background(), link, Setup{1, true}); e == nil {
			t.Fatal("symlink state written")
		}
		if e := Read(link, new(Setup)); e == nil {
			t.Fatal("symlink state read")
		}
	}
}
func TestLastSchemaAndArtifactSelection(t *testing.T) {
	p := filepath.Join(t.TempDir(), "last.json")
	if e := Write(context.Background(), p, Last{SchemaVersion: 2}); e != nil {
		t.Fatal(e)
	}
	if _, e := ReadLast(p); e == nil {
		t.Fatal("unknown last schema accepted")
	}
	r := domain.Report{Results: []domain.Result{{Output: "planned.png", Status: "planned"}, {Output: "-", Status: "optimized"}, {Output: "", Status: "failed"}, {Output: "actual.png", Status: "target-missed"}}}
	files := Outputs(r)
	if len(files) != 1 || files[0] != "actual.png" {
		t.Fatal(files)
	}
}
