package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/config"
	"github.com/Wolfsblvt/smolpng/internal/desktop"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/input"
	"github.com/Wolfsblvt/smolpng/internal/state"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
	"github.com/Wolfsblvt/smolpng/internal/ui"
	"github.com/spf13/cobra"
)

type promptStep struct {
	kind, title, choice string
	values              map[string]string
}
type promptFake struct {
	t     *testing.T
	steps []promptStep
	seen  []string
	menus [][]ui.Option
	forms [][]ui.Field
}

func (p *promptFake) take(kind, title string) promptStep {
	p.t.Helper()
	if len(p.steps) == 0 {
		p.t.Fatalf("unexpected %s %q; seen %v", kind, title, p.seen)
	}
	s := p.steps[0]
	p.steps = p.steps[1:]
	p.seen = append(p.seen, title)
	if s.kind != kind || s.title != "" && !strings.Contains(title, s.title) {
		p.t.Fatalf("got %s %q, expected %s %q", kind, title, s.kind, s.title)
	}
	return s
}
func (p *promptFake) Choose(_ context.Context, title, description string, options []ui.Option, initial string) (string, error) {
	s := p.take("choose", title)
	p.menus = append(p.menus, options)
	p.seen = append(p.seen, description)
	if s.choice == "/back" {
		return "", ui.ErrBack
	}
	if s.choice == "" {
		s.choice = initial
	}
	for _, o := range options {
		if o.Value == s.choice {
			return s.choice, nil
		}
	}
	p.t.Fatalf("choice %q not offered in %q: %+v", s.choice, title, options)
	return "", errors.New("unreachable")
}
func (p *promptFake) Edit(_ context.Context, title string, fields []ui.Field) (map[string]string, error) {
	s := p.take("edit", title)
	p.forms = append(p.forms, fields)
	if s.choice == "/back" {
		return nil, ui.ErrBack
	}
	out := map[string]string{}
	for _, f := range fields {
		v := f.Value
		if changed, ok := s.values[f.Key]; ok {
			v = changed
		}
		if f.Validate != nil {
			if e := f.Validate(v); e != nil {
				return nil, e
			}
		}
		out[f.Key] = v
	}
	return out, nil
}
func (p *promptFake) done() {
	p.t.Helper()
	if len(p.steps) > 0 {
		p.t.Fatalf("unused prompts: %+v", p.steps)
	}
}
func choose(title, value string) promptStep {
	return promptStep{kind: "choose", title: title, choice: value}
}
func edit(title string, values map[string]string) promptStep {
	return promptStep{kind: "edit", title: title, values: values}
}

type desktopFake struct {
	opened, urls  []string
	notifications int
	fail          bool
}

func (d *desktopFake) OpenFolder(_ context.Context, p string) error {
	d.opened = append(d.opened, p)
	return nil
}
func (d *desktopFake) OpenURL(_ context.Context, u string) error {
	d.urls = append(d.urls, u)
	return nil
}
func (d *desktopFake) Notify(context.Context, string, string) error {
	d.notifications++
	if d.fail {
		return fmt.Errorf("no notification service")
	}
	return nil
}

type copyFake struct {
	files   [][]string
	modes   []string
	changed chan struct{}
}

func (c *copyFake) Copy(_ context.Context, files []string, mode string) (desktop.CopyReceipt, error) {
	_, modes, e := desktop.CopyPayload(files, mode)
	if e != nil {
		return desktop.CopyReceipt{}, e
	}
	c.files = append(c.files, append([]string{}, files...))
	c.modes = append(c.modes, mode)
	return desktop.CopyReceipt{Modes: modes, NeedsOwner: c.changed != nil, Changed: c.changed}, nil
}

type clipboardFake struct {
	snapshot input.Snapshot
	reads    int
}

func (c *clipboardFake) Read(context.Context) (input.Snapshot, error) {
	c.reads++
	return c.snapshot, nil
}
func humanSetup(t *testing.T, steps ...promptStep) (*harness, *promptFake, *copyFake, *desktopFake) {
	t.Helper()
	h := setup(t)
	canonical, e := filepath.EvalSymlinks(h.dir)
	if e != nil {
		t.Fatal(e)
	}
	if canonical != h.dir {
		h.source = filepath.Join(canonical, "a.png")
		h.dir = canonical
		h.CWD = canonical
		h.Paths.Config = filepath.Join(canonical, "config.toml")
		h.Paths.Data = filepath.Join(canonical, "data")
		h.Paths.State = filepath.Join(canonical, "state")
		h.Paths.Cache = filepath.Join(canonical, "cache")
		h.Paths.Downloads = filepath.Join(canonical, "Downloads")
	}
	p := &promptFake{t: t, steps: steps}
	copy, actions := &copyFake{}, &desktopFake{}
	h.Prompt = p
	h.Copy = copy
	h.Desktop = actions
	h.Clipboard = &clipboardFake{}
	return h, p, copy, actions
}
func ack(t *testing.T, h *harness) {
	t.Helper()
	if e := state.Write(context.Background(), filepath.Join(h.Paths.State, "setup.json"), state.Setup{SchemaVersion: 1, Acknowledged: true}); e != nil {
		t.Fatal(e)
	}
}
func TestInteractiveFirstUseSharesEngineAndRetainsResult(t *testing.T) {
	h, p, copy, _ := humanSetup(t, choose("setup", "continue"), choose("recipe", "auto"), choose("Result", "done"))
	original, _ := os.ReadFile(h.source)
	if code := h.run(h.source); code != 0 {
		t.Fatal(code, h.out, h.err)
	}
	p.done()
	last, e := state.ReadLast(filepath.Join(h.Paths.State, "last-result.json"))
	if e != nil || last.Report.ExitCode != 0 || len(last.Report.Results) != 1 {
		t.Fatal(last, e)
	}
	out := last.Report.Results[0]
	if out.Output == "" || out.OutputBytes >= out.OriginalBytes {
		t.Fatal(out)
	}
	after, _ := os.ReadFile(h.source)
	if !bytes.Equal(after, original) {
		t.Fatal("human path changed original")
	}
	if len(copy.modes) != 1 || copy.modes[0] != "smart" || copy.files[0][0] != out.Output {
		t.Fatal(copy)
	}
	if len(h.runner.Calls()) == 0 {
		t.Fatal("human path did not invoke engine")
	}
}
func TestInteractiveSecondRecipeStillUsesOriginalClipboard(t *testing.T) {
	h, p, copy, _ := humanSetup(t, choose("recipe", "auto"), choose("Result", "again"), choose("recipe", "web"), choose("Result", "done"))
	ack(t, h)
	original := testutil.Corpus()["alpha-gradient.png"]
	clip := &clipboardFake{snapshot: input.Snapshot{Image: original}}
	h.Clipboard = clip
	if c := h.run(); c != 0 {
		t.Fatal(c, h.err)
	}
	p.done()
	if clip.reads != 1 || len(copy.files) != 2 {
		t.Fatal("clipboard reread after output copy", clip.reads, copy.files)
	}
	for _, call := range h.runner.Calls() {
		if strings.Contains(filepath.Base(call.Invocation.Executable), "pngquant") && len(call.Invocation.Args) > 1 {
			if !bytes.Equal(call.Input, original) {
				t.Fatal("second recipe consumed lossy result rather than original")
			}
		}
	}
}
func TestClipboardPresetImportAndPinnedRecipe(t *testing.T) {
	h, p, _, _ := humanSetup(t, choose("Preset found", "import"), edit("Choose an image", map[string]string{}), choose("recipe", "mcp"), choose("Result", "done"))
	ack(t, h)
	p.steps[1].values = map[string]string{"source": h.source}
	h.Clipboard = &clipboardFake{snapshot: input.Snapshot{Text: "[presets.mcp]\nschema=1\nextends='auto'\ntarget_size='10KiB'\n[presets.mcp.menu]\npinned=true\nposition=2\n"}}
	if c := h.run(); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	p.done()
	if _, e := os.Stat(filepath.Join(h.Paths.Data, "presets.toml")); e != nil {
		t.Fatal(e)
	}
	last, e := state.ReadLast(filepath.Join(h.Paths.State, "last-result.json"))
	if e != nil || last.Report.Results[0].Recipe != "mcp" || last.Report.Results[0].Target == nil || !last.Report.Results[0].Target.Met {
		t.Fatal(last, e)
	}
}
func TestClipboardImagePrecedesPresetAndUnrelatedTextNeverBecomesInput(t *testing.T) {
	h, p, _, _ := humanSetup(t, choose("recipe", "auto"), choose("Result", "done"))
	ack(t, h)
	h.Clipboard = &clipboardFake{snapshot: input.Snapshot{Image: testutil.Corpus()["flat-icon.png"], Text: "[presets.optional]\nschema=1\nextends='auto'\n"}}
	if c := h.run(); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	p.done()
	if _, e := os.Stat(filepath.Join(h.Paths.Data, "presets.toml")); !os.IsNotExist(e) {
		t.Fatal("image interrupted by automatic preset import")
	}
	if len(p.menus) == 0 {
		t.Fatal("no main menu")
	}
	found := false
	for _, o := range p.menus[0] {
		found = found || o.Value == "@import"
	}
	if !found {
		t.Fatal("nonblocking import hint missing")
	}
	h, p, _, _ = humanSetup(t, edit("Choose an image", nil))
	p.steps[0].choice = "/back"
	ack(t, h)
	h.Clipboard = &clipboardFake{snapshot: input.Snapshot{Text: "private-key-do-not-display"}}
	if c := h.run(); c != 0 {
		t.Fatal(c, h.err)
	}
	if strings.Contains(h.out.String(), "private-key") || strings.Contains(strings.Join(p.seen, " "), "private-key") {
		t.Fatal("unrelated clipboard text exposed")
	}
}
func TestInteractiveTargetMissRetainedWithoutOutputAndDryRunDoesNotCopy(t *testing.T) {
	h, p, copy, _ := humanSetup(t, choose("recipe", "@target"), edit("Strict maximum", map[string]string{"target": "1B"}), choose("Result", "done"))
	ack(t, h)
	if c := h.run(h.source); c != 0 {
		t.Fatal(c, h.err)
	}
	p.done()
	last, e := state.ReadLast(filepath.Join(h.Paths.State, "last-result.json"))
	if e != nil || last.Report.ExitCode != 6 || last.Report.Results[0].Output != "" || len(copy.files) != 0 {
		t.Fatal(last, copy, e)
	}
	h, p, copy, _ = humanSetup(t, choose("recipe", "auto"), choose("Result", "done"))
	ack(t, h)
	if c := h.run(h.source, "--dry-run"); c != 0 {
		t.Fatal(c, h.err)
	}
	p.done()
	if len(copy.files) != 0 {
		t.Fatal("dry run copied")
	}
	if _, e := os.Stat(filepath.Join(h.Paths.State, "last-result.json")); !os.IsNotExist(e) {
		t.Fatal("dry run replaced last result")
	}
}
func TestDropJSONNeverPromptsAndLastClearNeverDeletesFiles(t *testing.T) {
	h, p, copy, actions := humanSetup(t)
	if c := h.run("drop", h.source, "--json"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	p.done()
	var result domain.Report
	if e := json.Unmarshal(h.out.Bytes(), &result); e != nil {
		t.Fatal(e)
	}
	if actions.notifications != 1 || len(copy.files) != 1 || result.Results[0].Output == "" {
		t.Fatal(result, copy, actions)
	}
	output := result.Results[0].Output
	if c := h.run("last", "--json"); c != 0 || !json.Valid(h.out.Bytes()) {
		t.Fatal(c, h.out)
	}
	if c := h.run("last", "--clear"); c != 0 {
		t.Fatal(c, h.err)
	}
	if _, e := os.Stat(output); e != nil {
		t.Fatal("clearing metadata deleted artifact")
	}
	if c := h.run("last", "--json"); c != 3 {
		t.Fatal("missing last not identified", c)
	}
}
func TestDropDryRunHasNoClipboardStateOrNotificationEffects(t *testing.T) {
	h, p, copy, actions := humanSetup(t)
	if c := h.run("drop", h.source, "--dry-run", "--json"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	p.done()
	if len(copy.files) != 0 || actions.notifications != 0 {
		t.Fatal("dry-run native effect")
	}
	if _, e := os.Stat(h.Paths.State); !os.IsNotExist(e) {
		t.Fatal("dry-run created state")
	}
	if len(h.runner.Calls()) != 0 {
		t.Fatal("dry-run executed optimizer")
	}
}
func TestDropFailureAndDesktopWarningRemainRecoverable(t *testing.T) {
	h, _, _, actions := humanSetup(t)
	actions.fail = true
	if c := h.run("drop", h.source, "--target-size", "1B", "--json"); c != 6 {
		t.Fatal(c, h.out, h.err)
	}
	last, e := state.ReadLast(filepath.Join(h.Paths.State, "last-result.json"))
	if e != nil || last.Report.ExitCode != 6 || len(last.Warnings) == 0 {
		t.Fatal(last, e)
	}
	if len(state.Outputs(last.Report)) != 0 {
		t.Fatal("target miss output invented")
	}
}
func TestSettingsFormWritesOnlyChangesAndShowsBuiltins(t *testing.T) {
	h, p, _, _ := humanSetup(t, edit("User defaults", map[string]string{"defaults.preset": "web"}))
	os.WriteFile(h.Paths.Config, []byte("# keep my comment\n[defaults]\npreset='auto' # unchanged-note\n"), 0600)
	if c := h.run("settings"); c != 0 {
		t.Fatal(c, h.err)
	}
	p.done()
	raw, _ := os.ReadFile(h.Paths.Config)
	if !bytes.Contains(raw, []byte("# keep my comment")) || !bytes.Contains(raw, []byte("# unchanged-note")) {
		t.Fatal(string(raw))
	}
	if len(p.forms[0][0].Options) < 4 {
		t.Fatal("built-ins missing from settings")
	}
	d, e := config.Parse(raw)
	if e != nil || d.Defaults.Preset == nil || *d.Defaults.Preset != "web" {
		t.Fatal(string(raw), e)
	}
}
func TestNoTerminalAndInvalidStateDoNotPretendInteractiveSuccess(t *testing.T) {
	h := setup(t)
	if c := h.run(h.source); c != 2 {
		t.Fatal("redirected root should direct user to compress", c)
	}
	if _, e := os.Stat(h.Paths.State); !os.IsNotExist(e) {
		t.Fatal("machine root wrote setup")
	}
	h, _, _, _ = humanSetup(t)
	state.Write(context.Background(), filepath.Join(h.Paths.State, "setup.json"), state.Setup{SchemaVersion: 5, Acknowledged: true})
	if c := h.run(h.source); c != 3 {
		t.Fatal("unknown setup version allowed", c, h.err)
	}
	h, _, _, _ = humanSetup(t)
	if c := h.run(h.source, "--stdout"); c != 2 {
		t.Fatal(c)
	}
	if _, e := os.Stat(h.Paths.State); !os.IsNotExist(e) {
		t.Fatal("bad interactive binary mode created state")
	}
}
func TestInteractiveClonesSliceFlagsWithoutLosingLiteralCommas(t *testing.T) {
	base := &compressFlags{}
	c := &cobra.Command{}
	base.bind(c)
	c.Flags().Set("include", "a,b.png")
	c.Flags().Set("include", "*.png")
	c.Flags().Set("quality", "60-85")
	_, f, e := cloneInvocation(c)
	if e != nil {
		t.Fatal(e)
	}
	if len(f.Input.Options.Include) != 2 || f.Input.Options.Include[0] != "a,b.png" || f.Quality != "60-85" {
		t.Fatal(f.Input.Options.Include, f.Quality)
	}
}
func TestSourceLabelOmitsURLTokens(t *testing.T) {
	got := sourceLabel("https://user:secret@example.test/image.png?token=hidden#private")
	if got != "https://example.test/image.png" {
		t.Fatal(got)
	}
}

func TestExplicitMachineCopyDoesNotRetainHistoryOrPrompt(t *testing.T) {
	h, p, copy, actions := humanSetup(t)
	if c := h.run("compress", h.source, "--copy", "path", "--json"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	p.done()
	if len(copy.modes) != 1 || copy.modes[0] != "path" || actions.notifications != 0 || !json.Valid(h.out.Bytes()) {
		t.Fatal(copy, actions, h.out)
	}
	if _, e := os.Stat(h.Paths.State); !os.IsNotExist(e) {
		t.Fatal("canonical compression wrote interactive state")
	}
	if c := h.run("compress", h.source, "--copy", "image", "--stdout"); c != 2 {
		t.Fatal("binary artifact was offered on clipboard", c)
	}
}
func TestExplicitMachineCopyFailureKeepsArtifactAndFailsOperation(t *testing.T) {
	h, _, _, _ := humanSetup(t)
	h.Copy = brokenCopy{}
	if c := h.run("compress", h.source, "--copy", "path", "--json"); c != 5 {
		t.Fatal(c, h.out, h.err)
	}
	var report domain.Report
	json.Unmarshal(h.out.Bytes(), &report)
	if len(report.Results) != 1 || report.Results[0].Output == "" {
		t.Fatal("copy failure concealed saved file", report)
	}
	if _, e := os.Stat(report.Results[0].Output); e != nil {
		t.Fatal(e)
	}
}

type brokenCopy struct{}

func (brokenCopy) Copy(context.Context, []string, string) (desktop.CopyReceipt, error) {
	return desktop.CopyReceipt{}, errors.New("clipboard unavailable")
}
func TestSetupCancellationIsNotAnArgumentFailure(t *testing.T) {
	h, p, _, _ := humanSetup(t, choose("setup", "quit"))
	if c := h.run("setup"); c != 0 {
		t.Fatal(c, h.err)
	}
	p.done()
	if _, e := os.Stat(filepath.Join(h.Paths.State, "setup.json")); !os.IsNotExist(e) {
		t.Fatal("cancelled setup acknowledged")
	}
}
