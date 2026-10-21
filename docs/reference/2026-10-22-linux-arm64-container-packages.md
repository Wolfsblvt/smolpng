# Native ARM64 Fedora RPM and Alpine APK lifecycle

## Meaning

The maintained ARM64 RPM and APK now have installation, real PNG/JPEG use, package ownership and removal evidence in Fedora 44 and Alpine 3.24.3 containers on native hosted ARM64 Linux. Both installed CLIs passed all 79 checks and package removal preserved four per-user OxiPNG files. This record joins exact source, container/package/executable identities and retained reports; the containers share Ubuntu's hosted kernel. It does not qualify a distribution desktop, a release or the joined private candidate.

## Source and execution

- Source and embedded application/qualifier commit: `32341d917a418b93d11a4c3335be871920c1796b`.
- [Linux package run 39694952966](https://github.com/Wolfsblvt/smolpng/actions/runs/39694952966): all seven jobs passed, including both new ARM64 routes, both Ubuntu DEB routes and the three existing amd64 container routes.
- [Verify 39694952921](https://github.com/Wolfsblvt/smolpng/actions/runs/39694952921): all eleven jobs passed at that same source. Forgehand passed; the existing size check was skipped and is not counted as green.
- Native runner selection: `ubuntu-24.04-arm`; both containers reported `Linux 6.17.0-1022-azure aarch64`. Go `1.26.6`, `linux/arm64`, `CGO_ENABLED=0`; no QEMU route was configured. Package architecture and installed CLI platform were independently checked.
- nFPM module `github.com/goreleaser/nfpm/v2` v`2.48.0`; executable SHA-256 `2f807e9f9dced5e049c255a88dfba0fa74ff8e8a6557f27b9f86b6eb11de38f5`. Its module build information is retained rather than interpreting the source-build `dev` banner as a release version.
- `0.0.0` and the reserved `.invalid` asset address are qualification fixtures. They choose neither the public version nor a published endpoint.

## Installed package outcomes

| Native container | Package / manager | Package SHA-256 | Checks |
| --- | --- | --- | --- |
| Fedora 44 | `smolpng 0.0.0 aarch64`, RPM / DNF5 5.4.3.0 | `e7065f5470663e23243922f0e4b3e0fb7e4a5f9cc41b596521f5cb0641d35ad2` | 79 passed |
| Alpine 3.24.3 | `smolpng 0.0.0 aarch64`, APK / apk-tools 3.0.8-r0 | `0e9b47f4415cfaccf997e1fc111064def76b3b65a528ca89de4d7856051ccec5` | 79 passed |

Both package-owned `/usr/bin/smolpng` files matched their exact staging bytes, SHA-256 `e5ecbe8378a1715b1765e81d93d013db0c78daadf791786ccdef500688c76bad`. Primary/dependency licenses and notices matched their staged payloads. `version` reports the builder's `standalone` distribution marker; the separately inspected update channel correctly detects `system-package`, `canApply=false`, with no feed, trusted key or network check. These are separate observations.

RPM package integrity passed. APK ownership identified smolpng. Both native managers removed smolpng, the executable disappeared, the manager no longer reported it installed, and all four pre-removal OxiPNG files retained their hashes. No optimizer is bundled in either package.

The Fedora image was `fedora@sha256:43b29f65a41eb9c35e1cd5323e3bdf3b655c2357a9f4f1ff2f9c2798e5045d80`; Alpine was `alpine@sha256:c6b073f8e9e58d881f6ffba6ebf3b9ebd60b9e7d6e8337018c0e68bb5c356873`. These are recorded image digests, not an assertion that mutable `latest` will remain identical.

## Real optimizers and retained evidence

Fedora used checksum-verified upstream pngquant 2.18.0 built in the container; Alpine used native pngquant 3.0.3. Both used managed OxiPNG 10.2.1. jpegoptim 1.5.6 was built from the same pinned upstream source because its managed catalog has no ARM64 archive and its amd64 archive targets glibc. This reuses the established native source-build route without adding an application runtime or packaging dependency.

The [retained payload](2026-10-22-linux-arm64-container-packages/report.json) contains the complete two 79-check reports, source/build/module identities, package and installed hashes, image/OS/kernel facts, optimizer identities, install/removal transcripts and retained-state receipts. Each captured original UTF-8 file has its own SHA-256. Payload SHA-256: `2960b80355cfdc6c0c4aaa9acba6f96986f4b9717a6435d189b08aac9c3f7c2f`.

Original report SHA-256 values:

- Fedora: `21ad1cb0610ac47efcc1421a9e3e3519292d2f1afeeef1d11a9e105ad0545af5`, artifact `12390358102`.
- Alpine: `035c10a943a2c41cab6e7a6c73ab913a7c7141463791c74b8ddeb9bad3257a73`, artifact `12390168472`.

Hosted artifacts retain reports for 14 days. The durable payload preserves these selected receipts beyond expiry; original fixture packages, executables and images were not uploaded. The original evidence source remains the SHA above even when a later documentation-only commit carries this record.

## Remaining boundaries

This cut extends and qualifies existing distribution tooling. It supplies no source application change, package publication, signing, new package repository, physical desktop behavior or Arch Linux ARM claim. The [earlier amd64 record](2026-10-20-linux-amd64-container-package-lifecycle.md) retains its own source, tools and failures. The complete Windows graphical journey and next current-source private candidate remain in [Issue #36](https://github.com/Wolfsblvt/smolpng/issues/36); the complete product and public-experience sequence remain in Direction.
