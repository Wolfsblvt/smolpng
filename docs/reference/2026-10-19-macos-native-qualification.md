# Native macOS PNG and JPEG qualification

## Meaning

Source `e3c68a37a150d8809a62d504cb4d7bb2cc324e33` passed all 63 real PNG/JPEG qualifier checks on both Intel and ARM64 macOS hosts. This record preserves the first Intel failure, its catalog architecture repair and exact native source/tool identities. It qualifies exercised CLI and managed installation/use paths, not graphical desktop behavior, signing, package installation or the retained dev.9 archives.

## Failure and repair

The [first run](https://github.com/Wolfsblvt/smolpng/actions/runs/39322082451), at `82f6270670b3a53d691a90f70773226be3378a4b`, installed OxiPNG on Intel macOS but refused jpegoptim at the exact-version probe. It did not reach compression. Linux passed all 63 checks; the separate Verify run passed all 11 jobs.

The [upstream release](https://github.com/tjko/jpegoptim/releases/tag/v1.5.6) names its 205390-byte archive `jpegoptim-1.5.6-x64-osx.zip`, but the enclosed 541880-byte executable is thin ARM64 Mach-O: header `CFFAEDFE0C000001`, CPU type `0x0100000c`. The ZIP matches the published and catalog-pinned SHA-256 `5ca21c2f8f02ed815b00e453b0ea235d84ac6a99704d15bb587f952f700d6d44`. The executable hash is `f38a97a846c48e49bcfae16d997bf29ca73d5f4fa00b32efef8c2fcb98adc61f`.

The catalog now offers those unchanged authenticated bytes for macOS arm64 and refuses the incorrect Intel target. Windows/Linux amd64 selections are unchanged. Intel macOS continues through a separately installed native jpegoptim; the hosted runner uses Homebrew. Authentication, extraction checks and exact-version probing remain intact. The [owner-visible finding](https://github.com/Wolfsblvt/smolpng/issues/36#issuecomment-6277049476) and [Decisions](../DECISIONS.md#pinned-dependency-delivery-not-upstream-latest-execution) preserve the rationale and capability boundary.

The corrected catalog regression failed on the old source. Then `go test ./internal/toolchain -count=1` and `go vet ./internal/toolchain` passed locally with Go 1.27.2, BelowNormal priority, GOMAXPROCS=1 and a 512 MiB soft memory target. No desktop or clipboard interaction was used.

## Actual repaired hosts

The [native run](https://github.com/Wolfsblvt/smolpng/actions/runs/39323233182) built matching application and qualifier binaries from the repaired source. Each of its three jobs passed 63 checks. The separate [Verify run](https://github.com/Wolfsblvt/smolpng/actions/runs/39323233179) passed all 11 jobs. Applications reported `0.1.0-dev`, distribution `source`, the exact source commit and native architecture. All used Go 1.26.6 with `CGO_ENABLED=0`.

| Host | OS/kernel | Runner image | Optimizers |
| --- | --- | --- | --- |
| Intel Mac | macOS 26.6.1, build 25G76; Darwin 25.6.0 x86_64 | `macos26` / `20260824.0517.1` | Homebrew pngquant 3.0.3 and jpegoptim 1.5.6; managed OxiPNG 10.2.1 |
| ARM64 Mac | macOS 26.6.2, build 25G83; Darwin 25.6.0 arm64 | `macos26` / `20261013.0424.1` | Homebrew pngquant 3.0.3; managed OxiPNG 10.2.1 and jpegoptim 1.5.6 |
| Linux amd64 | Ubuntu 24.04.5 LTS; Linux 6.17.0-1022-azure x86_64 | `ubuntu24` / `20261012.334.1` | Extracted Ubuntu pngquant 2.18.0; managed OxiPNG 10.2.1 and jpegoptim 1.5.6 |

Managed tools reached `current` after archive/license/version checks. Actual ARM64 jpegoptim used by the successful corpus has the same executable hash as the inspected archive above. This proves installation/use, not managed removal or upgrade. Homebrew formula/config receipts retain its separate dependency choices; no optimizer is bundled with smolpng.

| macOS executable | Intel SHA-256 | ARM64 SHA-256 |
| --- | --- | --- |
| smolpng | `7c10a02ded7c4cbeeb7491a2f0434976932de9c50f0d9d974a2d422b55750ac3` | `425a4f5be50a0f588cf33cf13bf5be9f5690fde3b09372329d3000071434d9a1` |
| smolpng-qualify | `50a6d2481e2141009b5238df4ab68b11db45903bf9e2b39c750edd1c62d7b76b` | `364dc48c2572354b7a6d9d318e930466c543d2d72f9120fc4a881e6b5e774bbe` |
| pngquant | `2953010c4167a40b1bce8e7d5ad6c1890f3dd90455d93ea1ce5a999dd80cd220` | `f3f00f58339e28c6b6a28fefaf74d7685cd80970ee8bab748897195f763c2f1f` |
| OxiPNG | `b9f6bace13541441cd6dc9f3f80695f9dbf94db89c4f51209214d14936ae71c6` | `275a2049091caf576be74b838df4675fbb179c9f2d118cec44ea1521e54fc45e` |
| jpegoptim | `496b4515240d428b4a6e99166fecfd2bff32da580282793667dc3aeff2076990` | `f38a97a846c48e49bcfae16d997bf29ca73d5f4fa00b32efef8c2fcb98adc61f` |

Checks cover exact build/tool provenance, PNG/JPEG Auto and strict lossless, targets/misses/nearest, binary stdout, partial/mixed batches, original preservation, independently inspected output structure/pixels and final optimizer hashes. Every individual check passed; reports have `complete=true` and `passed=true`. Fixtures, isolated homes and local-only policy avoid personal configuration, uploads and graphical clipboard.

## Retained evidence and limits

The 14-day artifacts are named `native-{platform}-e3c68a37a150d8809a62d504cb4d7bb2cc324e33-39323233182`, with platforms `linux`, `macos-amd64` and `macos-arm64`. They contain reports and source/host/tool/installation receipts, excluding executables and images. Downloaded copies are under ignored `artifacts/validation/native-pr45-e3c68a3/{platform}/`. Full `journey/report.json` hashes are:

- Intel macOS: `297b7f319195bb5defc2fdb14820ac029904a4ae50d9d0e3df50ecc858a8d670`.
- ARM64 macOS: `48c627a1c3bf9dd1977b8f4f84e4c0fa8f8318345b9ab69e412e412d882449fa`.
- Linux: `1225d62cc6169af375dd3e11f0313274e1af3b4a456b64bc024a65403b9cc62d`.

The failed first attempt, verified upstream ZIP and static inspection remain separately under `artifacts/validation/native-pr45-82f6270/`. The renamed on-demand workflow retains ordinary Verify and self-qualifies when its own source changes; a PR run does not prove default-branch manual dispatch.

These source builds are distinct from dev.9's cross-built archives, version and toolchain. The catalog changes application source and requires a fresh later joined candidate. Neither Mac proves package installation, Finder/drop packaging, graphical clipboard/notifications, TUI rendering, keyring access, updater replacement or signing/notarization. Linux arm64 runtime also remains separate. The complete destination and joined candidate remain open in [Issue #36](https://github.com/Wolfsblvt/smolpng/issues/36). No public release, registry submission, production signing/feed, deployment or spend occurred.
