package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"fmt"
	"image/jpeg"
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
	if err := generate(root, false); err != nil {
		t.Fatal(err)
	}
	retained, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(retained, encoded) {
		t.Fatal("regeneration rewrote an equivalent stored PNG", err)
	}
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

func TestCorpusCheckAcceptsICCCompressionButRejectsProfileChanges(t *testing.T) {
	root := t.TempDir()
	if err := generate(root, false); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "png", "color-profile", "icc-srgb.png")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	chunks, err := pngimage.Chunks(original)
	if err != nil {
		t.Fatal(err)
	}
	profileIndex := -1
	for i, chunk := range chunks {
		if chunk.Type == "iCCP" {
			profileIndex = i
		}
	}
	if profileIndex < 0 {
		t.Fatal("fixture has no profile")
	}
	encodeProfile := func(profile []byte) []byte {
		var compressed bytes.Buffer
		writer, err := zlib.NewWriterLevel(&compressed, zlib.NoCompression)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write(profile); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		return append([]byte("smolpng\x00\x00"), compressed.Bytes()...)
	}
	chunks[profileIndex].Data = encodeProfile(testutil.ICCProfile(false))
	reencoded := pngimage.EncodeChunks(chunks)
	if bytes.Equal(original, reencoded) {
		t.Fatal("test requires different encodings of the same profile")
	}
	replaceManifestedFixture(t, root, path, original, reencoded)
	if err := generate(root, true); err != nil {
		t.Fatal("equivalent compressed profile refused:", err)
	}
	if err := generate(root, false); err != nil {
		t.Fatal(err)
	}
	retained, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(retained, reencoded) {
		t.Fatal("regeneration rewrote the equivalent stored profile", err)
	}
	chunks[profileIndex].Data = encodeProfile(testutil.ICCProfile(true))
	changed := pngimage.EncodeChunks(chunks)
	replaceManifestedFixture(t, root, path, reencoded, changed)
	if err := generate(root, true); err == nil {
		t.Fatal("changed ICC colorants admitted")
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
	var m corpusManifest
	if e = toml.Unmarshal(b, &m); e != nil {
		t.Fatal(e)
	}
	if m.Schema != 1 || len(m.Image) != 34 {
		t.Fatalf("unexpected inventory: %d", len(m.Image))
	}
	classes := map[string]bool{}
	generated, photographic := 0, 0
	for _, item := range m.Image {
		if !filepath.IsLocal(item.Path) || item.License != "CC0-1.0" || !item.Redistributable {
			t.Fatalf("unlicensed/unscoped entry: %+v", item)
		}
		b, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(item.Path)))
		if e != nil {
			t.Fatal(e)
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != item.SHA256 {
			t.Fatal("fixture digest differs", item.Path)
		}
		classes[item.Format+"/"+item.Class] = true
		if item.Source != "project-generated" {
			photographic++
			// generate/check above fully validates the photographs. An independent
			// header read checks dimensions here without decoding them a second time.
			config, err := jpeg.DecodeConfig(bytes.NewReader(b))
			if err != nil || item.Class != "photographic" || item.Creator == "" || item.DownloadURL == "" || config.Width != item.Width || config.Height != item.Height {
				t.Fatal("incomplete photographic admission", item.Path, err)
			}
			if strings.HasSuffix(item.Path, "food-refocus.jpg") && (item.Orientation != 3 || item.AppendedBytes != 1799549) {
				t.Fatal("camera attachment specimen lost its original structure", item)
			}
			continue
		}
		generated++
		a, e := imagefile.Inspect(b, 1_000_000)
		if e != nil {
			t.Fatal(item.Path, e)
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != item.SHA256 || a.Format != item.Format || a.Dimensions.Width != item.Width || a.Dimensions.Height != item.Height || a.Transparency != item.Alpha || a.Orientation != item.Orientation || a.BitDepth != item.BitDepth || a.Progressive != item.Progressive {
			t.Fatalf("lying fixture manifest: %s", item.Path)
		}
	}
	if generated != 32 || photographic != 2 {
		t.Fatal("incorrect generated/imported inventory", generated, photographic)
	}
	for _, class := range []string{"png/flat-icon", "png/alpha-16", "png/transparent-hidden-rgb", "png/interlaced", "png/color-profile", "jpeg/exif-orientation", "jpeg/metadata", "jpeg/progressive", "jpeg/color-profile", "jpeg/photographic"} {
		if !classes[class] {
			t.Fatal("missing class", class)
		}
	}
}

func addImportedSpecimen(t *testing.T, root string, item corpusImage, data []byte) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(item.Path))
	if filepath.IsLocal(item.Path) {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	encoded, err := toml.Marshal(struct{ Image []corpusImage }{[]corpusImage{item}})
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(filepath.Join(root, "manifest.toml"), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write(append([]byte("\n"), encoded...))
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		t.Fatal(err, closeErr)
	}
}

func importedSpecimen(data []byte) corpusImage {
	return corpusImage{Path: "jpeg/photographic/imported.jpg", Creator: "Fixture contributor", Source: "https://example.org/source", DownloadURL: "https://example.org/original.jpg", Method: "unmodified publisher original", License: "CC0-1.0", Redistributable: true, Purpose: "camera attachment preservation", SHA256: fmt.Sprintf("%x", sha256.Sum256(data))}
}

func TestImportedMediaRegenerationPreservesAdmissionAndExactBytes(t *testing.T) {
	root := t.TempDir()
	if err := generate(root, false); err != nil {
		t.Fatal(err)
	}
	data := append(testutil.JPEG(testutil.Pattern(7, 5, "flat"), 90), []byte("opaque camera data")...)
	item := importedSpecimen(data)
	addImportedSpecimen(t, root, item, data)
	if err := generate(root, false); err != nil {
		t.Fatal(err)
	}
	if err := generate(root, true); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, filepath.FromSlash(item.Path))
	kept, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(kept, data) {
		t.Fatal("regenerated an imported file", err)
	}
	manifestPath := filepath.Join(root, "manifest.toml")
	originalManifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest corpusManifest
	if err := toml.Unmarshal(originalManifest, &manifest); err != nil {
		t.Fatal(err)
	}
	last := manifest.Image[len(manifest.Image)-1]
	if last.Path != item.Path || last.Creator != item.Creator || last.Source != item.Source || last.SHA256 != item.SHA256 || last.Width != 7 || last.Height != 5 || last.AppendedBytes != int64(len("opaque camera data")) {
		t.Fatal("lost imported provenance/properties", last)
	}
	changed := append(bytes.Clone(data), 0)
	if err := os.WriteFile(path, changed, 0600); err != nil {
		t.Fatal(err)
	}
	if err := generate(root, false); err == nil {
		t.Fatal("regeneration silently adopted modified imported bytes")
	}
	after, _ := os.ReadFile(manifestPath)
	kept, _ = os.ReadFile(path)
	if !bytes.Equal(after, originalManifest) || !bytes.Equal(kept, changed) {
		t.Fatal("failed regeneration modified the imported specimen or its admission")
	}
}

func TestImportedMediaNeedsRightsAndUniqueLocalPath(t *testing.T) {
	data := testutil.JPEG(testutil.Pattern(7, 5, "flat"), 90)
	for _, failure := range []string{"license", "source", "download", "path", "duplicate", "missing"} {
		t.Run(failure, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "manifest.toml"), []byte("schema = 1\n"), 0600); err != nil {
				t.Fatal(err)
			}
			item := importedSpecimen(data)
			switch failure {
			case "license":
				item.License = ""
			case "source":
				item.Source = ""
			case "download":
				item.DownloadURL = ""
			case "path":
				item.Path = "../outside.jpg"
			}
			addImportedSpecimen(t, root, item, data)
			if failure == "duplicate" {
				addImportedSpecimen(t, root, item, data)
			}
			if failure == "missing" {
				if err := os.Remove(filepath.Join(root, filepath.FromSlash(item.Path))); err != nil {
					t.Fatal(err)
				}
			}
			if err := generate(root, false); err == nil {
				t.Fatal("invalid import admitted", failure)
			}
		})
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
