# Windows JPEG color-model qualification

## Meaning

Sixteen native Windows checks passed for generated grayscale, CMYK and Display P3 JPEGs using real jpegoptim and independent Pillow/LittleCMS observations. Four canonical CC0 fixtures and focused Go tests now preserve the exercised color-model and profile boundaries. This qualifies those controls and the existing refusal of tagged non-RGB resizing; it does not select a new codec, alter JPEG quality defaults, complete the graphical candidate or establish arbitrary print/display behavior.

## Exact runtime and inputs

The tested Windows amd64 development executable came from [Verify run 39653980910](https://github.com/Wolfsblvt/smolpng/actions/runs/39653980910), source `5aaa04001d1da4f7809b650a4c4afd1b67d1e7c9`. It reports `0.1.0-dev`, distribution `source`, and SHA-256 `8aa8e355f26368f10bbedd79285330495df37eaa0148de713ff2d53541b207aa`. It is distinct from private dev.11 and no release archive was rebuilt. All eleven exact-source Verify jobs passed.

The host was Windows 11 amd64, kernel `10.0.26200`. The independent observer was the existing bundled Python 3.12.14, Pillow 12.3.0 and LittleCMS 2.19. Existing jpegoptim 1.5.6 had SHA-256 `0a79681b56961ec2c5f45c4f9dff01a64e7274b4f3250c22cde5f09c0bb457f3`. Commands used a fresh portable home, `--no-project`, one image job, `--copy none`, remote policy `never`, and a credential-free optimizer environment. No desktop, clipboard, keyring or remote image service was used.

The canonical inputs are 128×96 generated controls. [Input provenance](../../internal/testutil/jpeg-color-inputs.md) carries the unchanged CC0 profiles, original CMYK seed construction and hashes. Normal corpus regeneration requires Go only. `testdata/corpus/manifest.toml` binds the four generated JPEGs:

| Fixture | SHA-256 |
| --- | --- |
| `jpeg-icc-gray.jpg` | `50ba3ea54dc9044b41c8b22ed86371bc48072252fe377bb027bf6e5b23b608eb` |
| `jpeg-icc-cmyk.jpg` | `e051ee60ad1d86f8579c6121990d1e6788108ee52122fd03497334cee91bf3a7` |
| `jpeg-icc-display-p3.jpg` | `b29a2b5fabddd97ff4c39537209fa27604f7ed636bfced71b10036e0ee315e3d` |
| `jpeg-cmyk.jpg` | `bccd6a50ff4d3284d85401e5a337bb1e3157c6b0b8e8c775695ffabacdb123d8` |

## Native observations

Each fixture exercised Auto, strict Lossless, strip-all and explicit 64×48 resize. Outputs were independently decoded and every original was rehashed after each operation.

| Input | Original bytes | Auto bytes | Lossless bytes | Strip-all bytes | Resize |
| --- | ---: | ---: | ---: | ---: | --- |
| Profiled grayscale | 4460 | 1835 | 3045 | 1457 | Exit 2; no output |
| Profiled CMYK | 16432 | 10915 | 12640 | 2433 | Exit 2; no output |
| Profiled Display P3 RGB | 3829 | 1786 | 2618 | 1288 | 1231 bytes, RGB, profile retained |
| Untagged CMYK | 7950 | 2433 | 4158 | 2433 | 779 bytes, converted to RGB |

Auto and Lossless preserved the exact embedded profile and source color model. Actual Auto/Lossless/strip-all pipelines included jpegoptim, rather than only retaining the original file. All four Lossless outputs preserved their decoded channels exactly; the three tagged Lossless outputs also had zero difference after independent conversion to sRGB. Strip-all deliberately removed profiles while retaining the original component count; this is not a claim that removing a profile preserves managed appearance.

The documented grayscale/CMYK tagged-resize refusal returned exit 2, explained the missing color-managed encoder, committed no output and preserved the original. P3 RGB resizing retained its profile. Untagged CMYK resizing produced the existing Go RGB conversion, with no invented profile attached.

Auto is lossy. Managed maximum channel differences were 74 for grayscale, 4 for CMYK and 80 for P3 on these synthetic gradients. Mean per-channel differences were approximately 1.58, 0.45–0.49 and 0.91–1.81 respectively. These are diagnostics, not a universal perceptual threshold or visual acceptance of the selected 85/80/70 defaults.

## Retained proof and limits

The complete canonical report is retained locally at `artifacts/validation/jpeg-colour-5aaa040/canonical-run/report.json`, SHA-256 `303f0c2acbce7378998fe3b6709c439ec3e5843c4caef00404e912c91a7edb7c`. It contains exact commands, exits, JSON results, source/output hashes, pipelines, profiles and CMS differences. The earlier Pillow-generated exploratory controls have their own separate `run/report.json`; they are not substituted for the canonical fixture result.

Source tests cover one-/three-/four-component decoding, exact ICC policy through cleanup/reconciliation, pixel preservation and refusal before tagged non-RGB conversion. They use the existing test runner without optimizers or a CMS dependency. Corpus generation/checking preserves the existing 30 fixtures and adds four controls.

Arbitrary ICC profiles, rendering intents, print workflows, real wide-gamut photographs, graphical consumers, other hosts and the joined candidate remain separate evidence. The CMYK profile is a perceptual source-to-display control only. No application behavior, codec, quality default, production signing/feed or publication effect changed.
