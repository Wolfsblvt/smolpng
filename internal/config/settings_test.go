package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSettingsPreserveHandwrittenCommentsAndUnchangedValues(t *testing.T) {
	original := []byte("# my defaults\nconfig_version=1\n[defaults] # note\npreset = 'auto' # keep this\nquality = '70-95'\n\n[output]\ntemplate='{stem}.smol{ext}'\n")
	b, e := PatchSettings(original, map[string]any{"defaults.preset": "web", "defaults.jobs": 3, "ui.header": "compact", "output.template": "{stem}-small{ext}"})
	if e != nil {
		t.Fatal(e)
	}
	for _, text := range []string{"# my defaults", "[defaults] # note", "# keep this", "quality = '70-95'"} {
		if !strings.Contains(string(b), text) {
			t.Fatalf("lost %q in %s", text, b)
		}
	}
	d, e := Parse(b)
	if e != nil || *d.Defaults.Preset != "web" || *d.Defaults.Jobs != 3 || *d.UI.Header != "compact" || *d.Output.Template != "{stem}-small{ext}" {
		t.Fatal(string(b), e)
	}
}
func TestSettingsHandleDottedInlineAndEmptyConfigWithoutBlindRewrite(t *testing.T) {
	for _, text := range []string{"", "defaults.preset='auto'\n", "defaults = { preset = 'auto' }\n", "[defaults]"} {
		b, e := PatchSettings([]byte(text), map[string]any{"defaults.preset": "web"})
		if e != nil {
			t.Fatalf("%q: %v", text, e)
		}
		d, e := Parse(b)
		if e != nil || d.Defaults.Preset == nil || *d.Defaults.Preset != "web" {
			t.Fatal(string(b), e)
		}
	}
	b, e := PatchSettings([]byte("defaults={preset='auto'}"), map[string]any{"defaults.jobs": 3})
	if e == nil {
		t.Fatalf("missing inline-field addition should be safely rejected, got %s", b)
	}
	if _, e = PatchSettings(nil, map[string]any{"providers.api_key": "secret"}); e == nil {
		t.Fatal("arbitrary secret setting accepted")
	}
}
func TestSettingsSaveChecksSnapshotAndKeepsProjectOverride(t *testing.T) {
	f := setup(t)
	write(t, f.options.Paths.Config, "[defaults]\npreset='auto'\n")
	w := f.load(t)
	if e := w.SaveUserSettings(context.Background(), map[string]any{"defaults.preset": "web"}); e != nil {
		t.Fatal(e)
	}
	data, e := os.ReadFile(w.User.Path)
	if e != nil || !strings.Contains(string(data), "web") {
		t.Fatal(string(data), e)
	}
	if e = os.WriteFile(w.User.Path, append(data, []byte("# changed externally\n")...), 0600); e != nil {
		t.Fatal(e)
	}
	if e = w.SaveUserSettings(context.Background(), map[string]any{"defaults.jobs": 3}); e == nil {
		t.Fatal("stale settings screen replaced new edits")
	}
	if _, e = os.Stat(filepath.Join(filepath.Dir(w.User.Path), "config.toml")); e != nil {
		t.Fatal(e)
	}
}

func TestUserSettingsCannotDependOnProjectOnlyPreset(t *testing.T) {
	f := setup(t)
	write(t, filepath.Join(f.options.CWD, ".smolpng.toml"), "[presets.projectonly]\nschema=1\nextends='auto'\n")
	w := f.load(t)
	if e := w.SaveUserSettings(context.Background(), map[string]any{"defaults.preset": "projectonly"}); e == nil {
		t.Fatal("user default stranded outside project")
	}
	if _, e := os.Stat(w.User.Path); !os.IsNotExist(e) {
		t.Fatal("invalid settings written")
	}
	if e := w.SaveUserSettings(context.Background(), map[string]any{}); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(w.User.Path); !os.IsNotExist(e) {
		t.Fatal("unchanged settings created file")
	}
}
