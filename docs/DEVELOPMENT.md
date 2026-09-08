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
go run ./cmd/smolpng inspect testdata/corpus/alpha-gradient.png --json
```

Output is `smolpng.exe` on Windows and `smolpng` elsewhere. Linker variables `internal/cli.Version`, `Commit`, and `Distribution` carry exact build provenance. Set `CGO_ENABLED=0`, `GOOS` and `GOARCH` for cross-builds; supported build targets are Windows/Linux/macOS on amd64 and arm64. In PowerShell use `$env:GOOS`, not POSIX assignment syntax, and clear overrides before native tests.

## Corpus

```sh
go run ./tools/corpus
go run ./tools/corpus --check
```

The generator owns ten project-original CC0 PNGs and a hash/structure/source manifest. Ordinary tests generate temporary specimens, not source-tree output. No lossy byte snapshots are asserted across optimizer versions.

## Real-host qualification

Install current compatible pngquant/OxiPNG separately; record their exact `deps status --json` output. Compress the corpus into a fresh directory, compare dimensions/alpha and strict lossless decoded channels, exercise target success/miss, cancellation, locked outputs and binary stdout. The admitted argument families are pngquant 2.12+/3.x and OxiPNG 9.x/10.x; an admitted version probe is not a claim that every version was actually exercised.

The authoring sandbox uses an isolated checksum-verified Go/module cache transported from an earlier private workflow because shell DNS is unavailable. This changes neither product installation nor source prerequisites. Repository checkpoints include Git history, not the compiler/cache or external optimizer binaries. Hosted CI results must name their exact commit rather than inherit a predecessor's green result.

## Configuration/preset verification

`go test ./internal/config ./internal/cli` covers source precedence, explicit false/zero, provenance, strict unknown-field/data-only validation, cycles, import snapshots/renames, dependency-aware removal, remote permission and full resolved export roundtrips. `presets/examples/*.toml` are parsed and resolved in ordinary tests. `config show --effective --provenance` provides local inspection without running an optimizer.

## Agent integration verification

`go test ./internal/skill ./internal/fsutil ./internal/cli` exercises actual temporary directories, native process-scoped locking, no-replace moves, installed hashes, custom-path consent, modified-file protection, replacement prompt conditions, interrupted swaps and deterministic ZIP entries. It does not install to the developer's actual harness folders. `skills/smolpng/` is the source of the embedded artifact; its content hash changes automatically when any resource changes. A rebuilt binary with unchanged resources retains the same bundle version.


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
