package pngimage_test

import (
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestBackgroundCompositesPartialAlphaAndPreservesPrecision(t *testing.T) {
	m := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	m.SetNRGBA(0, 0, color.NRGBA{R: 250, G: 100, B: 20, A: 0})
	m.SetNRGBA(1, 0, color.NRGBA{R: 255, A: 128})
	m.SetNRGBA(2, 0, color.NRGBA{R: 20, G: 90, B: 240, A: 255})
	a, err := pngimage.Decode(testutil.PNG(m, png.NoCompression), 100)
	if err != nil {
		t.Fatal(err)
	}
	data, err := pngimage.Flatten(a, "white", "safe")
	if err != nil {
		t.Fatal(err)
	}
	out, err := pngimage.Decode(data, 100)
	if err != nil {
		t.Fatal(err)
	}
	if out.Info.Alpha || out.Info.Transparency || out.Info.BitDepth != 8 {
		t.Fatalf("not opaque RGB8: %+v", out.Info)
	}
	want := []color.NRGBA{{R: 255, G: 255, B: 255, A: 255}, {R: 255, G: 127, B: 127, A: 255}, {R: 20, G: 90, B: 240, A: 255}}
	for x, w := range want {
		got := color.NRGBAModel.Convert(out.Image.At(x, 0))
		if got != w {
			t.Fatalf("pixel %d: %v want %v", x, got, w)
		}
	}
}
func TestProfiledMatteNeedsExplicitColorDisposition(t *testing.T) {
	data := testutil.AddChunk(testutil.PNG(testutil.Pattern(4, 4, "alpha"), png.NoCompression), "iCCP", []byte("test profile"))
	a, e := pngimage.Decode(data, 100)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = pngimage.Flatten(a, "black", "safe"); e == nil {
		t.Fatal("profiled matte silently accepted")
	}
	if _, e = pngimage.Flatten(a, "black", "all"); e != nil {
		t.Fatal(e)
	}
}
