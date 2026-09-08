package jpegimage

import (
	"bytes"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

func specimen() []byte {
	m := image.NewNRGBA(image.Rect(0, 0, 17, 11))
	for y := range 11 {
		for x := range 17 {
			m.Set(x, y, color.NRGBA{R: byte(x * 13), G: byte(y * 19), B: byte(x * y), A: 255})
		}
	}
	var b bytes.Buffer
	_ = jpeg.Encode(&b, m, &jpeg.Options{Quality: 97})
	return b.Bytes()
}
func withMetadata(data []byte, orientation int) []byte {
	meta := Marker(0xe1, append(MinimalEXIF(orientation), []byte("private GPS camera thumbnail")...))
	meta = append(meta, Marker(0xfe, []byte("private comment"))...)
	return append(append(bytes.Clone(data[:2]), meta...), data[2:]...)
}
func TestSafePreservesRasterAndOrientationButDropsPrivateBytes(t *testing.T) {
	for orient := 1; orient <= 8; orient++ {
		source := withMetadata(specimen(), orient)
		a, e := Decode(source, 10000)
		if e != nil {
			t.Fatal(e)
		}
		filtered, e := Filter(a, "safe")
		if e != nil {
			t.Fatal(e)
		}
		b, e := Decode(filtered, 10000)
		if e != nil || b.Orientation != orient || !pngimage.PixelEqual(a.Image, b.Image) {
			t.Fatal(orient, e)
		}
		if bytes.Contains(filtered, []byte("private")) {
			t.Fatal("privacy metadata survived")
		}
		keep, e := Filter(a, "keep")
		if e != nil || !bytes.Equal(keep, source) {
			t.Fatal("keep changed source")
		}
		if orient != 1 {
			if _, e := Filter(a, "all"); e == nil {
				t.Fatal("stripped orientation before normalizing")
			}
		}
	}
}
func TestAllEightOrientationMappings(t *testing.T) {
	m := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			m.SetNRGBA(x, y, color.NRGBA{R: byte(y*3 + x + 1), A: 255})
		}
	}
	expected := []string{"123456", "321654", "654321", "456123", "142536", "415263", "635241", "362514"}
	for orientation := 1; orientation <= 8; orientation++ {
		a := &Asset{Image: m, Orientation: orientation, Info: domain.ImageInfo{Dimensions: domain.Dimensions{Width: 3, Height: 2}}}
		u := Upright(a)
		got := ""
		for y := 0; y < u.Bounds().Dy(); y++ {
			for x := 0; x < u.Bounds().Dx(); x++ {
				c := color.NRGBAModel.Convert(u.At(x, y)).(color.NRGBA)
				got += string(rune('0' + c.R))
			}
		}
		if got != expected[orientation-1] {
			t.Fatal(orientation, got)
		}
	}
}
func TestMalformedJPEGAndMetadataAreBounded(t *testing.T) {
	for _, b := range [][]byte{nil, {0xff, 0xd8}, {0xff, 0xd8, 0xff, 0xe1, 0xff, 0xff}, append(specimen(), 0)} {
		if _, e := Decode(b, 1000); e == nil {
			t.Fatal("malformed JPEG accepted")
		}
	}
	if _, e := Decode(specimen(), 1); e == nil {
		t.Fatal("pixel bound ignored")
	}
	bad := MinimalEXIF(6)
	bad[10] = 255
	if _, e := Orientation(bad); e == nil {
		t.Fatal("invalid offset accepted")
	}
}
func TestResizeNormalizesOrientationAndICCIsPreserved(t *testing.T) {
	data := withMetadata(specimen(), 6)
	profile := Marker(0xe2, []byte("ICC_PROFILE\x00\x01\x01test-only-profile"))
	data = append(append(bytes.Clone(data[:2]), profile...), data[2:]...)
	a, e := Decode(data, 10000)
	if e != nil {
		t.Fatal(e)
	}
	filtered, e := Filter(a, "safe")
	if e != nil || !bytes.Contains(filtered, profile) {
		t.Fatal(e)
	}
	b, e := EncodeOriginal(a, domain.Dimensions{Width: 5, Height: 7}, 85, "safe")
	if e != nil {
		t.Fatal(e)
	}
	res, e := Decode(b, 10000)
	if e != nil || res.Orientation != 1 || res.Info.Dimensions != (domain.Dimensions{Width: 5, Height: 7}) || !bytes.Contains(b, profile) {
		t.Fatal(res, e)
	}
}
