// SPDX-License-Identifier: MIT
// Package testutil contains original specimen generators and process-boundary doubles.
// Generated images are dedicated to CC0-1.0; the generator itself remains MIT.
package testutil

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"github.com/Wolfsblvt/smolpng/internal/toolchain"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func Pattern(w, h int, kind string) *image.NRGBA {
	m := image.NewNRGBA(image.Rect(0, 0, w, h))
	n := uint32(0x534d4f4c)
	for y := range h {
		for x := range w {
			c := color.NRGBA{R: 20, G: 40, B: 70, A: 255}
			switch kind {
			case "flat":
				if x > w/4 && x < w*3/4 && y > h/4 && y < h*3/4 {
					c = color.NRGBA{R: 60, G: 200, B: 170, A: 255}
				}
			case "alpha":
				c = color.NRGBA{R: uint8(x * 255 / max(1, w-1)), G: uint8(y * 255 / max(1, h-1)), B: 137, A: uint8(x * 255 / max(1, w-1))}
			case "hidden":
				c = color.NRGBA{R: uint8(x*17 + 11), G: uint8(y*13 + 29), B: 97, A: 0}
				if x > w/2 {
					c.A = 128
				}
			case "noise":
				n ^= n << 13
				n ^= n >> 17
				n ^= n << 5
				c = color.NRGBA{R: uint8(n), G: uint8(n >> 8), B: uint8(n >> 16), A: 255}
			case "screenshot":
				if y < 8 {
					c = color.NRGBA{R: 30, G: 33, B: 41, A: 255}
				} else if y%8 < 4 && x%6 < 4 {
					c = color.NRGBA{R: 220, G: 225, B: 234, A: 255}
				}
			}
			m.SetNRGBA(x, y, c)
		}
	}
	return m
}
func PNG(m image.Image, level png.CompressionLevel) []byte {
	var b bytes.Buffer
	if e := (&png.Encoder{CompressionLevel: level}).Encode(&b, m); e != nil {
		panic(e)
	}
	return b.Bytes()
}
func AddChunk(b []byte, kind string, data []byte) []byte {
	cs, e := pngimage.Chunks(b)
	if e != nil {
		panic(e)
	}
	out := append([]pngimage.Chunk{cs[0], {Type: kind, Data: data}}, cs[1:]...)
	return pngimage.EncodeChunks(out)
}
func Adam7(m *image.NRGBA) []byte {
	h := make([]byte, 13)
	binary.BigEndian.PutUint32(h, uint32(m.Rect.Dx()))
	binary.BigEndian.PutUint32(h[4:], uint32(m.Rect.Dy()))
	h[8] = 8
	h[9] = 6
	h[12] = 1
	var raw bytes.Buffer
	for _, p := range [][4]int{{0, 0, 8, 8}, {4, 0, 8, 8}, {0, 4, 4, 8}, {2, 0, 4, 4}, {0, 2, 2, 4}, {1, 0, 2, 2}, {0, 1, 1, 2}} {
		if p[0] >= m.Rect.Dx() {
			continue
		}
		for y := p[1]; y < m.Rect.Dy(); y += p[3] {
			raw.WriteByte(0)
			for x := p[0]; x < m.Rect.Dx(); x += p[2] {
				v := m.NRGBAAt(x, y)
				raw.Write([]byte{v.R, v.G, v.B, v.A})
			}
		}
	}
	var zipped bytes.Buffer
	z, _ := zlib.NewWriterLevel(&zipped, zlib.NoCompression)
	_, _ = z.Write(raw.Bytes())
	_ = z.Close()
	return pngimage.EncodeChunks([]pngimage.Chunk{{Type: "IHDR", Data: h}, {Type: "IDAT", Data: zipped.Bytes()}, {Type: "IEND"}})
}
func Corpus() map[string][]byte {
	pal := image.NewPaletted(image.Rect(0, 0, 19, 23), color.Palette{color.NRGBA{R: 81, G: 19, B: 30, A: 0}, color.NRGBA{R: 55, G: 180, B: 140, A: 255}, color.NRGBA{R: 200, G: 30, B: 60, A: 128}})
	for i := range pal.Pix {
		pal.Pix[i] = byte(i % 3)
	}
	gray := image.NewGray16(image.Rect(0, 0, 17, 13))
	for y := range 13 {
		for x := range 17 {
			gray.SetGray16(x, y, color.Gray16{Y: uint16(x*2903 + y*71)})
		}
	}
	meta := PNG(Pattern(32, 32, "flat"), png.NoCompression)
	meta = AddChunk(meta, "gAMA", []byte{0, 0, 177, 143})
	meta = AddChunk(meta, "tEXt", []byte("Comment\x00Generated metadata specimen, not personal data"))
	return map[string][]byte{"flat-icon.png": PNG(Pattern(64, 64, "flat"), png.NoCompression), "alpha-gradient.png": PNG(Pattern(64, 64, "alpha"), png.NoCompression), "hidden-rgb.png": PNG(Pattern(16, 16, "hidden"), png.NoCompression), "synthetic-noise.png": PNG(Pattern(128, 96, "noise"), png.NoCompression), "screenshot.png": PNG(Pattern(96, 48, "screenshot"), png.NoCompression), "indexed.png": PNG(pal, png.NoCompression), "gray16.png": PNG(gray, png.NoCompression), "metadata.png": meta, "interlaced.png": Adam7(Pattern(17, 13, "alpha")), "tiny.png": PNG(Pattern(1, 1, "alpha"), png.BestCompression)}
}

type Call struct {
	Invocation process.Invocation
	Input      []byte
}
type Runner struct {
	mu      sync.Mutex
	calls   []Call
	Handler func(context.Context, process.Invocation) (process.Result, error, bool)
}

func (r *Runner) Run(ctx context.Context, i process.Invocation) (process.Result, error) {
	if e := ctx.Err(); e != nil {
		return process.Result{}, e
	}
	var data []byte
	if len(i.Args) > 1 {
		data, _ = os.ReadFile(i.Args[len(i.Args)-1])
	}
	r.mu.Lock()
	r.calls = append(r.calls, Call{i, bytes.Clone(data)})
	r.mu.Unlock()
	if r.Handler != nil {
		v, e, handled := r.Handler(ctx, i)
		if handled {
			return v, e
		}
	}
	if len(i.Args) == 1 && i.Args[0] == "--version" {
		v := "3.0.3"
		if strings.Contains(filepath.Base(i.Executable), "oxipng") {
			v = "10.2.1"
		}
		return process.Result{Stdout: v}, nil
	}
	if i.Stdout != nil {
		_, e := fmt.Fprint(i.Stdout, strings.Join(i.Args, "|"))
		return process.Result{}, e
	}
	a, e := pngimage.Decode(data, 256_000_000)
	if e != nil {
		return process.Result{ExitCode: 1}, e
	}
	out := PNG(a.Image, png.BestCompression)
	out, e = pngimage.Reconcile(data, out, "keep", false)
	if e != nil {
		return process.Result{}, e
	}
	dest := ""
	for n, s := range i.Args {
		if s == "--out" || s == "--output" {
			dest = i.Args[n+1]
		}
	}
	if dest == "" {
		return process.Result{ExitCode: 2}, fmt.Errorf("missing output argument")
	}
	return process.Result{}, os.WriteFile(dest, out, 0600)
}
func (r *Runner) Calls() []Call {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Call{}, r.calls...)
}
func Tools() toolchain.Set {
	return toolchain.Set{PNGQuant: toolchain.Tool{Name: "pngquant", Path: "pngquant", Available: true, Version: "3.0.3"}, OxiPNG: toolchain.Tool{Name: "oxipng", Path: "oxipng", Available: true, Version: "10.2.1"}}
}
