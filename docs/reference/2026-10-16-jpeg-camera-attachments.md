# JPEG camera attachment qualification, 16 October 2026

## Meaning

A real Samsung photograph exposed smolpng's rejection of valid primary JPEGs with appended camera content. The repaired Windows build produces a warned still image under Safe/all and preserves the complete original byte-for-byte under Keep. All seven native policy checks passed; independent Pillow decoding confirmed upright pixel identity for explicit Lossless. This record binds those results to their input, executable and tools. It does not select final JPEG quality defaults, qualify auxiliary camera features, or claim rendered desktop notifications.

## Subject and provenance

- Source commit: `3e102e135f8067ecc6dc6101bdf5ffd3e6414b2b`; Go 1.27.2, Windows amd64, CGO disabled, standalone `0.3.0-dev.1`, built with `-trimpath -buildvcs=false`.
- Executable SHA-256: `fdce1ea82dd82fdf18ee3ee34d4ab7c31453e90de9b5076acdf9f1295da7f427`.
- Separately configured jpegoptim 1.5.6, executable SHA-256 `0a79681b56961ec2c5f45c4f9dff01a64e7274b4f3250c22cde5f09c0bb457f3`. No optimizer was bundled or machine-wide configuration changed.
- Original [Food photography image by Santhosh annakili](https://commons.wikimedia.org/w/index.php?title=File:Food_photography_image.jpg&oldid=1256263239), explicitly CC0-1.0; [original download](https://upload.wikimedia.org/wikipedia/commons/c/c5/Food_photography_image.jpg), 2532117 bytes, SHA-256 `62842b3086bb8ec0ddd119b87a9c64055a28f0dcf3d890af7b0c4891248bcb8b`.
- The photo is 3264×1472 with EXIF orientation 3. Structural marker/scan traversal finds the primary EOI at offset 732568, followed by 1799549 bytes containing DualShot/ArcSoft refocus and SEFT identifiers. The payload was not interpreted as a complete camera-format specification.

The original failed all four built-in comparison recipes at predecessor `e611fd3ee768d349e841f87e6c008a2e49a461eb` with `trailing bytes after JPEG EOI`. A separate CC0 landscape passed that comparison; its results do not qualify this phone image.

## Native policy results

Each command used the original bytes, an isolated existing config home, `--no-project --offline --json` and a fresh explicit output path. Pillow 12.3.0 independently decoded results; no user images, clipboard, provider credentials or remote service were used.

| Selection | Exit | Output bytes | Observed result |
| --- | --- | --- | --- |
| Auto, Safe | 0 | 541651 | Still image, orientation 3 retained, attachment-removal warning; lossy output is permitted |
| Lossless, Safe | 0 | 657362 | Progressive JPEG; independently identical upright pixels; orientation 3 retained; warning |
| Auto, strip-all | 0 | 700262 | Original-derived upright re-encode, orientation tag absent, same display dimensions; warning |
| Keep | 0 | 2532117 | Complete original SHA-256 unchanged; empty encoder pipeline; preservation explanation |
| Keep, target 1 B | 6 | none | Target miss; no output committed |
| Keep, target 1 B, accept-nearest | 6 | 2532117 | Complete original SHA-256 unchanged; target still missed |
| Keep, resize 640×480 | 2 | none | Unsupported preservation/geometry combination refused |

Safe outputs contained only the orientation EXIF entry; Safe/all ended at the primary EOI with no attachment. Source hashes remained unchanged after every case. Native `inspect` reported `info.appendedBytes=1799549` for the original and omitted it for the cleaned Lossless result. Source and Auto output were visually inspected at preview resolution; that is not a general perceptual-quality or color-managed calibration.

The 78.61% Auto saving includes removal of the 1799549-byte attachment and private primary-image metadata. It is not a measurement of encoder savings alone. Auto's pixels need not be identical: the first disposable qualification attempt incorrectly imposed Lossless's pixel contract on Auto, stopped on that assertion, and remained incomplete. Inspection confirmed changed quantization tables; the corrected policy-specific run passed all seven cases. No production fix was made to disguise that harness mistake.

## Source and hosted evidence

Parser and local workflow regressions failed before the repair with the original trailing-data refusal, then passed. The complete local `internal/jpegimage`, `internal/workflow` and `internal/cli` suites passed, including exact EOI boundaries despite marker-like metadata bytes, Keep/target behavior, fake remote upload filtering, source preservation, and drop warning delivery to the notification adapter and retained result. The remote test used a fake boundary and consumed no quota.

[Verify run 38954741846](https://github.com/Wolfsblvt/smolpng/actions/runs/38954741846) passed all eleven jobs on the exact source commit: tests/vet/corpus checks on Windows, Linux and macOS; six platform cross-builds; two native Windows shell compile jobs. These are hosted source checks, not native clipboard or Explorer tests. Forgehand runs were deliberately skipped while activation remained disabled.

The ignored local receipts are under `artifacts/calibration/photographs-20261016/`: `builtins-e611fd3/` preserves the original refusal, `camera-3e102e1/` the incomplete Auto assertion and decoder observations, and `camera-3e102e1-policy/report.json` plus per-case stdout/stderr and outputs the completed run. The executable is under `artifacts/bin/3e102e1/`. These local working artifacts are reproducible supporting evidence; this committed account and original source coordinates carry the durable result.

## Remaining boundary

Keep preserves bytes, not a claim that depth/refocus/motion playback was exercised. Safe/all deliberately remove that capability from the output under Wolf's accepted policy while retaining the original by default. Native toast rendering, graphical clipboard, physical dragging, other host behavior, real ICC/CMYK calibration, broader photographic corpus admission and final quality selection remain separate. This compatibility repair does not complete the product or authorize publication/signing.
