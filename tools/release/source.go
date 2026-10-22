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

// releaseInputs reads exact ordinary-file blobs from the Git index. NUL-delimited
// records preserve every valid Git path byte except NUL. Rejecting symlink/gitlink
// modes and never following working-tree paths prevents package content from
// escaping the repository or changing after the clean-tree check.
func releaseInputs(index []byte, readBlob func(string) ([]byte, error)) (map[string]packaging.Content, error) {
	result := map[string]packaging.Content{}
	for _, record := range bytes.Split(index, []byte{0}) {
		if len(record) == 0 {
			continue
		}
		parts := bytes.SplitN(record, []byte{'\t'}, 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid git index record")
		}
		meta := strings.Fields(string(parts[0]))
		if len(meta) != 3 || meta[2] != "0" || !validObjectID(meta[1]) {
			return nil, fmt.Errorf("invalid git index metadata")
		}
		name := string(parts[1])
		if !fs.ValidPath(name) {
			return nil, fmt.Errorf("invalid tracked release path %q", name)
		}
		if !releaseInputPath(name) {
			continue
		}
		if meta[0] != "100644" && meta[0] != "100755" {
			return nil, fmt.Errorf("release input %q has unsupported tracked mode %s", name, meta[0])
		}
		if _, exists := result[name]; exists {
			return nil, fmt.Errorf("duplicate tracked release path %q", name)
		}
		data, err := readBlob(meta[1])
		if err != nil {
			return nil, fmt.Errorf("read release input %q: %w", name, err)
		}
		result[name] = packaging.Content{Data: data}
	}
	return result, nil
}
