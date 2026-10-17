# Windows locked-output qualification, 2026-10-17

## Meaning

The real Windows amd64 application preserved both the source and previous destination when native sharing locks prevented atomic replacement. Four PNG/JPEG overwrite/in-place cases refused with exit 5, omitted a committed output path and left no temporary output; identical retries succeeded after releasing the locks. This record also preserves an unavailable isolated-clipboard route, which supplied no clipboard-use evidence and did not touch the interactive clipboard. It does not qualify graphical interaction, other platforms or publication.

## Exact application and fixtures

The extracted local candidate reports version `0.3.0-dev.2`, standalone ownership, Windows amd64 and source commit `42176c30997e395287cb863f5b7a93a83f158876`. Its accepted source tree equals main `6e51028df288069668995a61c817af91a333cd2d`. The console executable SHA-256 is `a8023ddf70a6d079b9358ae0e21b5f1f197914a0b0cc5c7516001bed534b46d0`. No new application build or behavior change was needed for this qualification.

Inputs were disposable copies of the checked-in CC0 flat-icon PNG and landscape JPEG. An isolated user home selected separately installed OxiPNG 10.2.1 and jpegoptim 1.5.6. Project discovery, remote use, output copying and notifications were disabled. No personal image, credential, clipboard content or normal application setting was consumed.

## Sharing-lock and retry results

For each case, `CreateFileW` opened the intended destination with read access and `FILE_SHARE_READ`, allowing the application to inspect bytes while preventing replacement. The normal `compress` command used Auto with either explicit `--output PATH --overwrite` or explicit `--in-place`. The first invocation ran while the native handle remained open. The second used exactly the same arguments after `CloseHandle`.

| Case | Locked invocation | Unlocked retry | Output bytes |
| --- | --- | --- | ---: |
| PNG explicit overwrite | Exit 5; existing source/destination preserved | Exit 0; optimized | 12,425 → 97 |
| PNG explicit in-place | Exit 5; original preserved | Exit 0; optimized in-place | 12,425 → 97 |
| JPEG explicit overwrite | Exit 5; existing source/destination preserved | Exit 0; optimized | 399,226 → 276,012 |
| JPEG explicit in-place | Exit 5; original preserved | Exit 0; optimized in-place | 399,226 → 276,012 |

All four locked reports said `status=failed`, `error="cannot commit complete output"` and `exitCode=5`; none supplied an output path falsely identifying the older destination as newly committed. SHA-256 comparisons preserved every locked source/destination. No `.smolpng-*` temporary output remained. After unlock, actual file sizes matched each successful report, and non-in-place source hashes were unchanged. The PNG pipeline recorded OxiPNG 10.2.1; JPEG recorded jpegoptim 1.5.6. Every result reported no remote use.

The completed local receipt, arguments, stdout/stderr and fixture outputs remain under ignored `artifacts/validation/windows-locked-42176c3/`. The report has `complete=true`, `passed=true` and all four cases. This supports real Windows sharing-lock refusal and ordinary retry, not power-loss durability or protection from a malicious simultaneous directory writer.

## Isolated clipboard route was unavailable

A separate preflight attempted to create a new noninteractive window station with `CWF_CREATE_ONLY`, so it could not adopt an existing station. Microsoft documents that each [window station owns a clipboard](https://learn.microsoft.com/en-us/windows/win32/winstation/window-stations), and [CreateWindowStation](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-createwindowstationa) distinguishes new-only creation and named-station access.

The named-station attempt returned Windows error 5 (access denied). The system-assigned-name attempt returned error 183 (already exists); the new-only condition prevented adopting that existing station. Both stopped at station creation, before any `OpenClipboard`, format enumeration or clipboard read/write. Their local incomplete receipts remain under `artifacts/validation/windows-clipboard-42176c3/` and the corresponding `-system/` directory. No station was created, privilege elevated or interactive desktop/clipboard changed.

These results establish a limitation of this session's isolated test route. They do not show that smolpng clipboard support fails or succeeds. Native image/file-list/text publication, Windows persistence and real application paste remain unobserved; they require an authorized interactive specimen or another genuinely isolated host route. No fallback to reading or replacing the user's clipboard was selected.
