package main

import (
	"crypto/sha256"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/imagefile"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
	"github.com/pelletier/go-toml/v2"
)

func TestCorpusCheckAcceptsPNGEncodingChangesButRejectsChangedSpecimens(t *testing.T) {
	root := t.TempDir()
	if err := generate(root, false); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "png", "flat-icon", "flat-icon.png")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	asset, err := pngimage.Decode(original, 1_000_000)
	if err != nil {
		t.Fatal(err)
	}
	encoded := testutil.PNG(asset.Image, png.BestCompression)
	if string(original) == string(encoded) {
		t.Fatal("test requires different encodings of identical pixels")
	}
	if err := os.WriteFile(path, encoded, 0600); err != nil {
		t.Fatal(err)
	}
	if err := generate(root, true); err == nil {
		t.Fatal("changed bytes without a matching manifest digest admitted")
	}
	replaceManifestedFixture(t, root, path, original, encoded)
	if err := generate(root, true); err != nil {
		t.Fatal("equivalent PNG encoding refused:", err)
	}
	changed := testutil.PNG(testutil.Pattern(asset.Image.Bounds().Dx(), asset.Image.Bounds().Dy(), "noise"), png.BestCompression)
	replaceManifestedFixture(t, root, path, encoded, changed)
	if err := generate(root, true); err == nil {
		t.Fatal("manifested pixel change admitted")
	}
	replaceManifestedFixture(t, root, path, changed, original)
	chunks, err := pngimage.Chunks(original)
	if err != nil {
		t.Fatal(err)
	}
	chunks = append(chunks[:len(chunks)-1], pngimage.Chunk{Type: "tEXt", Data: []byte("Comment\x00changed metadata")}, chunks[len(chunks)-1])
	changed = pngimage.EncodeChunks(chunks)
	replaceManifestedFixture(t, root, path, original, changed)
	if err := generate(root, true); err == nil {
		t.Fatal("manifested metadata change admitted")
	}
}

func replaceManifestedFixture(t *testing.T, root, path string, old, replacement []byte) {
	t.Helper()
	manifestPath := filepath.Join(root, "manifest.toml")
	manifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.Replace(string(manifest), fmt.Sprintf("%x", sha256.Sum256(old)), fmt.Sprintf("%x", sha256.Sum256(replacement)), 1)
	if updated == string(manifest) {
		t.Fatal("fixture digest was not in manifest")
	}
	if err := os.WriteFile(path, replacement, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, []byte(updated), 0600); err != nil {
		t.Fatal(err)
	}
}

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
