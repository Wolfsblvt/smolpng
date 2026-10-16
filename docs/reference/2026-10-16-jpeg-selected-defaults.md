# Selected JPEG preset qualification, 16 October 2026

## Meaning

The Windows executable built from `e2747f2958846b77b57f2a36774ca10e0affb30f` produced the owner-selected JPEG Auto / Web balanced / Maximum squeeze outputs at 85 / 80 / 70. Six real compression paths and three effective-configuration reads passed. This record preserves exact source, executable and output evidence for that selection; it does not establish universal perceptual quality or native behavior on other hosts.

## Subject and method

Wolf selected **85 / 80 / 70** after inspecting original-scale photographic crops and asking for stronger size reduction. The [selection and rationale](https://github.com/Wolfsblvt/smolpng/pull/14#issuecomment-6225186257) supersede the earlier conservative recommendation in the [first comparison](2026-10-16-jpeg-photographic-comparison.md).

The actual Windows 11 amd64 build used Go 1.27.2, CGO disabled, `-trimpath -buildvcs=false`, and standalone version `0.3.0-dev.2`. Its SHA-256 is `f426f70ec58fe7b0871e7d22c0cfbb88749ab940570177605d3beccaacc3f458`. The separately installed jpegoptim 1.5.6 executable has SHA-256 `0a79681b56961ec2c5f45c4f9dff01a64e7274b4f3250c22cde5f09c0bb457f3`.

Every command read the original corpus `jpeg/photographic/food-refocus.jpg`, SHA-256 `62842b3086bb8ec0ddd119b87a9c64055a28f0dcf3d890af7b0c4891248bcb8b`, under an isolated configured home with project configuration disabled and offline compression. No TinyPNG credential or upload was used. Each command wrote a distinct output and captured JSON/stderr before independent Pillow 12.3.0 decoding.

## Observed outputs

| Actual command selection | Output bytes | Smaller than pixel-identical Safe |
| --- | ---: | ---: |
| Auto, default JPEG 70–85 | 287155 | 56.32% |
| Web balanced, default JPEG 70–80 | 227579 | 65.38% |
| Maximum squeeze, default JPEG 55–70 | 161307 | 75.46% |
| Lossless | 657362 | baseline |
| Auto with shared `--quality 70-95` | 541651 | 17.60% |
| Auto with `--quality 70-95 --jpeg-quality 55-70` | 161307 | 75.46% |

The complete original is 2532117 bytes, comprising a 732568-byte primary JPEG and 1799549 attached camera bytes. All six outputs reported the removal warning with that exact attached-byte count. Those removed auxiliary bytes are **not** counted as encoder savings in the table.

Independent decoding found unchanged upright dimensions of 3264 × 1472, retained orientation 3, and unchanged ICC standing (this photograph has no ICC profile). Lossless retained identical decoded pixels. The other five outputs were lossy, as intended. The original's SHA-256 remained unchanged. These are encoder outputs from the committed defaults, not renamed files from the earlier explicit-quality experiment.

| Output | SHA-256 |
| --- | --- |
| Auto | `7edbea1d5b50a43a11a186379212c1bda59b008d04e9dfbd229af4fe780295c0` |
| Web balanced | `26f5c56004ea7a958d32145379ae4626c0d88ed0c42d49638fc991e63b048d0c` |
| Maximum squeeze / JPEG override | `bbe3810334df4b0a817f65cc31a83af155d20a18caee949df8dee510d0279946` |
| Lossless | `da0eb42430323fe6ebe54a59e60027704aa8b6e8ec90c7acd8a73649e903a222` |
| Shared quality override | `556d0782b43554e83ba361482cd8ca7ad374aded443a957e5f5e4449128dab00` |

Three actual `config show --effective --provenance --preset …` invocations confirmed JPEG maxima 85 / 80 / 70, PNG maxima 95 / 95 / 85, and the JPEG bound provenance entries. Automated regressions separately cover shared/specific override precedence, zero bounds, PNG namespace isolation, resolved preset export/import, and persistence of the visible JPEG setting when changing PNG quality.

## Qualification corrections and limits

The qualification script initially asserted the noun “attachment” while the real warning says “attached”; the warning was present and the output correct. Its first config invocation also supplied compression-only `--offline`, which that command correctly rejected. Both harness mistakes were corrected. The saved six successful compression outputs were inspected and reused without recompression; the corrected config reads then passed. The complete local receipt is under `artifacts/calibration/photographs-20261016/defaults-e2747f2/receipt.json` in the operator checkout, with captured command output beside it. These ignored artifacts support this committed record and are not required for normal use.

Wolf's crop inspection and choice supply product judgment for this useful default balance. JPEG quality numbers are encoder-specific and are not Photoshop-equivalence or perceptual scores. Broader photographic/ICC/CMYK calibration, graphical desktop use and other native hosts remain separate evidence. This change does not alter PNG defaults, the lossless contract, or the full product destination.
