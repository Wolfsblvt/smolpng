# Current implementation direction

## Meaning

smolpng now has a substantial PNG/JPEG daily-use source candidate: one engine under deterministic CLI, keyboard and background/drop workflows, with inherited presets, native clipboard adapters, optional TinyPNG, tracked integrations, version-bound Agent Skills, managed OxiPNG and standalone update transactions. This is implemented source, not a claim of a publicly released or fully native-qualified 1.0.

## Current frontier

The source implements original-derived PNG/JPEG candidates, exact target semantics, explicit resizing, safe metadata/orientation handling, atomic outputs, data-only configuration/presets and structured reporting. TinyPNG requires explicit policy and native/environment credentials; fake HTTP tests never consume quota. Desktop installers produce Windows per-user `.lnk` and Linux `.desktop` targets with exact ownership hashes. The signed updater stages bounded archives, verifies all members, preserves prior files for guarded rollback and synchronizes only untouched app-managed Skills.

Local release tooling now builds six platform packages and optional offline Winget, Scoop, Homebrew, Chocolatey and nFPM inputs from actual archive hashes. A separate signing tool requires an explicitly supplied protected key. Neither tool publishes or creates a production trust authority.

## Current refinement boundary

Local `compare` evaluates presets and the separate contributions of pngquant/OxiPNG from identical originals, with JSON/CSV and offline visual reports. Explicit PNG matte preparation is shared across CLI/presets/Custom without adding subject segmentation. The legal corpus now has 24 classified structural specimens, including true progressive JPEG, valid synthetic GPS EXIF and 16-bit alpha.

Right-click configuration/dispatch and Dolphin service generation are implemented. Windows 11 primary-menu support has a small authored native IExplorerCommand bridge, sparse-package builder and protected activation route. Hosted CI has compiled the Windows SDK bridge for x64 and ARM64 and passed the x64 C++ protocol tests. Signature/identity validation and actual Explorer/Dolphin menu behavior remain separate native qualification. The ordinary Go archive does not contain a qualified DLL or signed menu package.

## Remaining source refinements and external boundaries

Final JPEG backend/quality calibration remains evidence-led; jpegoptim is the reversible first backend and Go encodes explicit original-derived geometry changes. Remote geometry/keep-all metadata combinations intentionally refuse unsupported promises. Richer macOS Finder integration, a final installer/MSIX/signing composition and package-manager submission refinements remain after corresponding host/publisher decisions. The website, final art and hosted sibling retain separate writer boundaries.

Wolf explicitly deferred intermediate native smoke tests to let this source line continue. Windows/Linux/macOS graphical clipboard, file-picker paste, real optimizer behavior, native shortcut execution, running-program update replacement and keyring availability still need direct observations. The repository is now public under its existing licenses. Public releases, package registration, signing/feed operation, Store submission and legal brand qualification have not occurred. These limits are not substitutes for implemented source or grounds to describe cross-builds as desktop tests.

## Continuation

The root Git checkout and GitHub implementation PR now carry the complete authored history; the earlier Git recovery ZIP is historical intake. Current Work belongs to the owning implementation conversation. Root commands, current architecture and qualification instructions live in Development, Architecture, Releasing, JPEG, Remote, Integration and Updates rather than an accumulating roadmap here.

## Local implementation evidence, 2026-10-16

The implementation candidate passed all 11 hosted Verify jobs at `f70486ceff89008e98a6d79f033472c4b793c94c`. Local Go 1.27.2 built the console executable. Real pngquant 2.17.0 and OxiPNG 10.2.1 passed every Auto PNG corpus case and eight of eleven strict lossless cases. Three lossless invocations refused during optimizer discovery; nearest-target qualification exceeded the helper deadline. These remain unresolved native observations, not pixel-corruption findings or a complete native pass. [The qualification record](reference/2026-10-16-local-implementation.md) separates exact candidates and boundaries.
