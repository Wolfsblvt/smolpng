// SPDX-License-Identifier: MIT
package toolchain

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"path"
	"strings"
)

const maxArchiveBytes int64 = 16 << 20
const maxExpandedBytes int64 = 128 << 20

// ArchiveExecutable extracts only the selected executable from an already
// authenticated archive. Every entry is inspected, not blindly joined to disk.
func ArchiveExecutable(data []byte, format, executable string) ([]byte, error) {
	return archiveMember(data, format, executable)
}

// archiveMember retains the same archive validation for a program or its notices.
func archiveMember(data []byte, format, executable string) ([]byte, error) {
	if len(data) == 0 || int64(len(data)) > maxArchiveBytes {
		return nil, fmt.Errorf("invalid archive size")
	}
	var binary []byte
	var expanded int64
	entries := 0
	readEntry := func(name string, size int64, regular, directory bool, reader io.Reader) error {
		entries++
		if entries > 1024 {
			return fmt.Errorf("archive has too many entries")
		}
		if name == "" || strings.ContainsAny(name, "\\:\x00") || strings.HasPrefix(name, "/") || path.Clean(name) == ".." || strings.HasPrefix(path.Clean(name), "../") {
			return fmt.Errorf("unsafe archive entry %q", name)
		}
		if !regular && !directory {
			return fmt.Errorf("links and special archive entries are not accepted")
		}
		if size < 0 || size > maxArchiveBytes || expanded > maxExpandedBytes-size {
			return fmt.Errorf("archive expands beyond limits")
		}
		expanded += size
		if !regular || path.Base(name) != executable {
			return nil
		}
		if binary != nil {
			return fmt.Errorf("archive contains multiple matching executables")
		}
		b, e := io.ReadAll(io.LimitReader(reader, maxArchiveBytes+1))
		if e != nil {
			return e
		}
		if int64(len(b)) != size || int64(len(b)) > maxArchiveBytes || len(b) == 0 {
			return fmt.Errorf("invalid executable size")
		}
		binary = b
		return nil
	}
	switch format {
	case "zip":
		z, e := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if e != nil {
			return nil, e
		}
		for _, f := range z.File {
			if f.UncompressedSize64 > uint64(maxArchiveBytes) {
				return nil, fmt.Errorf("archive member too large")
			}
			rd, e := f.Open()
			if e != nil {
				return nil, e
			}
			e = readEntry(f.Name, int64(f.UncompressedSize64), f.Mode().IsRegular(), f.FileInfo().IsDir(), rd)
			closeErr := rd.Close()
			if e != nil {
				return nil, e
			}
			if closeErr != nil {
				return nil, closeErr
			}
		}
	case "tar.gz":
		gz, e := gzip.NewReader(bytes.NewReader(data))
		if e != nil {
			return nil, e
		}
		defer gz.Close()
		// Header/data overhead is bounded as well as member sizes.
		limited := &io.LimitedReader{R: gz, N: maxExpandedBytes + (2 << 20)}
		tr := tar.NewReader(limited)
		for {
			h, e := tr.Next()
			if e == io.EOF {
				break
			}
			if e != nil {
				return nil, e
			}
			if e = readEntry(h.Name, h.Size, h.Typeflag == tar.TypeReg || h.Typeflag == tar.TypeRegA, h.Typeflag == tar.TypeDir, tr); e != nil {
				return nil, e
			}
		}
		if limited.N <= 0 {
			return nil, fmt.Errorf("archive stream exceeded limit")
		}
		// Read remaining gzip bytes to verify the trailer and reject hidden expansion.
		if _, e = io.Copy(io.Discard, limited); e != nil {
			return nil, e
		}
		if limited.N <= 0 {
			return nil, fmt.Errorf("archive trailer exceeded limit")
		}
	default:
		return nil, fmt.Errorf("unsupported archive format")
	}
	if binary == nil {
		return nil, fmt.Errorf("archive contains no %s", executable)
	}
	return binary, nil
}
