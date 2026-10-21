package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/jpegimage"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
)

func colorResult(original, output []byte) domain.Result {
	before, _, err := decodeImage(original)
	if err != nil {
		panic(err)
	}
	after, _, err := decodeImage(output)
	if err != nil {
		panic(err)
	}
	return domain.Result{Format: "jpeg", OriginalBytes: int64(len(original)), OutputBytes: int64(len(output)),
		Dimensions: domain.DimensionChange{
			Before: domain.Dimensions{Width: before.Bounds().Dx(), Height: before.Bounds().Dy()},
			After:  domain.Dimensions{Width: after.Bounds().Dx(), Height: after.Bounds().Dy()},
		}}
}

func TestQualifierRejectsDroppedJPEGProfileAndChangedComponentModel(t *testing.T) {
	for name, original := range testutil.JPEGColorCorpus() {
		t.Run(name, func(t *testing.T) {
			if failures := validateImage(original, original, colorResult(original, original), imageExpectation{Lossless: true}); len(failures) != 0 {
				t.Fatal(failures)
			}
			before, err := jpegimage.Decode(original, 1_000_000)
			if err != nil {
				t.Fatal(err)
			}
			stripped, err := jpegimage.Filter(before, "all")
			if err != nil {
				t.Fatal(err)
			}
			result := colorResult(original, stripped)
			if failures := validateImage(original, stripped, result, imageExpectation{StripICC: true, Lossless: true}); len(failures) != 0 {
				t.Fatal("explicit profile removal should preserve the channels", failures)
			}
			if len(before.ICC()) > 0 {
				if failures := validateImage(original, stripped, result, imageExpectation{}); len(failures) == 0 {
					t.Fatal("successful exit and identical pixels concealed a lost profile")
				}
			}
			if before.Components == 4 {
				rgb := testutil.JPEG(before.Image, 70)
				if failures := validateImage(original, rgb, colorResult(original, rgb), imageExpectation{StripICC: true}); len(failures) == 0 {
					t.Fatal("unrequested CMYK-to-RGB conversion passed")
				}
			}
		})
	}
}

func TestQualifierAcceptsOnlyRequestedJPEGGeometry(t *testing.T) {
	for _, name := range []string{"jpeg-icc-display-p3.jpg", "jpeg-cmyk.jpg"} {
		t.Run(name, func(t *testing.T) {
			original := testutil.JPEGColorCorpus()[name]
			before, err := jpegimage.Decode(original, 1_000_000)
			if err != nil {
				t.Fatal(err)
			}
			dimensions := domain.Dimensions{Width: 64, Height: 48}
			output, err := jpegimage.EncodeOriginal(before, dimensions, 85, "safe")
			if err != nil {
				t.Fatal(err)
			}
			result := colorResult(original, output)
			if failures := validateImage(original, output, result, imageExpectation{Resize: dimensions}); len(failures) != 0 {
				t.Fatal(failures)
			}
			if failures := validateImage(original, output, result, imageExpectation{}); len(failures) == 0 {
				t.Fatal("unrequested geometry change passed")
			}
		})
	}
}

func TestJPEGColorRefusalMustExplainBoundaryAndLeaveNoArtifact(t *testing.T) {
	for _, mode := range []string{"correct", "wrong-reason", "committed-output"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			s := suite{root: root, opt: options{Timeout: time.Second}, data: evidence{Passed: true},
				runner: runnerFunc(func(_ context.Context, invocation process.Invocation) (process.Result, error) {
					args := invocation.Args
					arg := func(key string) string {
						for i := range args {
							if args[i] == key && i+1 < len(args) {
								return args[i+1]
							}
						}
						return ""
					}
					name := filepath.Base(args[1])
					original := testutil.JPEGColorCorpus()[name]
					before, err := jpegimage.Decode(original, 1_000_000)
					if err != nil {
						t.Fatal(err)
					}
					dest := arg("--output")
					var result domain.Result
					code := domain.Success
					if arg("--resize") != "" && before.Components != 3 && len(before.ICC()) > 0 {
						code = domain.InvalidArguments
						result = domain.Result{ExitCode: code, Error: "non-RGB ICC JPEG requires a managed encoder"}
						if mode == "wrong-reason" {
							result.Error = "unknown flag"
						}
						if mode == "committed-output" {
							if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
								t.Fatal(err)
							}
							if err := os.WriteFile(dest, original, 0600); err != nil {
								t.Fatal(err)
							}
						}
					} else {
						var output []byte
						if arg("--resize") != "" {
							output, err = jpegimage.EncodeOriginal(before, domain.Dimensions{Width: 64, Height: 48}, 85, "safe")
						} else {
							output, err = jpegimage.Filter(before, "all")
						}
						if err != nil {
							t.Fatal(err)
						}
						if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(dest, output, 0600); err != nil {
							t.Fatal(err)
						}
						result = colorResult(original, output)
						result.Output = dest
					}
					payload, err := json.Marshal(domain.Report{ExitCode: code, Results: []domain.Result{result}})
					if err != nil {
						t.Fatal(err)
					}
					return process.Result{ExitCode: code, Stdout: string(payload)}, nil
				})}
			if err := s.exerciseJPEGColors(context.Background(), filepath.Join(root, "sources")); err != nil {
				t.Fatal(err)
			}
			if len(s.data.Checks) != 8 || s.data.Passed != (mode == "correct") {
				t.Fatal(mode, s.data.Checks)
			}
			if mode != "correct" {
				found := false
				for _, c := range s.data.Checks {
					found = found || strings.Contains(c.Name, "resize/") && !c.Passed
				}
				if !found {
					t.Fatal("known bad refusal was not attributed to its resize check")
				}
			}
		})
	}
}
