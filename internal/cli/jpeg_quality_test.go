package cli

import (
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/spf13/cobra"
)

func TestJPEGQualityFlagsSurviveBothResolutionStages(t *testing.T) {
	var flags compressFlags
	command := &cobra.Command{}
	flags.bind(command)
	if err := command.Flags().Set("quality", "60-90"); err != nil {
		t.Fatal(err)
	}
	if err := command.Flags().Set("jpeg-quality", "50-75"); err != nil {
		t.Fatal(err)
	}
	fields, err := flags.configOverrides(command)
	if err != nil {
		t.Fatal(err)
	}
	if fields.Quality == nil || *fields.Quality != "60-90" || fields.JPEGQuality == nil || *fields.JPEGQuality != "50-75" {
		t.Fatal(fields)
	}
	recipe, err := domain.Builtin("smallest")
	if err != nil {
		t.Fatal(err)
	}
	recipe, err = flags.apply(command, recipe)
	if err != nil {
		t.Fatal(err)
	}
	if recipe.QualityMin != 60 || recipe.QualityMax != 90 || recipe.JPEGQualityRange() != (domain.QualityRange{Min: 50, Max: 75}) {
		t.Fatal(recipe)
	}
}
