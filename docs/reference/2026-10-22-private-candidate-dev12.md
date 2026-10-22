# Private candidate 0.3.0-dev.12

## Meaning

Six private platform archives and an independently restored source-history ZIP bind `0.3.0-dev.12` to application source `61b3ea59a3dbf21275d2bc8ba3d29fdae566776e`. The Windows amd64 archive passed 79 native PNG/JPEG checks, dependency/Skill lifecycles, update/rollback and interrupted recovery, file-lock/deadline cases, active optimizer cancellation and three terminal flows. This record retains the original non-secret reports and the incomplete attempts that preceded them. It does not complete [Issue #36's joined graphical journey](https://github.com/Wolfsblvt/smolpng/issues/36), qualify foreign archive runtime or publish the candidate.

## Exact artifacts

Files remain under ignored `artifacts/candidates/0.3.0-dev.12-61b3ea5/` in the primary checkout. These are retained local coordinates, not download URLs. `build-index.json`, `SHA256SUMS` and the separate `SOURCE-SHA256SUMS` identify:

| Artifact | Bytes | SHA-256 |
| --- | ---: | --- |
| `smolpng-0.3.0-dev.12-windows-amd64.zip` | 31,967,410 | `99750ce9c86d044a6ebde1f2f9c2c47b5e9df2941bdf8ed3cfe4f59a19121461` |
| `smolpng-0.3.0-dev.12-windows-arm64.zip` | 29,385,658 | `f13ad3fe67fd4d39a82abed5cca3cf72f113024ac7370798397534e74418625b` |
| `smolpng-0.3.0-dev.12-linux-amd64.tar.gz` | 19,939,749 | `a6c1dfc9c5c07d8fd45265feca628cbf399e6e8ebb622ad4797085df3afc7678` |
| `smolpng-0.3.0-dev.12-linux-arm64.tar.gz` | 18,532,582 | `01f269e2b5ff6345f91d977f2de5ec7dde891d218d819d92dd91edd6729bfb39` |
| `smolpng-0.3.0-dev.12-darwin-amd64.tar.gz` | 19,958,337 | `351e06bf0831d34f5fe8853b75a9298cf69c41e93b5bf8c2cc9036fa49bc5aed` |
| `smolpng-0.3.0-dev.12-darwin-arm64.tar.gz` | 18,924,134 | `46167b6863fc8a77997d3efa5d187fc7f615cf26878a66c94e7f6e8c61d7ff2d` |
| `smolpng-0.3.0-dev.12-source-recovery.zip` | 12,467,283 | `cdf5b796911b3f7b86f6e8a31ef79620930239114c31ba39af40df7b0f8c29cf` |

Independent inspection verified every archive checksum, size, member digest, manifest identity, safe unique member path and target executable header. Each Windows ZIP has 183 members and 182 hashed payloads; each Unix archive has 182 members and 181 hashed payloads. All six carry the exact 60-file source notice set with matching bytes. Stage/archive executable hashes match. Native Windows metadata reads additionally verified all six x64/ARM64 executable version resources; reading ARM64 metadata is not ARM64 execution.

The extracted Windows amd64 CLI SHA-256 is `d53cc691cdab5c83ec7c8e0c32b8c03ce1f4d589c70d711833cac0e79bd807c6`. Its launcher and qualifier are `e3c99ef7197dcb6a147b8a165d1b8e2cbd384798a53b7988bc417b7cfd100fb9` and `726fcb5a6413faf00caedf34b513caa10b25d06cc8ccb64fc6cbeb373bd362ea`. Actual CLI readback reports dev.12, the selected full source, `windows/amd64` and standalone distribution.

The source ZIP has 424 members and carries the selected source's ancestral history plus a Git bundle. Independent ZIP and bundle restorations recovered the exact clean commit and passed object verification. This establishes local source recovery, not disaster qualification or recovery of later main commits.

## Build and source evidence

The clean primary checkout stayed frozen at the selected commit throughout the build and archive inspection. The maintained command was `go run ./tools/release --version 0.3.0-dev.12 --out artifacts/candidates/0.3.0-dev.12-61b3ea5`. Go 1.27.2 windows/amd64 has SHA-256 `b8cd508a0a7aac883cd750e9e75eb306d219e726bdf1cde29009490729437fbf`; the selected release builder is `5059efb57a86974c44cd4c48d82f32800c5982db59f118886e0e43a84ae4625b`, and the qualified quiet wrapper is `e8712f2ca46621220f8cb7785b7f53425a40b96f11ce5bacfd656b06812438d9`.

The owned build ran from 2026-10-22 01:01:02.559Z to 02:42:58.056Z and exited 0. Controls were `GOFLAGS=-p=1`, `GOMAXPROCS=1`, the soft `GOMEMLIMIT=512MiB` target and BelowNormal process/child priority. This is observed build provenance, not an independent rebuild claim. All eleven [Verify jobs at the exact candidate source](https://github.com/Wolfsblvt/smolpng/actions/runs/39702731167) passed; hosted source checks do not execute these private archives.

The later [release-input repair](https://github.com/Wolfsblvt/smolpng/pull/69) is a separate source contribution. The selected dev.12 input tree was enumerated as ordinary files/directories, and the checkout remained frozen. Later builder/source changes neither alter these artifacts nor silently mint another candidate.

## Windows consumer observations

The extracted amd64 candidate ran on Windows 11 Home 25H2 amd64, build `10.0.26200.9711`, with invocation-selected pngquant 2.17.0, OxiPNG 10.2.1 and jpegoptim 1.5.6. Their SHA-256 identities were checked around native qualification and after the terminal journeys:

| Tool | SHA-256 |
| --- | --- |
| pngquant | `7aadad6a50aa5c40cfcb20f9478c92f6030360a3424bc262ce383dc1e97fb86a` |
| OxiPNG | `35ae3980ab831af64f4aecc98f69b81bfa146fa750a61d17ebbb12520128cbc8` |
| jpegoptim | `0a79681b56961ec2c5f45c4f9dff01a64e7274b4f3250c22cde5f09c0bb457f3` |

All 79 native PNG/JPEG checks passed, including original preservation, targets, partial batches, binary output, profile/component controls and reason-specific geometry refusal. The isolated 26-command managed dependency/Skill lifecycle also passed. Every installed agents/Claude/custom Skill resource and the uploadable ZIP matched the candidate's bundled bytes; modified Skill content survived sync and protected removal/forget behavior.

The ordinary signed-fixture journey passed 11 commands. It refused untrusted and tampered inputs, applied the admitted update, observed locked rollback requiring recovery, and restored the original files after releasing the lock. The separate 13-command interruption journey observed committed new member bytes while other changed members remained old, terminated only the owned stage worker with exit 77, recovered every one of the 64 original files, then completed ordinary apply and rollback. Configuration and the edited project Skill survived throughout. Its target is deliberately the existing test-only stable `0.3.0` fixture at `b0ed6391c4d8c80fcb9d78c611aa42d8debb798f`, not the dev.12 source. Ephemeral test signing keys were not exported; no production trust or feed was used.

All four Win32 overwrite/in-place lock refusal/retry cases and both early deadline/refusal/retry cases passed. In an owned hidden ConPTY, Ctrl+C during an observed real OxiPNG compression child produced native and structured exit 130, preserved the original/configuration, committed no output and removed its job directory. The exact application, optimizer and shell processes were observed exited.

Three hidden ConPTY flows passed: ordinary keyboard navigation at 80 columns, accessible text choices at 80 columns and the keyboard flow at 40 columns. Each crossed first-run Continue, PNG/JPEG Auto and Done, retained two successful outputs and last-result/setup state, and preserved originals/configuration with empty job caches. These are functional terminal observations, not graphical layout/focus or final design acceptance.

## Incomplete attempts and recovery

The reused archive inspector initially expected 58 notices. The actual source and candidate contain 60; the corrected helper compared the complete source notice set and hashes instead of a stale count. No package bytes changed.

The first native run stopped at dependency preflight, and the first managed lifecycle stopped at the OxiPNG version probe, under BelowNormal helper priority. A direct pngquant comparison took 1,181 ms at Normal priority and 16,414 ms at BelowNormal; the identical application dependency read then passed at Normal priority for all three tools. Complete successful qualification used Normal priority, one worker and no-window controls while the bulk build remained BelowNormal. This is bounded host evidence; no application probe deadline changed.

One final JPEG in-place retry exceeded the helper's 90-second limit after three completed lock cases. The original bytes and absence of the affected child processes were reconciled before resuming only that retry and the untouched deadline cases with a five-minute invocation budget. The resumed retry passed in 36.069 seconds. The earlier incomplete lock report and failed native/lifecycle reports remain retained; they are not counted as passing runs.

## Retained reports and remainder

The [authored evidence payload](2026-10-22-private-candidate-dev12/report.json) retains 28 original non-secret UTF-8 inputs, each with its original SHA-256, including build/archive/recovery reports, native/lifecycle results, both updater journeys, cancellation and terminal transcripts, completed/resumed file-safety evidence and the earlier attempts. Payload size is 685,247 bytes; SHA-256 is `0a6b9559999e62dcc5fef2bf6f19dd3a2c0626bbb34d5228acdd7b05cf112e16`. It is a join of original inputs, not itself an original capture. Binary artifacts remain private local files. The [dev.11 record](2026-10-20-private-candidate-dev11.md) retains its separate identities and observations.

Katja retains the current candidate's integration-installer/removal follow-through and physical drag/drop, Send To, rendered notifications, smart clipboard/file-picker paste and sensitive-clipboard journey in Issue #36. The physical journey awaits another authorized desktop window. No launcher was activated or desktop/clipboard controlled in this cut. Foreign-host archive runtime, public package/signing/feed operation, final identity/design acceptance and publication remain separate. This qualification tranche does not reduce the complete selected product or declare the candidate accepted.
