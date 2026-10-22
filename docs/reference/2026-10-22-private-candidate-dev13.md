# Private candidate 0.3.0-dev.13

## Meaning

The single joined private `0.3.0-dev.13` candidate binds accepted Finder, Linux and Windows source at `319d5e87cef6d772020732107139ebf21fb9b0db` to six inspected archives, two actual source-history restores and 79 passing native Windows PNG/JPEG checks. Its desktop, TinyPNG and primary-menu preparation packets also passed against the extracted candidate. This dated record preserves those original reports and their limits. It does not complete [Issue #36's joined Windows journey](https://github.com/Wolfsblvt/smolpng/issues/36), relabel earlier lifecycle observations, qualify foreign archive runtime or publish a release.

## Exact artifacts

The files remain local under ignored `artifacts/candidates/0.3.0-dev.13-319d5e8/`. These coordinates are not download URLs. All platform filenames begin `smolpng-0.3.0-dev.13-`.

| Artifact suffix | Bytes | SHA-256 |
| --- | ---: | --- |
| `windows-amd64.zip` | 37,281,796 | `bf7b73c9ea1b69d5c57b02177bc550e51a8bde9f32b8b6f2843dcc9bae264fd3` |
| `windows-arm64.zip` | 34,676,698 | `525090a3ba09e6fa403daeec568b9b5f713735ccaddc64d31747dbe44f04789c` |
| `linux-amd64.tar.gz` | 24,990,701 | `850590c218afa612c91e0cf2a5099dccad4bd24e0a73a8e77ebf2e0e12eaf410` |
| `linux-arm64.tar.gz` | 23,584,018 | `91085c62066db83e7f6ef41d582c1c394ee4e81ec286dc694df49f78a1810b03` |
| `darwin-amd64.tar.gz` | 25,016,735 | `2199c79efcb73c2a20c3c0b6955a3e92e88ceb76ba18b0f7fb5b0da3571dc1ec` |
| `darwin-arm64.tar.gz` | 23,982,765 | `fe0e19517818a279cd3e0f510069196cfd94323dcfbb6b7dd7039ac08dbdd0ab` |
| `source-recovery.zip` | 17,968,525 | `49f42d6a90168333e33f3d2eff62f2b3c07552bc689781da5e04ef7cb6cd6d37` |

The separate complete-history `history.bundle` is 8,991,174 bytes, SHA-256 `95c4983a59922490af507e57fb6de2c61b96db57e1b8dd027a28b27a8db70625`. `SHA256SUMS`, `SOURCE-SHA256SUMS` and `build-index.json` preserve the artifact joins.

Independent inspection compared every actual archive hash/length, safe unique member path, complete payload set and member digest, manifest identity, staged bytes, committed docs/Skill/corpus/presets and all 63 current notice files. Windows archives have 197 members and 196 hashed payloads each; Unix archives have 196 and 195. PE/ELF/Mach-O architecture and Unix executable modes match the selected targets. Only `smolpngw.exe` uses the Windows GUI subsystem; CLI and qualifier use the console subsystem. Actual Windows VERSIONINFO reads passed for all six extracted x64/ARM64 executables. Reading ARM64 headers/resources is not executing ARM64 code.

The source ZIP contains the committed snapshot and a complete-history bundle. Independent clones from the standalone bundle and ZIP-extracted bundle recovered exact main, passed `git fsck --full` and stayed clean. All 425 actual working files in both restores matched the ZIP snapshot bytes. This establishes local source recovery, not disaster or dependency-cache recovery.

## Build and source evidence

The clean primary checkout remained at the selected source during composition. The maintained command ran once:

```text
go run ./tools/release --version 0.3.0-dev.13 --out artifacts/candidates/0.3.0-dev.13-319d5e8
```

The owned build started at 2026-10-22 12:30:49.9016002Z, ended at 12:34:21.0627462Z and exited 0. Go 1.27.2 windows/amd64 has SHA-256 `b8cd508a0a7aac883cd750e9e75eb306d219e726bdf1cde29009490729437fbf`. The receipt preserves `GOWORK=off`, `GOENV=off`, `GO111MODULE=on`, `GOTOOLCHAIN=local`, `GOFLAGS=-p=1`, empty `GOEXPERIMENT`, `CGO_ENABLED=0`, `GOMAXPROCS=1` and the soft `GOMEMLIMIT=512MiB` target. No hard RSS limit or special process priority is claimed.

The maintained builder's Windows child constructor suppresses console windows while preserving output and status. An attempted discovery of the older wrapper was refused before execution; the existing direct builder was used within its established quiet process boundary. No wrapper or machine configuration was changed.

[Linux PR #74's terminal Return](https://github.com/Wolfsblvt/smolpng/issues/73#issuecomment-6340838047) proves accepted/landed tree equality and carrier retirement. All eleven [post-main Verify jobs](https://github.com/Wolfsblvt/smolpng/actions/runs/39778182311) passed on the candidate source; these include the ordinary source checks and target builds, not runtime execution of the private archives. The accepted proposal's Linux GLib/Dolphin and native Finder observations retain their [Linux](2026-10-22-native-linux-desktop.md) and [Finder](2026-10-22-macos-finder-ownership.md) source boundaries. No intermediate candidate was built between those landings and dev.13.

## Native Windows result

The extracted Windows amd64 files have these SHA-256 identities:

| File | SHA-256 |
| --- | --- |
| `smolpng.exe` | `41ebbba74f01fbf61175345dd17af531ac04f8e0fbfe7aed118bcf264be25e78` |
| `smolpngw.exe` | `53686ab15cb1e97cf9b9cec8e0af5c5089e0471cdd5b2f9a9c48872db41c37cf` |
| `smolpng-qualify.exe` | `3dc42c66a28aeb44936d36b0d664cab3f60fc40c998d3b02c66f423e3bccd6d2` |

The actual host is Windows 11 Home/Core 25H2 amd64, kernel 10.0.26200 revision 9711. App, qualifier and expected source all bind the complete selected SHA. All **79 native PNG/JPEG checks passed**, including Auto/lossless, original preservation, exact targets and nearest, binary stdout, partial/mixed batches, orientation, profile/component controls and supported resize/refusal boundaries.

The separately installed real optimizers were checked before execution and by the qualifier at both provenance checkpoints:

| Tool | Version | SHA-256 |
| --- | --- | --- |
| pngquant | 2.17.0 | `7aadad6a50aa5c40cfcb20f9478c92f6030360a3424bc262ce383dc1e97fb86a` |
| OxiPNG | 10.2.1 | `35ae3980ab831af64f4aecc98f69b81bfa146fa750a61d17ebbb12520128cbc8` |
| jpegoptim | 1.5.6 | `0a79681b56961ec2c5f45c4f9dff01a64e7274b4f3250c22cde5f09c0bb457f3` |

The successful qualifier invocation ran from 13:00:38.7457986Z to 13:02:11.3311095Z and exited 0. It used a new portable home, disabled project lookup and forced offline/remote-never compression. Its complete passing report SHA-256 is `72e05337088520563f24f3bcb9cdd646e95856df282f96cbad131d1598d033b8`.

The preceding runner attempt was interrupted by a one-second tool timeout before a report existed. Its empty `native/` folder remains retained; no app/qualifier process remained when execution resumed into the separately named `native-attempt2/` folder. The interrupted attempt is not a pass.

## Preparation through the extracted candidate

The existing helpers were rebound to dev.13's app/launcher hashes, source, manifest and new isolated output roots. Their dev.12 originals remain unchanged. All preparations use direct hidden child processes.

| Packet | Observed boundary | Report SHA-256 |
| --- | --- | --- |
| Desktop | Nine actual CLI calls: identity/config/dependencies, install and shortcut previews, quiet single/mixed drops, last-result and integration status. Fixture originals unchanged. | `a0037120d0ae98c540f5baf9419c41cbdf60ebcc6ee3dd9b60b711bc9aedf50e` |
| TinyPNG | Eight dry-run/refusal calls, exact batch/report path, collision refusal, invalid child credential before upload, missing policy refusal; no output or original mutation. | `5965e42fac23c7bd0d5d9d6fa36aa428760d6b59c9292f70b02556cf9bdb71d5` |
| Primary menu | Eight calls with redirected LOCALAPPDATA: preview both modes, filesystem-only apply/remove, missing-package refusal; empty final ownership records and no remaining menu file. | `7e8e04df4b36a328d84f5a1e4ae71c6c4a68c8c46d8bd7610f398015bc5e8047` |

Both actual live-helper preflights passed: the synthetic marker reported ready without opening or reading the clipboard, and cleanup preview reported no installed entries. TinyPNG preparation explicitly empties only the child environment credential, preventing fallback to the account-global native keyring. No actual credential entry was read or changed. No upload, clipboard operation, graphical integration activation or native package registration occurred.

These reports remain under `artifacts/validation/desktop-dev13-20261022/`, `remote-preparation-dev13-319d5e8/` and `windows-menu-preparation-dev13-319d5e8/`. They prepare the real journey; they do not prove physical interaction, notification presentation or clipboard consumption.

## Retained evidence and interpretation

The [authored evidence payload](2026-10-22-private-candidate-dev13/report.json) preserves **18 original UTF-8 inputs** with their original SHA-256 values plus two explicitly derived live-preflight observations. Payload size is **335,456 bytes**, SHA-256 `7135b55f4569c1386ef6ccab735303cbc47f8c061e4aa2824658e6fb41a3ed90`. The two original preflight receipts also contain private helper paths, so they remain local; the payload preserves their hashes and selected non-secret results rather than claiming an unchanged original capture. One initial payload assembly refused that private-path inclusion before writing; the corrected assembly publishes only the stated boundary. The wrapper is an authored join, not itself an original capture.

[Dev.12](2026-10-22-private-candidate-dev12.md) separately retains dependency/Skill, updater/rollback, interrupted recovery, locks, deadlines, cancellation and terminal evidence. Its [integration installer record](2026-10-22-windows-dev12-integration.md) retains actual shortcut lifecycle. Those observations remain useful at their exact source and behavior boundaries; none is relabelled as a dev.13 run. No repeated archive inspection or new candidate is implied by adding this documentation after composition.

Katja retains the joined dev.13 qualification in Issue #36. Physical drop/Send To, rendered notifications/focus, smart clipboard consumption and file-picker paste require a newly authorized desktop window. Live TinyPNG retains an explicit credential/quota grant. Primary-menu activation retains trusted signing/registration and native observation. Foreign archive runtime, broader physical hosts, production signing/feed, public distribution and Wolf's final public-experience workshop remain separate accepted consequences. The full product destination remains in Vision and the owning workplace; this record does not close it.
