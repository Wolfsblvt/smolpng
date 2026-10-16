package config

import (
	"context"
	"testing"
)

func TestJPEGQualityDefaultsSurviveResolvedPresetRoundTrip(t *testing.T) {
	f := setup(t)
	w := f.load(t)
	for _, test := range []struct {
		name    string
		maximum int
	}{{"auto", 85}, {"web", 80}, {"smallest", 70}} {
		t.Run(test.name, func(t *testing.T) {
			before, err := w.Resolve(test.name, Fields{})
			must(t, err)
			if before.Recipe.JPEGQualityRange().Max != test.maximum || before.Provenance["jpeg_quality_max"].File != "builtin:"+test.name {
				t.Fatal(before.Recipe, before.Provenance)
			}
			data, err := w.Export(test.name, true)
			must(t, err)
			parsed, err := ParseBundle(data)
			must(t, err)
			plan, err := w.PlanImport(parsed, false)
			must(t, err)
			must(t, w.ApplyImport(context.Background(), &plan))
			after, err := w.Resolve(plan.Items[0].Name, Fields{})
			must(t, err)
			if after.Recipe.JPEGQualityRange() != before.Recipe.JPEGQualityRange() || after.Recipe.QualityMin != before.Recipe.QualityMin || after.Recipe.QualityMax != before.Recipe.QualityMax {
				t.Fatal("Resolved import changed format quality", before.Recipe, after.Recipe)
			}
		})
	}
}

func TestJPEGQualityOverridesFollowLayerAndInvocationPrecedence(t *testing.T) {
	f := setup(t)
	write(t, f.options.Paths.Config, "[defaults]\nquality='60-92'\njpeg_quality='50-83'\n[presets.photo]\nschema=1\nextends='auto'\njpeg_quality='55-80'\n")
	w := f.load(t)
	for _, test := range []struct {
		name      string
		fields    Fields
		png, jpeg int
		origin    string
	}{
		{"auto", Fields{}, 92, 83, "defaults.jpeg_quality"},
		{"photo", Fields{}, 92, 80, "presets.photo.jpeg_quality"},
		{"photo", Fields{Quality: Ptr("45-77")}, 77, 77, "flags.quality"},
		{"photo", Fields{JPEGQuality: Ptr("40-70")}, 92, 70, "flags.jpeg_quality"},
		{"photo", Fields{Quality: Ptr("45-77"), JPEGQuality: Ptr("40-70")}, 77, 70, "flags.jpeg_quality"},
		{"photo", Fields{JPEGQuality: Ptr("0-0")}, 92, 0, "flags.jpeg_quality"},
	} {
		eff, err := w.Resolve(test.name, test.fields)
		must(t, err)
		if eff.Recipe.QualityMax != test.png || eff.Recipe.JPEGQualityRange().Max != test.jpeg || eff.Provenance["jpeg_quality_max"].Key != test.origin {
			t.Fatal(eff.Recipe, eff.Provenance["jpeg_quality_max"], test)
		}
	}
	for _, invalid := range []string{"-1-80", "90-80", "0-101", "run-command"} {
		if _, err := w.Resolve("photo", Fields{JPEGQuality: Ptr(invalid)}); err == nil {
			t.Fatal("Accepted invalid JPEG quality", invalid)
		}
	}
}

func TestPNGAutoQualityDoesNotReplaceJPEGPolicy(t *testing.T) {
	f := setup(t)
	write(t, f.options.Paths.Config, "[defaults]\njpeg_quality='60-82'\n[png.auto]\nquality='90-99'\nmin_quality=91\n")
	eff, err := f.load(t).Resolve("auto", Fields{})
	must(t, err)
	jpeg := eff.Recipe.JPEGQualityRange()
	if eff.Recipe.QualityMin != 91 || eff.Recipe.QualityMax != 99 || jpeg.Min != 60 || jpeg.Max != 82 || eff.Provenance["jpeg_quality_max"].Key != "defaults.jpeg_quality" {
		t.Fatal(eff.Recipe, eff.Provenance)
	}
}
