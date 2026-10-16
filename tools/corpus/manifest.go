// SPDX-License-Identifier: MIT
package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/pelletier/go-toml/v2"
)

type corpusManifest struct {
	Schema int
	Image  []corpusImage
}

type corpusImage struct {
	Path, Format, Class, Creator, Source, Method, License, Purpose, SHA256 string
	Width, Height, Orientation                                             int
	BitDepth                                                               int `toml:"bit_depth"`
	Alpha, Progressive, Redistributable                                    bool
	DownloadURL                                                            string `toml:"download_url"`
	AppendedBytes                                                          int64  `toml:"appended_bytes"`
}

func writeImage(manifest *strings.Builder, item corpusImage, info domain.ImageInfo) {
	fmt.Fprintf(manifest, "\n[[image]]\npath = %q\nformat = %q\nclass = %q\ncreator = %q\nsource = %q\nmethod = %q\nlicense = %q\nredistributable = %t\npurpose = %q\nwidth = %d\nheight = %d\nalpha = %t\nbit_depth = %d\norientation = %d\nprogressive = %t\nsha256 = %q\n", item.Path, info.Format, strings.Split(item.Path, "/")[1], item.Creator, item.Source, item.Method, item.License, item.Redistributable, item.Purpose, info.Dimensions.Width, info.Dimensions.Height, info.Transparency, info.BitDepth, info.Orientation, info.Progressive, item.SHA256)
	if item.DownloadURL != "" {
		fmt.Fprintf(manifest, "download_url = %q\n", item.DownloadURL)
	}
	if info.AppendedBytes > 0 {
		fmt.Fprintf(manifest, "appended_bytes = %d\n", info.AppendedBytes)
	}
}

// The existing manifest is the sole admission source for imported media. Their
// provenance and fixed digests survive regeneration; the generator never fetches
// or rewrites them. Inspectable image properties are derived from actual bytes.
func importedImages(root string) ([]corpusImage, error) {
	data, err := os.ReadFile(filepath.Join(root, "manifest.toml"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var manifest corpusManifest
	if err := toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields().Decode(&manifest); err != nil {
		return nil, fmt.Errorf("corpus admission manifest: %w", err)
	}
	if manifest.Schema != 1 {
		return nil, fmt.Errorf("unsupported corpus manifest schema: %d", manifest.Schema)
	}
	imports := []corpusImage{}
	for _, item := range manifest.Image {
		if item.Source == "project-generated" {
			continue
		}
		parts := strings.Split(item.Path, "/")
		if !fs.ValidPath(item.Path) || !filepath.IsLocal(item.Path) || strings.ContainsAny(item.Path, "\\:") || len(parts) != 3 || (parts[0] != "png" && parts[0] != "jpeg") {
			return nil, fmt.Errorf("imported fixture needs a local format/class/file path: %s", item.Path)
		}
		if item.Creator == "" || item.Purpose == "" || item.Method == "" || !strings.HasPrefix(item.Source, "https://") || !strings.HasPrefix(item.DownloadURL, "https://") || item.License != "CC0-1.0" || !item.Redistributable {
			return nil, fmt.Errorf("imported fixture needs explicit creator, source/download, method, purpose and CC0 redistribution rights: %s", item.Path)
		}
		imports = append(imports, item)
	}
	sort.Slice(imports, func(i, j int) bool { return imports[i].Path < imports[j].Path })
	return imports, nil
}
