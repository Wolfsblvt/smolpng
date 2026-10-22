// SPDX-License-Identifier: MIT
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/packaging"
)

const (
	objectA = "1111111111111111111111111111111111111111"
	objectB = "2222222222222222222222222222222222222222"
	objectC = "3333333333333333333333333333333333333333"
)

func treeRecord(mode, kind, object, path string) []byte {
	return []byte(mode + " " + kind + " " + object + "\t" + path + "\x00")
}

func TestCommittedFilesCollectExactBlobs(t *testing.T) {
	blobs := map[string][]byte{objectA: []byte("readme\n\x00bytes"), objectB: []byte("package main\n"), objectC: []byte("ignored")}
	reads := []string{}
	read := func(object string) ([]byte, error) {
		reads = append(reads, object)
		data, ok := blobs[object]
		if !ok {
			return nil, fmt.Errorf("unknown object")
		}
		return data, nil
	}
	tree := bytes.Join([][]byte{
		treeRecord("100644", "blob", objectA, "README.md"),
		treeRecord("100755", "blob", objectB, "cmd/tool/main.go"),
		treeRecord("100644", "blob", objectC, "ignored.txt"),
	}, nil)
	got, err := committedFiles(tree, read)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got["README.md"].Data, blobs[objectA]) || !bytes.Equal(got["cmd/tool/main.go"].Data, blobs[objectB]) {
		t.Fatal("committed source bytes differ from selected blobs")
	}
	if !got["cmd/tool/main.go"].Executable || len(got) != 3 {
		t.Fatalf("unexpected committed source: %#v", got)
	}
	if strings.Join(reads, ",") != objectA+","+objectB+","+objectC {
		t.Fatalf("unexpected blob reads: %v", reads)
	}
	selected := releaseInputs(got)
	if len(selected) != 1 || !bytes.Equal(selected["README.md"].Data, blobs[objectA]) {
		t.Fatalf("unexpected package inputs: %#v", selected)
	}
}

func TestCommittedFilesRejectSpecialOrCollidingPaths(t *testing.T) {
	cases := map[string][]byte{
		"symlink":            treeRecord("120000", "blob", objectA, "docs/link"),
		"gitlink":            treeRecord("160000", "commit", objectA, "nested"),
		"parent path":        treeRecord("100644", "blob", objectA, "../README.md"),
		"backslash":          treeRecord("100644", "blob", objectA, `docs\..\outside`),
		"drive-like":         treeRecord("100644", "blob", objectA, "docs/C:outside"),
		"forbidden wildcard": treeRecord("100644", "blob", objectA, "docs/a?.md"),
		"trailing dot":       treeRecord("100644", "blob", objectA, "docs/name."),
		"trailing space":     treeRecord("100644", "blob", objectA, "docs/name "),
		"reserved device":    treeRecord("100644", "blob", objectA, "docs/con.txt"),
		"line control":       treeRecord("100644", "blob", objectA, "docs/line\nbreak.md"),
		"invalid object":     treeRecord("100644", "blob", "not-an-object", "README.md"),
		"case collision":     bytes.Join([][]byte{treeRecord("100644", "blob", objectA, "README.md"), treeRecord("100644", "blob", objectB, "readme.md")}, nil),
		"duplicate":          bytes.Repeat(treeRecord("100644", "blob", objectA, "README.md"), 2),
		"malformed":          []byte("100644 no-tab\x00"),
	}
	for name, tree := range cases {
		t.Run(name, func(t *testing.T) {
			called := false
			_, err := committedFiles(tree, func(string) ([]byte, error) {
				called = true
				return []byte("data"), nil
			})
			if err == nil {
				t.Fatal("expected committed source refusal")
			}
			if name != "case collision" && name != "duplicate" && called {
				t.Fatal("invalid source reached blob reader")
			}
		})
	}
}

func TestCommittedFilesRejectUnreadableBlob(t *testing.T) {
	_, err := committedFiles(treeRecord("100644", "blob", objectA, "README.md"), func(string) ([]byte, error) {
		return nil, fmt.Errorf("missing blob")
	})
	if err == nil || !strings.Contains(err.Error(), "missing blob") {
		t.Fatalf("expected blob-read refusal, got %v", err)
	}
}

func TestMaterializeCommittedTree(t *testing.T) {
	root := t.TempDir()
	files := map[string]packaging.Content{
		"README.md": {Data: []byte("readme")},
		"cmd/tool":  {Data: []byte("tool"), Executable: true},
	}
	if err := materializeCommittedTree(root, files); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{"README.md": "readme", "cmd/tool": "tool"} {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil || string(data) != want {
			t.Fatalf("materialized %s = %q, %v", name, data, err)
		}
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Join(root, "cmd", "tool"))
		if err != nil || info.Mode().Perm() != 0755 {
			t.Fatalf("executable mode = %v, %v", info, err)
		}
	}
	if err := materializeCommittedTree(root, files); err == nil {
		t.Fatal("expected existing materialized paths to refuse overwrite")
	}
}
