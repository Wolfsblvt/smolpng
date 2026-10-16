// SPDX-License-Identifier: MIT
// Package jpegimage owns JPEG structure, orientation and privacy-safe metadata.
package jpegimage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"golang.org/x/image/draw"
)

type Segment struct {
	Marker    byte
	Data, Raw []byte
}
type Asset struct {
	Data []byte
	// Trailer is opaque attached content after the primary JPEG's EOI. Camera
	// formats may contain offsets into the original file, so copying it after
	// rewriting the primary image does not guarantee preservation.
	Trailer     []byte
	Image       image.Image
	Info        domain.ImageInfo
	Segments    []Segment
	Orientation int
	Components  int
}

func Is(data []byte) bool { return len(data) >= 2 && data[0] == 0xff && data[1] == 0xd8 }

// Segments retains entropy bytes verbatim, including byte stuffing and restart
// markers, and handles multiple scans. It stops at the primary EOI, excluding
// opaque appended content, and never interprets JPEG coefficients.
func Segments(data []byte) ([]Segment, error) {
	if !Is(data) {
		return nil, fmt.Errorf("not a JPEG image")
	}
	out := []Segment{{Marker: 0xd8, Raw: data[:2]}}
	pos := 2
	for pos < len(data) {
		if len(out) > 65536 {
			return nil, fmt.Errorf("too many JPEG segments")
		}
		start := pos
		if data[pos] != 0xff {
			return nil, fmt.Errorf("invalid JPEG marker")
		}
		for pos < len(data) && data[pos] == 0xff {
			pos++
		}
		if pos == len(data) {
			return nil, fmt.Errorf("truncated JPEG marker")
		}
		marker := data[pos]
		pos++
		if marker == 0xd9 {
			return append(out, Segment{Marker: marker, Raw: data[start:pos]}), nil
		}
		if marker == 0 || marker == 0xd8 || marker >= 0xd0 && marker <= 0xd7 {
			return nil, fmt.Errorf("unexpected standalone JPEG marker")
		}
		if pos+2 > len(data) {
			return nil, fmt.Errorf("truncated JPEG length")
		}
		n := int(binary.BigEndian.Uint16(data[pos:]))
		if n < 2 || n > len(data)-pos {
			return nil, fmt.Errorf("invalid JPEG segment length")
		}
		out = append(out, Segment{Marker: marker, Data: data[pos+2 : pos+n], Raw: data[start : pos+n]})
		pos += n
		if marker == 0xda {
			start = pos
			for pos < len(data) {
				if data[pos] != 0xff {
					pos++
					continue
				}
				j := pos + 1
				for j < len(data) && data[j] == 0xff {
					j++
				}
				if j == len(data) {
					return nil, fmt.Errorf("unterminated JPEG scan")
				}
				if data[j] == 0 || data[j] >= 0xd0 && data[j] <= 0xd7 {
					pos = j + 1
					continue
				}
				break
			}
			out = append(out, Segment{Raw: data[start:pos]})
		}
	}
	return nil, fmt.Errorf("JPEG has no EOI")
}
func Marker(kind byte, data []byte) []byte {
	b := make([]byte, len(data)+4)
	b[0] = 0xff
	b[1] = kind
	binary.BigEndian.PutUint16(b[2:], uint16(len(data)+2))
	copy(b[4:], data)
	return b
}
func Decode(data []byte, maxPixels int64) (*Asset, error) {
	cs, e := Segments(data)
	if e != nil {
		return nil, e
	}
	primaryEnd := 0
	for _, s := range cs {
		primaryEnd += len(s.Raw)
	}
	primary := data[:primaryEnd]
	config, e := jpeg.DecodeConfig(bytes.NewReader(primary))
	if e != nil {
		return nil, fmt.Errorf("invalid JPEG header")
	}
	d := domain.Dimensions{Width: config.Width, Height: config.Height}
	if d.Area() < 1 || d.Area() > maxPixels {
		return nil, fmt.Errorf("JPEG exceeds the decoded pixel limit")
	}
	a := &Asset{Data: data, Segments: cs, Orientation: 1, Info: domain.ImageInfo{Format: "jpeg", Bytes: int64(len(data)), Dimensions: d, BitDepth: 8, ColorType: 3, Chunks: []string{}}}
	a.Trailer = data[primaryEnd:]
	a.Info.AppendedBytes = int64(len(a.Trailer))
	seenOrientation := false
	for _, s := range cs {
		if s.Marker == 0 {
			continue
		}
		a.Info.Chunks = append(a.Info.Chunks, fmt.Sprintf("FF%02X", s.Marker))
		if s.Marker == 0xc0 || s.Marker == 0xc1 || s.Marker == 0xc2 {
			if len(s.Data) < 6 {
				return nil, fmt.Errorf("short JPEG SOF")
			}
			a.Components = int(s.Data[5])
			a.Info.ColorType = a.Components
			a.Info.Progressive = s.Marker == 0xc2
		}
		if s.Marker == 0xe1 && bytes.HasPrefix(s.Data, []byte("Exif\x00\x00")) {
			v, e := Orientation(s.Data)
			if e != nil {
				return nil, e
			}
			if seenOrientation && v != a.Orientation {
				return nil, fmt.Errorf("conflicting JPEG orientations")
			}
			a.Orientation = v
			seenOrientation = true
		}
	}
	a.Info.Orientation = a.Orientation
	m, e := jpeg.Decode(bytes.NewReader(primary))
	if e != nil {
		return nil, fmt.Errorf("invalid or truncated JPEG scan")
	}
	a.Image = m
	return a, nil
}
func Orientation(exif []byte) (int, error) {
	if len(exif) < 14 {
		return 0, fmt.Errorf("truncated EXIF")
	}
	t := exif[6:]
	var order binary.ByteOrder
	if string(t[:2]) == "II" {
		order = binary.LittleEndian
	} else if string(t[:2]) == "MM" {
		order = binary.BigEndian
	} else {
		return 0, fmt.Errorf("unknown EXIF byte order")
	}
	if order.Uint16(t[2:]) != 42 {
		return 0, fmt.Errorf("invalid EXIF header")
	}
	offset := uint64(order.Uint32(t[4:]))
	if offset+2 > uint64(len(t)) {
		return 0, fmt.Errorf("EXIF IFD outside segment")
	}
	n := uint64(order.Uint16(t[offset:]))
	if offset+2+n*12+4 > uint64(len(t)) {
		return 0, fmt.Errorf("truncated EXIF IFD")
	}
	orientation := 1
	seen := false
	for i := uint64(0); i < n; i++ {
		entry := t[offset+2+i*12:]
		if order.Uint16(entry) != 0x112 {
			continue
		}
		if seen || order.Uint16(entry[2:]) != 3 || order.Uint32(entry[4:]) != 1 {
			return 0, fmt.Errorf("invalid EXIF orientation")
		}
		v := int(order.Uint16(entry[8:]))
		if v < 1 || v > 8 {
			return 0, fmt.Errorf("EXIF orientation must be 1..8")
		}
		orientation = v
		seen = true
	}
	return orientation, nil
}
func MinimalEXIF(orientation int) []byte {
	b := make([]byte, 32)
	copy(b, []byte("Exif\x00\x00II"))
	binary.LittleEndian.PutUint16(b[8:], 42)
	binary.LittleEndian.PutUint32(b[10:], 8)
	binary.LittleEndian.PutUint16(b[14:], 1)
	binary.LittleEndian.PutUint16(b[16:], 0x112)
	binary.LittleEndian.PutUint16(b[18:], 3)
	binary.LittleEndian.PutUint32(b[20:], 1)
	binary.LittleEndian.PutUint16(b[24:], uint16(orientation))
	return b
}
func icc(s Segment) bool {
	return s.Marker == 0xe2 && bytes.HasPrefix(s.Data, []byte("ICC_PROFILE\x00"))
}
func (a *Asset) ICC() []byte {
	var b []byte
	for _, s := range a.Segments {
		if icc(s) {
			b = append(b, s.Raw...)
		}
	}
	return b
}
func structural(s Segment) bool { return !(s.Marker >= 0xe0 && s.Marker <= 0xef || s.Marker == 0xfe) }

// Safe rebuilds EXIF with orientation only: GPS, camera info, thumbnails and
// private pointers cannot survive. Strict lossless therefore preserves raster
// pixels and display orientation without a lossy rotation/re-encode.
func Filter(a *Asset, policy string) ([]byte, error) {
	if policy == "keep" {
		return a.Data, nil
	}
	if policy != "all" && policy != "safe" {
		return nil, fmt.Errorf("invalid JPEG metadata policy")
	}
	if policy == "all" && a.Orientation != 1 {
		return nil, fmt.Errorf("normalize orientation before discarding EXIF")
	}
	var b bytes.Buffer
	for _, s := range a.Segments {
		if structural(s) {
			b.Write(s.Raw)
			if s.Marker == 0xd8 && policy == "safe" && a.Orientation != 1 {
				b.Write(Marker(0xe1, MinimalEXIF(a.Orientation)))
			}
			continue
		}
		if s.Marker == 0xe0 && bytes.HasPrefix(s.Data, []byte("JFIF\x00")) && len(s.Data) >= 14 {
			v := bytes.Clone(s.Data[:14])
			v[12] = 0
			v[13] = 0
			b.Write(Marker(0xe0, v))
		} else if s.Marker == 0xee && bytes.HasPrefix(s.Data, []byte("Adobe")) && len(s.Data) == 12 {
			b.Write(s.Raw)
		} else if policy == "safe" && icc(s) {
			b.Write(s.Raw)
		}
	}
	return b.Bytes(), nil
}
func DisplayDimensions(a *Asset) domain.Dimensions {
	d := a.Info.Dimensions
	if a.Orientation >= 5 {
		d.Width, d.Height = d.Height, d.Width
	}
	return d
}

// Upright produces display-oriented pixels directly from the original decode.
func Upright(a *Asset) image.Image {
	if a.Orientation == 1 {
		return a.Image
	}
	b := a.Image.Bounds()
	d := DisplayDimensions(a)
	m := image.NewNRGBA(image.Rect(0, 0, d.Width, d.Height))
	w, h := b.Dx(), b.Dy()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx, dy := x, y
			switch a.Orientation {
			case 2:
				dx = w - 1 - x
			case 3:
				dx = w - 1 - x
				dy = h - 1 - y
			case 4:
				dy = h - 1 - y
			case 5:
				dx = y
				dy = x
			case 6:
				dx = h - 1 - y
				dy = x
			case 7:
				dx = h - 1 - y
				dy = w - 1 - x
			case 8:
				dx = y
				dy = w - 1 - x
			}
			m.Set(dx, dy, a.Image.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return m
}
func EncodeOriginal(a *Asset, d domain.Dimensions, quality int, policy string) ([]byte, error) {
	if len(a.ICC()) > 0 && a.Components != 3 {
		return nil, fmt.Errorf("resizing a non-RGB ICC JPEG needs a color-managed encoder; refusing to mislabel converted pixels")
	}
	src := Upright(a)
	var img image.Image = src
	if src.Bounds().Dx() != d.Width || src.Bounds().Dy() != d.Height {
		dst := image.NewNRGBA(image.Rect(0, 0, d.Width, d.Height))
		draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
		img = dst
	}
	var b bytes.Buffer
	if e := jpeg.Encode(&b, img, &jpeg.Options{Quality: quality}); e != nil {
		return nil, e
	}
	data := b.Bytes()
	if policy != "all" && len(a.ICC()) > 0 {
		data = append(append(bytes.Clone(data[:2]), a.ICC()...), data[2:]...)
	}
	// Private EXIF cannot be copied after a geometry change. A keep request must
	// choose the untransformed path rather than preserving stale dimensions/thumbnails.
	if policy == "keep" {
		return nil, fmt.Errorf("JPEG resize with keep metadata is unsupported; choose safe or all")
	}
	return data, nil
}
func Reconcile(source, candidate *Asset, policy string, remote bool) ([]byte, error) {
	if policy == "keep" && len(source.Trailer) > 0 {
		if remote {
			return nil, fmt.Errorf("cannot preserve attached JPEG data through remote processing")
		}
		return source.Data, nil
	}
	if source.Orientation != candidate.Orientation {
		return nil, fmt.Errorf("JPEG orientation changed unexpectedly")
	}
	if remote {
		if len(source.ICC()) > 0 && !bytes.Equal(source.ICC(), candidate.ICC()) {
			return nil, fmt.Errorf("remote JPEG changed ICC profile; color-managed verification is required")
		}
		return Filter(candidate, policy)
	}
	if len(source.ICC()) > 0 && source.Components != candidate.Components {
		return nil, fmt.Errorf("JPEG color model changed beneath ICC profile")
	}
	// jpegoptim owns coefficients, smolpng owns preservation. Copy source metadata
	// onto the candidate then apply policy, never importing new private markers.
	var b bytes.Buffer
	for _, s := range candidate.Segments {
		if !structural(s) {
			continue
		}
		b.Write(s.Raw)
		if s.Marker == 0xd8 {
			for _, m := range source.Segments {
				if !structural(m) {
					b.Write(m.Raw)
				}
			}
		}
	}
	a, e := Decode(b.Bytes(), source.Info.Dimensions.Area())
	if e != nil {
		return nil, e
	}
	return Filter(a, policy)
}

// ClipboardPNG is decoded, orientation-correct imagery. With embedded ICC it is
// rejected until a color-managed JPEG-to-PNG path can preserve interpretation.
func ClipboardPNG(a *Asset) ([]byte, error) {
	if len(a.ICC()) > 0 {
		return nil, fmt.Errorf("JPEG has an ICC profile; copy file/path instead of an unmanaged image conversion")
	}
	var b bytes.Buffer
	e := png.Encode(&b, Upright(a))
	return b.Bytes(), e
}
