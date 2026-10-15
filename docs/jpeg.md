# JPEG implementation

## Meaning

JPEG is implemented through the existing input, recipe, output and result flow. This first backend uses separately installed `jpegoptim` 1.4+/1.5.x for ordinary entropy-only optimization and lossy candidates. It is a reversible implementation choice, not a claim that the final jpegli/jpegoptim quality benchmark has been decided. No JPEG optimizer is bundled or automatically downloaded.

## Use

```sh
smolpng compress photo.jpg --json
smolpng compress photo.jpeg --preset lossless --json
smolpng compress photo.jpg --target-size 100KiB --quality 55-85 --json
smolpng compress photo.jpg --resize 1024x768 --json
smolpng deps configure jpegoptim /path/to/jpegoptim --apply
```

Explicit `--jpegoptim-path` overrides configured and system discovery. Output remains JPEG with the original `.jpg` or `.jpeg` extension. Directories discover PNG and JPEG; a preset's formats are enforced before optimization. Numerical JPEG quality is the encoder's requested quality ceiling, not a perceptual score or the same metric as pngquant. The wrapper checks actual bytes rather than trusting jpegoptim's approximate target-size switch.

Every lossy attempt begins from the original. Explicit resize uses Go's JPEG encoder directly from the original decoded, oriented image and permits only a lossless optimizer tail. This is functional JPEG support, not a quality-equivalence claim with jpegli. Local entropy-only candidates must match decoded pixels exactly. No-growth and target-missed/accept-nearest semantics are shared with PNG.

## Metadata and orientation

Safe preserves ICC and display orientation while replacing EXIF with an orientation-only record. GPS, camera information, thumbnail offsets, comments and private EXIF fields do not survive. Pixel-identical cleanup therefore does not require a lossy rotation. A geometry change normalizes all eight EXIF orientations first, then encodes an upright JPEG. Strip-all similarly normalizes orientation before removing its tag. Strict lossless refuses orientation-normalizing re-encoding.

Keep retains source metadata on ordinary same-size jpegoptim candidates. Keep plus a geometry change is refused instead of preserving stale dimensions and thumbnails. ICC-tagged grayscale/CMYK images are not silently relabeled as RGB during a resize. Unknown or contradictory orientation and malformed headers fail safely. The decoder accepts baseline and progressive scans but rejects unbounded/trailing data.

TinyPNG accepts upright JPEG under explicit permission. A remotely changed orientation or ICC interpretation is rejected; remote keep and rotated JPEG input are refused before upload. Smart clipboard supplies image PNG when conversion is unambiguous; ICC-tagged JPEG falls back to actual file/path representations instead of copying a miscolored image. Explicit image-only copy reports the limitation.

## Evidence and qualification

Generated CC0 baseline, grayscale, high-chroma and eight-orientation specimens accompany parser, metadata, target, original-preservation and process-boundary tests. Ordinary tests do not execute jpegoptim or consume TinyPNG quota. JPEG quality, progressive output from the real optimizer, CMYK/ICC edge cases and native clipboard rendering still need corpus/desktop qualification.

Upstream command contract: https://github.com/tjko/jpegoptim/blob/main/jpegoptim.1 . Current source and older releases have differing GPL license vintages; the installed tool retains its own exact license. External invocation is not permission to relicense or bundle it.
