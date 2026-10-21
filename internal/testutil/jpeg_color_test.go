package testutil

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/jpegimage"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
)

func TestJPEGColorModelsSurviveMetadataCleanup(t *testing.T) {
	components := map[string]int{"jpeg-icc-gray.jpg": 1, "jpeg-icc-cmyk.jpg": 4, "jpeg-icc-display-p3.jpg": 3, "jpeg-cmyk.jpg": 4}
	for name, data := range JPEGColorCorpus() {
		t.Run(name, func(t *testing.T) {
			before, err := jpegimage.Decode(data, 1_000_000)
			if err != nil {
				t.Fatal(err)
			}
			if before.Components != components[name] || (len(before.ICC()) > 0) != strings.Contains(name, "icc-") {
				t.Fatal("fixture lost its real color model or profile")
			}
			for _, policy := range []string{"safe", "all"} {
				filtered, err := jpegimage.Filter(before, policy)
				if err != nil {
					t.Fatal(err)
				}
				after, err := jpegimage.Decode(filtered, 1_000_000)
				if err != nil || after.Components != before.Components || !pngimage.PixelEqual(before.Image, after.Image) {
					t.Fatal(policy, "changed decoded pixels or color model", err)
				}
				if policy == "safe" && !bytes.Equal(before.ICC(), after.ICC()) || policy == "all" && len(after.ICC()) != 0 {
					t.Fatal(policy, "violated the selected ICC policy")
				}
				reconciled, err := jpegimage.Reconcile(before, after, "safe", false)
				if err != nil {
					t.Fatal(err)
				}
				joined, err := jpegimage.Decode(reconciled, 1_000_000)
				if err != nil || !bytes.Equal(before.ICC(), joined.ICC()) || !pngimage.PixelEqual(before.Image, joined.Image) {
					t.Fatal("local reconciliation lost profile or pixels", err)
				}
			}
		})
	}
}

func TestJPEGResizeKeepsProfileInterpretationOrRefuses(t *testing.T) {
	for name, data := range JPEGColorCorpus() {
		t.Run(name, func(t *testing.T) {
			before, err := jpegimage.Decode(data, 1_000_000)
			if err != nil {
				t.Fatal(err)
			}
			resized, err := jpegimage.EncodeOriginal(before, domain.Dimensions{Width: 64, Height: 48}, 85, "safe")
			if before.Components != 3 && len(before.ICC()) > 0 {
				if err == nil || len(resized) != 0 || !strings.Contains(err.Error(), "non-RGB ICC JPEG") {
					t.Fatal("tagged non-RGB resize must refuse before relabeling pixels", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			after, err := jpegimage.Decode(resized, 1_000_000)
			if err != nil || after.Info.Dimensions != (domain.Dimensions{Width: 64, Height: 48}) || !bytes.Equal(before.ICC(), after.ICC()) {
				t.Fatal("supported resize lost geometry or profile", err)
			}
		})
	}
}
