// SPDX-License-Identifier: MIT
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func indexRecord(mode, path string) []byte {
	return []byte(mode + " " + strings.Repeat("0", 40) + " 0\t" + path + "\x00")
}

func TestReleaseInputsCollectOnlyAdmittedRegularFiles(t *testing.T) {
	root := t.TempDir()
	write := func(name, content string, mode os.FileMode) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), mode); err != nil {
			t.Fatal(err)
		}
	}
	write("README.md", "readme", 0o644)
	write("docs/a name.md", "docs", 0o644)
	write("cmd/ignored.go", "ignored", 0o644)
	index := bytes.Join([][]byte{
		indexRecord("100644", "README.md"),
		indexRecord("100755", "docs/a name.md"),
		indexRecord("100644", "cmd/ignored.go"),
	}, nil)
	got, err := releaseInputs(root, index)
	if err != nil {
		t.Fatal(err)
	}
	if string(got["README.md"].Data) != "readme" || string(got["docs/a name.md"].Data) != "docs" {
		t.Fatal("admitted release input bytes differ")
	}
	if _, ok := got["cmd/ignored.go"]; ok || len(got) != 2 {
		t.Fatalf("unexpected release inputs: %#v", got)
	}
}

func TestReleaseInputsRejectTrackedSymlinkOrGitlink(t *testing.T) {
	root := t.TempDir()
	for _, mode := range []string{"120000", "160000"} {
		t.Run(mode, func(t *testing.T) {
			if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "docs", mode), []byte("target"), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := releaseInputs(root, indexRecord(mode, "docs/"+mode))
			if err == nil || !strings.Contains(err.Error(), "unsupported tracked mode") {
				t.Fatalf("expected tracked-mode refusal, got %v", err)
			}
		})
	}
}

func TestReleaseInputsRejectWorkingTreeSymlink(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "docs", "link.md")
	if err := os.Symlink(outside, link); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlink creation unavailable: %v", err)
		}
		t.Fatal(err)
	}
	_, err := releaseInputs(root, indexRecord("100644", "docs/link.md"))
	if err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("expected working-tree symlink refusal, got %v", err)
	}
}

func TestReleaseInputsRejectInvalidOrDuplicatePaths(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("readme"), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, index := range map[string][]byte{
		"invalid":   indexRecord("100644", "../README.md"),
		"duplicate": bytes.Repeat(indexRecord("100644", "README.md"), 2),
		"malformed": []byte("100644 no-tab\x00"),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := releaseInputs(root, index); err == nil {
				t.Fatal("expected release input refusal")
			}
		})
	}
}
