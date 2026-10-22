// SPDX-License-Identifier: MIT
package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

// releaseInputs reads only regular files whose tracked index mode is an ordinary
// file. NUL-delimited index records preserve every valid Git path byte except NUL;
// refusing symlinks/gitlinks prevents a clean checkout from importing bytes outside
// the repository into a release archive.
func releaseInputs(root string, index []byte) (map[string]packaging.Content, error) {
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
		if len(meta) != 3 || meta[2] != "0" {
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
		path := filepath.Join(root, filepath.FromSlash(name))
		info, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return nil, fmt.Errorf("release input %q is not a regular file", name)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		result[name] = packaging.Content{Data: data}
	}
	return result, nil
}
