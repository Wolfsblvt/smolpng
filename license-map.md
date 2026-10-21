# License map

## Meaning

The application, data-only examples, generated and admitted test media, and future website/brand work have separate rights. This map identifies current material and the already-selected future boundaries; no custom brand policy is presented as reviewed legal text.

| Path/material | License |
| --- | --- |
| Project Go code, generator, ordinary docs and repository tooling | MIT, root `LICENSE` |
| `testdata/corpus/**` generated PNG/JPEG images, admitted original CC0 photographs, inventory documentation and manifest | CC0-1.0, local `LICENSE`; imported originals retain their publishers' CC0 dedications recorded in `manifest.toml` |
| `internal/testutil/profile-srgb.icc` generated ICC control data | CC0-1.0, as for the generated corpus; its embedded original copyright tag also permits free use |
| `internal/testutil/profile-gray.icc`, `profile-cmyk.icc`, `profile-display-p3.icc` and `jpeg-cmyk-seed.jpg` | CC0-1.0; exact profile sources and original seed construction are in `internal/testutil/jpeg-color-inputs.md` |
| `LICENSES/**` and third-party source notices | Their own unchanged terms |

Standard MIT, CC0-1.0 and AGPL-3.0-or-later texts are installed in `LICENSES/`. Compiled dependency notices are in `third-party-notices.md`.

Future copyable presets use CC0. Future `site/src/` uses AGPL-3.0-or-later; protected editorial copy stays separately under `site/content/`. Future sibling smolpng-web is AGPL-3.0-or-later. Brand files require the artist grant and qualified custom policy before they are added. Those future selections do not relicense the current MIT application.

The standard distribution includes neither pngquant, OxiPNG nor jpegoptim executables. The software license does not grant rights to nonexistent or later excluded official artwork. Plain product references are not an invented registered trademark claim.

`presets/examples/**` is now present and dedicated to CC0-1.0 by its local `LICENSE`. The user-facing configuration reference remains MIT.

`skills/smolpng/**` and its embedded/package representation are MIT; the package includes its own MIT `LICENSE`. The binary contains guidance, not optimizer executables.
