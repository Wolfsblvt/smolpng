# Local recipe and tool comparison

## Meaning

`smolpng compare` is the inspection and calibration workflow for the shared engine. It compares presets and the contribution of the two selected PNG tools without uploading, changing a setting, replacing an input, or creating an installed-user history entry. It is not a second compressor or a claim that minimum bytes means best visual quality.

```powershell
# Four built-in recipes, original bytes loaded once per input.
smolpng compare images --recursive --output-dir artifacts/review --json

# Isolate the PNG tool contribution under one quality policy.
smolpng compare icon.png --preset web --png-mode lossless --png-mode quantizer --png-mode pipeline --output-dir artifacts/png-tools

# Compare your own saved settings and a common target.
smolpng compare icons --preset web --preset mcp-icon --target-size 10KiB --repeat 3 --format csv

# Planning causes no probe, download or image write.
smolpng compare images --recursive --dry-run --json
```

`auto` admits ordinary engine selection; `lossless` enables OxiPNG-only candidates; `quantizer` permits pngquant without its OxiPNG tail; `pipeline` requires both tools. An original/lossless candidate may still beat quantized output. Read the returned pipeline, not just the requested mode. The two initial PNG tools are complementary: pngquant quantizes and OxiPNG optimizes DEFLATE/filtering. A third PNG codec was not silently added. JPEG uses the selected jpegoptim integration with `png-mode auto`. The completed bounded JPEGli comparison remains a separate calibration record; it did not add another application backend.

`--preset` and `--png-mode` are repeatable and form a matrix. Omitted presets select auto/web/smallest/lossless; omitted mode is auto. Contradictory combinations, missing required tools and preset-format exclusions remain explicit failed/skipped rows. Sources are bounded to 512, variants to 64 and repetitions to 5. Every repeat consumes the same original bytes. Timing is serial wall-clock engine time including process startup and validation, not a controlled hardware benchmark. Engine versions and resolved recipes are included.

## Reports and candidate preservation

With no output directory the report goes to stdout only. An explicit **new** `--output-dir` contains original snapshots, successful candidate images, `report.json`, `report.csv` and `index.html`. The offline HTML has before/after images on a checkerboard and no scripts, external styles or network references. File names in this bundle are bounded numeric IDs, not paths reconstructed from user input. JSON retains exact source labels; CSV guards cells that could be interpreted as spreadsheet formulas.

A strict target miss is still exit 6 for one row or 8 for a matrix with failures. Its nearest artifact is not written unless `--accept-nearest` was selected. Skipped unsupported format combinations are not invented zero-byte successes. Existing comparison directories are never overwritten. Snapshots and candidate images may contain user data and must be handled like the originals; nothing is uploaded by the command.

## Metrics and interpretation

Rows contain source/candidate hashes, exact byte counts, dimensions, alpha, orientation, selected pipeline, elapsed time, effective recipe, target status and errors. Pixel identity compares full decoded channels, including hidden RGB; JPEG pixels are first viewed upright.

The diagnostic scores sample at most 512 by 512 positions, compositing encoded-channel RGB over black and white. RMSE and mean absolute alpha error are normalized to 0..1. These are not perceptual scores, quality guarantees, or colour-managed evaluations. Geometry changes, matte preparation and differing colour metadata withhold the scores with a reason rather than comparing unlike representations. Compare transparent edges on light and dark surfaces and use actual corpus review before changing release defaults.

The `comparison-v1.schema.json` schema owns this distinct report shape. It does not change the existing compression report schema. The current authoring tests use generated legal specimens and fake optimizer boundaries; actual optimizer performance/visual findings must be recorded against exact versions on the target host.

## Use the corpus to refine defaults

The [regression inventory](../testdata/corpus/README.md) has format/class folders, actual progressive JPEG and all EXIF orientations. Use `--recursive` for the whole inventory, and keep evaluation outputs outside it. First compare under identical recipe/target/geometry constraints; then inspect saved edges against the HTML report's checkerboard, white and black backgrounds. CSV includes the pipeline and diagnostic scores; JSON remains the exact nested representation.

To compare optimizer versions, run the same command twice with explicit executable paths and separate new output directories. Preserve both reports and inspect their recorded versions and source hashes. A different library's quality number is not numerically comparable by itself. No public default is recalibrated from fake-process tests, synthetic noise, or minimum bytes alone. The current two-tool PNG stack remains pngquant plus OxiPNG; adding a different encoder requires its own dependency/license/behavior qualification.

The [first photographic JPEG comparison](reference/2026-10-16-jpeg-photographic-comparison.md) records native outputs from two CC0 originals, independent full-pixel/ICC checks, and the bounded visual judgment behind the selected 85 / 80 / 70 preset balance. The later [JPEG backend comparison](reference/2026-10-21-jpeg-backends.md) compares matched byte budgets and preservation contracts; it supports retaining jpegoptim rather than adding JPEGli. Camera-attachment removal remains separate from encoder savings.
