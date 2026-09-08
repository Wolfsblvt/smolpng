// SPDX-License-Identifier: MIT
package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/imagefile"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	root := flag.String("output", "testdata/corpus", "Output directory")
	check := flag.Bool("check", false, "Verify without writing")
	flag.Parse()
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
	manifest.WriteString("# SPDX-License-Identifier: CC0-1.0\nschema = 1\n")
	for _, name := range names {
		data := specimens[name]
		a, e := imagefile.Inspect(data, 1_000_000)
		must(e)
		fmt.Fprintf(&manifest, "\n[[image]]\npath = %q\ncreator = \"Wolfsblvt Works contributors\"\nsource = \"project-generated\"\nmethod = \"internal/testutil/testutil.go\"\nlicense = \"CC0-1.0\"\nredistributable = true\npurpose = %q\nwidth = %d\nheight = %d\nalpha = %t\nsha256 = \"%x\"\n", name, strings.TrimSuffix(name, ".png"), a.Dimensions.Width, a.Dimensions.Height, a.Transparency, sha256.Sum256(data))
		path := filepath.Join(*root, name)
		if *check {
			b, e := os.ReadFile(path)
			must(e)
			if !bytes.Equal(b, data) {
				must(fmt.Errorf("fixture changed: %s", path))
			}
		} else {
			must(os.MkdirAll(*root, 0755))
			must(os.WriteFile(path, data, 0644))
		}
	}
	path := filepath.Join(*root, "manifest.toml")
	if *check {
		b, e := os.ReadFile(path)
		must(e)
		if string(b) != manifest.String() {
			must(fmt.Errorf("corpus manifest differs"))
		}
	} else {
		must(os.WriteFile(path, []byte(manifest.String()), 0644))
	}
}
func must(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
