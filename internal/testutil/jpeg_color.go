// SPDX-License-Identifier: MIT
package testutil

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"

	"github.com/Wolfsblvt/smolpng/internal/jpegimage"
)

// The profiles are unmodified CC0 Compact-ICC-Profiles inputs. The frozen CMYK
// seed is an original CC0 fixture: Go's JPEG encoder only writes gray/RGB.
// See jpeg-color-inputs.md for exact upstream identities and seed construction.
//
//go:embed profile-gray.icc
var profileGray []byte

//go:embed profile-cmyk.icc
var profileCMYK []byte

//go:embed profile-display-p3.icc
var profileDisplayP3 []byte

//go:embed jpeg-cmyk-seed.jpg
var jpegCMYKSeed []byte

// JPEGColorCorpus supplies real one-, three- and four-component JPEG controls.
// Generation uses frozen inputs and Go alone, never an installed CMS or codec.
func JPEGColorCorpus() map[string][]byte {
	gray := image.NewGray(image.Rect(0, 0, 128, 96))
	rgb := image.NewNRGBA(image.Rect(0, 0, 128, 96))
	for y := 0; y < 96; y++ {
		for x := 0; x < 128; x++ {
			gray.SetGray(x, y, color.Gray{Y: uint8((x*2 + y*3) % 256)})
			rgb.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 2), G: uint8(y * 3), B: uint8(x + y), A: 255})
		}
	}
	profiled := func(base, profile []byte) []byte {
		marker := jpegimage.Marker(0xe2, append([]byte("ICC_PROFILE\x00\x01\x01"), profile...))
		return append(append(bytes.Clone(base[:2]), marker...), base[2:]...)
	}
	return map[string][]byte{
		"jpeg-icc-gray.jpg":       profiled(JPEG(gray, 97), profileGray),
		"jpeg-icc-cmyk.jpg":       profiled(jpegCMYKSeed, profileCMYK),
		"jpeg-icc-display-p3.jpg": profiled(JPEG(rgb, 97), profileDisplayP3),
		"jpeg-cmyk.jpg":           bytes.Clone(jpegCMYKSeed),
	}
}
