package testutil

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/jpegimage"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
)

func TestProgressiveSpecimenReallyDecodesTwoScans(t *testing.T) {
	b := ProgressiveFlatJPEG()
	a, e := jpegimage.Decode(b, 1000)
	if e != nil {
		t.Fatal(e)
	}
	scans := 0
	for _, s := range a.Segments {
		if s.Marker == 0xda {
			scans++
		}
	}
	if !a.Info.Progressive || scans != 2 || a.Info.Dimensions.Width != 8 || a.Info.Dimensions.Height != 8 {
		t.Fatal(a.Info, scans)
	}
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			r, g, b, a := a.Image.At(x, y).RGBA()
			if r != 128*257 || g != r || b != r || a != 65535 {
				t.Fatal(r, g, b, a)
			}
		}
	}
}
func TestSyntheticPrivateEXIFIsRealMetadataAndSafeRemovesIt(t *testing.T) {
	exif := PrivateEXIF()
	v, e := jpegimage.Orientation(exif)
	if e != nil || v != 6 {
		t.Fatal(v, e)
	}
	// Follow the actual GPS pointer rather than calling a JPEG comment GPS evidence.
	tiff := exif[6:]
	offset := int(binary.LittleEndian.Uint32(tiff[42:]))
	if binary.LittleEndian.Uint16(tiff[offset:]) != 5 {
		t.Fatal("missing GPS IFD")
	}
	a, e := jpegimage.Decode(JPEGCorpus()["jpeg-private-exif.jpg"], 1_000_000)
	if e != nil {
		t.Fatal(e)
	}
	safe, e := jpegimage.Filter(a, "safe")
	if e != nil {
		t.Fatal(e)
	}
	b, e := jpegimage.Decode(safe, 1_000_000)
	if e != nil {
		t.Fatal(e)
	}
	if b.Orientation != 6 || !pngimage.PixelEqual(a.Image, b.Image) || bytes.Contains(safe, []byte("synthetic camera")) {
		t.Fatal("safe metadata did not preserve only raster/orientation")
	}
	count := 0
	for _, s := range b.Segments {
		if s.Marker == 0xe1 {
			count++
			if !bytes.Equal(s.Data, jpegimage.MinimalEXIF(6)) {
				t.Fatal("private fields survived")
			}
		}
	}
	if count != 1 {
		t.Fatal("orientation carrier absent")
	}
}
func TestAlpha16SpecimenRetainsDepthAndTransparency(t *testing.T) {
	a, e := pngimage.Decode(Alpha16(), 1000)
	if e != nil {
		t.Fatal(e)
	}
	if a.Info.BitDepth != 16 || !a.Info.Transparency {
		t.Fatal(a.Info)
	}
}
