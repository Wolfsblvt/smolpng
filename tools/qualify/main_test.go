package main

import (
	"image/png"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
)

func TestQualifierValidatesArtifactNotJustSuccessfulExit(t *testing.T) {
	original := testutil.Corpus()["hidden-rgb.png"]
	a, e := pngimage.Decode(original, 1_000_000)
	if e != nil {
		t.Fatal(e)
	}
	output := testutil.PNG(a.Image, png.BestCompression)
	r := domain.Result{OriginalBytes: int64(len(original)), OutputBytes: int64(len(output)), Dimensions: domain.DimensionChange{Before: a.Info.Dimensions, After: a.Info.Dimensions}}
	if errors := validateImage(original, output, r, true); len(errors) != 0 {
		t.Fatal(errors)
	}
	if errors := validateImage(original, []byte("not a PNG"), r, true); len(errors) == 0 {
		t.Fatal("malformed result accepted")
	}
	r.RemoteUsed = true
	r.OutputBytes++
	if errors := validateImage(original, output, r, true); len(errors) != 2 {
		t.Fatal(errors)
	}
}
