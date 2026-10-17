# Regression image corpus

## Meaning

This tree contains **30 CC0-1.0 specimens: 28 project-generated structural images and two unmodified, explicitly licensed photographic originals**. Four PNG/JPEG controls carry complete RGB ICC profiles, including a deliberately different red/blue interpretation. `manifest.toml` binds each file to its class, creator, source/rights, method, hash, dimensions, depth, transparency, orientation, progressive state and any appended camera bytes. The photographs add reproducible real-image cases; these controls and two photos are not a general perceptual-quality or wide-gamut benchmark.

The directory layout now separates PNG and JPEG classes. The previous root-level filenames moved into these folders; the logical names used by in-memory tests/native qualification remain stable. Always use `--recursive` when evaluating this whole tree.

```text
png/
  flat-icon/             palette-friendly original
  alpha-gradient/        straight-alpha ramps
  alpha-16/              real 16-bit RGBA with hidden RGB
  transparent-hidden-rgb/
  screenshot-text/       synthetic text-like edges, not a captured desktop
  high-colour/           deterministic noise, not a photograph
  indexed-palette/
  grayscale-16/
  metadata/
  color-profile/         complete sRGB and synthetic red/blue-swapped ICC controls
  interlaced/            independently generated Adam7 PNG
  tiny/
jpeg/
  baseline/
  high-chroma/           synthetic noise, not a photograph
  grayscale/
  progressive/          actual two-scan, 8x8 flat-gray progressive JPEG
  exif-orientation/     all eight orientation values
  metadata/             valid EXIF maker/GPS IFDs at synthetic coordinates 0/0
  photographic/         original landscape and Samsung food/refocus photographs
  color-profile/        complete RGB ICC controls, separate from photographic diversity
```

## Regeneration and tests

From the repository root:

```sh
go run ./tools/corpus
go run ./tools/corpus --check
go test ./tools/corpus ./internal/testutil ./internal/pngimage ./internal/jpegimage
```

The generator uses Go and the checked-in source/media; it never downloads images or runs a native codec, font or image editor. Check mode verifies bytes against their manifest and rejects unmanifested PNG/JPEG files. PNG compression bytes may vary across Go versions: generated and stored specimens must retain identical decoded pixels (including hidden RGB), headers, palette, metadata and interlacing. This includes equivalent compression of the same complete ICC profile; its name, method and decompressed contents must still match. Existing equivalent encodings are preserved during regeneration, avoiding unrelated fixture/hash churn. Generated JPEG remains byte-exact. Changed generated specimens can be regenerated; imported originals are never rewritten or silently rehashed. Check mode writes nothing and neither mode deletes additional media.

The manifest is the single admission source for imported media. Their creator, permanent source-page URL, original download URL, method, CC0 rights, purpose and fixed SHA-256 are curated there. Regeneration preserves those entries, verifies the admitted hashes, and derives inspectable properties from the actual files under the application's decoded-pixel limit. A missing or modified original fails; restore it from Git or its recorded source and verify the hash. A fresh `--output` directory contains only the 28 generated specimens unless admitted originals and their manifest entries are deliberately supplied there.

Tests validate structure/rights/hashes, preserve imported bytes and admission through regeneration, reject missing/changed/unscoped imports, follow the synthetic GPS pointer, prove safe EXIF cleanup preserves orientation/pixels, and assert that the progressive specimen really has two scans. The 8x8 progressive image is deliberately a small structural case, not an optimizer-performance victory.

## Admitted photographic originals

| File | Creator and explicit source license | Coverage |
| --- | --- | --- |
| `jpeg/photographic/landscape-clouds.jpg` | Aammell, [Landscape with clouds](https://commons.wikimedia.org/w/index.php?title=File:Landscape_with_clouds.jpg&oldid=1048543422), CC0-1.0 | Natural texture/cloud gradients, 4000×3000, already-compressed JPEG/no-growth comparison |
| `jpeg/photographic/food-refocus.jpg` | Santhosh annakili, [Food photography image](https://commons.wikimedia.org/w/index.php?title=File:Food_photography_image.jpg&oldid=1256263239), CC0-1.0 | Real camera detail, 3264×1472, orientation 3, private EXIF and 1799549 appended DualShot/refocus bytes |

These files retain the original publisher bytes, including camera metadata and attachments, so privacy/preservation checks exercise the real input. The [native camera record](../../docs/reference/2026-10-16-jpeg-camera-attachments.md) records Safe/all removal, byte-exact Keep and independent Lossless decoding. Savings on the compound camera photo include attachment removal; compare that separately from encoder savings. Original download URLs and exact hashes live in the manifest. The publishers' CC0 dedications supply redistribution rights; the generator does not grant them.

Generate comparison output **outside this tree** so results never become fixture input:

```sh
smolpng compare testdata/corpus --recursive --output-dir artifacts/corpus-review --json
smolpng compare testdata/corpus/png --recursive --preset web --png-mode lossless --png-mode quantizer --png-mode pipeline --output-dir artifacts/png-review
```

Both commands use the real installed local tools. Ordinary unit tests use controlled process boundaries and do not establish the tools' visual quality or speed.

## Color-profile controls

`internal/testutil/profile-srgb.icc` is a frozen 588-byte LittleCMS RGB profile, originally generated through Pillow 12.3.0 `ImageCms.createProfile("sRGB")`. Its copyright tag says “No copyright, use freely”; the control data share this corpus's CC0 dedication. `internal/testutil/icc.go` embeds it, exchanges the red/blue XYZ colorants for the synthetic variant and clears the changed profile's ID. The inherited description in that variant is not a claim that its interpretation is sRGB. Go alone then creates the four swatch images and complete PNG iCCP/JPEG ICC_PROFILE carriers. No CMS or Python dependency is added to generation, tests or the application.

The 216-color swatches allow ordinary-quality quantization. In the swapped variant, stored red/blue channels change with the profile so the displayed image remains comparable. Ignoring the profile or restoring it after conversion to sRGB produces a large, independently measurable error. Unit tests check the complete profile, changed colorants and safe-metadata carriers; independent native CMS rendering is a separate qualification boundary.

## Admission gaps and future media

The first two natural photographs and four valid RGB profile controls are admitted. Broader photographic diversity, representative artwork, real wide-gamut distributions and grayscale/CMYK ICC coverage remain admission/qualification gaps. Small malformed/profile-marker fixtures in unit tests are **not** real color-managed corpus evidence; the controlled RGB cases do not establish those broader claims either. More complex progressive/chroma combinations also deserve native encoder qualification.

A contributed asset needs recorded creator/source, explicit CC0/public-domain or otherwise reviewed redistribution rights, exact hash and why it adds coverage. Keep privately supplied images outside this checked-in tree, for example an explicitly chosen `artifacts/private-corpus/`; review its comparison outputs as private data. Do not relicense imported media automatically.

For another explicitly CC0 original, copy its exact reviewed bytes into the appropriate format/class directory and add its admission fields to the existing manifest, then regenerate derived properties and run check mode. Other licenses need a deliberate rights/contract update; changing a string is not redistribution permission. Merely creating an empty class directory does not prove it is covered.
