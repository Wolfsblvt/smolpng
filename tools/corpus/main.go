// SPDX-License-Identifier: MIT
package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/imagefile"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
)

func main() {
	root := flag.String("output", "testdata/corpus", "Output directory")
	check := flag.Bool("check", false, "Verify without writing")
	flag.Parse()
	if e := generate(*root, *check); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func generate(root string, check bool) error {
	specimens := testutil.Corpus()
	for n, b := range testutil.JPEGCorpus() {
		specimens[n] = b
	}
	names := []string{}
	for n := range specimens {
		names = append(names, n)
	}
	sort.Strings(names)
	var manifest strings.Builder
	manifest.WriteString("# SPDX-License-Identifier: CC0-1.0\nschema = 1\n# Synthetic behavior corpus, not perceptual calibration or natural photographs.\n")
	expected := map[string]bool{}
	for _, name := range names {
		data := specimens[name]
		relative := testutil.CorpusPath(name)
		path := filepath.Join(root, filepath.FromSlash(relative))
		if check {
			stored, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !equivalentFixture(data, stored) {
				return fmt.Errorf("fixture changed: %s", path)
			}
			// The manifest binds the checked-in bytes. PNG encoder output can
			// change across Go versions without changing the specimen itself.
			data = stored
		}
		a, e := imagefile.Inspect(data, 1_000_000)
		if e != nil {
			return e
		}
		expected[filepath.FromSlash(relative)] = true
		class := strings.Split(relative, "/")[1]
		method := "internal/testutil/testutil.go"
		if a.Format == "jpeg" {
			method = "internal/testutil/jpeg.go"
		}
		if name == "alpha16.png" || name == "jpeg-progressive-flat.jpg" || name == "jpeg-private-exif.jpg" {
			method += " + internal/testutil/corpus.go"
		}
		fmt.Fprintf(&manifest, "\n[[image]]\npath = %q\nformat = %q\nclass = %q\ncreator = \"Wolfsblvt Works contributors\"\nsource = \"project-generated\"\nmethod = %q\nlicense = \"CC0-1.0\"\nredistributable = true\npurpose = %q\nwidth = %d\nheight = %d\nalpha = %t\nbit_depth = %d\norientation = %d\nprogressive = %t\nsha256 = \"%x\"\n", relative, a.Format, class, method, strings.TrimSuffix(name, filepath.Ext(name)), a.Dimensions.Width, a.Dimensions.Height, a.Transparency, a.BitDepth, a.Orientation, a.Progressive, sha256.Sum256(data))
		if !check {
			if e = os.MkdirAll(filepath.Dir(path), 0755); e != nil {
				return e
			}
			if e = os.WriteFile(path, data, 0644); e != nil {
				return e
			}
		}
	}
	path := filepath.Join(root, "manifest.toml")
	if check {
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		if string(b) != manifest.String() {
			return fmt.Errorf("corpus manifest differs")
		}
		return filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(p))
			if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
				return nil
			}
			rel, e := filepath.Rel(root, p)
			if e != nil {
				return e
			}
			if !expected[rel] {
				return fmt.Errorf("unmanifested image in checked-in corpus: %s", rel)
			}
			return nil
		})
	}
	return os.WriteFile(path, []byte(manifest.String()), 0644)
}

func equivalentFixture(generated, stored []byte) bool {
	if bytes.Equal(generated, stored) {
		return true
	}
	a, err := pngimage.Decode(generated, 1_000_000)
	if err != nil {
		return false
	}
	b, err := pngimage.Decode(stored, 1_000_000)
	if err != nil || !pngimage.PixelEqual(a.Image, b.Image) {
		return false
	}
	// Preserve exact headers, palette, metadata, interlacing and chunk order.
	// Only compression and subdivision of the image-data stream may differ.
	withoutImageData := func(chunks []pngimage.Chunk) []byte {
		retained := []pngimage.Chunk{}
		for _, chunk := range chunks {
			if chunk.Type != "IDAT" {
				retained = append(retained, chunk)
			}
		}
		return pngimage.EncodeChunks(retained)
	}
	return bytes.Equal(withoutImageData(a.Chunks), withoutImageData(b.Chunks))
}
