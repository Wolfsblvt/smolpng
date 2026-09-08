// SPDX-License-Identifier: MIT
package skill

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"
)

const MaxFiles = 64
const MaxBytes = 5 << 20

type Bundle struct {
	Version string            `json:"version"`
	Hash    string            `json:"hash"`
	Files   map[string][]byte `json:"-"`
	Hashes  map[string]string `json:"files"`
}

func digest(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func validFile(name string) bool {
	return fs.ValidPath(name) && name != "." && !strings.ContainsAny(name, "\\:\x00")
}
func ReadBundle(root fs.FS) (Bundle, error) {
	b := Bundle{Files: map[string][]byte{}, Hashes: map[string]string{}}
	total := 0
	e := fs.WalkDir(root, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !validFile(p) || !d.Type().IsRegular() {
			return fmt.Errorf("invalid Skill bundle path %q", p)
		}
		f, e := root.Open(p)
		if e != nil {
			return e
		}
		data, e := io.ReadAll(io.LimitReader(f, (1<<20)+1))
		f.Close()
		if e != nil {
			return e
		}
		total += len(data)
		if len(data) > 1<<20 || total > MaxBytes || len(b.Files) >= MaxFiles {
			return fmt.Errorf("Skill bundle exceeds resource limits")
		}
		b.Files[p] = data
		b.Hashes[p] = digest(data)
		return nil
	})
	if e != nil {
		return Bundle{}, e
	}
	if len(b.Files["SKILL.md"]) > 256<<10 {
		return Bundle{}, fmt.Errorf("SKILL.md exceeds 256KiB")
	}
	if len(b.Files["SKILL.md"]) == 0 {
		return Bundle{}, fmt.Errorf("Skill bundle needs SKILL.md")
	}
	var content bytes.Buffer
	for _, n := range sortedKeys(b.Hashes) {
		fmt.Fprintf(&content, "%s\x00%s\n", n, b.Hashes[n])
	}
	b.Hash = digest(content.Bytes())
	b.Version = b.Hash[:16]
	return b, nil
}
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func (b Bundle) Archive() ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, name := range sortedKeys(b.Files) {
		if !validFile(name) {
			return nil, fmt.Errorf("invalid resource path")
		}
		h := &zip.FileHeader{Name: path.Join("smolpng", name), Method: zip.Deflate}
		h.SetMode(0644)
		h.SetModTime(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC))
		f, e := w.CreateHeader(h)
		if e != nil {
			return nil, e
		}
		if _, e = f.Write(b.Files[name]); e != nil {
			return nil, e
		}
	}
	if e := w.Close(); e != nil {
		return nil, e
	}
	return buf.Bytes(), nil
}
