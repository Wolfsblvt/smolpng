// SPDX-License-Identifier: MIT
// Package packaging creates deterministic archives and offline distribution metadata.
// It never publishes, installs dependencies, signs without a supplied key, or mutates DNS.
package packaging

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/updates"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

type Content struct {
	Data       []byte
	Executable bool
}

func safe(name string) bool {
	return name != "" && path.Clean(name) == name && !strings.HasPrefix(name, "/") && !strings.HasPrefix(name, "../") && !strings.ContainsAny(name, "\\:\x00\r\n") && name != ".."
}
func Archive(files map[string]Content, format string, stamp time.Time) ([]byte, error) {
	names := []string{}
	seen := map[string]bool{}
	for n := range files {
		lower := strings.ToLower(n)
		if !safe(n) || seen[lower] {
			return nil, fmt.Errorf("unsafe/duplicate archive name")
		}
		seen[lower] = true
		names = append(names, n)
	}
	sort.Strings(names)
	var out bytes.Buffer
	if format == "zip" {
		w := zip.NewWriter(&out)
		for _, n := range names {
			c := files[n]
			h := zip.FileHeader{Name: n, Method: zip.Deflate}
			h.Modified = stamp.UTC()
			mode := os.FileMode(0644)
			if c.Executable {
				mode = 0755
			}
			h.SetMode(mode)
			f, e := w.CreateHeader(&h)
			if e != nil {
				return nil, e
			}
			if _, e = f.Write(c.Data); e != nil {
				return nil, e
			}
		}
		if e := w.Close(); e != nil {
			return nil, e
		}
	} else if format == "tar.gz" {
		gz := gzip.NewWriter(&out)
		gz.Header.ModTime = time.Time{}
		t := tar.NewWriter(gz)
		for _, n := range names {
			c := files[n]
			mode := int64(0644)
			if c.Executable {
				mode = 0755
			}
			if e := t.WriteHeader(&tar.Header{Name: n, Mode: mode, Size: int64(len(c.Data)), ModTime: stamp.UTC(), Typeflag: tar.TypeReg, Format: tar.FormatUSTAR}); e != nil {
				return nil, e
			}
			if _, e := t.Write(c.Data); e != nil {
				return nil, e
			}
		}
		if e := t.Close(); e != nil {
			return nil, e
		}
		if e := gz.Close(); e != nil {
			return nil, e
		}
	} else {
		return nil, fmt.Errorf("format must be zip or tar.gz")
	}
	return out.Bytes(), nil
}
func UpdateAsset(files map[string]Content, platform, url, format string, stamp time.Time) ([]byte, updates.Asset, error) {
	// Update distribution deliberately excludes configs, image fixtures and docs:
	// it contains only application executables and their license obligations.
	chosen := map[string]Content{}
	a := updates.Asset{URL: url}
	for n, c := range files {
		if n == "LICENSE" || n == "LICENSE-MAP.md" || n == "THIRD_PARTY_NOTICES.md" || strings.HasPrefix(n, "LICENSES/") || c.Executable && strings.HasPrefix(n, "smolpng") && !strings.Contains(n, "/") {
			chosen[n] = c
			a.Files = append(a.Files, updates.File{Name: n, Size: int64(len(c.Data)), SHA256: updates.SHA(c.Data)})
		}
	}
	sort.Slice(a.Files, func(i, j int) bool { return a.Files[i].Name < a.Files[j].Name })
	b, e := Archive(chosen, format, stamp)
	if e != nil {
		return nil, a, e
	}
	a.Size = int64(len(b))
	a.SHA256 = updates.SHA(b)
	if e = updates.ValidateAsset(platform, a); e != nil {
		return nil, a, e
	}
	return b, a, nil
}
func WriteNew(path string, data []byte) error {
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if e != nil {
		return e
	}
	_, e = io.Copy(f, bytes.NewReader(data))
	if e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e != nil {
		return e
	}
	return closeErr
}
