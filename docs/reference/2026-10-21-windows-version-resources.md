# Native Windows version resources, 21 October 2026

## Meaning

The local release builder now embeds exact version/source identity in all six Windows executables before receipts and archives are hashed. Native Windows APIs read the expected metadata on the hosted amd64 build host and Windows 11 amd64, with stage/ZIP/member/index digests joined. This qualifies resource construction and reading of x64/ARM64 files, not ARM64 execution, graphical use, signing or publication. `0.0.0-resource-test.1` is a fixture identity, not a selected public version.

## Exact source and native journey

Build source is `7da5d9fc1e4d7e3ca27b9b954dfda10ce702f286`, carried by [PR #60](https://github.com/Wolfsblvt/smolpng/pull/60) for [Issue #59](https://github.com/Wolfsblvt/smolpng/issues/59). [Native run 39681317779](https://github.com/Wolfsblvt/smolpng/actions/runs/39681317779) completed in about 93 seconds on Windows amd64, kernel observation `10.0.26100.0`, PowerShell 7.6.7, Go 1.26.6.

The existing builder compiled only `windows/amd64,windows/arm64` from clean source. The native reader checked all six files through `Diagnostics.FileVersionInfo`, machine types AMD64 `0x8664` / ARM64 `0xaa64`, GUI subsystem 2 for `smolpngw.exe` and console subsystem 3 otherwise, plus absence of an embedded certificate directory. Each stage digest matched its actual ZIP member and member receipt; each ZIP matched the build index. The amd64 console application ran and returned matching version, source and standalone ownership. The checkout stayed clean with no source-tree resource generation.

The unchanged downloaded ZIPs were then read through the **same** `tools/release/qualify-windows.ps1` entry on Windows 11 amd64, kernel observation `10.0.26200.0`, PowerShell 7.6.5. A bounded extractor first independently matched both archive digests and all six executable/member digests, extracting only those six files. The native report completed. No GUI, launcher, clipboard, installer or optimizer was invoked on the workstation.

The exact [hosted report](2026-10-21-windows-version-resources/hosted-report.json) has SHA-256 `31ae736941d1abf5aa8da5a8c5056f32e0440b239cc3f794981f6f0c54e7a2bd`; the [Windows 11 report](2026-10-21-windows-version-resources/windows-11-report.json) has `6ea1793c3ae2c70445c63eec987045e28ca1628c033a029be71f54fa4da60d1e`. These committed records outlive the 14-day artifact transport. The full fixture archives are in artifact `windows-identity-7da5d9fc1e4d7e3ca27b9b954dfda10ce702f286-39681317779`; source and the [Development entry](../DEVELOPMENT.md#windows-artifact-identity) reproduce the journey.

## Built executable identities

| Architecture / executable | Bytes | SHA-256 |
| --- | ---: | --- |
| amd64 / smolpng.exe | 20341248 | `85bd214553ea0dd552b35975b4fd5a83aa0b3c0eb4a30fda965cbd2c502aac3e` |
| amd64 / smolpngw.exe | 20347904 | `6a03cd1cb3976a581c40df04da9d837241e41f63e09a8b0f52363e0453f5451b` |
| amd64 / smolpng-qualify.exe | 7363584 | `0de45e8c5dec6201b0b832a87ba7f15d091b23ea3417b470f7616f521261d4ff` |
| arm64 / smolpng.exe | 18669056 | `cfdbddbd4dead96100dad374b5086c259a9784f917186da0392305e9d53e70b5` |
| arm64 / smolpngw.exe | 18673664 | `17973254eecb12b199d2d537068d1a3290ede7aca9cafee0b4ddccaa7627cef9` |
| arm64 / smolpng-qualify.exe | 6801408 | `707a2b6646aa211c36a7f41d3341779297bb6e3aeba6c3147e9996ee7ad565db` |

Fixture ZIP SHA-256: amd64 `9e3136a1b7664b1981da015ec85f77ee4b660fe03540a90e683f268297910ee1`; ARM64 `60095399a6275c26848ae39f59054d6ddd59ab06ef1d2482c0fe1912b84c4f11`.

All files returned ProductName `smolpng`, descriptive maker label `Wolfsblvt Works`, full FileVersion/ProductVersion `0.0.0-resource-test.1`, fixed numeric tuple `0.0.0.0`, the prerelease flag, their own original filename/description and the exact source/standalone comment. Maker text is not an authenticated publisher. Full strings distinguish prereleases sharing a numeric tuple; the application still selects updates through its semantic-version contract.

## Component evidence, failures and limits

All eleven [Verify jobs](https://github.com/Wolfsblvt/smolpng/actions/runs/39681317730) passed at the exact build source. Local Go 1.27.2 focused tests and vet passed for `internal/packaging` and `tools/release`. The Windows regression reuses its existing test PE and preserves input bytes, machine/subsystem/entry point, executable code and a non-version resource while adding version information. Its in-process duration was about 0.07 seconds; total command/compilation time was much longer. Two earlier local test attempts were canceled before verdict and are not passing evidence; the completed isolated run supplies the result.

The first pinned module fetch failed with a TLS handshake timeout; an unchanged retry succeeded. Build-only `github.com/tc-hib/winres v0.3.1` (0BSD, tag commit `bae858ad9c32978d161a2e37a46403b5a20fec76`) and `nfnt/resize` (ISC) have unchanged notices retained. Only release tooling imports the resource package; no application runtime dependency, native compiler or generated source-tree `.syso` file was added. This is not a live advisory audit or support for arbitrary PE inputs.

The initial Forgehand attempt returned GitHub HTTP 503 during its read-only plan before a writer step. Later automatic convergence passed at the same head. No writer control, source pin or sibling service was changed.

This does not mint the joined private candidate or choose `0.3.0` versus `1.0.0`. Dev.11 remains a held graphical behavior specimen; the next complete human/script/agent candidate needs current accepted source. Public release identity, SignPath admission, signing policy/workflow, certificate subject, package submission and actual graphical use retain their existing boundaries.
