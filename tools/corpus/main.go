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
		a, e := imagefile.Inspect(data, 1_000_000)
		if e != nil {
			return e
		}
		relative := testutil.CorpusPath(name)
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
		path := filepath.Join(root, filepath.FromSlash(relative))
		if check {
			b, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			if !bytes.Equal(b, data) {
				return fmt.Errorf("fixture changed: %s", path)
			}
		} else {
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
