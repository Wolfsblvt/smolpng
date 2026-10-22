// SPDX-License-Identifier: MIT
package main

import (
	"bytes"
	"fmt"
	"io/fs"
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

// releaseInputs reads exact ordinary-file blobs from the selected Git commit.
// NUL-delimited tree records preserve every valid Git path byte except NUL.
// Rejecting symlink/gitlink modes and non-portable paths before staging, then
// reading committed blobs instead of working-tree paths, binds common package
// content to the same commit named by the release receipt.
func releaseInputs(tree []byte, readBlob func(string) ([]byte, error)) (map[string]packaging.Content, error) {
	result := map[string]packaging.Content{}
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
		if !releaseInputPath(name) {
			continue
		}
		if (meta[0] != "100644" && meta[0] != "100755") || meta[1] != "blob" {
			return nil, fmt.Errorf("release input %q has unsupported tracked mode/type %s %s", name, meta[0], meta[1])
		}
		if _, exists := result[name]; exists {
			return nil, fmt.Errorf("duplicate tracked release path %q", name)
		}
		data, err := readBlob(meta[2])
		if err != nil {
			return nil, fmt.Errorf("read release input %q: %w", name, err)
		}
		result[name] = packaging.Content{Data: data}
	}
	return result, nil
}
