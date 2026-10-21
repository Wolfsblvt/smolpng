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

Hosted `Verify` runs the ordinary suite with `go test -race ./...` and cgo enabled on Linux; Windows and macOS run `go test ./...`. The Linux race run replaces that job's uninstrumented test step rather than adding another test job. It detects races reached by the tests, not all possible concurrency defects or native desktop behavior. Local race checking remains optional when the required C toolchain is unavailable; do not install a machine-wide compiler merely to duplicate hosted evidence.

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

The corpus contains 32 project-original CC0 structural specimens and two publisher-original CC0 photographs. The profile controls include complete RGB, grayscale, CMYK and Display P3 inputs. The single hash/structure/source manifest retains imported admission fields and fixed hashes; regeneration never downloads, rewrites or silently rehashes imports. Equivalent existing generated PNG encodings are preserved across Go versions; their pixels and metadata must match, including decoded contents of compressed ICC profiles. Generated JPEG remains byte-exact. Ordinary tests generate temporary specimens and read the checked-in corpus without writing it. No lossy byte snapshots are asserted across optimizer versions. The [corpus guide](../testdata/corpus/README.md) records provenance, regeneration and remaining photographic/arbitrary-profile gaps.

## Real-host qualification

Install current compatible pngquant/OxiPNG separately; record their exact `deps status --json` output. Compress the corpus into a fresh directory, compare dimensions/alpha and strict lossless decoded channels, exercise target success/miss, cancellation, locked outputs and binary stdout. The admitted argument families are pngquant 2.12+/3.x and OxiPNG 9.x/10.x; an admitted version probe is not a claim that every version was actually exercised.

The [Linux package qualifier](../.github/workflows/linux-packages.yml) exercises maintained nFPM packages in disposable hosted Linux environments. Ubuntu 24.04 native amd64 and arm64 runners install DEBs through dpkg. A separate Ubuntu 24.04 amd64 runner uses digest-recorded Fedora, Alpine and Arch Linux containers for RPM, APK and Arch package-manager lifecycles; these share the hosted runner's kernel and do not qualify a complete distribution installation or desktop. Each route builds an explicit `0.0.0` fixture with a reserved `.invalid` asset address, installs pinned nFPM in temporary runner storage, inspects the package, checks package ownership and disabled self-update, runs the real PNG/JPEG qualifier against `/usr/bin/smolpng`, removes the package and compares retained per-user dependency files. Catalog-managed OxiPNG and jpegoptim are used where their Linux amd64 runtime fits; Alpine uses the same checksum-pinned jpegoptim 1.5.6 source built natively for musl because the managed archive targets glibc. Each report binds the source, package and executable hashes, optimizer identities, distribution, container image and shared native host boundary. The [amd64 RPM/APK/Arch lifecycle record](reference/2026-10-20-linux-amd64-container-package-lifecycle.md), [amd64 DEB record](reference/2026-10-20-linux-deb-lifecycle.md) and [native arm64 DEB lifecycle record](reference/2026-10-20-linux-arm64-deb-lifecycle.md) preserve their exact runs. Fedora's RPM metadata did not match its installed pngquant CLI; the record documents the checksum-pinned upstream source build needed for the supported optimizer. The Arch record preserves pacman's existing `NoExtract` rule for `/usr/share/doc/*` rather than changing container policy. The workflow runs for its own source changes and can be dispatched after reaching the default branch. Only reports are uploaded; fixture packages, executables and images are not published. A workflow definition is not a passing hosted result, and package-manager observations do not qualify a full distribution desktop.

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

The on-demand **Native optimizer qualification** workflow builds the selected source and matching qualifier on Ubuntu 24.04 amd64, macOS 26 Intel/ARM64, and Windows 11 ARM64 hosts. All download catalog-pinned OxiPNG into an isolated application home; Linux and ARM64 macOS also install catalog-pinned jpegoptim. Intel macOS uses Homebrew's native jpegoptim because the upstream catalog archive is actually ARM64 despite its `x64-osx` filename. Linux extracts Ubuntu's separate pngquant/library packages under runner temporary storage without a system installation; macOS uses Homebrew's separate pngquant on its disposable CI machines, recording the selected formula and dependencies. Windows ARM64 builds the application and qualifier natively, then uses pinned x64 Windows pngquant/OxiPNG/jpegoptim archives under Windows x64 emulation and records both PE architectures. Run it with `gh workflow run native-optimizers.yml --repo Wolfsblvt/smolpng --ref main`. This replaces the earlier Linux-only filename; its historical run links remain evidence. The workflow also runs on PRs changing its own file, while ordinary source PRs retain the existing `Verify` surface.

Each host's 14-day artifact contains the complete or interrupted qualifier report, exact source/executable/tool/package identities and detailed host release. It excludes application binaries and generated images. The [first native Linux record](reference/2026-10-19-linux-native-qualification.md) and [both Mac architectures](reference/2026-10-19-macos-native-qualification.md) have all 63 checks passing. This establishes only behavior actually exercised on those hosts; package installation and graphical clipboard/desktop behavior remain separate from CLI qualification. A red preflight or incomplete report is not optimizer success. Consume the exact hosted result and preserve its durable qualification account before claiming the host qualified.

Normal tests use an in-memory HTTP transport and staged fake executable/version boundary, not upstream downloads. They cover catalog digests, malformed archives, verified-before-execute behavior, failed-probe cleanup, modified installations, retained versions, downgrade refusal, concurrent writers and protected removal. Actual artifact bytes and native OxiPNG/jpegoptim behavior require explicit installation on the target host. Follow the [managed dependency guide](dependencies.md), inspect `deps licenses` and an install preview before `--apply`, and record `deps status --json` plus exact tool versions alongside corpus reports. Native archive availability is host-specific; an x64 executable does not establish native arm64 qualification.

## Exercise the built candidate without rebuilding

The optional `smolpng-qualify` helper (`tools/qualify`) invokes the candidate executable with a fresh isolated home and project lookup disabled. It writes original CC0 PNG/JPEG specimens only under a new selected folder, exercises real Auto/lossless optimizers, exact targets, nearest acceptance, binary stdout, partial batches and a mixed-format batch, and checks actual artifact decoding/pixels/bytes and JPEG orientation. JPEG coverage includes the four grayscale/CMYK/Display P3 controls under Auto, Lossless, explicit profile removal and resizing. Ordinary JPEG paths must preserve component count and the exact profile; a requested supported resize may produce RGB. Tagged grayscale/CMYK resizing must explain its existing refusal and commit no output. These are encoded-channel/profile checks, not independent CMS conversion or visual quality acceptance. It records the tested executable's SHA-256, expected build commit, selected formats, host/platform, discovered tool paths/versions and per-check reports. Before corpus work it checkpoints SHA-256 identities for the selected optimizers; a final comparison refuses a passing qualification if those bytes changed. These are measurements at the two boundaries, not continuous monitoring or proof of execution before capture. It never downloads tools, accesses the clipboard, sends images to a remote service or publishes its report.

The `host` object records OS and architecture without machine names or user identifiers. Windows additionally records the native kernel build and available edition, display release and update revision; unavailable details are explicit. Other hosts currently require their detailed OS release to be recorded separately. `optimizerSHA256` keys match the selected executables in `tools.tools`; PNG-only and JPEG-only runs hash only their required tools. These fields do not retroactively fill omissions in older reports.

```sh
smolpng-qualify -executable ./smolpng -output ./qualification-run
```

Both formats run by default. `-formats png` keeps a PNG-only run; `-formats jpeg` selects JPEG alone. `-pngquant PATH`, `-oxipng PATH` and `-jpegoptim PATH` select separate real executables. Preflight requires pngquant/OxiPNG only for selected PNG checks and jpegoptim only for selected JPEG checks. Use `.exe` names and PowerShell's `./` or `\.\` form on Windows. The package builder embeds the expected source commit; a mismatch refuses to qualify the wrong build. `-expected-commit SHA` explicitly selects the candidate when using a separately built helper. The output directory must not already exist. Native optimizer checks are not part of ordinary CI; a preflight refusal is not native optimizer success.

`report.json` is saved incrementally and on an interrupted/refused run. `qualifierCommit` identifies the helper build independently from the candidate's `expectedCommit`. `complete` remains false until all selected checks finish; top-level `passed` is true only for a complete passing run. Individual passing checks remain visible in partial evidence. This prevents a successful prefix from being consumed as full qualification. [The native Windows run](reference/2026-10-17-windows-native-qualifier.md) passed all 61 PNG/JPEG checks and three separate refusal cases. [Hosted Windows ARM64 qualification](reference/2026-10-20-windows-arm64-native-qualification.md) passed all 63 PNG/JPEG checks with a native ARM64 app and qualifier and explicit x64 optimizer executables. These generated structural fixtures do not replace the separate photographic, ICC or graphical records.

The expanded [JPEG color-model qualifier](reference/2026-10-21-native-jpeg-color-qualifier.md) passed all 79 PNG/JPEG checks on hosted Linux amd64, both macOS architectures and Windows ARM64, plus all 51 JPEG-only checks on Windows amd64. Its record retains the first bad-flag failure, the real CLI parser regression and the corrected exact-source runs. The helper remains separate from the application under test and from private candidate composition.

Graphical clipboard, drag/drop, native notifications, windowless launch/console allocation and visual TUI behavior still require the manual checks in the candidate's local qualification packet. [Four real Windows locked-output cases](reference/2026-10-17-windows-output-locks.md) now establish PNG/JPEG overwrite/in-place refusal, byte preservation, temporary-file cleanup and successful retry after unlock. Other-host locks remain separately qualified. The helper exercises rather than rebuilds the app.

## JPEG backend comparison

`tools/qualify/jpeg-backends.py` is a separate native calibration journey for [Issue #57](https://github.com/Wolfsblvt/smolpng/issues/57), not an application runtime or an ordinary Go unit suite. It needs Python 3.12 with Pillow 12.3.0/LittleCMS, an already-built smolpng, jpegoptim and a separately built candidate cjpegli. The selected Python boundary supplies independent decoding, profile conversion and lossless pixel-crop artifacts without adding another application codec or runtime. The `JPEG backend comparison` workflow selects those tools on one disposable Ubuntu 24.04 host, pins upstream jpegli/Highway/skcms/libjpeg-turbo sources and retains build identities. It runs only manually or for changes to its own workflow/helper.

```sh
python tools/qualify/jpeg-backends.py --source . --smolpng /path/to/smolpng \
  --jpegoptim /path/to/jpegoptim --jpegli /path/to/cjpegli \
  --jpegli-source EXACT_UPSTREAM_COMMIT --output artifacts/validation/jpeg-backend-comparison
```

The helper validates the existing CC0 fixture manifest and refuses an existing output directory. It compares the two photographic inputs at the selected jpegoptim 85/80/70 anchors and actual matching jpegli byte budgets. Its prepared PNGs are losslessly decoded, display-oriented references; preparation is timed separately. Managed sRGB channel diagnostics and two 100% pixel crops per photograph accompany original/candidate hashes and outputs. These diagnostics are not perceptual acceptance. Untagged RGB assumes sRGB; untagged CMYK conversion remains unmanaged. Source color profiles and conversion assumptions stay visible.

Four grayscale/CMYK/Display P3 candidate calls separately record output, refusal or failure and profile/dimension observations. A complete comparison report means every outcome was captured, not that all candidate capabilities passed or that jpegli was accepted. Actual output size is measured; cjpegli's target-size request is not promoted to smolpng's strict target promise. Reports checkpoint before and after native calls so a failed late step retains earlier results. Qualify the exact hosted outcome before making a backend recommendation; codecs and defaults remain unchanged by this helper.

## Native Windows menu addon

The optional [primary context menu](context-menu.md) has a small Windows COM DLL outside the Go executable. `native/windows-shell/core.hpp` is portable; its parser and binary request protocol can be checked here:

```sh
cmake -S native/windows-shell -B artifacts/native-shell
cmake --build artifacts/native-shell
ctest --test-dir artifacts/native-shell --output-on-failure
```

Go also decodes and re-encodes a checked-in request generated by the C++ serializer. Linux core tests do not prove Windows API compilation or Explorer behavior. `native/windows-shell/build.ps1` owns the optional SDK build, source-bound receipt and sparse package staging. No ordinary test registers a package or changes certificate trust.

The [classified corpus](../testdata/corpus/README.md) records 28 generated structural specimens and two admitted photographic originals, including a compound Samsung camera file. The [native RGB ICC record](reference/2026-10-17-windows-icc-qualification.md) carries controlled PNG/JPEG color evidence; broader photographic/wide-gamut/CMYK calibration remains open. `go run ./tools/corpus --check` rejects missing, modified and unmanifested media. Run generation to completion before invoking checks; do not write evaluation outputs into the fixture tree.

## GitHub maintenance

`main` is the public pre-release integration line, protected from deletion, force pushes and merge commits. PRs are preferred; neither a second maintainer approval nor a metadata gate is required. `Verify` uses read-only permissions and consumes no remote optimizer/API quota. Public standard GitHub-hosted runner execution is free; this is not a claim that every possible Actions storage or larger-runner cost is free.

[The canonical label policy](../.github/label-policy.json) defines manual semantic types and automatic PR areas. [Automation standing](../.github/automation/README.md) records the enabled, provider-qualified Forgehand App caller and the separate unqualified sizing route. Grouped Go and Actions dependency updates are configured in `.github/dependabot.yml`. Sensitive reports use the root [security policy](../SECURITY.md). External contributions are not currently accepted; MIT permissions to use and fork remain intact.
