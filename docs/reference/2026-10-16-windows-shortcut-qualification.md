# Windows shortcut activation qualification, 2026-10-16

## Meaning

Native Windows amd64 activation of an application-created shortcut exposed a false failure after successful compression: the no-console executable tried to print its human report. Giving that entry point discard writers repaired the observed path. This record binds the before/after evidence to exact builds and preserves its limits: ShellExecute file-argument delivery was exercised; a physical Explorer drag, clipboard paste and rendered notifications were not.

## Native observations

The application created its own per-user Windows desktop `.lnk` through `integrate shortcut create`, using Auto, an isolated portable configuration and `--no-project`. The configuration selected real pngquant 2.17.0, OxiPNG 10.2.1 and jpegoptim 1.5.6, denied remote use, disabled clipboard copying and disabled ordinary notifications. Inputs were copies of the checked-in CC0 flat-icon PNG and progressive JPEG, renamed with spaces, an ampersand and Unicode.

| Invocation | Build | Observed result |
| --- | --- | --- |
| Console `drop --json` | `adabddefe241673ee041f01a9302335aba888cef` | Exit 0; both outputs and last-result state written |
| Native application-created shortcut, file arguments appended through Windows ShellExecute | Same build | Both outputs and successful last-result state written, but process exited 2 |
| Direct GUI-executable ShellExecute with `--quiet` | Same build | Exit 0; both outputs and last-result state written |
| Native application-created shortcut without `--quiet` | `e611fd3ee768d349e841f87e6c008a2e49a461eb` | Exit 0; both outputs, collision handling and public `last --json` recovery passed |

The PNG result was 97 bytes from 12,425 bytes; the already-small JPEG remained 152 bytes. All original copies matched their corpus source bytes. Repeating the original shortcut inputs after the repair created `.smol-2` outputs without replacing the first outputs. Both outputs matched their earlier counterparts byte for byte, and `last --json` reported both current files present. Reports recorded no remote use or warnings.

## Repair and build identity

Only the Windows background entry point changed its output environment to `io.Discard`. Successful compression no longer depends on terminal handles that a GUI launch may lack. Selected interactive mode retains its existing explicit console acquisition; the ordinary console executable retains its reporting behavior.

Both repaired Windows amd64 binaries were built with Go 1.27.2, CGO disabled, `-trimpath -buildvcs=false`, version `0.3.0-dev.1`, distribution `standalone` and embedded commit `e611fd3ee768d349e841f87e6c008a2e49a461eb`. Only the GUI executable used `-H=windowsgui`.

| Binary | PE subsystem | SHA-256 |
| --- | --- | --- |
| Repaired `smolpng.exe` | Windows console (3) | `7a6bb20dc11bd50145f7c5ab804d7b249ab5877e055136785fa30f3bfff43927` |
| Repaired `smolpngw.exe` | Windows GUI (2) | `4243a80b89e2ae2e2b35b88f28e8ee2434cafa4cf5c8cb01dcee7cb2dc1899c8` |
| Original `smolpngw.exe` | Windows GUI (2) | `f72a35aa679328ab28ba44229326863a3c187cf4ac8bd9f29858f8cc236fff3f` |

Focused `go test ./cmd/smolpngw ./internal/cli ./internal/integration` passed on Windows Go 1.27.2. The command package has no unit tests; the real GUI executable and native shortcut activation above exercise its changed boundary.

## Receipt and cleanup

Local ignored evidence lives under `artifacts/validation/drop-adabdde/` and `artifacts/validation/drop-e611fd3/`: exact activation scripts, native link properties, process/exit receipts, last-result snapshots, file hashes and a completed passing repaired report. Build logs and exit receipts live beside `artifacts/bin/e611fd3/`. These local artifacts are supporting evidence, not published downloads.

The application removed each test shortcut through its exact managed ID; the native path was then observed absent. No optimizer was bundled or installed into PATH. User images, ordinary application settings and clipboard contents were not used.

## Remaining consumer boundary

ShellExecute activation with appended file arguments proves native shortcut resolution, argument delivery, GUI-entry-point execution, output preservation and result recovery on this host. It does not prove physical pointer dragging in Explorer, visible absence of a console flash, rendered notifications, clipboard representations or paste into another application. PE subsystem inspection supports the no-console build claim without substituting for those visual observations. Interactive console behavior, other operating systems, publication and signed distribution remain separately qualified.
