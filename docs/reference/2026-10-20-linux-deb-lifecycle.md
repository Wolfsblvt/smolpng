# Ubuntu DEB installation and removal qualification

## Meaning

The maintained Linux package definition has now produced a DEB that was installed, exercised and removed on an actual Ubuntu amd64 runner. All 63 PNG/JPEG checks passed through the package-owned `/usr/bin/smolpng`, the application recognized system-package update ownership, and removal preserved eight per-user dependency files. This record binds the first observation to its exact fixture source, package, host and reports. It is a native DEB component result, not acceptance of the private Windows candidate, other Linux formats or a public release.

## Exact first observation

- Source and matching qualifier: `336df9e3278711e9c027bf9be9f5bd4b8db858e3`.
- [Native lifecycle run](https://github.com/Wolfsblvt/smolpng/actions/runs/39339372764): successful Ubuntu amd64 DEB lifecycle job.
- [Ordinary Verify](https://github.com/Wolfsblvt/smolpng/actions/runs/39339372554): all 11 jobs passed at that source, including native test hosts, Linux race tests, vet, corpus checks and six build targets.
- Host: Ubuntu **24.04.5 LTS**, Linux **6.17.0-1022-azure**, x86_64.
- Application and tool compiler: Go **1.26.6**, native Linux amd64, cgo disabled.
- Fixture package/version: `smolpng`, `0.0.0`, `amd64`. The stable version and reserved `.invalid` asset URL exist only to exercise the existing metadata generator; neither is a selected public release identity or download address.
- Native journey began `2026-10-19T23:06:50Z`, which is 20 October in Berlin.

The [workflow at the observed source](https://github.com/Wolfsblvt/smolpng/blob/336df9e3278711e9c027bf9be9f5bd4b8db858e3/.github/workflows/linux-packages.yml) builds only the Linux amd64 fixture through the existing release builder and nFPM definition. It installs nFPM into the disposable runner's temporary tool directory. The nFPM source module is pinned to **v2.48.0**, module sum `h1:Yvd07lX8YqDD+rtTwVqyeQiDX+7cnun/jFbZD501uB8=`. Its source-build version output says `dev`/`unknown`; the Go module build information, rather than that friendly version label, establishes the selected tool source. This follows the [upstream Go installation route](https://nfpm.goreleaser.com/docs/install/).

## Observed package journey

1. Generate a fresh Linux archive and offline package definition; check its complete index and source SHA.
2. Produce the DEB with nFPM, inspect its package fields and payload listing, compare its extracted CLI with staging, and verify executable mode `0755`.
3. Install with dpkg. Provider readback names `smolpng` as the installed package and owner of `/usr/bin/smolpng`; installed CLI bytes match staging and report the expected source/version.
4. Read `update status`: distribution metadata remains `standalone`, but the actual system path takes precedence. The channel is `system-package`, `canApply` is false, and no feed or trusted signing keys are configured. No network check is performed by that status command.
5. Use separately installed pngquant and isolated managed OxiPNG/jpegoptim with the installed CLI. The complete report contains **63 passing checks**, including actual PNG/JPEG compression, lossless pixels, targets and misses, binary stdout, mixed/partial batches, original preservation and final optimizer identities.
6. Remove the package with dpkg. The workflow checks that `/usr/bin/smolpng` is absent and package status is no longer installed. All **eight** pre-removal per-user dependency files pass their retained hash checks afterward.

This first observation establishes the listed payload checks. The current workflow additionally compares installed primary/dependency licenses and third-party notices with staging; later exact-head runs carry that added check separately rather than attributing it retroactively to this run.

The real optimizer identities were:

| Tool | Version | Executable SHA-256 |
| --- | --- | --- |
| pngquant | 2.18.0 | `cc9b1d68e0cd5d371e58ce3ba891ddca179a02b6c4c02dee7dbf4e99a0f473c5` |
| OxiPNG | 10.2.1 | `df8f3e79e6ef06d8515d4949d19d872ae4e82fe16bc1fe6ad1efe181e3a21780` |
| jpegoptim | 1.5.6 | `279800dabce2e03916560cefbefb307cd05c8f75430f24e79729ede1d8a5da8a` |

## Artifact identities

| Artifact | SHA-256 |
| --- | --- |
| Fixture DEB | `41d840083ce8c125f0f3ae3b0adbcba1f3e96223fd0ee2796d4fbac6584fc117` |
| Installed CLI | `7be0ca422024272df53f338ebfb749ac3b2a468cb2d30f9be74b8f6dfdbd9959` |
| Source-built nFPM | `a467d5f3ac01155c45082cb61823d6210e28535c6b43434c0680cd26a9acb126` |
| Complete native report | `8d1f31cb186b0482e309f1fc9ae3221e1c4917e4a1c0babafa2400ad3b1e5a39` |

The downloaded reports are retained locally under ignored `artifacts/validation/linux-deb-pr48-336df9e/`. Hosted artifacts contain reports only and have 14-day retention. Fixture DEBs, executable payloads and image outputs are not uploaded by this workflow. The ordinary Verify development-binary artifacts remain a separate existing surface.

## Boundaries and continuation

This proves an installed DEB CLI and removal journey on this Ubuntu amd64 host. RPM, APK, Arch, Linux ARM64 installation/runtime, graphical Linux integrations and other distributions remain separately unobserved. The related private dev.11 application remains bound to its own `61adb45` source, Go 1.27.2 compiler and archive hashes; this fixture does not become that candidate by association.

The job is on demand or scoped to changes in its own workflow. It uses read-only repository permissions, existing pinned Actions and no publishing credential. No registry submission, public release, signature, production update feed, service deployment or owner-machine package installation occurred. Katja retains the remaining distribution and joined-candidate work in [Issue #36](https://github.com/Wolfsblvt/smolpng/issues/36).
