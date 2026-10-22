# Local packaging and distribution inputs

## Meaning

The repository contains executable local packaging tooling, not just proposed release commands. It builds exact-source standalone candidates, reproducible ZIP/tar.gz archives, member receipts and checksums. With an explicit future asset URL and stable version, it also generates offline Winget, Scoop, Homebrew, Chocolatey and nFPM inputs plus an unsigned update payload. The builder does not publish, submit, register a package, create a signing authority or contact a package manager. Hosted qualification installs and removes fixture packages through dpkg on native Ubuntu amd64/arm64 runners, RPM/APK package managers inside digest-recorded Fedora/Alpine containers on native hosted Ubuntu amd64/arm64 runners, and the Arch package manager inside an Arch container on hosted Ubuntu amd64. The container cases share the host kernel; they do not claim full distribution or desktop qualification.

## Candidate builds

From a clean committed source tree with Go available:

```sh
go run ./tools/release --version 0.3.0-dev.1 --out artifacts/candidate
```

The output directory must be new. All six platform targets are built by default. `--platforms linux/amd64` selects a subset. Build arguments explicitly embed version, source SHA and standalone ownership; archives use source-commit time, sorted paths and normalized modes. Files include the main CLI, Windows windowless launcher, native qualification helper, docs, legal corpus, presets, instruction-only Skill sources, license map and full dependency notices. Optimizers, toolchain caches, private profiles and credentials are never collected.

Candidate compilation isolates the Go command from ambient workspaces, per-user Go configuration, build flags, experiments, toolchain switching and target microarchitecture overrides. It uses `GOWORK=off`, `GOENV=off`, `GO111MODULE=on`, `GOTOOLCHAIN=local`, empty `GOFLAGS`/`GOEXPERIMENT`, `CGO_ENABLED=0`, and the portable `GOAMD64=v1` or `GOARM64=v8.0` baseline. This prevents an external `go.work`, local replacement, user build tag or host-specific tuning from changing candidate bytes while the archive still names only the repository commit. The selected `go` executable and dependency transport remain operator inputs and belong in the candidate evidence record.

The builder enumerates every ordinary file from the selected commit, reads its exact blob bytes, and materializes a temporary isolated source root. All target binaries are built there. The same committed map supplies the admitted docs, Skills, presets, schemas, corpus and notices. Symlinks, gitlinks, malformed objects/paths, non-portable names and case collisions refuse before any write. Neither the mutable index nor working-tree paths supply release bytes after the clean-source gate.

The archive receipt binds individual member hashes. `SHA256SUMS` binds every archive. `build-index.json` is written incrementally and marks complete only after the entire selected build succeeds. Failure leaves inspectable local output and returns nonzero, not a nominal release.

## Windows executable identity

The local builder adds VERSIONINFO to `smolpng.exe`, `smolpngw.exe` and `smolpng-qualify.exe` on both Windows architectures, before member/archive hashes are calculated. ProductName is `smolpng`; CompanyName is the descriptive maker label `Wolfsblvt Works`. Each file has its own description, original filename and internal name. FileVersion and ProductVersion strings retain the complete explicit semantic version, including its prerelease suffix; Comments retains the exact source commit and standalone distribution.

The platform's fixed numeric version is `MAJOR.MINOR.PATCH.0`. Each component must fit Windows' 16-bit field; an unrepresentable version fails instead of wrapping. Prerelease builds set the prerelease flag. The full strings distinguish prereleases sharing one numeric tuple; Windows file-version numbers are not smolpng's update-selection contract.

Resource processing uses pinned Go build tooling, not an additional native compiler or an application runtime dependency. Existing non-version resources are preserved. No generated `.syso` file is written into the source tree, and the builder does not alter manifest/DPI/privilege settings or add an icon. Ordinary direct `go build` remains a development build without these release resources.

VERSIONINFO is descriptive metadata, not a signature or an authenticated publisher. Source compilation does not prove Windows can read the actual files; inspect the built PE resources, machine type and subsystem on x64 and ARM64 as part of candidate qualification. For a headless native string read on Windows:

```powershell
[Diagnostics.FileVersionInfo]::GetVersionInfo('C:\ABSOLUTE\BUILD\stage\windows-amd64\smolpng.exe') |
    Select-Object ProductName, CompanyName, FileVersion, ProductVersion, OriginalFilename, Comments, IsPreRelease
```

An unsigned Preview remains unsigned after this metadata is embedded. The separate signing sequence and retained release decisions are in [Releasing](RELEASING.md#windows-signing-sequence).

[Native qualification](reference/2026-10-21-windows-version-resources.md) records six exact-source files, matching stage/archive receipts, hosted amd64 CLI execution and Windows 11 resource reads. ARM64 execution and graphical use remain distinct.

## Package-manager inputs

For a separately authorized stable distribution, use the same builder with a real selected asset base URL:

```sh
go run ./tools/release --version 1.0.0 --base-url https://HOST/CHOSEN-RELEASE-PATH --out artifacts/release-inputs
```

The URL above is a placeholder, not a live endpoint. The supplied URL is recorded, not uploaded to. Generated manager IDs are proposed `Wolfsblvt.smolpng`/`smolpng`, not proof those names are registered or accepted.

Winget receives a multi-file portable ZIP manifest for x64/ARM64. Scoop receives per-architecture URLs/hashes and executable aliases. Homebrew receives macOS/Linux architecture-specific artifacts. Chocolatey receives a checksum-verified x64 ZIP installation recipe. nFPM receives amd64/arm64 definitions for `/usr/bin` and license/doc placement, consumable by its deb/rpm/apk/Arch builders after a separate tool qualification. No installer silently downloads pngquant, jpegoptim or OxiPNG; users explicitly run setup or use package-manager choices.

These files need the destination ecosystem's actual schema validation/install tests and submission review. The [native Winget validation record](reference/2026-10-17-winget-manifest-validation.md) now qualifies the generated three-file portable manifest shape with Windows Package Manager 1.29.380, using explicit synthetic URLs/hashes. It repaired an unsupported `Scope` field; real assets, installation and submission are not established by that metadata check. Other ecosystem schema/install checks remain separate. MSIX/Store and signed native installers remain separate packaging decisions requiring native and publication qualification.

## Linux package files

Provide a separately qualified nFPM binary, then run it from the generated output root where `stage/` and `nfpm/` live. The following examples use fresh output filenames:

```sh
nfpm package --config nfpm/amd64.yaml --packager deb --target smolpng-amd64.deb
nfpm package --config nfpm/amd64.yaml --packager rpm --target smolpng-amd64.rpm
nfpm package --config nfpm/amd64.yaml --packager apk --target smolpng-amd64.apk
nfpm package --config nfpm/amd64.yaml --packager archlinux --target smolpng-amd64.pkg.tar.zst
```

Use `nfpm/arm64.yaml` and distinct arm64 output names for the other architecture. These commands produce local files; they do not install them, submit them to a registry or sign them. nFPM is a build tool, not an application dependency.

The definitions set `umask: 0o022` to remove group/world write permissions inherited from staging, including Windows staging. The CLI retains its explicit `0755` mode. Regular license and notice files use `0644`, and the dependency-license directory uses `0755` with the maintained builder's staging. The package contains the CLI, primary license, complete dependency-license tree and third-party notices; it does not bundle optimizers, the qualification helper, user configuration or images.

Inspect actual package headers, executable architecture, paths, modes and extracted file hashes before treating package production as qualified. Format/payload inspection does not establish installation, package-manager ownership or runtime behavior on a Linux host.

The [Windows-built Linux package record](reference/2026-10-17-linux-package-artifacts.md) qualifies all eight amd64/arm64 format combinations with nFPM 2.48.0 and independent payload/header readers. It records the staging-permission repair, exact source and artifact identities, all 61 payload files per package, and the remaining host/publication limits.

The [Ubuntu amd64 DEB lifecycle record](reference/2026-10-20-linux-deb-lifecycle.md) and [native Ubuntu ARM64 DEB lifecycle record](reference/2026-10-20-linux-arm64-deb-lifecycle.md) establish actual installation, package/update ownership, all 63 PNG/JPEG checks through the installed CLI, and removal with per-user dependency state preserved on their respective native runners. The [hosted Fedora RPM, Alpine APK and Arch amd64 lifecycle record](reference/2026-10-20-linux-amd64-container-package-lifecycle.md) establishes the same installed-CLI exercise and removal boundary through those package managers inside digest-recorded distribution containers on Ubuntu amd64. The [native ARM64 Fedora RPM and Alpine APK lifecycle record](reference/2026-10-22-linux-arm64-container-packages.md) adds both installed ARM64 routes, all 79 PNG/JPEG checks per route, package/update ownership, and removal preserving four per-user OxiPNG files. The same [dedicated workflow](../.github/workflows/linux-packages.yml) retains reports only. Every fixture uses version `0.0.0`, which is not a public release identity. These hosted lifecycles do not qualify Arch Linux ARM, full distribution installations, graphical integrations, package submission or signing.

## Standalone update archive and signing

The update archive is deliberately different from the full development package. It contains only allowlisted application executables and license obligations. User config, presets, images, Skills and managed tools cannot appear as replacement targets. The builder verifies its generated member list against the same runtime update validation contract.

`release.UNSIGNED.json` is **not** accepted by the application updater. An operator who already holds the independently qualified production signing key can sign that payload:

```sh
go run ./tools/sign-update --payload artifacts/release-inputs/release.UNSIGNED.json --key-file /PROTECTED/PATH/release-key.pem --key-id CHOSEN-KEY-ID --out artifacts/release-inputs/release.signed.json
```

The command reads an existing protected PKCS#8 Ed25519 key, signs exact payload bytes, self-validates the result and writes a new envelope. It neither generates the key nor publishes it; no secret value belongs in argv, the repository or archives. The production public key/feed must be bound through a separate trustworthy channel. Unit tests use ephemeral test keys only. Authenticode/notarization are distinct from this manifest signature.

## Evidence boundary

Pure tests cover archive determinism, member modes, runtime-compatible update extraction, no user-data members and package-manager hash propagation. Source compilation and local archive verification do not substitute for live installers, signing infrastructure or desktop/optimizer qualification. No hosted CI result or public release is inferred merely from running this builder.
