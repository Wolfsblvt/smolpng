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

Keep retains source metadata on ordinary same-size jpegoptim candidates. Keep plus a geometry change is refused instead of preserving stale dimensions and thumbnails. ICC-tagged grayscale/CMYK images are not silently relabeled as RGB during a resize. Unknown or contradictory orientation and malformed headers fail safely. The decoder validates baseline and progressive markers/scans through the primary JPEG's end-of-image marker; input byte and decoded-pixel limits still apply.

Some camera JPEGs attach depth/refocus, motion or other auxiliary data after the primary still image. Default Safe (`--strip safe`) and strip-all (`--strip all`) remove this attachment before optimization and report its byte count in a warning. The output is a still image; attached camera features do not survive. The original is preserved unless in-place replacement was explicitly requested. Safe still preserves display orientation and ICC; strip-all may need an upright re-encode as described above.

For these compound files, `--strip keep` preserves the complete file **unchanged**, including the attachment. It explains why compression is skipped; an unmet size target remains exit 6. Geometry changes with Keep remain refused. smolpng does not interpret the attachment or claim that copying opaque camera bytes after a rewrite would preserve internal offsets. `inspect --json` includes `info.appendedBytes` when such bytes exist. Results include `jpegAttachment.bytes` and the selected `jpegAttachment.keep` policy; output/status establish whether a file was committed. Enabled drop notifications show a short camera-data summary, distinguish saved outputs from unsuccessful attempts, and direct you to `smolpng last` for the complete warning and report. A target miss without an output reports detection rather than completed removal.

TinyPNG accepts upright JPEG under explicit permission. A remotely changed orientation or ICC interpretation is rejected; remote keep and rotated JPEG input are refused before upload. Smart clipboard supplies image PNG when conversion is unambiguous; ICC-tagged JPEG falls back to actual file/path representations instead of copying a miscolored image. Explicit image-only copy reports the limitation.

## Evidence and qualification

Generated CC0 baseline, grayscale, high-chroma and eight-orientation specimens accompany parser, metadata, target, original-preservation and process-boundary tests. Ordinary tests do not execute jpegoptim or consume TinyPNG quota. [Native Windows structural qualification](reference/2026-10-16-windows-jpeg-qualification.md) exercised real jpegoptim, targets, progressive output and orientation cases. [Real camera attachment qualification](reference/2026-10-16-jpeg-camera-attachments.md) independently checked Safe/Lossless, strip-all and byte-exact Keep against a CC0 Samsung photograph. Final quality selection, broader photographic/ICC/CMYK calibration and native clipboard rendering remain separate.

Upstream command contract: https://github.com/tjko/jpegoptim/blob/main/jpegoptim.1 . Current source and older releases have differing GPL license vintages; the installed tool retains its own exact license. External invocation is not permission to relicense or bundle it.
