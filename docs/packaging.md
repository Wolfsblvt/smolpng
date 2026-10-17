# Local packaging and distribution inputs

## Meaning

The repository contains executable local packaging tooling, not just proposed release commands. It builds exact-source standalone candidates, reproducible ZIP/tar.gz archives, member receipts and checksums. With an explicit future asset URL and stable version, it also generates offline Winget, Scoop, Homebrew, Chocolatey and nFPM inputs plus an unsigned update payload. It does not publish, submit, register a package, create a signing authority or contact a package manager.

## Candidate builds

From a clean committed source tree with Go available:

```sh
go run ./tools/release --version 0.2.0-dev.1 --out artifacts/candidate
```

The output directory must be new. All six platform targets are built by default. `--platforms linux/amd64` selects a subset. Build arguments explicitly embed version, source SHA and standalone ownership; archives use source-commit time, sorted paths and normalized modes. Files include the main CLI, Windows windowless launcher, native qualification helper, docs, legal corpus, presets, instruction-only Skill sources, license map and full dependency notices. Optimizers, toolchain caches, private profiles and credentials are never collected.

The archive receipt binds individual member hashes. `SHA256SUMS` binds every archive. `build-index.json` is written incrementally and marks complete only after the entire selected build succeeds. Failure leaves inspectable local output and returns nonzero, not a nominal release.

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

## Standalone update archive and signing

The update archive is deliberately different from the full development package. It contains only allowlisted application executables and license obligations. User config, presets, images, Skills and managed tools cannot appear as replacement targets. The builder verifies its generated member list against the same runtime update validation contract.

`release.UNSIGNED.json` is **not** accepted by the application updater. An operator who already holds the independently qualified production signing key can sign that payload:

```sh
go run ./tools/sign-update --payload artifacts/release-inputs/release.UNSIGNED.json --key-file /PROTECTED/PATH/release-key.pem --key-id CHOSEN-KEY-ID --out artifacts/release-inputs/release.signed.json
```

The command reads an existing protected PKCS#8 Ed25519 key, signs exact payload bytes, self-validates the result and writes a new envelope. It neither generates the key nor publishes it; no secret value belongs in argv, the repository or archives. The production public key/feed must be bound through a separate trustworthy channel. Unit tests use ephemeral test keys only. Authenticode/notarization are distinct from this manifest signature.

## Evidence boundary

Pure tests cover archive determinism, member modes, runtime-compatible update extraction, no user-data members and package-manager hash propagation. Source compilation and local archive verification do not substitute for live installers, signing infrastructure or desktop/optimizer qualification. No hosted CI result or public release is inferred merely from running this builder.
