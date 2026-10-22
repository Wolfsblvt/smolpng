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

func indexRecord(mode, object, path string) []byte {
	return []byte(mode + " " + object + " 0\t" + path + "\x00")
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
	index := bytes.Join([][]byte{
		indexRecord("100644", objectA, "README.md"),
		indexRecord("100755", objectB, "docs/a name.md"),
		indexRecord("100644", objectC, "cmd/ignored.go"),
	}, nil)
	got, err := releaseInputs(index, read)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got["README.md"].Data, blobs[objectA]) || !bytes.Equal(got["docs/a name.md"].Data, blobs[objectB]) {
		t.Fatal("admitted release input bytes differ from indexed blobs")
	}
	if _, ok := got["cmd/ignored.go"]; ok || len(got) != 2 {
		t.Fatalf("unexpected release inputs: %#v", got)
	}
	if strings.Join(reads, ",") != objectA+","+objectB {
		t.Fatalf("unexpected blob reads: %v", reads)
	}
}

func TestReleaseInputsRejectTrackedSymlinkOrGitlink(t *testing.T) {
	for _, mode := range []string{"120000", "160000"} {
		t.Run(mode, func(t *testing.T) {
			called := false
			_, err := releaseInputs(indexRecord(mode, objectA, "docs/"+mode), func(string) ([]byte, error) {
				called = true
				return nil, nil
			})
			if err == nil || !strings.Contains(err.Error(), "unsupported tracked mode") {
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
		index []byte
		read  func(string) ([]byte, error)
	}{
		"invalid path": {index: indexRecord("100644", objectA, "../README.md")},
		"invalid object": {index: indexRecord("100644", "not-an-object", "README.md")},
		"duplicate": {index: bytes.Repeat(indexRecord("100644", objectA, "README.md"), 2)},
		"malformed": {index: []byte("100644 no-tab\x00")},
		"unreadable": {index: indexRecord("100644", objectA, "README.md"), read: func(string) ([]byte, error) { return nil, fmt.Errorf("missing blob") }},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			read := tc.read
			if read == nil {
				read = func(string) ([]byte, error) { return []byte("data"), nil }
			}
			if _, err := releaseInputs(tc.index, read); err == nil {
				t.Fatal("expected release input refusal")
			}
		})
	}
}
