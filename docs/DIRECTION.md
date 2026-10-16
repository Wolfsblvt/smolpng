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

Native Windows PNG/JPEG compression and application-created shortcut activation now have direct observations, with the exact boundaries recorded below. Windows/Linux/macOS graphical clipboard, physical file dragging and paste, rendered notifications, running-program update replacement and keyring availability still need direct observations. The repository is now public under its existing licenses. Public releases, package registration, signing/feed operation, Store submission and legal brand qualification have not occurred. These limits are not substitutes for implemented source or grounds to describe cross-builds as desktop tests.

## Continuation

Wolf clarified on 16 October 2026: “The general goal is implement everything that was planned anyway.” The complete Vision remains the destination without an urgency-driven temporary scope. His selected first everyday Windows workflow is “Drop files onto a shortcut”; that orders native qualification without replacing the remaining product plan.

The root Git checkout and GitHub implementation PR now carry the complete authored history; the earlier Git recovery ZIP is historical intake. Current Work belongs to the owning implementation conversation. Root commands, current architecture and qualification instructions live in Development, Architecture, Releasing, JPEG, Remote, Integration and Updates rather than an accumulating roadmap here.

## Local implementation evidence, 2026-10-16

The repaired native Windows amd64 build at `2abae9390167d35dbca942c6a69ec83bca90cfd9` passed all 29 real PNG checks with pngquant 2.17.0 and OxiPNG 10.2.1: Auto and strict lossless across eleven specimens, target success/refusal/nearest, binary stdout and partial batches, with originals preserved. A captured three-second version-probe deadline supplied one demonstrated cause of intermittent optimizer refusal; discovery now permits fifteen seconds while respecting earlier caller cancellation. Native Go 1.27.2 root tests/vet and all eleven hosted Verify jobs passed. [The complete PNG record](reference/2026-10-16-windows-png-qualification.md) binds this evidence to its executable and distinguishes the earlier partial runs from the still-unobserved JPEG, graphical and distribution boundaries.

Native JPEG qualification then exposed a separate parser defect: jpegoptim's `v1.5.6` banner was misread as `5.6`. The repaired Windows amd64 build at `b90a5165c7816c7cee3fc8c397872898c058807d` passed 43 JPEG checks, including all thirteen structural specimens under Auto/lossless, all eight orientation resizes, metadata/refusal and target contracts, binary stdout and partial batches. [The JPEG record](reference/2026-10-16-windows-jpeg-qualification.md) preserves exact provenance and independent decoding evidence. Photographic/ICC calibration, other native hosts and graphical/distribution qualification remain separate.

Native shortcut activation at `e611fd3ee768d349e841f87e6c008a2e49a461eb` then passed with a PNG/JPEG batch, Unicode/file-argument delivery, original preservation, incremented collision outputs and public last-result recovery. It repaired a false error after successful compression when the GUI entry point tried to print without terminal handles. [The shortcut record](reference/2026-10-16-windows-shortcut-qualification.md) distinguishes ShellExecute activation from the remaining physical drag, clipboard/paste and visual observations.
