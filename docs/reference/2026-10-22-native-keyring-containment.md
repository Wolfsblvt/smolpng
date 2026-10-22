# Native credential containment throughout the lifecycle

## Meaning

The amended native credential journey passed on six disposable Windows, Linux and macOS hosts. It now scans the isolated application home after every completed CLI invocation, including while a generated credential is stored, rather than observing only the final post-logout state. This record preserves fresh source-bound outcomes for that coverage correction. It changes no production credential behavior and does not qualify a personal store, every locked session, a live TinyPNG account or the complete private candidate.

## Source and review disposition

Source: `5ffd90444a26dbc426ceec71a2d4043b759f27a3`. All six jobs passed in [native run 39709619345](https://github.com/Wolfsblvt/smolpng/actions/runs/39709619345), and all eleven jobs passed in [Verify 39709619231](https://github.com/Wolfsblvt/smolpng/actions/runs/39709619231). Forgehand passed; the size check was skipped.

[Nyxara's independent review](https://github.com/Wolfsblvt/smolpng/issues/66#issuecomment-6330524844) identified that a temporary plaintext file deleted by logout could evade the previous final scan. Katja accepted the finding and reused the existing raw/base64 scanner after each CLI invocation, retaining stdout/stderr withholding and the explicit final scan. No adapter, precedence, upload permission, workflow or dependency changed. The [earlier native record](2026-10-22-native-keyring.md) and its original evidence remain historical results with that post-logout-only filesystem boundary.

## Fresh observed outcomes

Each host passed the eight named lifecycle steps: missing status, stdin login/native readback, cross-process resolution, replacement, environment precedence, invalid stdin preservation, logout/idempotent removal and final home containment. The added scans run throughout those steps. Both Linux hosts also passed the unavailable-store environment route: **50 successful leaf checks** overall, excluding parent/package bookkeeping.

| Platform | Observed host and native store |
| --- | --- |
| Linux amd64 and arm64 | Ubuntu 24.04.5 LTS, Secret Service |
| Windows amd64 | Windows Server 2025 Datacenter, build 26100, Credential Manager |
| Windows arm64 | Windows 11 Enterprise, build 26200, Credential Manager |
| macOS amd64 | macOS 26.6.1, build 25G76, Keychain |
| macOS arm64 | macOS 26.6.2, build 25G83, Keychain |

All hosts used Go 1.26.6 with `CGO_ENABLED=0`. Login remained local-only and reported no upload-policy change. Existing credentials refuse the journey, and cleanup removes only its owned generated fixture. Ordinary root verification excludes the native build tag; no workstation-store operation, real credential or provider quota was used.

## Retained evidence and limits

The [payload](2026-10-22-native-keyring-containment/report.json) retains all **68 original UTF-8 inputs** from the six successful artifacts, with independently computed input hashes, source/application/test identities, OS/tool facts and complete test JSONL. Payload size: **99,563 bytes**; SHA-256: `57250cd06c7bd58e323968b01cd14db5742dba0c25496ed08de0ab5d5e0c8c68`. The authored JSON joins runner inputs; it is not itself an original runner artifact. Credential-store files, fixture values and application binaries are excluded.

These are unlocked hosted-session observations. Native runner success does not establish every interactive/locked keychain state, masked interactive login, Wolf's credential store, live TinyPNG use, graphical behavior or [the joined candidate](https://github.com/Wolfsblvt/smolpng/issues/36).
