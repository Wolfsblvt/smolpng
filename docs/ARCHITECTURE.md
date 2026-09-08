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

`internal/config` now resolves strict TOML user/project layers, app-managed presets and invocation fields into these same recipe and job options. Pointer fields preserve explicit false/zero. Provenance records the owning file and TOML key. Imports write a separately owned `presets.toml`, preserving hand-edited configuration comments. The user remote grant caps project/preset requests. Native clipboard, TUI, last-result state, installed Skill hashes and updater decisions are separate boundaries, not another compressor. The full JPEG and hosted sibling destination remains in Vision; it does not require an empty registry or local MCP server.

## Installed Agent Skill

`skills/` embeds the user-facing package; `internal/skill` owns version-independent bundle hashes, explicit destination management and a journaled directory transaction. Inspection is read-only. Replacement and removal require exact last-installed hashes, preserve user changes, and use an OS advisory lock plus atomic no-replace directory moves. Standard destinations may synchronize during a future in-app update; custom destinations require consent. The keyboard UI can consume a persistent changed-bundle prompt without contaminating the deterministic CLI. [Agent Skill](agent-skill.md) owns the exact lifecycle and recovery contract.


## Shared human workflow

`internal/cli` resolves exactly the same config, source/output plan and `workflow.Batch` for deterministic, interactive and background calls. Human/drop presentation receives a typed report through the same orchestration boundary, not by parsing its own JSON or rebuilding a compressor. `internal/ui` owns keyboard menus and Huh forms. Clipboard reads are frozen for a human recipe-retry loop so automatic output copying cannot turn the next attempt into lossy recompression of the previous result.

`internal/desktop` owns native clipboard formats and browser/folder/notification commands. Smart output makes one multi-representation clipboard write. Native actions use fixed argument arrays; the Windows notification helper is fixed code with escaped data. The no-console Windows entry point selects drop explicitly and requests a console only for the configured interactive variant. No generic daemon or local MCP server exists.

`internal/state` bounds human-result retention. Ordinary `compress`, `doctor` and `skill status` neither create a first-run wizard nor silently write human history. User setting edits use the pinned TOML AST's value ranges under an advisory lock with a byte-snapshot recheck, preserving comments and unrelated fields. Preset storage remains a separate owned file.
