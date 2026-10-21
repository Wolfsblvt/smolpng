// SPDX-License-Identifier: MIT
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/jpegimage"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
)

// exerciseJPEGColors supplements Auto/Lossless with explicit metadata removal
// and geometry requests. Tagged non-RGB refusal must explain the actual boundary
// and leave no output; an arbitrary exit 2 is not successful qualification.
func (s *suite) exerciseJPEGColors(ctx context.Context, sourceDir string) error {
	fixtures := testutil.JPEGColorCorpus()
	names := make([]string, 0, len(fixtures))
	for name := range fixtures {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		original := fixtures[name]
		before, err := jpegimage.Decode(original, 1_000_000)
		if err != nil {
			return err
		}
		for _, action := range []string{"strip-all", "resize"} {
			if err := ctx.Err(); err != nil {
				return err
			}
			dest := filepath.Join(s.root, "outputs", action, name)
			args := []string{"compress", filepath.Join(sourceDir, name), "--output", dest, "--json"}
			want := imageExpectation{StripICC: true}
			code := domain.Success
			if action == "strip-all" {
				args = append(args, "--strip", "all")
			} else {
				args = append(args, "--resize", "64x48")
				want = imageExpectation{Resize: domain.Dimensions{Width: 64, Height: 48}}
				if before.Components != 3 && len(before.ICC()) > 0 {
					code = domain.InvalidArguments
				}
			}
			c := s.invoke(ctx, action+"/"+name, code, nil, args...)
			var report domain.Report
			if err := json.Unmarshal(c.Report, &report); err != nil || len(report.Results) != 1 {
				c.Errors = append(c.Errors, "missing JPEG color result")
			} else {
				result := report.Results[0]
				if report.ExitCode != code || result.ExitCode != code {
					c.Errors = append(c.Errors, "JPEG color result exit differs from the command")
				}
				if code != domain.Success {
					if !strings.Contains(result.Error, "non-RGB ICC JPEG") {
						c.Errors = append(c.Errors, "resize refused for an unrelated reason")
					}
					if _, err := os.Stat(dest); !os.IsNotExist(err) || result.Output != "" {
						c.Errors = append(c.Errors, "refused resize committed an output")
					}
				} else if output, err := os.ReadFile(dest); err != nil || result.Output != dest {
					c.Errors = append(c.Errors, "JPEG color result lacks its requested artifact")
				} else {
					c.Errors = append(c.Errors, validateImage(original, output, result, want)...)
				}
			}
			s.add(c)
			if err := s.save(); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateJPEGColor(original, output []byte, want imageExpectation) []string {
	before, err := jpegimage.Decode(original, 1_000_000)
	if err != nil {
		return []string{"invalid original JPEG color specimen"}
	}
	after, err := jpegimage.Decode(output, 1_000_000)
	if err != nil {
		return []string{"invalid JPEG color output"}
	}
	failures := []string{}
	components := before.Components
	if want.Resize.Area() > 0 {
		components = 3 // The supported Go geometry encoder emits RGB JPEG.
	}
	if after.Components != components {
		failures = append(failures, "JPEG component model changed without the requested RGB resize")
	}
	profile := before.ICC()
	if want.StripICC {
		profile = nil
	}
	if !bytes.Equal(profile, after.ICC()) {
		failures = append(failures, "JPEG ICC profile differs from the selected metadata policy")
	}
	return failures
}
