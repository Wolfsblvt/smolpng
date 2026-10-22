# Native macOS Finder and Terminal qualification, 22 October 2026

## Meaning

Unsigned per-user applets at source `04f2eaa1429a12ca42b35c8605010037a2bd7dd1` completed actual native file-open, background PNG/JPEG, interactive Terminal and owned installation/removal journeys on macOS 15.7.9 Intel and macOS 26.6.2 ARM64. This record preserves their original evidence and the failed preparations, including a separately reproduced macOS 26 Intel runner-service failure. It qualifies those exact native consumers and hosts, without claiming physical Finder dragging, a public package, final appearance, production signing or every Mac session.

## Exact native cut

The contribution starts at `d91b9818494e08adce4cca8d802e0363d2b62d8b`. Both native application and matching tagged fixture were built with Go 1.26.6 and `CGO_ENABLED=0`. The application reports source distribution, version `0.3.0-dev` and the exact source commit above.

| Native specimen | Runner image | Application SHA-256 | Result |
| --- | --- | --- | --- |
| macOS 15.7.9, build 24G830, amd64 | `macos15` / `20260824.0482.1` | `f8377e22264f6e02f4bbc69d0b4e47f3b732ff443a0eca75d4ab27dcc6210b59` | Complete, 20.62 s |
| macOS 26.6.2, build 25G83, arm64 | `macos26` / `20261013.0424.1` | `b8c7f7bd9c822d47dc8b1ed0db08683cc9e32bb29d5e7814534a1393dc73ddf6` | Complete, 16.69 s |

[Hosted run 39721257991](https://github.com/Wolfsblvt/smolpng/actions/runs/39721257991) contains both successful jobs and a failed macOS 26 Intel job. The run's aggregate conclusion is **failure**, not green. Its artifacts are `finder-macos15-amd64-04f2eaa1429a12ca42b35c8605010037a2bd7dd1-39721257991` and `finder-macos26-arm64-04f2eaa1429a12ca42b35c8605010037a2bd7dd1-39721257991`. The [ordinary Verify run](https://github.com/Wolfsblvt/smolpng/actions/runs/39721258066) separately passed all eleven jobs on this source.

Each successful native journey retains 22 command receipts. Twenty commands succeeded; two deliberately refused modified-bundle replacement/removal with exit 5. The complete journey exercised:

- Native compilation, four previewed applet destinations, repeat installation and current ownership status.
- Launch Services file-open events with one PNG, then a PNG/JPEG batch. Executable, data and input paths contain spaces, quotes or Unicode; an input includes a literal shell-substitution fragment. The exact original filesystem objects remained unchanged and no injection marker appeared.
- Decodable output files, collision increments and public `last --json` recovery.
- An explicit interactive launcher and an Auto launcher configured with `drop.mode=interactive`. Each reached the actual Terminal recipe screen, accepted Auto, produced two outputs, reached Result and completed Done. Original TTY and screen-history transcripts are retained.
- Preservation of an added user file, whole-plan replacement/removal refusal, restoration of the owned fixture and removal of all four unchanged applets. The final registry was empty and app configuration remained present.

The applets invoke the existing Go engine. No optimizer is bundled. The receipts bind the separate Homebrew pngquant/jpegoptim builds and catalog OxiPNG binary by hash. The native account is a disposable hosted account; only application data uses the explicit portable home. Pre-existing native destinations are refused. Remote use, notifications and clipboard writes are disabled for this journey.

## Preserved failures and corrections

The first ARM64 file-open succeeded, then the fixture incorrectly required one Unicode spelling. Native aliases returned a decomposed spelling of the same original file. The comparison now uses filesystem identity; the original failed report is preserved.

An external hosted-runner Apple-event consumer requested macOS Automation consent to read Terminal. The failure screenshot shows the real application already at its recipe screen. The corrected independent consumer runs in its own Terminal session and controls only the exact fixture window. No privacy setting or permission grant was changed. Subsequent fixture failures exposed AppleScript iterator dereferencing and Terminal's window-versus-tab close contract; these were corrected without changing product behavior. A transient editing error removed the foreground loop in `246880f`; `32b3082` restored it before any passing claim. That head is not qualification evidence.

A separate local Windows compile-only attempt exceeded its tool transport timeout and returned no verdict. It supplies no native or cross-build pass; the actual hosted Mac compilations above establish their own boundary.

The macOS 26 Intel hosted image repeatedly stalled in AppKit nib initialization before the Go handoff. Captured stacks wait in IconServices XPC; a preflight records 18 consecutive agent crashes and `Abort trap: 6`. The same Go application bytes completed the comparison on macOS 15 Intel. [Upstream runner-image issue 14751](https://github.com/actions/runner-images/issues/14751#issuecomment-6214439665) independently reports the matching standard-runner fault and the unaffected host pair. This supports an image-service diagnosis; it does not prove behavior on a healthy physical macOS 26 Intel host. No system-service workaround was applied and every affected specimen remains incomplete.

Ordinary qualification therefore uses macOS 15 Intel and macOS 26 ARM64. The workflow retains a manual macOS 26 Intel selection for a new healthy-image observation after the upstream repair or a qualified-host occasion. This changes the evidence host, not the implemented macOS/Intel capability or the standing of the failed observations. Current custody and remaining judgment stay in [Work #71](https://github.com/Wolfsblvt/smolpng/issues/71).

## Retained original bytes

[report.json](2026-10-22-macos-finder/report.json) retains **388 original files from 15 specimens** across seven hosted attempts: source/host/tool identities, command receipts, native dictionaries, screen-history transcripts, complete/incomplete reports and selected failure surfaces. UTF-8 payloads and base64 PNG payloads reconstruct their original bytes; every file's length and SHA-256 was checked after serialization. Fixture executables and app bundles are excluded.

The payload is **3,741,386 bytes**, SHA-256 **`183759da4aaba246ebc4737c7c195f0670e80d334b3d02ef58eccf6976d823cd`**. The successful test-log hashes are `65090a394010a0588befadde2be13529860fe0c7abc50a5bd64ffa9baf9daa55` (Intel) and `4af093418eadbf7ed40ac21d8272ad26859800895eb12d1b6d02432256028aa8` (ARM64). Failed runs remain failed; cancelled runs and an interrupted local transport are not promoted to passes.

Deterministic integration tests separately exercise interrupted bundle publication/restoration, whole-plan preservation and configuration argument separation. They do not establish an actual process interruption of the native GUI journey. Neither this record nor passing source checks complete the separate Windows graphical candidate, select final product assets or authorize publication.
