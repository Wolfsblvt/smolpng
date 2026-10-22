// SPDX-License-Identifier: MIT
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/packaging"
)

func releaseInputPath(name string) bool {
	return name == "LICENSE" ||
		name == "README.md" ||
		name == "license-map.md" ||
		name == "third-party-notices.md" ||
		strings.HasPrefix(name, "LICENSES/") ||
		strings.HasPrefix(name, "docs/") ||
		strings.HasPrefix(name, "schemas/") ||
		strings.HasPrefix(name, "presets/examples/") ||
		strings.HasPrefix(name, "testdata/corpus/") ||
		strings.HasPrefix(name, "skills/smolpng/")
}

func validObjectID(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, c := range value {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func portableReleasePath(name string) bool {
	return fs.ValidPath(name) && !strings.ContainsAny(name, "\\:\x00\r\n")
}

// committedFiles reads exact ordinary-file blobs from the selected Git commit.
// NUL-delimited tree records preserve every valid Git path byte except NUL.
// Rejecting symlink/gitlink modes, non-portable paths and case collisions before
// any write keeps both the build source and packaged content inside one source
// identity on Windows, Linux and macOS.
func committedFiles(tree []byte, readBlob func(string) ([]byte, error)) (map[string]packaging.Content, error) {
	result := map[string]packaging.Content{}
	seen := map[string]string{}
	for _, record := range bytes.Split(tree, []byte{0}) {
		if len(record) == 0 {
			continue
		}
		parts := bytes.SplitN(record, []byte{'\t'}, 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid git tree record")
		}
		meta := strings.Fields(string(parts[0]))
		if len(meta) != 3 || !validObjectID(meta[2]) {
			return nil, fmt.Errorf("invalid git tree metadata")
		}
		name := string(parts[1])
		if !portableReleasePath(name) {
			return nil, fmt.Errorf("invalid tracked release path %q", name)
		}
		folded := strings.ToLower(name)
		if previous, exists := seen[folded]; exists {
			return nil, fmt.Errorf("case-colliding tracked release paths %q and %q", previous, name)
		}
		seen[folded] = name
		if (meta[0] != "100644" && meta[0] != "100755") || meta[1] != "blob" {
			return nil, fmt.Errorf("release source %q has unsupported tracked mode/type %s %s", name, meta[0], meta[1])
		}
		data, err := readBlob(meta[2])
		if err != nil {
			return nil, fmt.Errorf("read release source %q: %w", name, err)
		}
		result[name] = packaging.Content{Data: data, Executable: meta[0] == "100755"}
	}
	return result, nil
}

func releaseInputs(committed map[string]packaging.Content) map[string]packaging.Content {
	result := map[string]packaging.Content{}
	for name, content := range committed {
		if releaseInputPath(name) {
			result[name] = content
		}
	}
	return result
}

func materializeCommittedTree(root string, committed map[string]packaging.Content) error {
	names := make([]string, 0, len(committed))
	for name := range committed {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		content := committed[name]
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		mode := os.FileMode(0644)
		if content.Executable {
			mode = 0755
		}
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
		if err != nil {
			return err
		}
		written, writeErr := file.Write(content.Data)
		if writeErr == nil && written != len(content.Data) {
			writeErr = io.ErrShortWrite
		}
		if writeErr == nil {
			writeErr = file.Sync()
		}
		closeErr := file.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
