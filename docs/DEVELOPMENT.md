# Development

## Meaning

This guide owns root build/test commands and the distinction between deterministic source verification, actual optimizers and graphical desktop qualification. It does not claim a published installer or release channel.

## Toolchain and ordinary verification

Use Go 1.26.6 or a compatible newer version. `go.mod` and `go.sum` pin source dependencies. Ordinary tests need no optimizer, credential, network service or Docker daemon.

```sh
go mod download
go mod verify
go test ./...
go vet ./...
go test -race ./...
```

`go test ./...` discovers every ordinary `*_test.go`; helper and executable packages need no decorative tests. Race checking requires a native C toolchain. Source tests do not spend TinyPNG quota. Focus a change with `go test ./internal/workflow ./internal/cli`.

## Build and smoke

```sh
go build -o artifacts/bin/ ./cmd/smolpng
go run ./cmd/smolpng capabilities --json
go run ./cmd/smolpng inspect testdata/corpus/png/alpha-gradient/alpha-gradient.png --json
```

Output is `smolpng.exe` on Windows and `smolpng` elsewhere. Linker variables `internal/cli.Version`, `Commit`, and `Distribution` carry exact build provenance. Set `CGO_ENABLED=0`, `GOOS` and `GOARCH` for cross-builds; supported build targets are Windows/Linux/macOS on amd64 and arm64. In PowerShell use `$env:GOOS`, not POSIX assignment syntax, and clear overrides before native tests.

## Corpus

```sh
go run ./tools/corpus
go run ./tools/corpus --check
```

The corpus contains 24 project-original CC0 structural specimens and two publisher-original CC0 photographs. The single hash/structure/source manifest retains imported admission fields and fixed hashes; regeneration never downloads, rewrites or silently rehashes imports. Equivalent existing generated PNG encodings are preserved across Go versions; their pixels and non-image-data chunks must match. Generated JPEG remains byte-exact. Ordinary tests generate temporary specimens and read the checked-in corpus without writing it. No lossy byte snapshots are asserted across optimizer versions. The [corpus guide](../testdata/corpus/README.md) records provenance, regeneration and remaining photographic/ICC gaps.

## Real-host qualification

Install current compatible pngquant/OxiPNG separately; record their exact `deps status --json` output. Compress the corpus into a fresh directory, compare dimensions/alpha and strict lossless decoded channels, exercise target success/miss, cancellation, locked outputs and binary stdout. The admitted argument families are pngquant 2.12+/3.x and OxiPNG 9.x/10.x; an admitted version probe is not a claim that every version was actually exercised.

The original Web authoring host used a transported checksum-verified Go/module cache because its shell could not resolve the module proxy. That is historical authoring evidence, not a development prerequisite. Normal local development uses the installed SDK and the ordinary checksum-verified module download route. Repository checkpoints include Git history, not the compiler/cache or external optimizer binaries. Hosted CI results must name their exact commit rather than inherit a predecessor's green result.

## Configuration/preset verification

`go test ./internal/config ./internal/cli` covers source precedence, explicit false/zero, provenance, strict unknown-field/data-only validation, cycles, import snapshots/renames, dependency-aware removal, remote permission and full resolved export roundtrips. `presets/examples/*.toml` are parsed and resolved in ordinary tests. `config show --effective --provenance` provides local inspection without running an optimizer.

## Agent integration verification

`go test ./internal/skill ./internal/fsutil ./internal/cli` exercises actual temporary directories, native process-scoped locking, no-replace moves, installed hashes, custom-path consent, modified-file protection, replacement prompt conditions, interrupted swaps and deterministic ZIP entries. It does not install to the developer's actual harness folders. `skills/smolpng/` is the source of the embedded artifact; its content hash changes automatically when any resource changes. A rebuilt binary with unchanged resources retains the same bundle version.

The [native Windows post-update Skill record](reference/2026-10-17-windows-update-skill-sync.md) separately qualifies a real signed-fixture update with managed, edited and custom installations in a disposable user profile. It proves filesystem synchronization and preservation, not an agent client's reload.


## Keyboard and desktop verification

`go test ./internal/ui ./internal/desktop ./internal/state ./internal/cli` exercises actual menu key messages, typed-name accessibility, multi-format clipboard payloads, exact native command arguments, last-result retention, first-run acknowledgement, preset import, settings edits and shared-engine human/drop behavior. Graphical clipboard, file-picker and notification APIs are substituted by focused external-boundary doubles, not claimed as observed desktops.

For Windows's no-console executable, build natively with:

```powershell
go build -ldflags=-H=windowsgui -o artifacts/bin/smolpngw.exe ./cmd/smolpngw
```

The normal console executable is still required for scripts, diagnostics and ordinary interactive launch. The no-console entry point always invokes `drop`. Only an explicitly configured interactive drop workflow allocates a console. The Windows GUI subsystem and both architectures must be checked before distribution; successful cross-builds do not qualify Start-menu/Explorer behavior.

A Linux pseudo-terminal smoke has exercised actual menu rendering, Enter/Done, original-copy target success and last-result recording without fake terminal drivers. It used an already-fitting PNG, so it is not evidence of optimizer compression, graphical clipboard or a human's visual-quality assessment. Qualification records must bind that experiment to the exact rebuilt commit.

## Hosted verification source

`Verify` replaces the one-time build-cache transport workflow. It runs root tests/vet/corpus checks on Windows, macOS and Linux, plus six CGO-free cross-builds and Windows GUI-launcher builds. Its three Actions are pinned to exact official upstream commits, checked on 2026-09-08. It uses read-only repository permission, persists no checkout credential, consumes no paid provider quota and publishes no release. A workflow file is not a hosted green run; do not inherit the bootstrap run's status.

## Managed optimizer qualification

Normal tests use an in-memory HTTP transport and staged fake executable/version boundary, not upstream downloads. They cover catalog digests, malformed archives, verified-before-execute behavior, failed-probe cleanup, modified installations, retained versions, downgrade refusal, concurrent writers and protected removal. Actual artifact bytes and native OxiPNG corpus behavior still require an explicit install on the target host. Run `deps licenses`, `deps install oxipng`, then `deps install oxipng --apply` only after reviewing the source. Record the resulting `deps status --json` and exact tool versions alongside corpus reports. On Windows arm64, configure a separately qualified native program; do not claim the x64 archive was a native build.

## Exercise the built candidate without rebuilding

The optional `smolpng-qualify` helper (`tools/qualify`) invokes the candidate executable with a fresh isolated home and project lookup disabled. It writes original CC0 specimens only under a new selected folder, exercises real Auto/lossless optimizers, exact targets, nearest acceptance, binary stdout and partial batches, and checks actual artifact decoding/pixels/bytes. It records the tested executable's SHA-256, build commit, platform, discovered tool versions and per-check reports. It never downloads tools, accesses the clipboard, sends images to a remote service or publishes its report.

```sh
smolpng-qualify -executable ./smolpng -output ./qualification-run
```

Use `.exe` names and PowerShell's `./` or `\.\` form on Windows. `-pngquant PATH` and `-oxipng PATH` select separate real executables; both are required for complete qualification. The package builder embeds the expected source commit; a mismatch refuses to qualify the wrong build. The output directory must not already exist. Native optimizer checks are not part of ordinary CI. A preflight refusal on a headless authoring host with no optimizers is not counted as real optimizer success.

Graphical clipboard, drag/drop, native notifications, windowless launch/console allocation, locked files and visual TUI behavior still require the manual checks in the candidate's local qualification packet. The helper exercises rather than rebuilds the app.

## Native Windows menu addon

The optional [primary context menu](context-menu.md) has a small Windows COM DLL outside the Go executable. `native/windows-shell/core.hpp` is portable; its parser and binary request protocol can be checked here:

```sh
cmake -S native/windows-shell -B artifacts/native-shell
cmake --build artifacts/native-shell
ctest --test-dir artifacts/native-shell --output-on-failure
```

Go also decodes and re-encodes a checked-in request generated by the C++ serializer. Linux core tests do not prove Windows API compilation or Explorer behavior. `native/windows-shell/build.ps1` owns the optional SDK build, source-bound receipt and sparse package staging. No ordinary test registers a package or changes certificate trust.

The [classified corpus](../testdata/corpus/README.md) records 24 generated structural specimens and two admitted photographic originals, including a compound Samsung camera file. Broader photographic/ICC calibration remains open. `go run ./tools/corpus --check` rejects missing, modified and unmanifested media. Run generation to completion before invoking checks; do not write evaluation outputs into the fixture tree.

## GitHub maintenance

`main` is the public pre-release integration line, protected from deletion, force pushes and merge commits. PRs are preferred; neither a second maintainer approval nor a metadata gate is required. `Verify` uses read-only permissions and consumes no remote optimizer/API quota. Public standard GitHub-hosted runner execution is free; this is not a claim that every possible Actions storage or larger-runner cost is free.

[The canonical label policy](../.github/label-policy.json) defines manual semantic types and automatic PR areas. [Automation standing](../.github/automation/README.md) records the enabled, provider-qualified Forgehand App caller and the separate unqualified sizing route. Grouped Go and Actions dependency updates are configured in `.github/dependabot.yml`. Sensitive reports use the root [security policy](../SECURITY.md). External contributions are not currently accepted; MIT permissions to use and fork remain intact.
