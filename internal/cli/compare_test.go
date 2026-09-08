package cli

import (
	"bytes"
	"encoding/json"
	"github.com/Wolfsblvt/smolpng/internal/comparison"
	"os"
	"path/filepath"
	"testing"
)

func TestCompareProducesInspectableLocalBundleWithoutChangingOriginal(t *testing.T) {
	h := setup(t)
	before, _ := os.ReadFile(h.source)
	dest := filepath.Join(h.dir, "comparison")
	if code := h.run("compare", h.source, "--preset", "web", "--png-mode", "quantizer", "--png-mode", "lossless", "--repeat", "2", "--output-dir", dest, "--json"); code != 0 {
		t.Fatal(code, h.out, h.err)
	}
	var r comparison.Report
	if e := json.Unmarshal(h.out.Bytes(), &r); e != nil {
		t.Fatal(e)
	}
	if len(r.Rows) != 4 {
		t.Fatal("missing runs", len(r.Rows))
	}
	for _, name := range []string{"report.json", "report.csv", "index.html", r.Rows[0].OriginalPath, r.Rows[0].CandidatePath} {
		if _, e := os.Stat(filepath.Join(dest, name)); e != nil {
			t.Fatal(name, e)
		}
	}
	now, _ := os.ReadFile(h.source)
	if !bytes.Equal(before, now) {
		t.Fatal("comparison modified original")
	}
	if code := h.run("compare", h.source, "--output-dir", dest); code != 2 {
		t.Fatal("overwrote comparison output")
	}
}
func TestCompareDryRunNeverProbesOrWrites(t *testing.T) {
	h := setup(t)
	dest := filepath.Join(h.dir, "never-created")
	if c := h.run("compare", h.source, "--dry-run", "--output-dir", dest, "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	if len(h.runner.Calls()) != 0 {
		t.Fatal("dry-run invoked optimizer")
	}
	if _, e := os.Stat(dest); !os.IsNotExist(e) {
		t.Fatal("dry run output created")
	}
}
func TestCompareMissDoesNotSaveRejectedTarget(t *testing.T) {
	h := setup(t)
	dest := filepath.Join(h.dir, "miss")
	if c := h.run("compare", h.source, "--preset", "lossless", "--target-size", "1B", "--output-dir", dest, "--json"); c != 6 {
		t.Fatal(c, h.out)
	}
	var r comparison.Report
	_ = json.Unmarshal(h.out.Bytes(), &r)
	if r.Rows[0].CandidatePath != "" {
		t.Fatal("missed target became a successful artifact")
	}
}
func TestBackgroundPresetAndInvocationShareOneOption(t *testing.T) {
	h := setup(t)
	if c := h.run("compress", h.source, "--background", "white", "--json"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	if c := h.run("config", "show", "--json"); c != 0 {
		t.Fatal(c)
	}
}
