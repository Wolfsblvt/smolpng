# Architecture

## Meaning

The current static-PNG CLI has one format-neutral recipe/result domain and one compression engine. This document describes actual components and the accepted seams that later TUI, drop and agent consumers must reuse.

## Data flow

`cmd/smolpng` installs cancellation. `internal/cli` maps Cobra flags into recipe/input/output options. `input.Resolver` plans sources without reading images or downloading; `output.Plan` resolves all destinations before optimization. `workflow.Batch` owns per-image deadlines and original preservation. `workflow.Engine` creates original-derived candidates and selects an actual validated result.

`process.Runner` uses direct argument arrays without a shell and removes `TINIFY_API_KEY` from optimizer environments. Discovery resolves invocation, configured, system and managed paths with local version probes. Managed downloads are not yet implemented. The clipboard is an injected boundary, not yet a native desktop claim.

## Image invariants

PNG chunks and CRCs are validated; dimensions are bounded before decode. Every quantization starts from the original, or a fresh authorized resize of the original. Only the lossless OxiPNG tail receives a quantized candidate. Strict lossless compares unassociated 16-bit channels including invisible RGB. Target preference is satisfaction, area, declared recipe rank and bytes, not a measured perceptual score. Only the incumbent candidate is retained between attempts.

Safe metadata removes text/time fields and preserves rendering metadata. Pixel-identical output can recover original ICC/color information. Quantized output must retain the encoder's color interpretation because pngquant may convert to sRGB; blindly restoring the old profile would mislabel transformed pixels. Physical and orientation metadata are recovered separately.

## Files and future consumers

Outputs use a synced temporary sibling and atomic rename/MoveFileEx, or exclusive Unix hard-link publication for no-clobber writes. A race-time collision fails instead of changing a reserved batch plan. In-place checks original bytes again before replacement; it is not a lock against a hostile directory owner.

User/project configuration and presets will produce these same recipe and job options. Native clipboard, TUI, last-result state, installed Skill hashes and updater decisions are separate boundaries, not another compressor. The full JPEG and hosted sibling destination remains in Vision; it does not require an empty registry or local MCP server.
