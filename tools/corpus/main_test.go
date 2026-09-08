package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/imagefile"
	"github.com/pelletier/go-toml/v2"
)

func TestCheckedInCorpusHasExactLicensedStructuralManifest(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "corpus")
	if e := generate(root, true); e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(filepath.Join(root, "manifest.toml"))
	if e != nil {
		t.Fatal(e)
	}
	var m struct {
		Schema int
		Image  []struct {
			Path, Format, Class, Source, License, SHA256 string
			Width, Height, Orientation                   int
			BitDepth                                     int `toml:"bit_depth"`
			Alpha, Progressive, Redistributable          bool
		}
	}
	if e = toml.Unmarshal(b, &m); e != nil {
		t.Fatal(e)
	}
	if m.Schema != 1 || len(m.Image) != 24 {
		t.Fatalf("unexpected inventory: %d", len(m.Image))
	}
	classes := map[string]bool{}
	for _, item := range m.Image {
		if !filepath.IsLocal(item.Path) || item.Source != "project-generated" || item.License != "CC0-1.0" || !item.Redistributable {
			t.Fatalf("unlicensed/unscoped entry: %+v", item)
		}
		b, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(item.Path)))
		if e != nil {
			t.Fatal(e)
		}
		a, e := imagefile.Inspect(b, 1_000_000)
		if e != nil {
			t.Fatal(item.Path, e)
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != item.SHA256 || a.Format != item.Format || a.Dimensions.Width != item.Width || a.Dimensions.Height != item.Height || a.Transparency != item.Alpha || a.Orientation != item.Orientation || a.BitDepth != item.BitDepth || a.Progressive != item.Progressive {
			t.Fatalf("lying fixture manifest: %s", item.Path)
		}
		classes[item.Format+"/"+item.Class] = true
	}
	for _, class := range []string{"png/flat-icon", "png/alpha-16", "png/transparent-hidden-rgb", "png/interlaced", "jpeg/exif-orientation", "jpeg/metadata", "jpeg/progressive"} {
		if !classes[class] {
			t.Fatal("missing class", class)
		}
	}
}
func TestCorpusCheckRejectsExtraOrModifiedImages(t *testing.T) {
	root := t.TempDir()
	if e := generate(root, false); e != nil {
		t.Fatal(e)
	}
	extra := filepath.Join(root, "unlicensed.jpg")
	if e := os.WriteFile(extra, []byte("no"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := generate(root, true); e == nil {
		t.Fatal("unmanifested input admitted")
	}
	os.Remove(extra)
	target := filepath.Join(root, "png", "flat-icon", "flat-icon.png")
	if e := os.WriteFile(target, []byte("modified"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := generate(root, true); e == nil {
		t.Fatal("fixture corruption ignored")
	}
}
