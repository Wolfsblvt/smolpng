# JPEG color fixture inputs

## Meaning

These frozen CC0 inputs let Go-only corpus generation exercise real grayscale, CMYK and Display P3 JPEGs. They are test controls, not application color-management dependencies or a perceptual-quality benchmark. Go's JPEG encoder cannot write CMYK, so that one original encoded seed is preserved instead of adding an encoder to normal development.

## Profile provenance

The three profiles are unchanged files from [Compact-ICC-Profiles at `bdd84663061bc4ae95ca70decff54f581e27f702`](https://github.com/saucecontrol/Compact-ICC-Profiles/tree/bdd84663061bc4ae95ca70decff54f581e27f702). `saucecontrol` dedicates the collection to [CC0-1.0](https://github.com/saucecontrol/Compact-ICC-Profiles/blob/bdd84663061bc4ae95ca70decff54f581e27f702/license).

| Local input | Upstream filename | SHA-256 |
| --- | --- | --- |
| `profile-gray.icc` | `sGrey-v4.icc` | `00c0f94e09127520a17dc0e1d9264b5702081d96dbdb1549ee88e3631ce42a9d` |
| `profile-cmyk.icc` | `CGATS001Compat-v2-micro.icc` | `73e1ba37d2bad5bab2a964f40a9eed96209666efc067c3322626214bbef234a0` |
| `profile-display-p3.icc` | `DisplayP3-v4.icc` | `cb51de38e482ee974c0c76b9689e16aad04bad16e226fed2f30c842d15ff3a3d` |

Each upstream file is under `profiles/` at that pinned revision. The CMYK profile supplies a perceptual CMYK-to-display mapping only, not conversion into CMYK or arbitrary print-workflow evidence. The P3 profile uses actual Display P3 primaries, including its adapted negative red Z value; these controls do not claim every desktop consumer supports it.

## Original CMYK seed

`jpeg-cmyk-seed.jpg` is an original 128×96 four-component Adobe-marked JPEG generated on 21 October 2026 with Pillow 12.3.0 and dedicated to CC0-1.0 by Wolfsblvt Works contributors. Its SHA-256 is `bccd6a50ff4d3284d85401e5a337bb1e3157c6b0b8e8c775695ffabacdb123d8`. Its construction is:

```python
from PIL import Image

image = Image.new("CMYK", (128, 96))
image.putdata([
    (x * 2 % 256, y * 3 % 256, (x + y) % 256,
     (x // 8 + y // 6) * 7 % 256)
    for y in range(96) for x in range(128)
])
image.save("jpeg-cmyk-seed.jpg", quality=97, icc_profile=None)
```

Pillow is needed only to reconstruct this historical seed. Normal generation and checks consume its fixed bytes with Go. Do not silently replace it after a codec change. `jpeg_color.go` embeds the seed and profiles, generates grayscale/RGB controls in Go, and adds ordinary ICC_PROFILE segments. The generated fixture hashes live in `testdata/corpus/manifest.toml`.
