# Native Windows update recovery, 17 October 2026

## Meaning

Source `3453a38e470d75d5d3ac78102c86eac708c3e32d` passed the real Windows standalone update and explicit rollback journey in a disposable installation. A deliberately held file lock produced an observed `recovery-required` error without losing the installed version; a normal retry then restored both original executable hashes and reached `rolled-back`. This record preserves that recovery evidence and its limits, separately from production signing, publication or a real-user installation.

## Original failure and correction

The earlier native experiment at `30104c1623ec81a4ddbc1e310bbbd436d5eda5f9` successfully applied a signed local update, but `update recover --apply` from the installed executable failed with `publish output: Access is denied.` Invoking that same rollback from the retained old helper succeeded. [Issue #15](https://github.com/Wolfsblvt/smolpng/issues/15) records the discriminating observation: Windows still held the executing destination image open.

The corrected command schedules a verified copy of its invoking application in a fresh worker directory, then exits. The helper waits for that process to finish before the guarded restore. The response says `scheduled`, returns the helper PID/log/journal paths and explicitly says `completed: false`. The original update helper and its evidence remain intact. Journal outcomes distinguish `rolling-back`, successful `rolled-back`, and failed `recovery-required` with an error.

## Exact native specimen

Windows 11 amd64, Go 1.27.2, CGO disabled, `-trimpath -buildvcs=false`. Both fixture versions were built from the same source commit above: standalone 1.0.0 and 1.0.1, each with `smolpng.exe` and the windowless `smolpngw.exe`. The local ZIP contained exactly those new executables and the repository license. An ephemeral Ed25519 signing key existed only in the qualification process memory; the retained trust file contains its public key. No production key, release feed, publication, administrator privilege, real user installation or optimizer installation was used.

| Executable | Old 1.0.0 SHA-256 | New 1.0.1 SHA-256 |
| --- | --- | --- |
| `smolpng.exe` | `9e36228001b8a84c1e81965c19e0a7e73a5646ffb861bf9a95e7d08a3fe8c32d` | `efda478c61335473988935468b720c9583b9eb583ebfba15fcbe285c9dcf834e` |
| `smolpngw.exe` | `88dfdc1f13d9ea4995f9abefd1745423cd95bbc78cf9a4a8c499dd8d4734d644` | `09be283cd1e685f56041eb4acb09b575d918f37ddbea99c897defc48a0ffca89` |

The installation and app home were created under a new temporary fixture root outside any Git checkout. Config, preset, optimizer-marker and unrelated-file sentinels were recorded before the operation. The Skill registry path was explicitly isolated and empty; no real managed/custom Skill installation was exercised.

## Observed sequence

1. The actual old executable reported version 1.0.0. `update apply` verified the local signed manifest/archive and returned a scheduling receipt. Its original helper was observed to exit; the journal reached `complete`, the actual installed executable reported 1.0.1, and both new executable hashes matched the signed members.
2. A Win32 read handle on the installed main executable permitted execution/read while denying replacement. The actual installed new executable's ordinary `update recover --apply` returned a scheduling receipt. Its recovery helper exited, the journal became `recovery-required`, and the recorded error was `publish output: Access is denied.` All installed archive-member hashes still matched the new version; no original was pre-deleted.
3. After releasing only that disposable handle, the same ordinary recovery command was invoked again. It returned a new helper/log path in a distinct recovery directory. That helper exited, the journal reached `rolled-back`, both executable hashes matched the original 1.0.0 hashes above, and the actual installed executable reported 1.0.0.
4. Config/preset and unrelated-file sentinels remained byte-exact across all three operations. The license remained unchanged. Each command's stdout/stderr was captured and scheduling data checkpointed before waiting for the original helper; no update was relaunched because an observation timed out.

The complete secret-free receipt and captured outputs are retained locally under `artifacts/validation/windows-update-3453a38/`. Its three cases (`native-apply`, `native-locked-rollback`, `native-explicit-rollback`) all passed. The observer reads the journal after its original worker exits and tolerates transient sharing/rename visibility, incorporating the earlier observer correction without repeating the earlier effect.

## Supporting evidence and limits

The updates suite covers preparation from verified old/new invokers, separate helper copies, wrong trust, modified installed files/backups, untrusted invokers and failed-recovery journal standing. Focused CLI update tests and affected-package vet passed; all eleven hosted Verify jobs passed on this exact source commit in [run 38978287054](https://github.com/Wolfsblvt/smolpng/actions/runs/38978287054). Hosted builds are not native desktop or publication evidence.

This qualifies the ordinary executable-lock repair and one real locked-file/retry boundary. It does not establish production signing/feed operation, Authenticode, installer/MSIX behavior, other native hosts, a power-loss guarantee or real installed-Skill synchronization. Helper launches use the existing hidden/no-console route and do not kill unrelated processes or elevate; no claim of instrumented foreground/window observation is made.
