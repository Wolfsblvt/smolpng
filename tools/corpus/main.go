// SPDX-License-Identifier: MIT
package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/domain"
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
	imports, err := importedImages(root)
	if err != nil {
		return err
	}
	specimens := testutil.Corpus()
	for n, b := range testutil.JPEGCorpus() {
		specimens[n] = b
	}
	for n, b := range testutil.ICCCorpus() {
		specimens[n] = b
	}
	names := []string{}
	for n := range specimens {
		names = append(names, n)
	}
	sort.Strings(names)
	var manifest strings.Builder
	manifest.WriteString("# SPDX-License-Identifier: CC0-1.0\nschema = 1\n# Generated structural specimens and explicitly admitted original media.\n")
	expected := map[string]bool{}
	for _, name := range names {
		data := specimens[name]
		relative := testutil.CorpusPath(name)
		path := filepath.Join(root, filepath.FromSlash(relative))
		needsWrite := !check
		if stored, err := os.ReadFile(path); err == nil {
			if equivalentFixture(data, stored) {
				// Retain valid checked-in encodings across Go versions, including
				// during regeneration. The manifest binds these exact stored bytes.
				data = stored
				needsWrite = false
			} else if check {
				return fmt.Errorf("fixture changed: %s", path)
			}
		} else if check || !os.IsNotExist(err) {
			return err
		}
		a, e := imagefile.Inspect(data, 1_000_000)
		if e != nil {
			return e
		}
		expected[filepath.FromSlash(relative)] = true
		method := "internal/testutil/testutil.go"
		if a.Format == "jpeg" {
			method = "internal/testutil/jpeg.go"
		}
		if name == "alpha16.png" || name == "jpeg-progressive-flat.jpg" || name == "jpeg-private-exif.jpg" {
			method += " + internal/testutil/corpus.go"
		}
		if strings.HasPrefix(name, "icc-") || strings.HasPrefix(name, "jpeg-icc-") {
			method = "internal/testutil/icc.go + internal/testutil/profile-srgb.icc"
		}
		writeImage(&manifest, corpusImage{Path: relative, Creator: "Wolfsblvt Works contributors", Source: "project-generated", Method: method, License: "CC0-1.0", Redistributable: true, Purpose: strings.TrimSuffix(name, filepath.Ext(name)), SHA256: fmt.Sprintf("%x", sha256.Sum256(data))}, a)
		if needsWrite {
			if e = os.MkdirAll(filepath.Dir(path), 0755); e != nil {
				return e
			}
			if e = os.WriteFile(path, data, 0644); e != nil {
				return e
			}
		}
	}
	for _, item := range imports {
		relative := filepath.FromSlash(item.Path)
		if expected[relative] {
			return fmt.Errorf("duplicate corpus path: %s", item.Path)
		}
		data, err := os.ReadFile(filepath.Join(root, relative))
		if err != nil {
			return err
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != item.SHA256 {
			return fmt.Errorf("imported fixture changed: %s; restore the admitted original instead of regenerating its hash", item.Path)
		}
		info, err := imagefile.Inspect(data, domain.DefaultLimits().MaxPixels)
		if err != nil {
			return fmt.Errorf("imported fixture %s: %w", item.Path, err)
		}
		if info.Format != strings.Split(item.Path, "/")[0] {
			return fmt.Errorf("imported fixture format/path mismatch: %s", item.Path)
		}
		expected[relative] = true
		writeImage(&manifest, item, info)
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
	// Preserve headers, palette, metadata, interlacing and chunk order. Go's
	// DEFLATE encoding may change for both image data and compressed ICC data.
	withoutImageData := func(chunks []pngimage.Chunk) []pngimage.Chunk {
		retained := []pngimage.Chunk{}
		for _, chunk := range chunks {
			if chunk.Type != "IDAT" {
				retained = append(retained, chunk)
			}
		}
		return retained
	}
	left, right := withoutImageData(a.Chunks), withoutImageData(b.Chunks)
	if len(left) != len(right) {
		return false
	}
	for i, chunk := range left {
		other := right[i]
		if chunk.Type != other.Type {
			return false
		}
		if bytes.Equal(chunk.Data, other.Data) {
			continue
		}
		if chunk.Type != "iCCP" || !equivalentICCData(chunk.Data, other.Data) {
			return false
		}
	}
	return true
}

func equivalentICCData(a, b []byte) bool {
	separator := bytes.IndexByte(a, 0)
	if separator < 1 || separator+2 > len(a) || separator+2 > len(b) || a[separator+1] != 0 || !bytes.Equal(a[:separator+2], b[:separator+2]) {
		return false
	}
	decode := func(data []byte) ([]byte, error) {
		reader, err := zlib.NewReader(bytes.NewReader(data[separator+2:]))
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return io.ReadAll(reader)
	}
	left, err := decode(a)
	if err != nil {
		return false
	}
	right, err := decode(b)
	return err == nil && bytes.Equal(left, right)
}
