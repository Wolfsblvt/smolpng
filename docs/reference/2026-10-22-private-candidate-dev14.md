# Private candidate 0.3.0-dev.14

## Meaning

Private `0.3.0-dev.14` carries the Windows Explorer startup repair at `825270987bfee3eb2064b23d5398162c24871ffa` into six inspected archives. Both source-history carriers restored successfully, all 79 native Windows PNG/JPEG checks passed, and desktop, TinyPNG and primary-menu preparation passed through the extracted candidate. This dated record preserves the exact reports and their limits. Physical drag/Send To, smart clipboard consumers and rendered notifications remain open in [Issue #36](https://github.com/Wolfsblvt/smolpng/issues/36); this is neither complete candidate acceptance nor a public release.

## Explorer repair and source boundary

The owner-operated dev.13 drop delivered correct arguments but exited 1 before compression. Cobra's Windows Explorer-parent guard printed its command-line-only warning, waited five seconds and exited before smolpng's dispatch. The windowless launcher concealed that warning. [PR #79](https://github.com/Wolfsblvt/smolpng/pull/79) disables that guard once during Windows CLI initialization, preserving the intended Explorer entry paths.

The native subprocess regression reproduces Explorer-parent detection without controlling the desktop. A temporary overlay removing only the repair made that regression fail with the warning and exit 1; the repaired regression and full Windows CLI suite passed. A normal, uninstrumented proposal launcher also completed a generated PNG job when activated by a real Explorer parent, with independently decoded output, report and last-result readback. That was command activation, not physical dragging. The [source Return](https://github.com/Wolfsblvt/smolpng/issues/36#issuecomment-6348058355) preserves exact proposal, counterfactual, native observation, integration and carrier-retirement evidence.

All eleven [post-main Verify jobs](https://github.com/Wolfsblvt/smolpng/actions/runs/39829300910) passed at the complete candidate SHA. Those jobs cover ordinary tests, Linux race detection, vet, module/corpus checks, six target builds and two native shell compilation targets. They do not execute the private foreign-platform archives. The proposal's original Windows job was cancelled after an unexplained overlong run; only that job was rerun at unchanged source and passed. The cancelled attempt remains incomplete evidence.

## Immutable artifacts

Files remain local under ignored `artifacts/releases/private-candidate-dev14-8252709/`. These are retained coordinates, not download URLs. Platform filenames begin `smolpng-0.3.0-dev.14-`.

| Artifact suffix | Bytes | SHA-256 |
| --- | ---: | --- |
| `windows-amd64.zip` | 37,367,740 | `1a3498bbd6861bd06e201857dfc4d54d8c0a6f4969d508ad8f6445bbd1459c18` |
| `windows-arm64.zip` | 34,862,304 | `6c1a63e956772e8a95551855592cc3b98655ac9058cc6f4c91c159f266175412` |
| `linux-amd64.tar.gz` | 25,124,033 | `89d57840d6d1abfe23651ca5b621b61334db067e3602f433a9dd6c3bc8b0412e` |
| `linux-arm64.tar.gz` | 23,714,078 | `636b679501ec375d1b823d2991cdf32873f29839a8cef1ba6c7df47053272151` |
| `darwin-amd64.tar.gz` | 25,142,729 | `719a9d3d4aa1cbab6ca8f9042af6aab85baa9018c1ae74d3adce4510835ecb76` |
| `darwin-arm64.tar.gz` | 24,110,814 | `c49df4175e5e3068bc9eb54331eee41b463f37410a231256f0af531389a91fc9` |
| `source-recovery.zip` | 18,127,619 | `996b4ae9c5a18d3ca9e29a97970060c36ed71e3ce7c7b5bedd5edee542fde828` |

The separate `history.bundle` is 9,036,503 bytes, SHA-256 `0340161dff62e347793d5dfe4f5b627d15148efb52f4792b2c489e26425033d0`. The ZIP contains the committed source snapshot, that same complete-history bundle and recovery instructions. Separate clones from the standalone and ZIP-extracted bundles recovered exact `main`, passed `git fsck --full` and remained clean. All **431 actual source files** in both restores matched the ZIP snapshot bytes, with Git conversion disabled during restoration. This establishes local source recovery, not disaster or dependency-cache qualification.

## Build, archive and native evidence

The clean primary checkout was at the selected SHA. The maintained builder ran with one Go compiler lane:

```text
go run -p=1 ./tools/release --version 0.3.0-dev.14 --out artifacts/releases/private-candidate-dev14-8252709
```

The owned process started at 2026-10-22 18:18:47.6932498Z, ended at 18:23:28.2174547Z and exited 0. The actual Go 1.27.2 Windows amd64 executable has SHA-256 `b8cd508a0a7aac883cd750e9e75eb306d219e726bdf1cde29009490729437fbf`. The invocation used `GOMAXPROCS=1`, `CGO_ENABLED=0`, `GOWORK=off`, `GOENV=off`, `GOTOOLCHAIN=local`, `GOPROXY=off` and `GOSUMDB=off`. No production credential or signature was used.

Independent inspection matched every archive/member hash with its manifest, staged file and common committed source payload. Windows archives each contain 201 files; Linux/macOS archives each contain 200. Binary headers, architecture, Windows console/GUI subsystem and Unix executable permissions matched. Native Windows version-resource reads verified all six x64/ARM64 executables against the source/version, package hashes and standalone identity; no embedded Authenticode certificate was present. ARM64 metadata reads are not ARM64 execution.

The extracted Windows amd64 console app has SHA-256 `eba0e6e1db780234441e9a6373488483e07002438846744c1d5cb63710e49aae`; its windowless launcher has SHA-256 `d65b8f8df7b40042741e90f15a9ac9cb00642a1422c807f961c3dfbc56676115`. App, packaged qualifier and expected source all bind the selected complete SHA. On Windows 11 Home/Core 25H2 amd64, kernel 10.0.26200 revision 9711, **all 79 PNG/JPEG checks passed**: Auto/lossless, original preservation, exact targets and nearest, binary stdout, partial/mixed batches, orientation, ICC/component models and supported resizing/refusal.

The qualifier process ran from 21:51:14.7412619Z to 21:51:33.8956279Z and exited 0. Its report SHA-256 is `f7c6715d25ea234b3ccc558ba332fbeb704045a44174a52acf4a34578b98dd76`. It used a new portable home, no project lookup and separately installed pngquant 2.17.0, OxiPNG 10.2.1 and jpegoptim 1.5.6. Their paths, version reports and binary hashes are retained in the original report. No clipboard, graphical consumer or remote-provider operation occurred.

## Prepared owner-use boundary

The existing preparation helpers were rebound to dev.14's actual app/launcher hashes, source, manifest and fresh output roots. Their earlier versions were preserved. Direct hidden child processes passed nine desktop preparation calls, eight TinyPNG dry-run/refusal calls and eight filesystem-only menu calls. Quiet single/mixed jobs produced results, last-result recovery worked, originals were preserved, integration previews found no occupied destinations, and menu settings installed/removed only under redirected local application data. TinyPNG preparation supplied an invalid empty child credential so it could not fall back to the native keyring; no upload was attempted.

The actual live-helper `check` entry also passed: its synthetic-marker preflight did not open/read the clipboard, and cleanup preview found no installed entries. These observations do not install a shortcut, register a native package, arm desktop capture or authorize another physical test window. Local reports remain under `artifacts/validation/desktop-dev14-20261022/`, `remote-preparation-dev14-8252709/` and `windows-menu-preparation-dev14-8252709/`.

## Retained evidence and continuing consequence

The [evidence payload](2026-10-22-private-candidate-dev14/report.json) joins **19 original UTF-8 inputs** with their original SHA-256 values and two explicitly derived preflight observations. Its size is **311,197 bytes**, SHA-256 `50d90b0fe21fe533a8f67dd09618425854b7fb549d557b983cdf704a5a7cc335`. Original preflight receipts also name private helper paths and remain local; their hashes and selected non-secret results are preserved without calling the derived fields unchanged captures. The wrapper is an authored join, not an original capture.

The immutable [dev.13 composition](2026-10-22-private-candidate-dev13.md) and [lifecycle](2026-10-22-windows-dev13-lifecycle.md) retain their own observations. They have not been relabelled as dev.14 execution. No foreign archive runtime, production signing/feed, package submission, public distribution or complete joined Windows journey is established here.

Katja retains candidate qualification and the next fully prepared, newly authorized physical Windows encounter in Issue #36. Nyxara retains product/review/release judgment. Wolf's website, documentation voice and public identity workshop remains selected before accepted 1.0 and explicit publication. The full Vision remains the destination; the Explorer repair and fresh candidate advance its daily-use path without closing its unobserved consumer consequences.
