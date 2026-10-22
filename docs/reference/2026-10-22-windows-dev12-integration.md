# Windows dev.12 integration lifecycle

## Meaning

The immutable Windows amd64 `0.3.0-dev.12` archive now has a native per-user launcher installation and removal observation. Five shortcuts installed, repeat installation stayed duplicate-safe, modified content survived removal and reinstall refusals, and all unchanged owned targets were removed. This closes the installer ownership portion of [candidate Work #36](https://github.com/Wolfsblvt/smolpng/issues/36), without claiming shortcut activation, physical drag/drop, clipboard use or notification rendering.

## Exact candidate and host

The application remains source `61b3ea59a3dbf21275d2bc8ba3d29fdae566776e`, separately identified in the [complete dev.12 artifact record](2026-10-22-private-candidate-dev12.md). Its extracted `smolpng.exe` SHA-256 is `d53cc691cdab5c83ec7c8e0c32b8c03ce1f4d589c70d711833cac0e79bd807c6`; adjacent `smolpngw.exe` is `e3c99ef7197dcb6a147b8a165d1b8e2cbd384798a53b7988bc417b7cfd100fb9`. Both actual files were hashed before the journey. Native `version --json` reported dev.12, the full source, `windows/amd64` and standalone distribution. No candidate file was rebuilt or patched.

The journey ran on Windows 11 Home 25H2 amd64, build `10.0.26200.9711`, from **2026-10-22 07:27:18.437Z to 07:28:15.498Z**. It used a new portable state home, `--no-project` and JSON output. Every child application invocation used a direct executable and argument array with Windows no-window controls. The candidate's native renderer invokes fixed profile-free, noninteractive PowerShell through its existing hidden process route. No launcher was activated, desktop input sent or clipboard accessed.

## Native lifecycle

The initial preview named five native destinations: interactive and Auto Start-menu entries, background and interactive Desktop entries, and Send To. Preflight verified that each destination was absent before installation. Only these generated test shortcuts were subsequently modified or removed.

| Command receipt | Exit | Observed consequence |
| --- | ---: | --- |
| Version | 0 | Exact private candidate identity |
| Install with `--desktop --send-to --apply` | 0 | Five real native `.lnk` files |
| Installed status | 0 | Five `current` entries; actual file hashes match recorded hashes; no pending recovery |
| Repeat install | 0 | Existing owned entries replaced through the ordinary installer |
| Reinstalled status | 0 | Exactly five current entries; no pending recovery |
| Status after generated modification | 0 | Interactive entry is `modified` |
| Remove modified interactive entry | 5 | Modified bytes retained unchanged |
| Reinstall over modified entry | 5 | Modified bytes retained unchanged |
| Remove after exact fixture restoration | 0 | All five owned targets removed |
| Final status | 0 | Empty entries, no pending recovery; every previewed path absent |

Before adding the generated modification marker, the original interactive shortcut was preserved separately. After both expected refusals, the helper restored those exact original bytes and verified their recorded hash before ordinary application removal. It did not force-delete a modified file, adopt unrelated content or clear another integration registry.

## Retained evidence and remainder

The original 47,600-byte UTF-8 command/ownership receipt is retained at ignored `artifacts/validation/candidate-integration-dev12-61b3ea5/integration.json`, SHA-256 `7ca0ce404c18ef0110bcfd40437aa1672d7eb39d4fbc42dde36f15facaecd1e7`. It includes actual arguments, stdout/stderr, timestamps, installed hashes, modified-content identity and final status. The generated original shortcut backup is retained alongside it. These are local evidence coordinates, not public download links; this document is an authored account rather than the unchanged raw capture.

Katja retains the joined Windows physical drag/drop, Send To selection, rendered notification, smart clipboard/file-picker paste and sensitive-clipboard journey in #36 for the next explicitly authorized desktop window. This installer result does not complete that journey, foreign-host archive runtime, the whole private candidate, final design, signing/feed operation or publication.
