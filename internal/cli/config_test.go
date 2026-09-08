package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/config"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/input"
)

func TestNamedPresetDrivesRealCompressionPath(t *testing.T) {
	h := setup(t)
	text := `[defaults]
preset="site"
[output]
directory="export"
[presets.site]
schema=1
extends="auto"
colors=64
quality="65-90"
`
	if e := os.WriteFile(h.Paths.Config, []byte(text), 0600); e != nil {
		t.Fatal(e)
	}
	if c := h.run("compress", h.source, "--json"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	var report domain.Report
	if e := json.Unmarshal(h.out.Bytes(), &report); e != nil {
		t.Fatal(e)
	}
	if report.Results[0].Recipe != "site" || report.Results[0].Output != filepath.Join(h.dir, "export", "a.smol.png") {
		t.Fatal(report)
	}
	found := false
	for _, call := range h.runner.Calls() {
		args := call.Invocation.Args
		if len(args) > 0 && strings.Contains(strings.Join(args, " "), "--quality=65-90") {
			found = true
			if args[0] != "64" {
				t.Fatal(args)
			}
		}
	}
	if !found {
		t.Fatal("config did not reach the engine")
	}
	if c := h.run("compress", h.source, "--output", "one.png", "--dry-run", "--json"); c != 0 {
		t.Fatal("explicit output did not override inherited directory", c, h.out)
	}
}
func TestCLIImportPreviewApplyAndExportRoundtrip(t *testing.T) {
	h := setup(t)
	file := filepath.Join(h.dir, "preset.toml")
	text := "[presets.tiny]\nschema=1\nextends='auto'\ntarget_size='10KiB'\n"
	if e := os.WriteFile(file, []byte(text), 0600); e != nil {
		t.Fatal(e)
	}
	if c := h.run("preset", "import", file, "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	var p config.ImportPlan
	if e := json.Unmarshal(h.out.Bytes(), &p); e != nil || p.Applied {
		t.Fatal(p, e)
	}
	if _, e := os.Stat(filepath.Join(h.Paths.Data, "presets.toml")); !os.IsNotExist(e) {
		t.Fatal("preview created store")
	}
	if c := h.run("preset", "import", file, "--apply", "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	if c := h.run("compress", h.source, "--preset", "tiny", "--dry-run", "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	if c := h.run("preset", "export", "tiny", "--resolved"); c != 0 {
		t.Fatal(c, h.err)
	}
	b, e := config.ParseBundle(h.out.Bytes())
	if e != nil || b.Presets["tiny"].Extends != "" {
		t.Fatal(string(h.out.Bytes()), e)
	}
	if c := h.run("preset", "pin", "tiny", "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	if c := h.run("preset", "move", "tiny", "--position", "2", "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	if c := h.run("preset", "remove", "tiny", "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	if c := h.run("config", "validate", "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
}
func TestConfigShowProvenanceAndBrokenConfigDiagnostics(t *testing.T) {
	h := setup(t)
	if c := h.run("config", "init", "--project", "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	if c := h.run("config", "show", "--effective", "--provenance", "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	var eff config.Effective
	if e := json.Unmarshal(h.out.Bytes(), &eff); e != nil || eff.Provenance["output.template"].File != filepath.Join(h.dir, ".smolpng.toml") {
		t.Fatal(e, h.out)
	}
	if c := h.run("config", "init", "--project", "--json"); c != 2 {
		t.Fatal("init overwrote existing settings")
	}
	if e := os.WriteFile(h.Paths.Config, []byte("[defaults]\ncommmand='oops'"), 0600); e != nil {
		t.Fatal(e)
	}
	if c := h.run("compress", h.source, "--json"); c != 2 || !json.Valid(h.out.Bytes()) {
		t.Fatal(c, h.out)
	}
	if c := h.run("version", "--json"); c != 0 {
		t.Fatal("broken config disabled version", c, h.out)
	}
	if c := h.run("config", "path", "--json"); c != 0 {
		t.Fatal("broken config disabled path recovery")
	}
}
func TestImportedRemoteRecipeCannotCauseAnUnattendedUpload(t *testing.T) {
	h := setup(t)
	p := filepath.Join(h.dir, "remote.toml")
	os.WriteFile(p, []byte("[presets.remote]\nschema=1\nengine='tinypng'\nremote='always'\n"), 0600)
	if c := h.run("preset", "import", p, "--apply", "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	if c := h.run("compress", h.source, "--preset", "remote", "--json"); c != 2 {
		t.Fatal("ungranted provider accepted", c, h.out)
	}
	if len(h.runner.Calls()) != 0 {
		t.Fatal("process ran before remote permission validation")
	}
}

type presetClipboard struct{ reads int }

func (c *presetClipboard) Read(context.Context) (input.Snapshot, error) {
	c.reads++
	return input.Snapshot{Text: "```toml\n[presets.clip]\nschema=1\n```", Image: []byte("untouched image")}, nil
}
func TestClipboardPresetImportDoesNotReplaceImageRepresentations(t *testing.T) {
	h := setup(t)
	c := &presetClipboard{}
	h.Clipboard = c
	if exit := h.run("preset", "import", "--clipboard", "--apply", "--json"); exit != 0 || c.reads != 1 {
		t.Fatal(exit, h.out)
	}
}
