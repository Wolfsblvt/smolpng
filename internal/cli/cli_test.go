package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/paths"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type harness struct {
	Environment
	out, err    *bytes.Buffer
	runner      *testutil.Runner
	source, dir string
	tools       []string
}

func setup(t *testing.T) *harness {
	t.Helper()
	d, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	out, er := &bytes.Buffer{}, &bytes.Buffer{}
	p := &paths.Paths{Config: filepath.Join(d, "config.toml"), Data: filepath.Join(d, "data"), State: filepath.Join(d, "state"), Cache: filepath.Join(d, "cache"), Downloads: filepath.Join(d, "Downloads")}
	r := &testutil.Runner{}
	h := &harness{Environment: Environment{In: bytes.NewReader(nil), Out: out, Err: er, Runner: r, Paths: p, CWD: d}, out: out, err: er, runner: r, dir: d, source: filepath.Join(d, "a.png")}
	if e := os.WriteFile(h.source, testutil.Corpus()["flat-icon.png"], 0600); e != nil {
		t.Fatal(e)
	}
	for _, n := range []string{"pngquant", "oxipng"} {
		p := filepath.Join(d, n)
		_ = os.WriteFile(p, []byte("test boundary"), 0700)
		h.tools = append(h.tools, "--"+n+"-path", p)
	}
	return h
}
func (h *harness) run(args ...string) int {
	h.out.Reset()
	h.err.Reset()
	return Run(context.Background(), append(append([]string{}, h.tools...), args...), h.Environment)
}
func TestCLICompressionJSONOriginalAndArgumentMapping(t *testing.T) {
	h := setup(t)
	original, _ := os.ReadFile(h.source)
	if c := h.run("compress", h.source, "--json", "--quality", "65-90", "--colors", "64", "--verbose"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	var r domain.Report
	if e := json.Unmarshal(h.out.Bytes(), &r); e != nil {
		t.Fatal(e)
	}
	if len(r.Results) != 1 || r.Results[0].Output == "" || r.Results[0].OutputBytes >= r.Results[0].OriginalBytes || r.Results[0].RemoteUsed {
		t.Fatal(r)
	}
	b, _ := os.ReadFile(h.source)
	if !bytes.Equal(b, original) {
		t.Fatal("original overwritten")
	}
	found := false
	for _, call := range h.runner.Calls() {
		if strings.Contains(strings.Join(call.Invocation.Args, " "), "--quality=65-90") {
			found = true
			if call.Invocation.Args[0] != "64" {
				t.Fatal(call)
			}
		}
	}
	if !found {
		t.Fatal("CLI mapping not observed")
	}
}
func TestBinaryStdoutAndSeparateReport(t *testing.T) {
	h := setup(t)
	p := filepath.Join(h.dir, "result.json")
	if c := h.run("compress", h.source, "--stdout", "--quiet", "--report-file", p); c != 0 {
		t.Fatal(c, h.err)
	}
	if _, e := pngimage.Decode(h.out.Bytes(), 1000000); e != nil {
		t.Fatal("stdout not pure PNG", e)
	}
	b, e := os.ReadFile(p)
	if e != nil || !json.Valid(b) {
		t.Fatal(e)
	}
	if c := h.run("compress", h.source, "--stdout", "--json"); c != 2 {
		t.Fatal("binary/JSON conflict accepted", c)
	}
}
func TestTargetMissAndPartialBatchJSONL(t *testing.T) {
	h := setup(t)
	if c := h.run("compress", h.source, "--target-size", "1B", "--json"); c != 6 {
		t.Fatal(c, h.out, h.err)
	}
	if _, e := os.Stat(filepath.Join(h.dir, "a.smol.png")); !os.IsNotExist(e) {
		t.Fatal("target miss wrote output")
	}
	if c := h.run("compress", h.source, "missing.png", "--jsonl", "--jobs", "2"); c != 8 {
		t.Fatal(c, h.out)
	}
	lines := strings.Split(strings.TrimSpace(h.out.String()), "\n")
	if len(lines) != 2 {
		t.Fatal(h.out)
	}
	for _, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Fatal(line)
		}
	}
}
func TestDryRunNoProcessAndRelativeOutput(t *testing.T) {
	h := setup(t)
	if c := h.run("compress", h.source, "--output-dir", "export", "--dry-run", "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	if len(h.runner.Calls()) != 0 {
		t.Fatal("dry run probed processes")
	}
	if _, e := os.Stat(filepath.Join(h.dir, "export")); !os.IsNotExist(e) {
		t.Fatal("dry run wrote directory")
	}
	if c := h.run("compress", h.source, "--output", "relative.png", "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	var r domain.Report
	_ = json.Unmarshal(h.out.Bytes(), &r)
	if r.Results[0].Output != filepath.Join(h.dir, "relative.png") {
		t.Fatal(r)
	}
}
func TestInvalidOptionsCannotClobberSource(t *testing.T) {
	h := setup(t)
	for _, args := range [][]string{{"compress", h.source, "--report-file", h.source, "--overwrite"}, {"compress", h.source, "--quality", "90-70"}, {"compress", h.source, "--pngquant-arg=--output=evil"}, {"compress", h.source, "--unknown", "--json"}} {
		if c := h.run(args...); c != 2 {
			t.Fatal(args, c, h.out, h.err)
		}
	}
	if !json.Valid(h.out.Bytes()) {
		t.Fatal("parse error broke machine output")
	}
}
func TestCapabilitiesInspectCompletions(t *testing.T) {
	h := setup(t)
	for _, cmd := range []string{"capabilities", "version", "doctor"} {
		if c := h.run(cmd, "--json"); c != 0 || !json.Valid(h.out.Bytes()) {
			t.Fatal(cmd, c, h.out)
		}
	}
	if c := h.run("inspect", h.source, "--json"); c != 0 || !json.Valid(h.out.Bytes()) {
		t.Fatal(c, h.out)
	}
	for _, s := range []string{"bash", "zsh", "fish", "powershell"} {
		if c := h.run("completion", s); c != 0 || h.out.Len() == 0 {
			t.Fatal(s, c)
		}
	}
}
