# Native Windows RGB ICC qualification

## Meaning

Native Windows compression now has complete RGB ICC controls and independent color-managed observations. Eight checks against the four new Go-generated corpus files passed: PNG quantizer conversion and strict lossless preservation, plus JPEG Auto/Lossless preservation. Earlier gradient controls separately exercised genuinely lossy JPEG Auto output. This record preserves that evidence and its observer correction; it does not claim broad photographic, wide-gamut, grayscale/CMYK, clipboard or display-device qualification.

## Exact executable and controls

The Windows amd64 application was built from `053586fd48f3cb8cc5567046d14f98af8a801615`, with that commit embedded. Its SHA-256 is `684c4ff331d293be0f2c2da46db1df4590f20fc3d71649807544970b9f2d15ea`. Separately installed pngquant 2.17.0, OxiPNG 10.2.1 and jpegoptim 1.5.6 were selected through explicit paths. A disposable portable home, `--no-project`, offline/remote-never compression and a credential-free child environment kept ordinary user configuration and provider quota outside this run. Originals were hashed and preserved.

Pillow 12.3.0 with LittleCMS 2.19 supplied the independent decoder/CMS. Its generated 588-byte sRGB profile has SHA-256 `e254bed42f94d9ca0674a15e3befceeeff11dc50956e8bdad7ea1bfcc8543aec`; the complete bytes are frozen in `internal/testutil/profile-srgb.icc`. Its copyright tag permits free use. The corpus generator needs only Go and this checked-in input.

The synthetic variant exchanges the red/blue XYZ colorants and clears the profile ID. It deliberately means something different from sRGB, despite inheriting the original descriptive tag. Its swatch channels are exchanged correspondingly. This makes ignored profiles and profiles incorrectly restored after conversion visibly wrong, without importing a proprietary profile or claiming a particular real-world gamut. ICC's [profile structure](https://github.com/InternationalColorConsortium/iccDEV/blob/master/iccdev-mcp/docs/icc-format-guide.md) and Pillow's [CMS interface](https://pillow.readthedocs.io/en/stable/reference/ImageCms.html) describe the underlying carrier and transform interfaces.

## Canonical corpus result

| Input and operation | Original → output bytes | Observed contract |
| --- | --- | --- |
| sRGB PNG, Auto quantizer-only comparison | 42030 → 1249 | Actual pngquant pipeline; converted output declares sRGB |
| Swapped-profile PNG, same comparison | 42030 → 1249 | Same sRGB result; the old profile is absent |
| sRGB PNG, Lossless | 42030 → 1247 | Decoded pixels and full ICC preserved; OxiPNG palette output |
| Swapped-profile PNG, Lossless | 42030 → 1236 | Decoded pixels and full ICC preserved |
| sRGB JPEG, Auto and Lossless | 3611 → 2126 each | Full ICC and decoded pixels preserved |
| Swapped-profile JPEG, Auto and Lossless | 3396 → 1980 each | Full ICC and decoded pixels preserved |

Each JPEG row includes Auto and Lossless, for eight checks altogether. Both canonical JPEG Auto results selected pixel-identical candidates; their small swatch sizes are not evidence of a lossy quality win.

The quantized outputs carry an sRGB intent of 0 and gamma 0.45455. Mean absolute color-managed channel error on the 0–255 scale is 0.1763 for each. Deliberately applying the old swapped profile to the converted output raises that error to 66.1765. This is a discriminating color-interpretation check, not a universal perceptual threshold or a claim that numerical channel error models human vision.

The canonical image hashes are in `testdata/corpus/manifest.toml`. The machine receipt is `artifacts/validation/windows-icc-053586f/canonical-qualification.json`; it reports `complete: true`, `passed: true`, eight checks and each exact fixture digest. The application and capture files live under the same artifact directory. The source generator and fixtures, rather than these local artifacts, preserve the reproducible inputs.

## Earlier controls and observer correction

Eight initial sRGB/swapped-profile gradient PNG/JPEG Auto/Lossless checks passed. JPEG Auto produced genuinely lossy outputs at the existing selected default, retained the complete profile and had mean managed errors 3.9213 and 4.5566. PNG Auto chose a lossless candidate, so those runs did not establish quantizer conversion.

Explicit quantizer-only gradient comparisons also retained the original under ordinary quality bounds. They are preserved as unsuccessful attempts to reach that path, not counted as quantizer qualification. A direct permissive pngquant call established that conversion was available. Finite-color swatches then reached the actual application quantizer path at ordinary Auto quality; no application default was relaxed. Two independently generated swatch checks and the canonical two checks agreed on converted sRGB output and the wrong-profile control.

The first canonical observer stopped on an OxiPNG-produced palette image: Pillow cannot build a palette-mode CMS transform directly. Its ICC bytes still matched the input exactly. Expanding the palette to RGB before applying that RGB profile repaired the observer. The existing successful compression output and its command receipt were reused; missing or unusable receipts are refused. That observer exception is not an application defect or a passing first attempt.

## Remaining boundary

This carries complete RGB matrix profiles and controlled interpretation through real native optimizers. It does not qualify arbitrary ICC variants, real wide-gamut distributions, grayscale/CMYK transforms, all profile segment arrangements, native clipboard rendering or another host. The two photographic originals remain the same admitted bytes. JPEG 85 / 80 / 70 and all PNG defaults are unchanged. No production signing, release, package, installation or hosted writer effect occurred.
