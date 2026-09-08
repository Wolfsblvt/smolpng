// SPDX-License-Identifier: MIT
package testutil

import (
	"bytes"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/jpegimage"
	"image"
	"image/color"
	"image/jpeg"
)

func JPEG(m image.Image, quality int) []byte {
	var b bytes.Buffer
	if e := jpeg.Encode(&b, m, &jpeg.Options{Quality: quality}); e != nil {
		panic(e)
	}
	return b.Bytes()
}
func JPEGCorpus() map[string][]byte {
	base := JPEG(Pattern(96, 64, "noise"), 95)
	out := map[string][]byte{"jpeg-baseline.jpg": base, "jpeg-high-chroma.jpg": JPEG(Pattern(128, 96, "noise"), 98)}
	gray := image.NewGray(image.Rect(0, 0, 32, 16))
	for y := range 16 {
		for x := range 32 {
			gray.SetGray(x, y, color.Gray{Y: uint8(x * 8)})
		}
	}
	out["jpeg-grayscale.jpg"] = JPEG(gray, 90)
	for i := 1; i <= 8; i++ {
		meta := jpegimage.Marker(0xe1, append(jpegimage.MinimalEXIF(i), []byte("private metadata fixture")...))
		meta = append(meta, jpegimage.Marker(0xfe, []byte("GPS test annotation (synthetic)"))...)
		out[fmt.Sprintf("jpeg-orientation-%d.jpg", i)] = append(append(bytes.Clone(base[:2]), meta...), base[2:]...)
	}
	return out
}
