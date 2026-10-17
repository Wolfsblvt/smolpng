# Windows PNG/JPEG qualifier evidence, 17 October 2026

## Meaning

The maintained `smolpng-qualify` command passed all 61 real PNG/JPEG checks on Windows amd64. Three separate preflight refusals retained incomplete reports and correctly scoped dependency requirements. This record binds the helper and tested application independently: it qualifies the new helper against an existing frozen development executable, rather than claiming a rebuilt application, graphical behavior or a published release.

## Executed sources

Katja ran helper `349ef0319709442f87235d4c1ba7c0e13aa8cf98` against the packaged application from `04ff79b22cb2028941913a092741d8cdea032d26` (version `0.3.0-dev.4`). The helper was built with Go 1.27.2, `-trimpath` and `-X main.Commit=349ef0319709442f87235d4c1ba7c0e13aa8cf98`; its build metadata reported that revision and `vcs.modified=false`.

| Executable | SHA-256 |
| --- | --- |
| Qualifier | `4fa2f671e5a23fb51e765187242b9ce751de1ecf489eb7b58ca54f465fe9d2a4` |
| Application | `c28d85606720c6aafca126893df0fa6aefd280c01fccf659c057ab3744ecbdf9` |

Explicit external tools were pngquant 2.17.0, OxiPNG 10.2.1 and jpegoptim 1.5.6. No optimizer was bundled or installed by this qualification. Each invocation used an isolated home and disabled project lookup. Generated CC0 specimens and outputs stayed under the selected proof directory.

The native invocation had this shape, with absolute executable paths and a new output directory:

```text
smolpng-qualify.exe -executable smolpng.exe -output RUN/full
  -expected-commit 04ff79b22cb2028941913a092741d8cdea032d26
  -pngquant PNGQUANT -oxipng OXIPNG -jpegoptim JPEGOPTIM
```

The local capture wrapper used direct argument arrays with Windows `CREATE_NO_WINDOW`, captured stdout/stderr and exit status, and persisted each receipt before judging it. It excluded inherited application configuration and credential environment entries without recording their values. The wrapper exited zero after writing its completed aggregate receipt.

## Results

| Case | Observed outcome |
| --- | --- |
| Default PNG/JPEG run | Exit 0; complete and passed; all 61 checks passed in 257.031 seconds |
| JPEG-only preflight, PNG tools unavailable, intentionally wrong expected source | Exit 1; provenance refused, JPEG dependency check passed; incomplete and not passed |
| PNG-only preflight, JPEG tool unavailable, intentionally wrong expected source | Exit 1; provenance refused, PNG dependency check passed; incomplete and not passed |
| Default formats, JPEG tool unavailable, correct expected source | Exit 1; provenance passed, dependency check refused; incomplete and not passed |

The full run covers two preflights; Auto and strict lossless across eleven PNG and thirteen JPEG specimens; five target/stdout/partial-batch checks per format; and one mixed-format batch. JPEG includes progressive/grayscale images, private EXIF and orientations 1–8. Validation decodes artifacts and compares format, dimensions, orientation, byte accounting and promised pixel identity; source files are checked for preservation. Mixed results must include each requested input exactly once. Strict targets distinguish success, refusal and explicitly accepted nearest output.

The format-isolated cases qualify dependency selection and refusal evidence only, not complete single-format compression runs. Cancellation retention also has a focused source test; this native attempt did not interrupt a running optimizer or qualify hard-kill recovery.

Local `go test ./tools/qualify` and `go vet ./tools/qualify` passed. A negative test demonstrated that duplicate PNG results could conceal an omitted JPEG result before its correction. Hosted [Verify run 39009774261](https://github.com/Wolfsblvt/smolpng/actions/runs/39009774261) passed all eleven jobs for helper source `349ef0319709442f87235d4c1ba7c0e13aa8cf98`: root verification on Windows, Linux and macOS, six cross-builds and two native shell compilation jobs.

## Evidence boundary

Local captures remain under `artifacts/validation/native-qualifier-349ef03/`: `qualification.json`, per-case process receipts/stdout/stderr and each case's `report.json`. The full report records helper source, candidate source, actual application hash, selected formats, discovered tools and individual checks. These ignored captures are local evidence, not files supplied by a public checkout; this record preserves the attributable result and executable identities.

Photographic quality, wider ICC/CMYK behavior, other-host runtime behavior, clipboard/paste, physical dragging, rendered notifications, UI design, signing, publication and production update feeds retain their separate standing. The [development guide](../DEVELOPMENT.md#exercise-the-built-candidate-without-rebuilding) owns the maintained command contract; [issue #31](https://github.com/Wolfsblvt/smolpng/issues/31) owns the contribution's discussion and Return.
