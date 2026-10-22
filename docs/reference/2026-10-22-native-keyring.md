# Native credential-store qualification

## Meaning

The existing native credential adapter and actual provider commands passed generated-fixture lifecycles on six disposable Windows, Linux and macOS hosts, including both supported architectures. Linux also passed the unavailable-store environment route. This record preserves exact source, executable/test identities and non-secret outcomes, including the initial Linux fixture failure. It establishes these unlocked native sessions, not every keychain state, a live TinyPNG account, Wolf's credential store or the joined private candidate.

## Source and hosts

Implementation/evidence source: `7689dae9139e2ac3e7dccb5bebde56b75f8d932b`. All six jobs passed in [native run 39698838819](https://github.com/Wolfsblvt/smolpng/actions/runs/39698838819); all eleven jobs passed in [Verify 39698838808](https://github.com/Wolfsblvt/smolpng/actions/runs/39698838808). Forgehand passed. The existing size check was skipped, not green.

Each host built its native console CLI and tagged test executable with Go 1.26.6 and `CGO_ENABLED=0`. The CLI reported `0.3.0-dev`, `distribution=source` and the exact source above. Test and CLI hashes, runner-image facts, OS/kernel identity and complete Go test JSONL are retained in the [evidence payload](2026-10-22-native-keyring/report.json).

| Native platform | Observed host | CLI SHA-256 |
| --- | --- | --- |
| Linux amd64 | Ubuntu 24.04.5 LTS / Secret Service | `be75132998d27e408a2b5530b6fe851d7118e054361d2d193148090a52a0569a` |
| Linux arm64 | Ubuntu 24.04.5 LTS / Secret Service | `82eb06724d82a3b7a2d35bd16a756a2433e646377905de5ebfd2631fe80f0fc5` |
| Windows amd64 | Windows Server 2025 Datacenter, 10.0.26100 / Credential Manager | `e8ff3ffeb0b032088d0fecd2f4852af565bcd95efdb273444a768c87c7d10084` |
| Windows arm64 | Windows 11 Enterprise, 10.0.26200 / Credential Manager | `f455a5f56a6d90ddff5f478a1eeb302aab90c34051c10bf86186f1cb83a536b3` |
| macOS amd64 | macOS 26.6.1, 25G76 / Keychain | `9d4077be3150e9ad3327fd36bbb25cb58bfd6c1255b78477a71ad66b5d5a51b9` |
| macOS arm64 | macOS 26.6.2, 25G83 / Keychain | `4329e73c6ea881a6a47bf718c232043444bc3f7c8a79c49c0342cd8b7bfd6f22` |

## Exercised consumer behavior

Each native journey passed eight named steps: missing status; stdin login with native readback; cross-process status and resolver lookup; native replacement visible to the CLI; valid and invalid environment precedence without store mutation; invalid stdin preserving the stored fixture; logout removing only stored credentials with idempotent removal; and absence of the generated raw/base64 credential representations in application-home files. Every CLI invocation also checked captured stdout/stderr for those representations before releasing any output.

Login/status receipts reported `not-probed`, and login retained `remotePolicyChanged=false`. Only local provider credential commands were selected; no compression upload, provider quota or real account was used. Environment values were passed inside the child process and stdin carried login input; neither was a command-line key. Existing native credentials refuse the journey. Cleanup deletes only a value matching one of that journey's two generated fixtures.

The two additional Linux unavailable-store tests used a nonexistent bus socket below an isolated application-home path. Native access returned the existing actionable unavailable error, stdin login exited 5 without a plaintext fallback, and an explicit process environment credential still configured the provider. These checks do not claim that a locked interactive desktop behaves identically.

## Linux fixture failure and repair

The first [run 39697831823](https://github.com/Wolfsblvt/smolpng/actions/runs/39697831823) at `040986c` refused Linux before credential mutation. The [diagnostic run 39698347578](https://github.com/Wolfsblvt/smolpng/actions/runs/39698347578) at `a8a32d7` distinguished unavailable storage from an existing credential. Both Linux hosts exposed only the session collection; reading the login alias returned `/`.

The helper had supplied EOF to the foreground daemon's unlock input. Supplying the disposable fixture password through stdin created the login collection. In the corrected run, both native collection readbacks contain `/org/freedesktop/secrets/collection/login`, and the original complete journeys pass. The helper owns and stops its foreground daemon inside `dbus-run-session`; it never selects a workstation's bus or keyring. The application credential implementation was unchanged.

## Retained boundary

Payload SHA-256: `9b3c3d5c1a7d296d7a406baa036067c297687a4538079c0a7ecd63bae87cb515`. Each captured UTF-8 input carries its original SHA-256; the authored JSON joins those inputs without claiming to be an original runner artifact. It retains all six successful report sets and both failed Linux/diagnostic attempts beyond hosted artifact expiry. Fixture values, actual credential-store files, application binaries and personal credentials are not included.

Ordinary `go test ./...` excludes the `nativekeyring` build tag. The native journey additionally requires the explicit disposable-host sentinel and GitHub Actions execution marker, and refuses an existing credential before mutation. No local native credential-store operation was run for this contribution. Native keyring availability does not complete graphical use, locked-session behavior, a production TinyPNG account or [Issue #36's joined private candidate](https://github.com/Wolfsblvt/smolpng/issues/36).
