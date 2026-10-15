# Candidate builds and the release boundary

## Meaning

smolpng currently produces private development candidates, not a public release. This document owns the build provenance, archive contents and qualification obligations for those candidates. No published package-manager identity, signed updater feed, Store listing or release automation is implied.

## Exact provenance

Build a clean source commit with Go 1.26.6 or a separately qualified compatible toolchain. Bind the application version report to that complete SHA, not just a friendly version label. The native qualification helper must be bound to the same SHA.

PowerShell, from the source root on Windows:

```powershell
$sha = git rev-parse HEAD
$flags = "-X github.com/Wolfsblvt/smolpng/internal/cli.Commit=$sha -X github.com/Wolfsblvt/smolpng/internal/cli.Distribution=standalone"
go build -trimpath -ldflags $flags -o artifacts/bin/ ./cmd/smolpng
go build -trimpath -ldflags "-H=windowsgui $flags" -o artifacts/bin/smolpngw.exe ./cmd/smolpngw
go build -trimpath -ldflags "-X main.Commit=$sha" -o artifacts/bin/smolpng-qualify.exe ./tools/qualify
```

On Linux/macOS, build `cmd/smolpng` and `tools/qualify` with the corresponding linker values and no `.exe` suffix. The `smolpngw` entry point is Windows-only. Explicit `GOOS`, `GOARCH` and `CGO_ENABLED=0` can cross-build the six supported source targets; compilation is not execution on the target host.

`Distribution=standalone` records who owns these portable development files. Self-update still needs an independently trusted signed manifest; the candidate has no production feed/key. Replacing the executable manually preserves external config, state, presets, managed tools and tracked Skills; human startup handles changed Skill contents. The application updater is implemented; its production signing/feed operation is not configured. See updates.md.

## What accompanies a candidate

A runnable platform archive includes the console application, Windows windowless launcher when applicable, native qualification helper, primary license, complete third-party license notices, relevant usage/qualification documents, original CC0 test corpus, and a manifest of file hashes plus source SHA. It must contain no Go toolchain/cache, credentials, private agent Profile, personal config, external optimizer executable or commissioned brand asset. Standard packages do not bundle pngquant or OxiPNG.

The source recovery ZIP includes the complete Git repository and a second Git bundle. A plain working-tree ZIP is not a substitute for recoverable commit history. Verify archive integrity and the extracted source's Git object database; recheck HEAD and a clean working tree. Preserve a separate SHA-256 checksum for each download.

The workflow retains notices beside private CI binary artifacts as well as building exact-head code. A workflow file is not hosted CI evidence. Quote a hosted run only after reading its actual conclusion and matching source SHA.

## Qualification, not nominal success

Run ordinary tests, race detection where supported, vet, module verification and the corpus integrity check. Compile all claimed targets. Check Windows PE machine/subsystem metadata: console applications use the console subsystem; `smolpngw.exe` uses the GUI subsystem. Inspect an actual native version report and real JSON/schema output.

Use `smolpng-qualify` against the built candidate and separately installed real optimizers, then manually exercise native clipboard representations, file-picker paste, graphical terminal rendering, drag/drop, windowless/console transitions, notifications, locked files and cancellation. Keep exact binary and optimizer versions with reports. A fake-process regression test, pseudo-terminal run, cross-build or archive checksum proves only its own boundary.

## Remaining outward effects

A public release requires separate exact authority for repository visibility, release upload, packages, signing, website deployment, donations and Store effects. Pin and review release-tool dependencies when that machinery is selected. Do not publish install commands for unregistered packages or claim a signed binary when none was signed. Native platform observations, legal brand/artist grant, optimizer distribution composition and current package-manager acceptance remain distinct checks.

## Executable local builder

`go run ./tools/release --version 0.2.0-dev.1 --out artifacts/candidate` now creates the exact platform packages and checksum/index files described here. `--base-url` plus a stable version additionally authors package-manager inputs and a narrow standalone update payload. A separate `tools/sign-update` command requires an already held protected key; normal candidate builds do not sign. See [Packaging](packaging.md) for the implemented invocation and strict publication boundary.
