// SPDX-License-Identifier: MIT
// Package pngimage owns static PNG validation, metadata, and exact pixel identity.
package pngimage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"golang.org/x/image/draw"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
)

var signature = []byte{137, 80, 78, 71, 13, 10, 26, 10}

type Chunk struct {
	Type string
	Data []byte
}
type Asset struct {
	Data   []byte
	Image  image.Image
	Info   domain.ImageInfo
	Chunks []Chunk
}

func Chunks(data []byte) ([]Chunk, error) {
	if len(data) < 8 || !bytes.Equal(data[:8], signature) {
		return nil, fmt.Errorf("not a PNG image")
	}
	out := []Chunk{}
	pos := 8
	for pos < len(data) {
		if len(data)-pos < 12 {
			return nil, fmt.Errorf("truncated PNG chunk")
		}
		n := uint64(binary.BigEndian.Uint32(data[pos : pos+4]))
		if n > uint64(len(data)-pos-12) {
			return nil, fmt.Errorf("PNG chunk length exceeds input")
		}
		end := pos + 12 + int(n)
		kind := string(data[pos+4 : pos+8])
		for _, c := range []byte(kind) {
			if c < 'A' || c > 'Z' && c < 'a' || c > 'z' {
				return nil, fmt.Errorf("invalid PNG chunk type")
			}
		}
		if crc32.ChecksumIEEE(data[pos+4:end-4]) != binary.BigEndian.Uint32(data[end-4:end]) {
			return nil, fmt.Errorf("invalid PNG %s checksum", kind)
		}
		if len(out) == 0 && (kind != "IHDR" || n != 13) {
			return nil, fmt.Errorf("PNG must start with a 13-byte IHDR")
		}
		if len(out) > 0 && kind == "IHDR" {
			return nil, fmt.Errorf("duplicate IHDR")
		}
		if kind[0] >= 'A' && kind[0] <= 'Z' && kind != "IHDR" && kind != "PLTE" && kind != "IDAT" && kind != "IEND" {
			return nil, fmt.Errorf("unknown critical PNG chunk")
		}
		out = append(out, Chunk{kind, data[pos+8 : end-4]})
		pos = end
		if kind == "IEND" {
			if n != 0 || pos != len(data) {
				return nil, fmt.Errorf("invalid IEND or trailing data")
			}
			return out, nil
		}
	}
	return nil, fmt.Errorf("missing IEND")
}
func EncodeChunks(cs []Chunk) []byte {
	var b bytes.Buffer
	b.Write(signature)
	for _, c := range cs {
		var n [4]byte
		binary.BigEndian.PutUint32(n[:], uint32(len(c.Data)))
		b.Write(n[:])
		b.WriteString(c.Type)
		b.Write(c.Data)
		crc := crc32.NewIEEE()
		_, _ = crc.Write([]byte(c.Type))
		_, _ = crc.Write(c.Data)
		binary.BigEndian.PutUint32(n[:], crc.Sum32())
		b.Write(n[:])
	}
	return b.Bytes()
}
func Decode(data []byte, maxPixels int64) (*Asset, error) {
	cs, e := Chunks(data)
	if e != nil {
		return nil, e
	}
	h := cs[0].Data
	w, ht := uint64(binary.BigEndian.Uint32(h)), uint64(binary.BigEndian.Uint32(h[4:]))
	if w == 0 || ht == 0 || maxPixels <= 0 || w*ht > uint64(maxPixels) {
		return nil, fmt.Errorf("PNG dimensions %dx%d exceed the %d-pixel limit", w, ht, maxPixels)
	}
	info := domain.ImageInfo{Format: "png", Bytes: int64(len(data)), Dimensions: domain.Dimensions{Width: int(w), Height: int(ht)}, BitDepth: int(h[8]), ColorType: int(h[9]), Alpha: h[9] == 4 || h[9] == 6, Interlaced: h[12] == 1, Chunks: []string{}}
	for _, c := range cs {
		info.Chunks = append(info.Chunks, c.Type)
		if c.Type == "tRNS" {
			info.Alpha = true
		}
		if c.Type == "acTL" {
			info.Animated = true
		}
	}
	m, e := png.Decode(bytes.NewReader(data))
	if e != nil {
		return nil, fmt.Errorf("decode PNG: %w", e)
	}
	if opaque, ok := m.(interface{ Opaque() bool }); ok {
		info.Transparency = !opaque.Opaque()
	} else {
		for y := 0; y < m.Bounds().Dy() && !info.Transparency; y++ {
			for x := 0; x < m.Bounds().Dx(); x++ {
				if straight(m.At(x, y)).A < 65535 {
					info.Transparency = true
					break
				}
			}
		}
	}
	return &Asset{data, m, info, cs}, nil
}
func straight(c color.Color) color.NRGBA64 {
	switch v := c.(type) {
	case color.NRGBA:
		return color.NRGBA64{R: uint16(v.R) * 257, G: uint16(v.G) * 257, B: uint16(v.B) * 257, A: uint16(v.A) * 257}
	case color.NRGBA64:
		return v
	default:
		return color.NRGBA64Model.Convert(c).(color.NRGBA64)
	}
}

// PixelEqual compares unassociated channels, including hidden RGB below alpha zero.
func PixelEqual(a, b image.Image) bool {
	if a.Bounds().Size() != b.Bounds().Size() {
		return false
	}
	for y := 0; y < a.Bounds().Dy(); y++ {
		for x := 0; x < a.Bounds().Dx(); x++ {
			if straight(a.At(a.Bounds().Min.X+x, a.Bounds().Min.Y+y)) != straight(b.At(b.Bounds().Min.X+x, b.Bounds().Min.Y+y)) {
				return false
			}
		}
	}
	return true
}
func structural(s string) bool {
	return s == "IHDR" || s == "PLTE" || s == "tRNS" || s == "IDAT" || s == "IEND"
}
func Filter(data []byte, policy string) ([]byte, error) {
	cs, e := Chunks(data)
	if e != nil {
		return nil, e
	}
	if policy == "keep" {
		return data, nil
	}
	if policy != "safe" && policy != "all" {
		return nil, fmt.Errorf("invalid metadata policy")
	}
	out := []Chunk{}
	for _, c := range cs {
		keep := structural(c.Type)
		if policy == "safe" {
			keep = c.Type != "tEXt" && c.Type != "zTXt" && c.Type != "iTXt" && c.Type != "tIME"
		}
		if keep {
			out = append(out, c)
		}
	}
	return EncodeChunks(out), nil
}

var display = map[string]bool{"gAMA": true, "cHRM": true, "iCCP": true, "sRGB": true, "cICP": true}

// Reconcile never reattaches an old ICC profile to pixels transformed by pngquant.
func Reconcile(original, candidate []byte, policy string, quantized bool) ([]byte, error) {
	source, e := Chunks(original)
	if e != nil {
		return nil, e
	}
	dest, e := Chunks(candidate)
	if e != nil {
		return nil, e
	}
	if policy == "all" {
		return Filter(candidate, policy)
	}
	copyTypes := map[string]bool{"pHYs": true, "eXIf": true}
	if !quantized {
		for k := range display {
			copyTypes[k] = true
		}
		for _, c := range source {
			if c.Type == "iCCP" && (source[0].Data[9] == 0 || source[0].Data[9] == 4) != (dest[0].Data[9] == 0 || dest[0].Data[9] == 4) {
				return nil, fmt.Errorf("ICC color model changed")
			}
		}
	} else {
		a, b := false, false
		for _, c := range source {
			a = a || display[c.Type]
		}
		for _, c := range dest {
			b = b || display[c.Type]
		}
		if a && !b {
			return nil, fmt.Errorf("quantizer dropped color interpretation; candidate rejected")
		}
	}
	if policy == "keep" {
		for _, k := range []string{"tEXt", "zTXt", "iTXt", "tIME"} {
			copyTypes[k] = true
		}
	}
	out := []Chunk{}
	for _, c := range dest {
		if copyTypes[c.Type] {
			continue
		}
		out = append(out, c)
		if c.Type == "IHDR" {
			for _, s := range source {
				if copyTypes[s.Type] {
					out = append(out, s)
				}
			}
		}
	}
	return Filter(EncodeChunks(out), policy)
}
func Resize(a *Asset, d domain.Dimensions, policy string) ([]byte, error) {
	dest := image.NewNRGBA64(image.Rect(0, 0, d.Width, d.Height))
	draw.CatmullRom.Scale(dest, dest.Bounds(), a.Image, a.Image.Bounds(), draw.Src, nil)
	var b bytes.Buffer
	e := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(&b, dest)
	if e != nil {
		return nil, e
	}
	return Reconcile(a.Data, b.Bytes(), policy, false)
}
func Fit(d, box domain.Dimensions) domain.Dimensions {
	if box.Width == 0 || d.Width <= box.Width && d.Height <= box.Height {
		return d
	}
	s := math.Min(float64(box.Width)/float64(d.Width), float64(box.Height)/float64(d.Height))
	return domain.Dimensions{Width: max(1, int(float64(d.Width)*s)), Height: max(1, int(float64(d.Height)*s))}
}
func ReadBounded(r io.Reader, n int64) ([]byte, error) {
	if n <= 0 {
		return nil, fmt.Errorf("byte limit must be positive")
	}
	b, e := io.ReadAll(io.LimitReader(r, n))
	if e != nil {
		return nil, e
	}
	if int64(len(b)) == n {
		var extra [1]byte
		count, tail := io.ReadFull(r, extra[:])
		if count > 0 {
			return nil, fmt.Errorf("input exceeds %d bytes", n)
		}
		if tail != nil && tail != io.EOF {
			return nil, tail
		}
	}
	return b, nil
}
