// SPDX-License-Identifier: MIT
package pngimage

import (
	"bytes"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"image"
	"image/color"
	"image/png"
)

// Flatten applies source-over to unassociated channels, with round-to-nearest at
// 16-bit precision. This is encoded-channel compositing, not an ICC conversion,
// object segmentation, color-key eraser, or linear-light renderer.
func Flatten(a *Asset, background, policy string) ([]byte, error) {
	matte, err := domain.ParseBackground(background)
	if err != nil {
		return nil, err
	}
	if a.Info.Animated {
		return nil, fmt.Errorf("cannot flatten animation into one frame")
	}
	if policy != "all" {
		for _, c := range a.Chunks {
			if c.Type == "iCCP" || c.Type == "cICP" || c.Type == "cHRM" {
				return nil, fmt.Errorf("flattening a profiled PNG needs color conversion; use an sRGB export or explicit --strip all")
			}
		}
	}
	b := a.Image.Bounds()
	dest := image.NewNRGBA64(image.Rect(0, 0, b.Dx(), b.Dy()))
	blend := func(v, alpha, bg uint16) uint16 {
		return uint16((uint64(v)*uint64(alpha) + uint64(bg)*uint64(65535-alpha) + 32767) / 65535)
	}
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			c := straight(a.Image.At(b.Min.X+x, b.Min.Y+y))
			dest.SetNRGBA64(x, y, color.NRGBA64{R: blend(c.R, c.A, matte[0]), G: blend(c.G, c.A, matte[1]), B: blend(c.B, c.A, matte[2]), A: 65535})
		}
	}
	var encoded bytes.Buffer
	// Preserve original precision rather than emitting a 16-bit file for 8-bit input.
	var img image.Image = dest
	if a.Info.BitDepth <= 8 {
		small := image.NewNRGBA(dest.Bounds())
		for y := 0; y < b.Dy(); y++ {
			for x := 0; x < b.Dx(); x++ {
				v := dest.NRGBA64At(x, y)
				small.SetNRGBA(x, y, color.NRGBA{R: uint8((uint32(v.R) + 128) / 257), G: uint8((uint32(v.G) + 128) / 257), B: uint8((uint32(v.B) + 128) / 257), A: 255})
			}
		}
		img = small
	}
	if err = (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(&encoded, img); err != nil {
		return nil, err
	}
	return Reconcile(a.Data, encoded.Bytes(), policy, false)
}
