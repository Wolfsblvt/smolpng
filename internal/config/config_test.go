package config

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/paths"
)

type fixture struct {
	options LoadOptions
	root    string
}

func setup(t *testing.T) fixture {
	t.Helper()
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	must(t, os.MkdirAll(filepath.Join(cwd, ".git"), 0700))
	p := paths.Paths{Config: filepath.Join(root, "user", "config.toml"), Data: filepath.Join(root, "data"), State: filepath.Join(root, "state"), Cache: filepath.Join(root, "cache"), Downloads: filepath.Join(root, "Downloads")}
	return fixture{LoadOptions{Paths: p, CWD: cwd}, root}
}
func must(t *testing.T, e error) {
	t.Helper()
	if e != nil {
		t.Fatal(e)
	}
}
func write(t *testing.T, p, s string) {
	t.Helper()
	must(t, os.MkdirAll(filepath.Dir(p), 0700))
	must(t, os.WriteFile(p, []byte(s), 0600))
}
func (f fixture) load(t *testing.T) *Workspace {
	t.Helper()
	w, e := Load(f.options)
	must(t, e)
	return w
}

const validIcon = `[presets.mcp-icon]
schema = 1
label = "MCP icon · ≤10 KiB"
extends = "auto"
formats = ["png"]
target_size = "10KiB"
min_quality = 55
allow_resize = true
min_dimensions = "64x64"
remote = "never"
clipboard = "smart"
[presets.mcp-icon.menu]
pinned = true
position = 2
`

func bundle(t *testing.T, s string) Bundle {
	t.Helper()
	b, e := ParseBundle([]byte(s))
	must(t, e)
	return b
}
func TestConfigurationPrecedenceAndExactFieldProvenance(t *testing.T) {
	f := setup(t)
	write(t, f.options.Paths.Config, `config_version=1
[defaults]
preset="child"
quality="60-90"
recursive=true
[output]
template="{stem}.from-user{ext}"
[presets.parent]
schema=1
quality="80-100"
[presets.child]
schema=1
extends="parent"
min_quality=85
recursive=false
`)
	project := filepath.Join(f.options.CWD, ".smolpng.toml")
	write(t, project, `[defaults]
quality="75-98"
[output]
directory="public/assets"
`)
	w := f.load(t)
	eff, e := w.Resolve("", Fields{})
	must(t, e)
	if eff.Recipe.QualityMin != 85 || eff.Recipe.QualityMax != 100 || eff.Recursive || eff.Output.Template != "{stem}.from-user{ext}" {
		t.Fatalf("wrong inheritance: %+v", eff)
	}
	if eff.Output.Directory != filepath.Join(f.options.CWD, "public", "assets") {
		t.Fatal(eff.Output.Directory)
	}
	if src := eff.Provenance["quality_min"]; src.File != f.options.Paths.Config || src.Key != "presets.child.min_quality" {
		t.Fatal(src)
	}
	if src := eff.Provenance["output.directory"]; src.File != project {
		t.Fatal(src)
	}
	eff, e = w.Resolve("", Fields{Quality: Ptr("50-70"), Recursive: Ptr(true)})
	must(t, e)
	if eff.Recipe.QualityMin != 50 || eff.Recipe.QualityMax != 70 || !eff.Recursive || eff.Provenance["quality_min"].File != "CLI" {
		t.Fatal(eff)
	}
}
func TestZeroAndFalseValuesAreNotMistakenForMissing(t *testing.T) {
	f := setup(t)
	write(t, f.options.Paths.Config, `[defaults]
min_quality=0
optimize_alpha=true
[ui]
animations=false
[presets.no-alpha]
schema=1
optimize_alpha=false
oxipng_level=0
timeout="0s"
`)
	eff, e := f.load(t).Resolve("no-alpha", Fields{})
	must(t, e)
	if eff.Recipe.QualityMin != 0 || eff.Recipe.OptimizeAlpha || eff.Recipe.OxiLevel != 0 || eff.Recipe.Timeout != 0 || eff.Animations {
		t.Fatal(eff)
	}
}
func TestProjectSearchStopsAtRepositoryBoundary(t *testing.T) {
	f := setup(t)
	write(t, filepath.Join(f.root, ".smolpng.toml"), "invalid TOML :::")
	f.options.CWD = filepath.Join(f.options.CWD, "nested")
	must(t, os.MkdirAll(f.options.CWD, 0700))
	w := f.load(t)
	if w.Project.Path != "" {
		t.Fatal("searched above repository", w.Project.Path)
	}
	project := filepath.Join(filepath.Dir(f.options.CWD), ".smolpng.toml")
	write(t, project, "[defaults]\npreset='web'\n")
	w = f.load(t)
	if w.DefaultPreset() != "web" || w.Project.Path != project {
		t.Fatal(w.Project)
	}
	f.options.NoProject = true
	w = f.load(t)
	if w.DefaultPreset() != "auto" {
		t.Fatal("no-project ignored")
	}
}
func TestExplicitConfigMustExistAndRelativePathsUseItsDirectory(t *testing.T) {
	f := setup(t)
	f.options.ConfigPath = "missing.toml"
	if _, e := Load(f.options); e == nil {
		t.Fatal("explicit absence treated as defaults")
	}
	f.options.ConfigPath = "settings/custom.toml"
	path := filepath.Join(f.options.CWD, f.options.ConfigPath)
	write(t, path, "[output]\ndirectory='images'\n[tools]\noxipng='tools/oxipng'\n")
	e, err := f.load(t).Resolve("auto", Fields{})
	must(t, err)
	if e.Output.Directory != filepath.Join(filepath.Dir(path), "images") || e.Tools.OxiPNG != filepath.Join(filepath.Dir(path), "tools", "oxipng") {
		t.Fatal(e)
	}
}
func TestImportIsDataOnlyAndBuiltinsAreImmutable(t *testing.T) {
	for name, s := range map[string]string{
		"script": "[presets.x]\nschema=1\ncommand='echo compromised'", "hook": "[presets.x]\nschema=1\n[ presets.x.hooks ]\npost='command'", "credential": "[presets.x]\nschema=1\napi_key='secret'", "exe": "[presets.x]\nschema=1\npngquant_path='/usr/bin/evil'", "env": "[presets.x]\nschema=1\n[ presets.x.env ]\nHOME='/tmp'", "raw": "[presets.x]\nschema=1\npngquant_args=['--output=x']", "global": "[defaults]\nremote_policy='always'\n[presets.x]\nschema=1", "builtin": "[presets.auto]\nschema=1", "absolute": "[presets.x]\nschema=1\noutput_directory='C:\\user\\secret'", "traversal": "[presets.x]\nschema=1\noutput_directory='../outside'", "unknown-schema": "[presets.x]\nschema=9", "duplicate": "[presets.x]\nschema=1\nschema=1"} {
		t.Run(name, func(t *testing.T) {
			if _, e := ParseBundle([]byte(s)); e == nil {
				t.Fatal("accepted unsafe or ambiguous input")
			}
		})
	}
	if _, e := ParseBundle([]byte("```toml\n" + validIcon + "```")); e != nil {
		t.Fatal(e)
	}
	if _, e := ParseBundle(bytes.Repeat([]byte("x"), MaxDocumentBytes+1)); e == nil {
		t.Fatal("input bound ignored")
	}
}
func TestCyclicAndMissingParentAreRejected(t *testing.T) {
	f := setup(t)
	w := f.load(t)
	for _, s := range []string{"[presets.a]\nschema=1\nextends='b'\n[presets.b]\nschema=1\nextends='a'", "[presets.a]\nschema=1\nextends='missing'"} {
		if _, e := w.PlanImport(bundle(t, s), false); e == nil {
			t.Fatal("broken graph accepted")
		}
	}
}
func TestPresetCannotUpgradeRemotePermission(t *testing.T) {
	f := setup(t)
	write(t, f.options.Paths.Config, "[presets.remote-test]\nschema=1\nengine='tinypng'\nremote='always'\n")
	w := f.load(t)
	e, err := w.Resolve("remote-test", Fields{})
	must(t, err)
	if e.CheckUsable() == nil || len(e.Warnings) == 0 {
		t.Fatal("preset granted upload permission")
	}
	e, err = w.Resolve("remote-test", Fields{Remote: Ptr("always")})
	must(t, err)
	must(t, e.CheckUsable())
	write(t, filepath.Join(f.options.CWD, ".smolpng.toml"), "[defaults]\nremote_policy='always'\n")
	w, err = Load(f.options)
	if err == nil {
		_, err = w.Resolve("auto", Fields{})
	}
	if err == nil {
		t.Fatal("project granted user upload permission")
	}
}
func TestProjectCannotChooseAnExecutable(t *testing.T) {
	f := setup(t)
	write(t, filepath.Join(f.options.CWD, ".smolpng.toml"), "[tools]\noxipng='evil'\n")
	if _, e := Load(f.options); e == nil {
		t.Fatal("project can execute arbitrary path")
	}
}
func TestPreviewImportRemapsInternalParentsAndPreservesConfig(t *testing.T) {
	f := setup(t)
	original := "# my handwritten config\n[defaults]\npreset='auto'\n"
	write(t, f.options.Paths.Config, original)
	w := f.load(t)
	b := bundle(t, "[presets.parent]\nschema=1\ncolors=128\n[presets.child]\nschema=1\nextends='parent'\n")
	p, e := w.PlanImport(b, false)
	must(t, e)
	if _, e = os.Stat(w.StorePath); !os.IsNotExist(e) {
		t.Fatal("preview wrote the store")
	}
	must(t, w.ApplyImport(context.Background(), &p))
	if !p.Applied {
		t.Fatal("apply state missing")
	}
	p, e = w.PlanImport(b, false)
	must(t, e)
	if p.Items[0].Name != "child-2" || p.Items[0].After.Extends != "parent-2" || p.Items[1].Name != "parent-2" {
		t.Fatal(p)
	}
	must(t, w.ApplyImport(context.Background(), &p))
	got, _ := os.ReadFile(f.options.Paths.Config)
	if string(got) != original {
		t.Fatal("import reformatted user config")
	}
	reload := f.load(t)
	resolved, e := reload.Resolve("child-2", Fields{})
	must(t, e)
	if resolved.Recipe.Colors != 128 {
		t.Fatal(resolved)
	}
}
func TestReplacementPreviewAndModifiedSnapshotProtection(t *testing.T) {
	f := setup(t)
	w := f.load(t)
	p, e := w.PlanImport(bundle(t, validIcon), false)
	must(t, e)
	must(t, w.ApplyImport(context.Background(), &p))
	replacement := strings.ReplaceAll(validIcon, "10KiB", "8KiB")
	p, e = w.PlanImport(bundle(t, replacement), true)
	must(t, e)
	if !p.Items[0].Replaces || p.Items[0].Before == nil || *p.Items[0].Before.Target != "10KiB" || *p.Items[0].After.Target != "8KiB" {
		t.Fatal(p)
	}
	write(t, w.StorePath, string(w.StoreBytes)+"\n# changed after preview\n")
	if e = w.ApplyImport(context.Background(), &p); e == nil {
		t.Fatal("stale preview overwrote edited bytes")
	}
	b, _ := os.ReadFile(w.StorePath)
	if !strings.Contains(string(b), "changed after preview") {
		t.Fatal("user change destroyed")
	}
}
func TestExportClosureAndResolvedRoundTrip(t *testing.T) {
	f := setup(t)
	w := f.load(t)
	p, e := w.PlanImport(bundle(t, "[presets.parent]\nschema=1\ncolors=128\n[presets.child]\nschema=1\nextends='parent'\ntarget_size='1KiB'\n"), false)
	must(t, e)
	must(t, w.ApplyImport(context.Background(), &p))
	normal, e := w.Export("child", false)
	must(t, e)
	b := bundle(t, string(normal))
	if len(b.Presets) != 2 || b.Presets["child"].Extends != "parent" {
		t.Fatal(string(normal))
	}
	resolved, e := w.Export("child", true)
	must(t, e)
	b = bundle(t, string(resolved))
	if len(b.Presets) != 1 || b.Presets["child"].Extends != "" {
		t.Fatal(string(resolved))
	}
	other := setup(t).load(t)
	p, e = other.PlanImport(b, false)
	must(t, e)
	must(t, other.ApplyImport(context.Background(), &p))
	before, e := w.Resolve("child", Fields{})
	must(t, e)
	after, e := other.Resolve("child", Fields{})
	must(t, e)
	if !reflect.DeepEqual(before.Recipe, after.Recipe) || before.Jobs != after.Jobs {
		t.Fatalf("resolved export changed recipe\nbefore %+v\nafter %+v", before.Recipe, after.Recipe)
	}
}
func TestMenuPositionAndRemovalProtectInheritance(t *testing.T) {
	w := setup(t).load(t)
	p, e := w.PlanImport(bundle(t, validIcon), false)
	must(t, e)
	must(t, w.ApplyImport(context.Background(), &p))
	menu := w.MenuEntries()
	if menu[0].Name != "auto" || menu[1].Name != "mcp-icon" {
		t.Fatal(menu)
	}
	must(t, w.ChangeMenu(context.Background(), "mcp-icon", Ptr(false), nil))
	if len(w.MenuEntries()) != 4 {
		t.Fatal("unpin ignored")
	}
	if e = w.ChangeMenu(context.Background(), "auto", Ptr(true), nil); e == nil {
		t.Fatal("edited builtin")
	}
	p, e = w.PlanImport(bundle(t, "[presets.dependent]\nschema=1\nextends='mcp-icon'\n"), false)
	must(t, e)
	must(t, w.ApplyImport(context.Background(), &p))
	if e = w.Remove(context.Background(), "mcp-icon"); e == nil {
		t.Fatal("removed parent")
	}
	must(t, w.Remove(context.Background(), "dependent"))
	must(t, w.Remove(context.Background(), "mcp-icon"))
	if len(w.Store.Presets) != 0 {
		t.Fatal(w.Store)
	}
	f, e := os.ReadFile(w.StorePath)
	must(t, e)
	if !IsEmptyStore(f) {
		t.Fatal("empty store cannot reload")
	}
}
func TestImportCannotDependOnTransientProjectPreset(t *testing.T) {
	f := setup(t)
	write(t, filepath.Join(f.options.CWD, ".smolpng.toml"), "[presets.project-only]\nschema=1\n")
	w := f.load(t)
	b := bundle(t, "[presets.imported]\nschema=1\nextends='project-only'\n")
	if _, e := w.PlanImport(b, false); e == nil {
		t.Fatal("import retains inaccessible parent outside this directory")
	}
}

func TestPublishedExamplesParseAndResolve(t *testing.T) {
	files, e := filepath.Glob("../../presets/examples/*.toml")
	must(t, e)
	if len(files) == 0 {
		t.Fatal("no examples found")
	}
	w := setup(t).load(t)
	for _, p := range files {
		t.Run(filepath.Base(p), func(t *testing.T) {
			data, e := os.ReadFile(p)
			must(t, e)
			b, e := ParseBundle(data)
			must(t, e)
			plan, e := w.PlanImport(b, false)
			must(t, e)
			must(t, w.ApplyImport(context.Background(), &plan))
			for _, i := range plan.Items {
				eff, e := w.Resolve(i.Name, Fields{})
				must(t, e)
				must(t, eff.CheckUsable())
			}
		})
	}
}
