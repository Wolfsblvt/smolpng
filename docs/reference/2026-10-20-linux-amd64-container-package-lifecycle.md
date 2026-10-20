# Hosted Linux amd64 RPM, APK and Arch package lifecycle qualification

## Meaning

The maintained Linux amd64 RPM, APK and Arch packages were installed, exercised with real optimizers and removed through their native package managers in digest-recorded Fedora 44, Alpine 3.24.2 and Arch Linux containers on hosted Ubuntu amd64. Each package-owned CLI passed all 63 PNG/JPEG checks, reported system-package update ownership, and left retained per-user dependency files unchanged after package removal. This record binds those results to one exact source, package, image and report set. The containers share the hosted runner kernel; this is not a full distribution installation or desktop qualification. Arch's pacman `NoExtract` policy also intentionally prevents extraction of the package's `/usr/share/doc` file while retaining its package ownership.

## Exact source and hosted runs

- Source: [`0c080d256246916acf1546b5dbadcd2105a470de`](https://github.com/Wolfsblvt/smolpng/commit/0c080d256246916acf1546b5dbadcd2105a470de), clean hosted checkout. The workflow's expected source and package-embedded commit matched.
- [Linux package qualification run](https://github.com/Wolfsblvt/smolpng/actions/runs/39354075858), completed at `2026-10-20T01:34:48Z`; Fedora RPM, Alpine APK, Arch Linux Arch, Ubuntu amd64 DEB and native Ubuntu arm64 DEB jobs all passed.
- [Exact-source Verify run](https://github.com/Wolfsblvt/smolpng/actions/runs/39354075762): all 11 jobs passed at the same source. This is the existing repository Verify workflow, not a second package lifecycle.
- Native host: GitHub-hosted `ubuntu-24.04`, Linux `6.17.0-1022-azure`, x86_64. The three named distribution environments were Docker containers pinned in the run evidence by image digest. Each container reports that same Linux kernel; none boots a separate distribution kernel.
- Application and fixture build: Go `1.26.6`, Linux amd64, `CGO_ENABLED=0`. The version `0.0.0` and reserved `.invalid` asset URL are fixture-only inputs, not a public release identity or endpoint.
- nFPM: module `github.com/goreleaser/nfpm/v2` v`2.48.0`, module sum `h1:Yvd07lX8YqDD+rtTwVqyeQiDX+7cnun/jFbZD501uB8=`, built with Go `1.26.6`; executable SHA-256 `a467d5f3ac01155c45082cb61823d6210e28535c6b43434c0680cd26a9acb126`. Its source-build banner says `dev`/`unknown`; the module build information establishes its pinned source.

## Package and host results

| Distribution / manager | Container image and release | Package identity | Package SHA-256 | Installed CLI SHA-256 | Optimizer evidence | Retained state after removal |
| --- | --- | --- | --- | --- | --- | --- |
| Fedora 44 / DNF5 5.4.3.0 | `fedora@sha256:43b29f65a41eb9c35e1cd5323e3bdf3b655c2357a9f4f1ff2f9c2798e5045d80` | `smolpng 0.0.0 x86_64` RPM | `09d7897472e607f169099e02bc73945941c00c3a1d758f0021185b752ea42ecb` | `8a6b4ec0f38ce6195032578ae9801cea12112e5f372f5c338af28aa40dfa30fb` | pngquant 2.18.0 `3e95089aae07401bd4076a1c3d8a56fedd787c1ee10e26849ae1a28bfda9441b`; OxiPNG 10.2.1 `df8f3e79e6ef06d8515d4949d19d872ae4e82fe16bc1fe6ad1efe181e3a21780`; jpegoptim 1.5.6 `279800dabce2e03916560cefbefb307cd05c8f75430f24e79729ede1d8a5da8a` | 8 files, all hashes preserved |
| Alpine 3.24.2 / apk-tools 3.0.8-r0 | `alpine@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6` | `smolpng 0.0.0 x86_64` APK | `c0b7d255535baeac3e09646d545552584cdcaa84a9e1a0a373c89ae749793307` | `8a6b4ec0f38ce6195032578ae9801cea12112e5f372f5c338af28aa40dfa30fb` | pngquant 3.0.3 `a0e2c9d3de5445043b2a3aea30d948ed3143924bfb32e282b91a327d1f189055`; OxiPNG 10.2.1 `df8f3e79e6ef06d8515d4949d19d872ae4e82fe16bc1fe6ad1efe181e3a21780`; jpegoptim 1.5.6 built for musl, `9463a782690bbc8e1cde54e7d34cb923e4bb290df39f9d84fb4d9291fe8510ed` | 4 files, all hashes preserved |
| Arch Linux / pacman 7.1.0, libalpm 16.0.1 | `archlinux@sha256:e43f1fd80e073992c5f173efbbbf65144170bf60aeb7369940087d63eacea8f4`; rolling image `20261018.0.612668` | `smolpng 0.0.0-1 x86_64` Arch package | `5b8cadd4094545575d065ff8765fdd758e20a6bc66d722308f13f3792311e1f3` | `8a6b4ec0f38ce6195032578ae9801cea12112e5f372f5c338af28aa40dfa30fb` | pngquant 3.0.3 `71440ba36ac0e128a68bd82d7efdf13bf0d8f0154eaac0c11434b3b7387f6e23`; OxiPNG 10.2.1 `df8f3e79e6ef06d8515d4949d19d872ae4e82fe16bc1fe6ad1efe181e3a21780`; jpegoptim 1.5.6 `279800dabce2e03916560cefbefb307cd05c8f75430f24e79729ede1d8a5da8a` | 8 files, all hashes preserved |

All three package-owned CLIs reported `system-package`, `canApply=false`, no configured update feed or signing keys, and no network check. Each complete real-image report contains 63 passing checks and zero failures:

| Distribution | Hosted report artifact | Artifact ID | Report SHA-256 |
| --- | --- | ---: | --- |
| Fedora RPM | `linux-rpm-amd64-0c080d256246916acf1546b5dbadcd2105a470de-39354075858` | `12234248520` | `fb1cb65f46d21d5699084faf2ca5ca7579cb2dd5e3caafe7234cf752ac4783b0` |
| Alpine APK | `linux-apk-amd64-0c080d256246916acf1546b5dbadcd2105a470de-39354075858` | `12234980175` | `2b923e549855f34a8cc097303ad2aa4b1ed97f84b307750ae782e8db27ff895e` |
| Arch Linux | `linux-archlinux-amd64-0c080d256246916acf1546b5dbadcd2105a470de-39354075858` | `12234108637` | `3121925755d2cedb2838f4c7659fac281b29160786da77246e50a4f997b7b304` |

GitHub-hosted artifacts retain reports only for 14 days. The fixture packages, application executables, optimizer binaries and generated images were not uploaded or published. The exact package SHA-256 values above were read from both the packaging and container copies before installation; installed executable hashes were independently read after installation. The complete reports and run outputs are also available from the linked run while GitHub retains them.

## Fedora optimizer qualification route

Fedora's RPM database reported a `pngquant` package version `2.18.0-12.fc44`, but the actual executable reported `4.0.3` and the application correctly refused its incompatible command-line family. The workflow now builds upstream pngquant 2.18.0 from its release archive in the disposable Fedora container. The archive SHA-256 is `e72194b52b36f040deaec49a1ddd5dcd8d4feecc3a5fe6c5e9589a9707b233d4`, as listed by the [pngquant release archive](https://pngquant.org/releases.html). GCC `16.2.1`, GNU Make `4.4.1`, the configure/build/install transcript, source checksum, binary version and executable SHA-256 are retained. This preserves the supported optimizer input without changing smolpng's package or runtime dependencies.

Alpine uses its APK-provided pngquant 3.0.3. Fedora and Arch use the application-managed OxiPNG 10.2.1 and jpegoptim 1.5.6 binaries. Alpine builds the pinned jpegoptim 1.5.6 source for musl because the maintained managed archive targets glibc; its source archive SHA-256 is `661a808dfffa933d78c6beb47a2937d572b9f03e94cbaaab3d4c0d72f410e9be`. These optimizer binaries are separate test tools and are not included in smolpng packages.

## Arch notices and pacman policy

The Arch package archive lists `usr/share/doc/smolpng/third-party-notices.md`, and extracting that member before installation produced bytes identical to the fixture staging file. The Arch container's existing `pacman.conf` contains `NoExtract = usr/share/doc/*`. Under that policy, pacman lists the notices path as owned in `pacman -Ql smolpng` but does not extract it to disk; pacman reports `69 total files, 0 missing files`. The package preserves the notices file in its archive. The lifecycle therefore records the system policy and checks the package archive rather than overriding `NoExtract` or calling the omitted on-disk copy a package omission. The primary license and dependency-license tree remain installed and byte-identical.

## Repaired hosted failures

Earlier runs retained the following distinct observations and corrections:

- A minimal Fedora image lacked `cmp`; adding its native `diffutils` package restored payload comparison. Container-created report directories were root-owned, so the host runner now reclaims only the artifact path before upload.
- Fedora's package version metadata did not match its executable CLI behavior. The hosted `4.0.3` executable was unsupported despite RPM metadata saying 2.18.0; the checksum-verified upstream 2.18.0 source build above supplied the compatible real optimizer.
- `pacman -Qkk` flagged package files whose only observed difference was modification time. The lifecycle now uses pacman's missing-file check and compares the executable and legal payload bytes directly.
- The Arch archive contained the notices file but the container's configured `NoExtract` policy omitted it from disk. The final run retains the configuration and package ownership evidence and reports this boundary explicitly.

The unsuccessful package runs include [39352707430](https://github.com/Wolfsblvt/smolpng/actions/runs/39352707430), [39353501889](https://github.com/Wolfsblvt/smolpng/actions/runs/39353501889) and [39353778914](https://github.com/Wolfsblvt/smolpng/actions/runs/39353778914). Their partial observations are not counted as completed lifecycle results. The final exact-source package run is [39354075858](https://github.com/Wolfsblvt/smolpng/actions/runs/39354075858).

## Boundaries and continuation

This establishes Linux amd64 RPM/APK/Arch package-manager lifecycles in these hosted images, plus the separate Ubuntu amd64 and arm64 DEB lifecycles recorded in [the Ubuntu amd64 report](2026-10-20-linux-deb-lifecycle.md) and [the Ubuntu arm64 report](2026-10-20-linux-arm64-deb-lifecycle.md). It does not establish a full Fedora, Alpine or Arch host, a distribution-specific kernel, other package formats or architectures, desktop/graphical integrations, the private dev.11 Linux ARM64 archive runtime, signing, registry submission, package publication or a public release. The selected complete PNG/JPEG application and Windows graphical journey remain open in [Issue #36](https://github.com/Wolfsblvt/smolpng/issues/36).
