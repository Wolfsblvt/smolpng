# Decisions

## Meaning

These are durable smolpng choices whose reasons cannot be reconstructed from source alone. Wolf's frozen product and licensing briefs select the destination; this record separates that destination from provisional implementation and sequence. Routine progress is not recorded here.

## Go, one engine, native boundaries

**Decision:** Retain Go and share one image-job engine across CLI, TUI, and drop workflows. **Why:** A portable compiled orchestrator with native clipboard support fits the selected human/agent experience without an additional runtime or frontend framework. **Rejected:** Reopening the language for preference, separate interactive compression logic, and a local MCP server around one deterministic command. **Consequence:** Interfaces are earned at process, clipboard, HTTP, secrets, and OS boundaries; the domain admits JPEG without an empty provider framework.

## Local-first is permission, not credential detection

**Decision:** Auto defaults to local-only. TinyPNG participates only through explicit policy, and imported presets cannot override a global prohibition. **Why:** A key says the provider is usable, not that images may leave the machine. **Rejected:** Silent upload fallback. **Consequence:** Policy is checked before any remote operation; reports expose actual remote use; ordinary tests never spend quota.

## Originals and exact targets

**Decision:** Every lossy candidate starts from the original, outputs are atomic, and targets mean `bytes <= requested`. **Why:** Success must be dependable for hostile upload limits without compounding generation loss or corrupting source assets. **Rejected:** Repeated recompression, approximate targets labeled success, and automatic resizing without permission. **Consequence:** Target-missed is distinct; nearest output requires explicit acceptance; strict lossless checks decoded pixels including hidden RGB.

## Presets are the extension surface

**Decision:** Immutable built-ins plus data-only inherited user/project recipes, including clipboard preview/import and pinned placement. **Why:** Shareable text provides useful customization without code execution or a plugin infrastructure. **Rejected:** Hooks, scripts, secrets, arbitrary environment, executable overrides, and unrestricted provider arguments inside imported recipes. **Consequence:** Parse and validate before import; exports preserve inheritance or explicitly resolve it.

## Mixed licensing and separate optimizers

**Decision:** Application/Skill/docs MIT; project examples and corpus CC0; future site code and hosted sibling AGPL-3.0-or-later; protected editorial and brand assets separate. Standard distributions bundle neither optimizer. **Why:** The application should be easy to embed and sell, presets easy to copy, hosted code improvements available, and commissioned identity protected under an actual grant. **Rejected:** Calling AGPL a ban on reuse, imposing branding restrictions on MIT code, or publishing an unreviewed custom license. **Consequence:** Exact path map and third-party notices; artist/legal and publication boundaries remain unclaimed.

## Installed Skills are maintained user files

**Decision:** Track every app-created destination and hashes, including custom paths. **Why:** Updating instructions without silently destroying local edits is part of agent usability. **Rejected:** Fire-and-forget copies and unconditional overwrites. **Consequence:** Only exact unmodified managed installations auto-sync; custom destinations are offered synchronization; manual executable replacement prompts only for changed embedded content.

## Source provenance

These decisions adopt the owner-supplied `smolpng-product-brief.md` (SHA-256 `0b3eeacb11cd63f816e76034ac1b510f5c6f3d4f62d6b6a19ec916c1cb581232`) and licensing boundary (`8b3aecb93ce0a36c8151d818423f37ea635f9c895506ac48024f76fd19cb032b`). The direct implementation assignment authorizes private source delivery and a local Git/ZIP alternative when a connector fails. The raw design conversation is supporting history, not a competing live specification.


## Human workflow uses one orchestration path

The keyboard and drop commands consume typed results from the same config/input/output/engine orchestration as `compress`. Clipboard source is snapshotted once per human retry loop; output copying cannot become the next lossy input. Deterministic commands do not infer clipboard input, prompt, retain history or copy outputs without explicit selection.

Huh/Bubble Tea and the cgo-free desktop clipboard package are retained for real keyboard forms and multi-representation copy. Native runtime availability is reported/qualified, not inferred from compilation. A small project-written text header remains replaceable. Plain and typed-name accessibility modes are available without inventing numbered-choice menus.

Settings preserve handwritten TOML via byte-range edits from the pinned parser, validate both the project and standalone-user view, and recheck the file snapshot under a native lock. An unsupported document shape refuses instead of rewriting the whole file. App-owned preset storage remains separate.

## Pinned dependency delivery, not upstream-latest execution

Managed installation currently admits OxiPNG 10.2.1 only, on the five platforms with a selected native upstream archive/digest. Full archive verification precedes any version probe. Immutable retained versions and a small active index avoid overwriting a running Windows executable. Modified/untracked bytes are refused, and old app catalogs cannot downgrade a newer managed version. Explicit saved paths still win; system/managed preference is a user-only setting. Automatic pngquant and native Windows arm64 OxiPNG remain unqualified, not guessed from a nearby asset name. This does not authorize bundling, Store distribution or application self-update.

## Explicit TinyPNG implementation

The optional HTTP adapter uses no provider SDK or retries. Authorization is obtained only after policy chooses remote use. A single original-derived upload is preferred over a quota-consuming resize search. `keep` and explicit local resize reject remote processing before upload rather than pretending to meet preservation contracts. Unknown output hosts/redirects are rejected. The native secret library is pinned and license-noticed; no service credentials were used during implementation.

## JPEG first implementation

Use separately installed jpegoptim for ordinary local JPEG optimization, with a pure Go original-derived path for explicit resizing. Do not add a native codec-linking dependency to the CLI for this candidate. Final codec/quality selection remains subject to corpus evidence. Safe strips private metadata by reconstructing minimal orientation EXIF, not by preserving the whole camera record or re-encoding pixels merely to remove GPS.

## Desktop installer boundary

Use per-user native `.lnk`/`.desktop` launchers with exact recorded file hashes. Do not install context-menu registry handlers, MIME defaults, machine PATH entries or privileged integration implicitly. Keep the no-console launcher distinct from interactive execution. Journal each file installation so interrupted operations cannot silently claim unowned artifacts.
