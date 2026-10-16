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

## Update trust and ownership

Use standard-library Ed25519 over exact manifest payload bytes, plus signed hashes for archives and each member. Ship no invented signing key or feed. A short-lived copy of the current executable handles locked-program replacement, exact snapshots permit rollback, and the new binary synchronizes its own embedded Skill only after identity verification. Package-manager ownership always wins over self-update.

## Explicit alpha flattening, not background segmentation

Wolf's final refinement request raised alpha/background removal. The reversible implementation is a named opaque matte in CLI/Custom/presets; it never changes the default preservation policy. Subject cutout and colour-key deletion are distinct editing products and were not silently added. Strict lossless and remote-source equivalence remain enforced. See `background.md` for the exact encoded-channel and colour-profile limits.

## Optional primary context menus, 2026-09-08

Wolf selected an opt-in one-action image menu and an alternative submenu with explicit replacement, interactive mode and user presets. Single-action Make smol preserves originals. The submenu's destructive first action says replace original. All shell invocations remain local and receive exact file selections, never clipboard fallback or project-discovered command behavior.

Windows 11 primary-menu support uses a minimal C++ IExplorerCommand bridge plus separately signed sparse package identity. Static registry verbs are not presented as equivalent primary-menu support. The Go application remains the single compression/config/result engine; the native component only supplies menu/selection/launch ABI. DLL compilation/signing/activation are native host boundaries, not completed by cross-building Go. KDE Dolphin gets file-type-scoped, owned service files; its whole-service Terminal setting requires separate grouped background and interactive services. Other Linux file managers remain unclaimed.

## Classified calibration corpus and evidence scope

The canonical checked-in media now live under PNG/JPEG class directories. Twenty-four project-generated CC0 specimens include an actual minimal two-scan progressive JPEG, true TIFF GPS metadata with synthetic coordinates, and 16-bit transparency. These are structural tests, not substitutes for licensed photographs or real ICC color-managed evaluation. Generator check mode detects unmanifested media; normal evaluation output belongs outside the corpus. No third encoder or quality default is selected from fake-process timing or a diagnostic RMSE.

## Public source and protected main, 2026-10-16

**Decision:** Wolf authorized making `Wolfsblvt/smolpng` public and applying Leitsatz GitHub standards. The existing MIT application and CC0 example/corpus licenses remain unchanged. **Why:** The product was selected as open source; public standard-runner CI avoids consuming the private-repository Actions allowance. Wolf also explicitly rejected leaving `main` unprotected. **Superseded:** The source-publication hold in the original private bootstrap grant. **Current consequence:** Public pre-release source, deletion/non-fast-forward/linear-history safeguards for `main`, read-only CI and private security reporting. Releases, package publication, signing, hosting and external-contribution acceptance remain distinct decisions. Mechanical automation activation still needs a qualified App route. **Sources:** Wolf’s direct instructions in the implementation conversation on 2026-10-16; [GitHub public-runner billing](https://docs.github.com/en/actions/concepts/billing-and-usage); the installed Leitsatz GitHub repository defaults.

## JPEG attachments are a separate preservation boundary, 2026-10-16

**Decision (ADD):** Safe and strip-all produce the primary still image from a JPEG with appended camera data and visibly warn that the attachment is removed. Keep returns the complete original file unchanged with an explanation; explicit geometry changes with Keep remain unsupported and an unsatisfied target remains a target miss. **Why:** A real Samsung photograph has a valid primary JPEG followed by a substantial DualShot/refocus payload. Rejecting the file prevents ordinary still-image compression; rewriting the primary JPEG while copying opaque camera data could invalidate internal offsets or auxiliary behavior. **Rejected:** Treating every byte after EOI as malformed, silently discarding the attachment under Keep, or claiming that opaque payload copying preserves camera-specific structures after a rewrite. **Current consequence:** Validate the primary JPEG structurally through its EOI, treat subsequent bytes as opaque bounded input, strip them before optimization under Safe/all, and never optimize a compound file under Keep. Originals retain their existing output-preservation contract. The application does not interpret or promise playback/editing of depth, refocus, motion or other attached content. **Sources:** Wolf accepted the proposed Auto/Safe versus Keep policy in this implementation conversation on 2026-10-16 (“Yes, that fits”); the original CC0 [Food photography image](https://commons.wikimedia.org/w/index.php?title=File:Food_photography_image.jpg&oldid=1256263239), SHA-256 `62842b3086bb8ec0ddd119b87a9c64055a28f0dcf3d890af7b0c4891248bcb8b`, has a primary JPEG ending at byte 732568 and 1799549 appended bytes containing DualShot/ArcSoft/SEFT identifiers. This is an observed compatibility case, not a complete Samsung format specification.
