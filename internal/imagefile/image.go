// SPDX-License-Identifier: MIT
// Package imagefile selects implemented formats, without a plugin registry.
package imagefile

import (
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/jpegimage"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
)

func Format(b []byte) string {
	if jpegimage.Is(b) {
		return "jpeg"
	}
	return "png"
}
func Inspect(b []byte, maxPixels int64) (domain.ImageInfo, error) {
	if jpegimage.Is(b) {
		a, e := jpegimage.Decode(b, maxPixels)
		if e != nil {
			return domain.ImageInfo{}, e
		}
		return a.Info, nil
	}
	a, e := pngimage.Decode(b, maxPixels)
	if e != nil {
		return domain.ImageInfo{}, e
	}
	return a.Info, nil
}
func ClipboardPNG(b []byte, maxPixels int64) ([]byte, error) {
	if jpegimage.Is(b) {
		a, e := jpegimage.Decode(b, maxPixels)
		if e != nil {
			return nil, e
		}
		return jpegimage.ClipboardPNG(a)
	}
	_, e := pngimage.Decode(b, maxPixels)
	return b, e
}
