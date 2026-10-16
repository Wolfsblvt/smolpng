# JPEG photographic comparison, 2026-10-16

## Meaning

Two CC0 photographs now have native built-in-preset observations with independently decoded outputs. The Samsung image distinguishes ordinary lossy Auto/Web output from the smaller Maximum squeeze choice; the landscape produces one pixel-identical result across all four presets while preserving its embedded ICC profile. These results support retaining the current choices while broadening calibration. They do not establish a general perceptual-quality default or color-managed fidelity.

## Inputs and executed boundary

The [corpus manifest](../../testdata/corpus/manifest.toml) carries original publisher, creator, download and CC0 provenance. No input was cleaned or re-encoded before comparison.

| Original | Bytes | SHA-256 |
| --- | ---: | --- |
| `jpeg/photographic/food-refocus.jpg` | 2532117 | `62842b3086bb8ec0ddd119b87a9c64055a28f0dcf3d890af7b0c4891248bcb8b` |
| `jpeg/photographic/landscape-clouds.jpg` | 399226 | `d266b5bfa1179a6298f1c084e8c539ef81698643f7c45d9adb34d43a30ed2758` |

Windows amd64, separately installed jpegoptim 1.5.6, isolated portable home, `--no-project --offline`, no clipboard or remote service. The camera build is source `3e102e135f8067ecc6dc6101bdf5ffd3e6414b2b`, Go 1.27.2, `0.3.0-dev.1`, executable SHA-256 `fdce1ea82dd82fdf18ee3ee34d4ab7c31453e90de9b5076acdf9f1295da7f427`. Existing Auto/Lossless outputs are reused from the [complete camera-policy run](2026-10-16-jpeg-camera-attachments.md); the additional command compared Web balanced and Maximum squeeze from the same original:

```text
smolpng compare <food-refocus.jpg> --preset web --preset smallest --output-dir <new-directory> --home <isolated-home> --no-project --offline --json
```

Landscape evidence is reused from the earlier four-preset comparison at `e611fd3ee768d349e841f87e6c008a2e49a461eb`, with the same jpegoptim version. It establishes that earlier source boundary; this record does not invent a fresh landscape run on the camera repair. Independent Pillow 12.3.0 inspection of every retained output was performed for this record, checking hashes against comparison receipts, upright dimensions, full decoded RGB identity, quantization tables and ICC bytes.

## Observed results

| Image | Preset | Output bytes | Upright pixels identical to original |
| --- | --- | ---: | --- |
| Camera | Auto | 541651 | no |
| Camera | Web balanced | 541651 | no; byte-identical to Auto |
| Camera | Maximum squeeze | 287155 | no |
| Camera | Lossless | 657362 | yes |
| Landscape | Auto, Web balanced, Maximum squeeze, Lossless | 276012 each | yes; all four outputs byte-identical |

Camera outputs remain 3264×1472 with orientation 3; landscape outputs remain 4000×3000 with orientation 1. Every output matches its independently recorded SHA-256. Camera Auto/Web and Maximum squeeze change the source quantization tables; Lossless and all landscape outputs retain them. The landscape's 560-byte ICC profile is byte-identical in every output. The camera original has no ICC profile. This exercises one real profile's preservation, not color conversion, wide-gamut display or CMYK behavior.

The camera original contains 1799549 appended depth/refocus bytes after a 732568-byte primary JPEG. Every Safe result warns that the output is a still image without that attachment. Total-input reduction includes attachment and private-metadata removal and must not be advertised as encoder savings. Against the 657362-byte pixel-identical Safe result, Auto/Web are about 17.6% smaller and Maximum squeeze about 56.3% smaller. Maximum squeeze is about 47.0% smaller than Auto/Web in this image.

The application's sampled encoded-channel RMSE is approximately 0.00326 for Web and 0.00558 for Maximum squeeze. These diagnostics are not perceptual or color-managed scores and did not select a default. Serial engine times in the additional run were 11.21 and 5.62 seconds; earlier runs on this host were substantially slower. There was no controlled timing experiment, so no performance-improvement claim follows.

## Visual judgment and product consequence

Katja inspected three unscaled 512×512 upright crops spanning food texture, dark background and wooden board, with Original, Auto/Web and Maximum squeeze side by side. Auto/Web remain close to the source in those crops; Maximum squeeze shows slight texture simplification. The original already has camera processing and shallow-focus artifacts, which must not be blamed on compression. Preview display and these crops are a bounded visual judgment, not a human study.

The current Auto and explicit size-first choice remain sensible for these samples. No preset values, codec, pipeline, privacy policy or default changed. Wolf was offered the actual crop comparison for taste input while that judgment can still change calibration direction. More subjects, stronger texture/text examples, additional real ICC profiles and CMYK still need evidence before final backend/quality selection; two photographs cannot represent that complete set.

## Reproducible local evidence

Ignored receipts under `artifacts/calibration/photographs-20261016/` retain `builtins-e611fd3/` (earlier landscape comparison), `camera-3e102e1-policy/` (complete camera policy run), and `builtins-camera-3e102e1/` (two additional recipes, original snapshot, candidates, JSON/CSV and offline HTML). The latter also holds `independent-inspection.json` and `food-crops.png`. Original corpus hashes remained unchanged. These local files support the committed result; their presence is not a release, integration, rendered-toast or native-drag claim.
