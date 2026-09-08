// SPDX-License-Identifier: MIT
package updates

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Extract(data []byte, a Asset, platform, dest string) error {
	if int64(len(data)) != a.Size || SHA(data) != a.SHA256 {
		return fmt.Errorf("update archive size/checksum mismatch")
	}
	expected := map[string]File{}
	for _, f := range a.Files {
		expected[f.Name] = f
	}
	seen := map[string]bool{}
	save := func(name string, size int64, reader io.Reader) error {
		f, ok := expected[name]
		if !ok || !allowedFile(name, platform) || seen[name] || f.Size != size {
			return fmt.Errorf("unexpected, duplicate or wrong-sized archive member %q", name)
		}
		b, e := pngimage.ReadBounded(reader, f.Size)
		if e != nil || int64(len(b)) != f.Size || SHA(b) != f.SHA256 {
			return fmt.Errorf("archive member checksum/length mismatch: %s", name)
		}
		p := filepath.Join(dest, filepath.FromSlash(name))
		if e = os.MkdirAll(filepath.Dir(p), 0700); e != nil {
			return e
		}
		out, e := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			return e
		}
		_, e = out.Write(b)
		if e == nil {
			e = out.Sync()
		}
		closeErr := out.Close()
		if e != nil {
			return e
		}
		if closeErr != nil {
			return closeErr
		}
		seen[name] = true
		return nil
	}
	if len(data) >= 4 && string(data[:2]) == "PK" {
		z, e := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if e != nil {
			return e
		}
		for _, f := range z.File {
			if f.FileInfo().IsDir() {
				if f.Name != "LICENSES/" {
					return fmt.Errorf("unexpected update directory")
				}
				continue
			}
			if !f.Mode().IsRegular() || f.UncompressedSize64 > MaxFile {
				return fmt.Errorf("unsafe ZIP member")
			}
			reader, e := f.Open()
			if e != nil {
				return e
			}
			e = save(f.Name, int64(f.UncompressedSize64), reader)
			reader.Close()
			if e != nil {
				return e
			}
		}
	} else {
		gz, e := gzip.NewReader(bytes.NewReader(data))
		if e != nil {
			return fmt.Errorf("update archive must be ZIP or tar.gz")
		}
		defer gz.Close()
		gz.Multistream(false)
		// Bound unpacked headers, padding and content as well as per-member payloads.
		limited := &io.LimitedReader{R: gz, N: MaxArchive + 2<<20}
		t := tar.NewReader(limited)
		for {
			h, e := t.Next()
			if e == io.EOF {
				break
			}
			if e != nil {
				return e
			}
			if h.Typeflag == tar.TypeDir {
				if strings.TrimSuffix(h.Name, "/") != "LICENSES" {
					return fmt.Errorf("unexpected update directory")
				}
				continue
			}
			if h.Typeflag != tar.TypeReg || len(h.PAXRecords) > 0 {
				return fmt.Errorf("links and extended tar entries are not accepted")
			}
			if e = save(h.Name, h.Size, t); e != nil {
				return e
			}
		}
		// Consume the gzip trailer so CRC errors are observable.
		if _, e = io.Copy(io.Discard, limited); e != nil {
			return e
		}
		if limited.N <= 0 {
			return fmt.Errorf("expanded archive exceeds bound")
		}
	}
	if len(seen) != len(expected) {
		return fmt.Errorf("update archive is missing signed members")
	}
	return nil
}
