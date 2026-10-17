// SPDX-License-Identifier: MIT
package testutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/jpegimage"
)

// CorpusPath maps stable in-memory specimen names to purpose-scoped disk paths.
// Names remain stable for the native qualifier and process-boundary tests.
func CorpusPath(name string) string {
	classes := map[string]string{
		"flat-icon.png": "png/flat-icon", "alpha-gradient.png": "png/alpha-gradient",
		"alpha16.png": "png/alpha-16", "hidden-rgb.png": "png/transparent-hidden-rgb",
		"synthetic-noise.png": "png/high-colour", "screenshot.png": "png/screenshot-text",
		"indexed.png": "png/indexed-palette", "gray16.png": "png/grayscale-16",
		"metadata.png": "png/metadata", "interlaced.png": "png/interlaced", "tiny.png": "png/tiny",
		"jpeg-baseline.jpg": "jpeg/baseline", "jpeg-high-chroma.jpg": "jpeg/high-chroma",
		"jpeg-grayscale.jpg": "jpeg/grayscale", "jpeg-progressive-flat.jpg": "jpeg/progressive",
		"jpeg-private-exif.jpg": "jpeg/metadata",
		"icc-srgb.png":          "png/color-profile", "icc-swapped.png": "png/color-profile",
		"jpeg-icc-srgb.jpg": "jpeg/color-profile", "jpeg-icc-swapped.jpg": "jpeg/color-profile",
	}
	if c, ok := classes[name]; ok {
		return c + "/" + name
	}
	if strings.HasPrefix(name, "jpeg-orientation-") && strings.HasSuffix(name, ".jpg") {
		return "jpeg/exif-orientation/" + name
	}
	panic(fmt.Sprintf("unclassified specimen %q", name))
}

func Alpha16() []byte {
	m := image.NewNRGBA64(image.Rect(0, 0, 19, 13))
	for y := 0; y < 13; y++ {
		for x := 0; x < 19; x++ {
			m.SetNRGBA64(x, y, color.NRGBA64{R: uint16(177 + x*3137), G: uint16(901 + y*4139), B: 49999, A: uint16(x * 65535 / 18)})
		}
	}
	return PNG(m, png.NoCompression)
}

// ProgressiveFlatJPEG is an original, minimal 8x8 grayscale progressive JPEG.
// Two scans encode zero DC difference then AC end-of-block. It exercises actual
// progressive parsing/decoding, not a baseline image relabelled in the header.
// Its constant pixels are not a photographic-quality benchmark.
func ProgressiveFlatJPEG() []byte {
	out := []byte{0xff, 0xd8}
	quant := append([]byte{0}, bytes.Repeat([]byte{1}, 64)...)
	out = append(out, jpegimage.Marker(0xdb, quant)...)
	out = append(out, jpegimage.Marker(0xc2, []byte{8, 0, 8, 0, 8, 1, 1, 0x11, 0})...)
	for _, kind := range []byte{0x00, 0x10} {
		table := make([]byte, 18)
		table[0] = kind
		table[1] = 1 // one one-bit code; symbol 0
		out = append(out, jpegimage.Marker(0xc4, table)...)
	}
	out = append(out, jpegimage.Marker(0xda, []byte{1, 1, 0, 0, 0, 0})...)
	out = append(out, 0x7f) // code zero with one-bit padding
	out = append(out, jpegimage.Marker(0xda, []byte{1, 1, 0, 1, 63, 0})...)
	out = append(out, 0x7f, 0xff, 0xd9)
	return out
}

// PrivateEXIF contains valid TIFF IFD pointers, a synthetic camera maker and GPS
// coordinates at 0/0 (not a person's location). Orientation remains meaningful.
func PrivateEXIF() []byte {
	maker := []byte("smolpng synthetic camera\x00")
	// Header, IFD0 (3 entries), maker, GPS IFD (5 entries), two rational triples.
	gpsOffset := 8 + 2 + 3*12 + 4 + len(maker)
	if gpsOffset%2 != 0 {
		gpsOffset++
	}
	rationals := gpsOffset + 2 + 5*12 + 4
	t := make([]byte, rationals+48)
	copy(t, "II")
	o := binary.LittleEndian
	o.PutUint16(t[2:], 42)
	o.PutUint32(t[4:], 8)
	o.PutUint16(t[8:], 3)
	entry := func(at int, tag, kind uint16, count, value uint32) {
		o.PutUint16(t[at:], tag)
		o.PutUint16(t[at+2:], kind)
		o.PutUint32(t[at+4:], count)
		o.PutUint32(t[at+8:], value)
	}
	entry(10, 0x10f, 2, uint32(len(maker)), 50)
	entry(22, 0x112, 3, 1, 6)
	entry(34, 0x8825, 4, 1, uint32(gpsOffset))
	copy(t[50:], maker)
	o.PutUint16(t[gpsOffset:], 5)
	entry(gpsOffset+2, 0, 1, 4, 0x00000302)
	entry(gpsOffset+14, 1, 2, 2, uint32('N'))
	entry(gpsOffset+26, 2, 5, 3, uint32(rationals))
	entry(gpsOffset+38, 3, 2, 2, uint32('E'))
	entry(gpsOffset+50, 4, 5, 3, uint32(rationals+24))
	for i := 0; i < 6; i++ {
		o.PutUint32(t[rationals+i*8+4:], 1)
	}
	return append([]byte("Exif\x00\x00"), t...)
}
