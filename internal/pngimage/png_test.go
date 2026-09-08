package pngimage_test

import (
	"bytes"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
	"image"
	"image/color"
	"testing"
)

func TestCorpusDecodesIncludingAdam7AndGray16(t *testing.T) {
	for name, b := range testutil.Corpus() {
		t.Run(name, func(t *testing.T) {
			a, e := pngimage.Decode(b, 1_000_000)
			if e != nil {
				t.Fatal(e)
			}
			if name == "interlaced.png" && !a.Info.Interlaced {
				t.Fatal("lost interlace")
			}
			if name == "gray16.png" && a.Info.BitDepth != 16 {
				t.Fatal("lost bit depth")
			}
		})
	}
}
func TestStrictComparisonPreservesHiddenRGB(t *testing.T) {
	a := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	b := image.NewNRGBA(a.Rect)
	a.SetNRGBA(0, 0, color.NRGBA{R: 17, G: 29, B: 41, A: 0})
	b.SetNRGBA(0, 0, color.NRGBA{R: 19, G: 29, B: 41, A: 0})
	if pngimage.PixelEqual(a, b) {
		t.Fatal("hidden RGB ignored")
	}
	hi := image.NewNRGBA64(a.Rect)
	hi.SetNRGBA64(0, 0, color.NRGBA64{R: 17 * 257, G: 29 * 257, B: 41 * 257, A: 0})
	if !pngimage.PixelEqual(a, hi) {
		t.Fatal("equivalent bit depths differ")
	}
}
func TestMetadataAndMalformedInputs(t *testing.T) {
	b := testutil.Corpus()["metadata.png"]
	safe, e := pngimage.Filter(b, "safe")
	if e != nil || bytes.Contains(safe, []byte("tEXt")) || !bytes.Contains(safe, []byte("gAMA")) {
		t.Fatal("metadata policy", e)
	}
	bad := bytes.Clone(b)
	bad[len(bad)-1] ^= 1
	for _, v := range [][]byte{bad, b[:20], append(bytes.Clone(b), 0)} {
		if _, e = pngimage.Decode(v, 1_000_000); e == nil {
			t.Fatal("accepted malformed PNG")
		}
	}
	if _, e = pngimage.Decode(b, 1); e == nil {
		t.Fatal("ignored pixel bound")
	}
}
func TestQuantizerDoesNotRecoverOldICCOverConvertedPixels(t *testing.T) {
	b := testutil.Corpus()["flat-icon.png"]
	src := testutil.AddChunk(b, "iCCP", []byte("profile\x00\x00"))
	dst := testutil.AddChunk(b, "sRGB", []byte{0})
	out, e := pngimage.Reconcile(src, dst, "safe", true)
	if e != nil || bytes.Contains(out, []byte("iCCP")) || !bytes.Contains(out, []byte("sRGB")) {
		t.Fatal("reintroduced old color space", e)
	}
}
func TestBoundedRead(t *testing.T) {
	for _, n := range []int64{4, 9223372036854775807} {
		b, e := pngimage.ReadBounded(bytes.NewBufferString("abcd"), n)
		if e != nil || len(b) != 4 {
			t.Fatal(e)
		}
	}
	if _, e := pngimage.ReadBounded(bytes.NewBufferString("abcde"), 4); e == nil {
		t.Fatal("byte limit ignored")
	}
}
