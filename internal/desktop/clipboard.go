// SPDX-License-Identifier: MIT
package desktop

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/input"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	native "golang.design/x/clipboard"
)

type CopyReceipt struct {
	Modes      []string        `json:"modes"`
	NeedsOwner bool            `json:"needsOwner"`
	Changed    <-chan struct{} `json:"-"`
}
type ClipboardWriter interface {
	Copy(context.Context, []string, string) (CopyReceipt, error)
}
type NativeClipboard struct{}

func (NativeClipboard) Read(ctx context.Context) (input.Snapshot, error) {
	if e := ctx.Err(); e != nil {
		return input.Snapshot{}, e
	}
	if e := native.Init(); e != nil {
		return input.Snapshot{}, e
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	s := input.Snapshot{}
	image, e1 := native.Read(ctx, native.FmtImage)
	files, e2 := native.ReadFiles(ctx)
	text, e3 := native.Read(ctx, native.FmtText)
	if len(image) > 64<<20 || len(text) > 1<<20 || len(files) > 1024 {
		return s, fmt.Errorf("clipboard representation exceeds safe input limit")
	}
	s.Image = image
	s.Files = files
	s.Text = string(text)
	if len(image) == 0 && len(files) == 0 && len(text) == 0 {
		for _, e := range []error{e1, e2, e3} {
			if e != nil && !errors.Is(e, native.ErrUnsupported) {
				return s, e
			}
		}
	}
	return s, nil
}
func ValidCopy(mode string) bool {
	switch mode {
	case "none", "path", "image", "file", "files", "smart", "directory":
		return true
	}
	return false
}

// CopyPayload is a pure policy boundary. It publishes all smart formats in one
// write; repeated single-format writes would erase the earlier representations.
func CopyPayload(files []string, mode string) ([]native.Option, []string, error) {
	if !ValidCopy(mode) {
		return nil, nil, fmt.Errorf("copy must be none, path, image, file, files, smart, or directory")
	}
	if mode == "none" {
		return nil, []string{}, nil
	}
	if len(files) == 0 {
		return nil, nil, fmt.Errorf("no existing output to copy")
	}
	paths := make([]string, 0, len(files))
	for _, f := range files {
		p, e := filepath.Abs(f)
		if e != nil {
			return nil, nil, e
		}
		st, e := os.Stat(p)
		if e != nil || !st.Mode().IsRegular() {
			return nil, nil, fmt.Errorf("output is not a readable file: %s", p)
		}
		paths = append(paths, p)
	}
	if mode == "image" && len(paths) != 1 {
		return nil, nil, fmt.Errorf("image copy requires exactly one output")
	}
	if mode == "directory" {
		return []native.Option{native.Item{Format: native.FmtText, Bytes: []byte(filepath.Dir(paths[0]))}}, []string{"directory-text"}, nil
	}
	options := []native.Option{}
	modes := []string{}
	if mode == "image" || mode == "smart" && len(paths) == 1 {
		f, e := os.Open(paths[0])
		if e != nil {
			return nil, nil, e
		}
		b, e := pngimage.ReadBounded(f, 64<<20)
		f.Close()
		if e != nil {
			return nil, nil, e
		}
		if _, e := pngimage.Decode(b, 64_000_000); e != nil {
			return nil, nil, fmt.Errorf("cannot copy invalid or oversized PNG: %w", e)
		}
		options = append(options, native.Item{Format: native.FmtImage, Bytes: b})
		modes = append(modes, "image")
	}
	if mode == "smart" || mode == "file" || mode == "files" {
		uris := []string{}
		for _, p := range paths {
			uris = append(uris, fileURI(p))
		}
		options = append(options, native.Item{Format: native.FmtFiles, Bytes: []byte(strings.Join(uris, "\r\n") + "\r\n")})
		modes = append(modes, "files")
	}
	if mode != "image" {
		options = append(options, native.Item{Format: native.FmtText, Bytes: []byte(strings.Join(paths, "\n"))})
		modes = append(modes, "path-text")
	}
	return options, modes, nil
}
func fileURI(p string) string {
	p = filepath.ToSlash(p)
	u := url.URL{Scheme: "file"}
	if strings.HasPrefix(p, "//") {
		host, rest, _ := strings.Cut(strings.TrimPrefix(p, "//"), "/")
		u.Host = host
		u.Path = "/" + rest
	} else {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		u.Path = p
	}
	return u.String()
}
func (NativeClipboard) Copy(ctx context.Context, files []string, mode string) (CopyReceipt, error) {
	payload, modes, e := CopyPayload(files, mode)
	if e != nil {
		return CopyReceipt{}, e
	}
	if len(payload) == 0 {
		return CopyReceipt{Modes: modes}, nil
	}
	if e = native.Init(); e != nil {
		return CopyReceipt{}, e
	}
	changed, e := native.WriteAll(ctx, payload...)
	if e != nil {
		return CopyReceipt{}, e
	}
	return CopyReceipt{Modes: modes, NeedsOwner: runtime.GOOS == "linux", Changed: changed}, nil
}

// Hold is used only by the explicitly selected background workflow. It is not a
// daemon: ownership ends when replaced, cancelled or the selected bound expires.
func Hold(ctx context.Context, r CopyReceipt, limit time.Duration) {
	if !r.NeedsOwner || r.Changed == nil || limit <= 0 {
		return
	}
	timer := time.NewTimer(limit)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-r.Changed:
	case <-timer.C:
	}
}
