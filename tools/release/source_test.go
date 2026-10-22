// SPDX-License-Identifier: MIT
package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

const (
	objectA = "1111111111111111111111111111111111111111"
	objectB = "2222222222222222222222222222222222222222"
	objectC = "3333333333333333333333333333333333333333"
)

func treeRecord(mode, kind, object, path string) []byte {
	return []byte(mode + " " + kind + " " + object + "\t" + path + "\x00")
}

func TestReleaseInputsCollectExactAdmittedBlobs(t *testing.T) {
	blobs := map[string][]byte{objectA: []byte("readme\n\x00bytes"), objectB: []byte("docs"), objectC: []byte("ignored")}
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
		treeRecord("100755", "blob", objectB, "docs/a name.md"),
		treeRecord("100644", "blob", objectC, "cmd/ignored.go"),
	}, nil)
	got, err := releaseInputs(tree, read)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got["README.md"].Data, blobs[objectA]) || !bytes.Equal(got["docs/a name.md"].Data, blobs[objectB]) {
		t.Fatal("admitted release input bytes differ from committed blobs")
	}
	if _, ok := got["cmd/ignored.go"]; ok || len(got) != 2 {
		t.Fatalf("unexpected release inputs: %#v", got)
	}
	if strings.Join(reads, ",") != objectA+","+objectB {
		t.Fatalf("unexpected blob reads: %v", reads)
	}
}

func TestReleaseInputsRejectTrackedSymlinkOrGitlink(t *testing.T) {
	for _, tc := range []struct{ mode, kind string }{{"120000", "blob"}, {"160000", "commit"}} {
		t.Run(tc.mode, func(t *testing.T) {
			called := false
			_, err := releaseInputs(treeRecord(tc.mode, tc.kind, objectA, "docs/"+tc.mode), func(string) ([]byte, error) {
				called = true
				return nil, nil
			})
			if err == nil || !strings.Contains(err.Error(), "unsupported tracked mode/type") {
				t.Fatalf("expected tracked-mode refusal, got %v", err)
			}
			if called {
				t.Fatal("unsupported tracked mode reached blob reader")
			}
		})
	}
}

func TestReleaseInputsRejectInvalidDuplicateOrUnreadableRecords(t *testing.T) {
	cases := map[string]struct {
		tree []byte
		read func(string) ([]byte, error)
	}{
		"invalid path":   {tree: treeRecord("100644", "blob", objectA, "../README.md")},
		"invalid object": {tree: treeRecord("100644", "blob", "not-an-object", "README.md")},
		"duplicate":      {tree: bytes.Repeat(treeRecord("100644", "blob", objectA, "README.md"), 2)},
		"malformed":      {tree: []byte("100644 no-tab\x00")},
		"unreadable":     {tree: treeRecord("100644", "blob", objectA, "README.md"), read: func(string) ([]byte, error) { return nil, fmt.Errorf("missing blob") }},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			read := tc.read
			if read == nil {
				read = func(string) ([]byte, error) { return []byte("data"), nil }
			}
			if _, err := releaseInputs(tc.tree, read); err == nil {
				t.Fatal("expected release input refusal")
			}
		})
	}
}
