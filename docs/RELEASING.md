# Candidate builds and the release boundary

## Meaning

smolpng currently produces private development candidates, not a public release. This document owns the build provenance, archive contents and qualification obligations for those candidates. No published package-manager identity, signed updater feed, Store listing or release automation is implied.

## Exact provenance

Build a clean source commit with Go 1.26.6 or a separately qualified compatible toolchain. Bind the application version report to that complete SHA, not just a friendly version label. A packaged native qualification helper is built from that same SHA. A separately built helper may exercise an existing candidate with explicit `-expected-commit SHA`; record both source identities and executable hashes rather than claiming the helper and application share a build.

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

A runnable platform archive includes the console application, Windows windowless launcher when applicable, native qualification helper, version-bound Agent Skill source, primary license, complete third-party license notices, relevant usage/qualification documents, original CC0 test corpus, and a manifest of file hashes plus source SHA. It must contain no Go toolchain/cache, credentials, private agent Profile, personal config, external optimizer executable or commissioned brand asset. Standard packages do not bundle pngquant, OxiPNG or jpegoptim.

The source recovery ZIP includes the complete Git repository and a second Git bundle. A plain working-tree ZIP is not a substitute for recoverable commit history. Verify archive integrity and the extracted source's Git object database; recheck HEAD and a clean working tree. Preserve a separate SHA-256 checksum for each download.

The workflow retains notices beside private CI binary artifacts as well as building exact-head code. A workflow file is not hosted CI evidence. Quote a hosted run only after reading its actual conclusion and matching source SHA.

## Qualification, not nominal success

Run ordinary tests, race detection where supported, vet, module verification and the corpus integrity check. Compile all claimed targets. Check Windows PE machine/subsystem metadata: console applications use the console subsystem; `smolpngw.exe` uses the GUI subsystem. Inspect an actual native version report and real JSON/schema output.

Use `smolpng-qualify` against the built candidate and separately installed real optimizers, then manually exercise native clipboard representations, file-picker paste, graphical terminal rendering, drag/drop, windowless/console transitions, notifications, locked files and cancellation. Keep exact binary and optimizer versions with reports. A fake-process regression test, pseudo-terminal run, cross-build or archive checksum proves only its own boundary.

## Remaining outward effects

The repository source is already public under its selected licences. A public application release still requires separate exact authority for release upload, package-manager or Store submission, production signing and update-feed operation, website deployment, donations and spend. Pin and review release-tool dependencies when that machinery is selected. Do not publish install commands for unregistered packages or claim a signed binary when none was signed. Native platform observations, legal brand/artist grant, optimizer distribution composition and current package-manager acceptance remain distinct checks.

## Windows signing sequence

Wolf selected the open-source signing route on 21 October 2026. The first public Preview may be distributed **unsigned** when its release page, checksums and Windows guidance say so plainly. An unsigned Preview is not a signed release with an inconvenient warning attached; do not claim an authenticated publisher or SmartScreen reputation that does not exist.

After the first documented public release, smolpng may apply to SignPath Foundation. Treat provider acceptance as unobserved until it happens. At the decision date, the published route expects an already released and documented open-source project, a public code-signing policy, MFA, trusted source/build provenance and manual approval for each release; recheck the current terms at application and activation time. The certificate publisher is SignPath Foundation rather than Wolfsblvt Works.

Before a SignPath-signed cut, carry one coherent signing contribution:

- verify the maintained [Windows version resources](packaging.md#windows-executable-identity), including ProductName, CompanyName, FileVersion and ProductVersion values derived from the exact candidate identity;
- inspect the actual PE resources on x64 and ARM64;
- publish the code-signing policy and bind the trusted build/origin route;
- configure no signing secret outside the accepted provider custody;
- sign only exact accepted artifacts and read back the resulting signatures and certificate subject; and
- keep Authenticode, SHA-256 checksums and the Ed25519 standalone-update manifest as distinct trust layers.

Microsoft Azure Artifact Signing is deferred, not selected. It may be reconsidered if Wolfsblvt Works later has an eligible legal organization and the paid Azure operating consequence is deliberately accepted. Traditional commercial certificates and self-signed public artifacts are not the current route.

The current pre-1.0 machine and product posture is **`0.3.0` Public Preview**. This is a provisional maturity identity, not release authorization and not a smaller destination. Finish every ratified application capability and its honest qualification first. Wolf then owns the final public-experience workshop and acceptance: website, documentation voice and language, and the actual icon/brand identity. The official stable release is **`1.0.0`** only after the complete product and that public experience are accepted and publication is explicitly authorized. Green source, packages or native evidence do not promote the version by themselves.

Current packages still have placeholder visual identity. No accepted `.ico`/`.icns` or brand asset is present; Windows shortcut creation sets no explicit `IconLocation`, while Linux desktop entries use `image-x-generic`. The maintained Windows VERSIONINFO resources identify executable product/version/source facts, but they do not supply the final Squishling/app icon. Keep these seams replaceable until Wolf's design workshop and accepted artist/icon package.

## Executable local builder

`go run ./tools/release --version 0.2.0-dev.1 --out artifacts/candidate` now creates the exact platform packages and checksum/index files described here. `--base-url` plus a stable version additionally authors package-manager inputs and a narrow standalone update payload. A separate `tools/sign-update` command requires an already held protected key; normal candidate builds do not sign. See [Packaging](packaging.md) for the implemented invocation and strict publication boundary.

The builder uses the application's direct process constructor for Git and Go. On Windows those children start without a console window; build output, explicit cross-compilation environment and exit-status handling remain available. The native process tests check both a direct child and inherited console behavior. This process qualification does not establish graphical application or desktop-integration behavior.

The [private dev.9 record](reference/2026-10-19-private-candidate-dev9.md) binds one completed six-platform build to source restoration, 63 native Windows checks, dependency/Skill lifecycle and actual interrupted-update recovery. Its explicit remaining joined-desktop and other-host boundaries still govern; a completed build is not complete candidate acceptance. A later documentation-only record does not change the observed candidate source or require rebuilding it to include its own evidence.

The [private dev.10 record](reference/2026-10-19-private-candidate-dev10.md) binds the later macOS catalog repair to a fresh six-platform build, independent archive/source restoration and dependency/Skill lifecycle. Native invocation failures and their focused reproductions remain explicit; related earlier observations do not become a clean new joined candidate. The application's per-image timeout and the qualifier's whole-invocation deadline are distinct budgets.

The [private dev.11 record](reference/2026-10-20-private-candidate-dev11.md) binds the next exact six-platform archive set to source restoration, a clean 63-check Windows run, dependency/Skill lifecycle, update interruption/recovery, active cancellation, hidden-terminal control flow and tracked shortcut installation/removal. Later native platform and Linux package lifecycle records strengthen their own source boundaries. Dev.11 remains a private behavior specimen rather than the final joined human/script/agent candidate after current embedded Skill guidance changed; the next complete candidate must include the accepted current source.

The [joined private dev.13 record](reference/2026-10-22-private-candidate-dev13.md) binds the accepted Finder/Linux/Windows source to six inspected archives, two complete-history restores, all 79 native Windows PNG/JPEG checks and the exercised desktop/TinyPNG/primary-menu preparation packets. Earlier [dev.12 lifecycle evidence](reference/2026-10-22-private-candidate-dev12.md) remains separately identified. The joined physical Windows journey is still open; this record is neither candidate acceptance nor a public release. Adding its evidence after composition does not change the immutable candidate or require another build.

The subsequent [dev.13 Windows lifecycle record](reference/2026-10-22-windows-dev13-lifecycle.md) retains the same candidate's 28 managed dependency/Skill calls, independently inspected instruction ZIP and 12 local signed-fixture update/recovery calls, including independent restoration of all 197 original package files. It distinguishes scheduled acknowledgement, actual transaction completion, helper exit and retained user edits. This is a local test-fixture observation, not production trust/feed or joined graphical acceptance.
