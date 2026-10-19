# Native Linux PNG and JPEG qualification

## Meaning

Source `5cb1132e5ab9d0706f6e1c02694cd8ffe5570192` passed all 63 real PNG/JPEG qualifier checks on an Ubuntu 24.04 amd64 GitHub-hosted machine, using separately obtained pngquant, OxiPNG and jpegoptim. This record binds the first Linux optimizer/runtime observation to its source, host, executable and dependency identities. It qualifies the exercised CLI and managed-tool installation paths, not a Linux package installation, graphical desktop or the immutable dev.9 archives.

## Source and actual host

The [native run](https://github.com/Wolfsblvt/smolpng/actions/runs/39318158234) built both application and qualifier natively from the same commit. The separate [Verify run](https://github.com/Wolfsblvt/smolpng/actions/runs/39318158203) passed all eleven ordinary test/build/shell jobs at that exact head. The application reports the existing source-build identity `0.1.0-dev`, distribution `source`; this is not public release naming.

The host reported Ubuntu **24.04.5 LTS**, Noble, Linux kernel `6.17.0-1022-azure`, x86_64. Runner image was `ubuntu24` / `20261012.334.1`. Go was **1.26.6**, native `linux/amd64`, with `CGO_ENABLED=0`. The native application/qualifier were not race-instrumented; Verify owns its separate Linux race result.

The application source under `cmd`, `internal`, `skills`, `tools`, `testdata`, module files and license inputs has no difference from dev.9's `b0ed6391c4d8c80fcb9d78c611aa42d8debb798f`. This relationship does not make these Linux source-build bytes the retained cross-built dev.9 archive: build commit, version and toolchain identities remain distinct.

## Tool installation and identities

The workflow extracted authenticated Ubuntu package downloads into runner temporary storage without a system installation or PATH modification. Packages were pngquant `2.18.0-1build2`, libimagequant0 `2.18.0-1build1`, libpng16-16t64 `1.6.43-5ubuntu0.7` and libgomp1 `14.2.0-4ubuntu2~24.04.1`. Their archive hashes and pngquant copyright were retained.

The built application previewed and installed its existing pinned OxiPNG 10.2.1 musl archive and jpegoptim 1.5.6 Linux archive in a fresh disposable app home. Both reached `current` after archive/license/version checks; their actual executables then supplied the native qualification. The catalog and its download policy were not changed. This is installation/use evidence, not the managed removal or upgrade lifecycle.

| Executable | Version or role | SHA-256 |
| --- | --- | --- |
| smolpng | Source application | `0d78eb746b656d216eb9df1147bf965b076bf9d4abf8d918e809b70b139774d3` |
| smolpng-qualify | Matching helper | `ce5e927120235d7e20844f303a4dc8cee6a9debf1a8e4f0e285de2e2872b7a75` |
| pngquant | 2.18.0 | `cc9b1d68e0cd5d371e58ce3ba891ddca179a02b6c4c02dee7dbf4e99a0f473c5` |
| OxiPNG | 10.2.1 | `df8f3e79e6ef06d8515d4949d19d872ae4e82fe16bc1fe6ad1efe181e3a21780` |
| jpegoptim | 1.5.6 | `279800dabce2e03916560cefbefb307cd05c8f75430f24e79729ede1d8a5da8a` |

## Exercised behavior and retained evidence

All 63 checks passed, including exact build provenance, optimizer identities before and after the corpus, PNG/JPEG Auto and strict lossless, strict target satisfaction/miss/nearest, clean binary stdout, partial batches and a mixed-format batch. The helper independently checks actual output files, image structure, decoded lossless pixels and original preservation. Top-level `complete` and `passed` are both true; no individual check failed. Explicit fixture inputs and isolated configuration avoid graphical clipboard, real user settings and remote image providers.

The report artifact `native-linux-5cb1132e5ab9d0706f6e1c02694cd8ffe5570192-39318158234` contains 15 report/receipt files, with no application/optimizer executable or image. It retains OS/kernel/image details, Go and jq versions, application/helper hashes, package hashes/versions, optimizer previews/installation receipts and the full qualifier report. `journey/report.json` SHA-256 is `f10a34127ae2653dc89d3ba55556199e4015c5904393cda450884caac554a835`. The downloaded copy is under ignored `artifacts/validation/native-linux-pr44-5cb1132/`. Hosted retention is 14 days; this record preserves the durable qualification account.

The maintained workflow is on demand and also exercises PRs changing its own file; ordinary source PRs keep the existing Verify surface. Source authoring and the successful PR event do not prove the separate default-branch manual-dispatch route. That route's result must be consumed at its own exact commit.

## Remaining boundaries

This host does not establish DEB/RPM/APK/Arch package installation or package-manager ownership, X11/Wayland clipboard, desktop launcher/menu behavior, notifications, interactive terminal rendering, updater replacement, keyring access or arm64/macOS runtime. Those remain separate from the [earlier Linux package-file evidence](2026-10-17-linux-package-artifacts.md) and [Windows dev.9 record](2026-10-19-private-candidate-dev9.md). The complete destination and joined candidate remain open in [Issue #36](https://github.com/Wolfsblvt/smolpng/issues/36). No release, package registration, production signing/feed, hosted sibling, deployment or spend occurred.
