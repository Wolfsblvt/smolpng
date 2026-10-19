# Private candidate 0.3.0-dev.9

## Meaning

Source `b0ed6391c4d8c80fcb9d78c611aa42d8debb798f` produced six private platform archives and a restored source-history ZIP. Its extracted Windows amd64 application passed 63 native compression/provenance checks, 26 dependency/Skill lifecycle checks and a 13-command signed-fixture update journey, including recovery from actual interrupted replacement and a locked rollback. This record binds those observations to their artifacts and preserves an earlier failed observer attempt. It does not accept the complete candidate: joined desktop use and other remaining evidence stay open in [Issue #36](https://github.com/Wolfsblvt/smolpng/issues/36).

## Exact source and build

The clean source includes the [PNG clipboard repair](https://github.com/Wolfsblvt/smolpng/pull/40) and [automation-guidance correction](https://github.com/Wolfsblvt/smolpng/pull/41). The maintained `tools/release` builder used Windows Go 1.27.2, executable SHA-256 `b8cd508a0a7aac883cd750e9e75eb306d219e726bdf1cde29009490729437fbf`, explicit version `0.3.0-dev.9` and standalone ownership. The build completed from 19:03:41Z to 19:13:31Z on 19 October.

Wolf selected one low-priority build worker while using the PC. `GOMAXPROCS=1` and a `512MiB` Go soft memory target were inherited; the parent and sampled descendants reported Windows Below Normal scheduling priority. The builder clears ambient `GOFLAGS` for its child builds; Go 1.27.2 derives default package parallelism from the preserved `GOMAXPROCS=1`. This is a scheduling/soft-memory observation, not a hard CPU or RSS quota. Build, archive/source inspection and native qualification ran sequentially. No desktop or clipboard input occurred in these runs.

The existing module cache was reused. All eleven [Verify jobs on the exact source](https://github.com/Wolfsblvt/smolpng/actions/runs/39307361743) passed: module verification, ordinary tests, vet and corpus checks on Windows/Linux/macOS; six target builds; and Windows SDK shell compilation, including the x64 protocol tests. Native race checking was not rerun for this cut. Compilation and CI do not establish foreign-host runtime or graphical integration.

## Retained private artifacts

The seven archives are under ignored `artifacts/candidates/0.3.0-dev.9-b0ed639/`. They are local evidence, not public downloads. `build-index.json` and `SHA256SUMS` bind the six platform packages; `SOURCE-SHA256SUMS` binds source recovery separately. The Skill ZIP is under `artifacts/validation/candidate-lifecycle-dev9-b0ed639/`.

| Artifact | Bytes | SHA-256 |
| --- | ---: | --- |
| `smolpng-0.3.0-dev.9-windows-amd64.zip` | 31090160 | `b01820a974f2e13a22d297e95faf5301581678c6ab97ad6eca18444957cf2868` |
| `smolpng-0.3.0-dev.9-windows-arm64.zip` | 28522497 | `553d3d77078899d76b8e4d77e5edc385941503ab6b546cab5a5466b8cac87554` |
| `smolpng-0.3.0-dev.9-linux-amd64.tar.gz` | 19064243 | `99b9b094f657be3ca0e45cfbcaae2ccfadbbee82e6ebd869b94211994ce0082c` |
| `smolpng-0.3.0-dev.9-linux-arm64.tar.gz` | 17682533 | `2b30ed388d54978e3fe6a14e8f83b5eba85ba4db28fe73fab6e8f3a6b2fda9f5` |
| `smolpng-0.3.0-dev.9-darwin-amd64.tar.gz` | 19094395 | `d23b67c69ed0cf27fec9b2bd52dcff9fb56c77a332a395acdbe4335691ecb94d` |
| `smolpng-0.3.0-dev.9-darwin-arm64.tar.gz` | 18059908 | `29fc498f6ce55da9b660cb0887d1f6cf1cafd7bc47dce7d8181d0ecccc0f251a` |
| `smolpng-0.3.0-dev.9-source-recovery.zip` | 9669821 | `2234ef8b8099e8bcdda78da3544542bd00a90cc4b4e16c5937f742cf8bbdd899` |
| `smolpng-skill.zip` | 6832 | `0bf6c51847499f34ffbdea487fbfd3d6456113a9e3b14c45a971c8d34d5e7c82` |

Independent ZIP/tar readers checked every manifest member/hash, duplicate and unsafe paths, all 58 dependency notices, and the expected PE/ELF/Mach-O architecture headers. Windows has 155 members / 154 hashed payload files; console and qualification binaries use the console subsystem and `smolpngw.exe` uses the GUI subsystem. External optimizers, personal settings and private agent material are not bundled. The extracted application reports the exact version, commit and standalone identity.

The recovery ZIP contains only selected `main` history, a clean repository and a second Git bundle. ZIP integrity, extracted-repository restore and independent bundle clone all recovered the exact clean commit and passed full Git object verification. This establishes isolated local restoration, not disaster recovery. Earlier candidates were not patched or relabelled.

## Native Windows and managed lifecycle

The host was Windows amd64, Core/Home 25H2, kernel/build `10.0.26200.9711`; the updater preflight also recorded boot time `2026-10-17T19:10:33.5000000Z`. Candidate application SHA-256 is `f44cb8399f42af212ed5f3137fd6d535b3d77ecdfb10cb1e5846a309e9e057bc`; GUI launcher is `2ad4c168bf8fe63acde1d6e6c185b03b953e2d2b25036a583a6aa59aaaefe909`; same-source packaged qualifier is `9718b5574bad1193507f0f357d433874a65538d4005bb6e1572bde97533b1afd`.

All 63 native checks passed, including optimizer identity before and after the corpus, PNG/JPEG Auto/lossless/target behavior, target miss, mixed/partial batches and clean binary output. Optimizer executable hashes were:

| Tool | SHA-256 |
| --- | --- |
| pngquant | `7aadad6a50aa5c40cfcb20f9478c92f6030360a3424bc262ce383dc1e97fb86a` |
| OxiPNG | `35ae3980ab831af64f4aecc98f69b81bfa146fa750a61d17ebbb12520128cbc8` |
| jpegoptim | `0a79681b56961ec2c5f45c4f9dff01a64e7274b4f3250c22cde5f09c0bb457f3` |

The 26-command lifecycle used an isolated profile/home, local-only processing, disabled copying/notifications and one configured image job. It previewed, installed, reused and selected managed OxiPNG/jpegoptim with retained licenses and no PATH change; compressed owned PNG/JPEG copies; installed standard/custom Skills; preserved an edited Skill through sync and removal refusal; forgot that managed record without deleting edited bytes; removed untouched installations; produced the Skill ZIP; and removed the managed optimizers while preserving config and image originals. No real user profile or installed agent client was changed. Filesystem Skill installation is not client reload evidence.

## Update, interruption and rollback

A separate same-source Windows-only fixture used test identity `0.3.0`, not public release naming. Its narrow update archive contains 64 admitted members, SHA-256 `427db73c1027db48f52604171f67a342d34b254148317860dcc14f9f41764cdf`, size 28232635 bytes. The fixture's application hash is `b3a9ed567914529ff4c637831da64122ac7da1abf28da2eeec016351bdb8b07d`. An ephemeral Ed25519 key stayed in process memory; only public trust and signed envelopes were retained. No production key or feed was used.

The first observer repeatedly read the live journal. The worker exited at `staged` with `smolpng: publish output: Access is denied.`; all 64 installed members remained original. Ordinary `update recover --apply` reached `rolled-back`, with config and edited Skill intact. The incomplete six-command prefix and its reconciliation remain separate evidence. Repeated journal opens may have contended with atomic publication; this run does not establish that as the sole cause.

The second observer watched the known member's committed timestamp without opening the journal while its writer ran. It bound a process handle to this stage's verified original application, observed replacement, terminated only that owned worker with test exit code 77, and then read the settled journal and actual bytes. The journal was `applying`; `smolpng-qualify.exe` had the admitted new hash while main and GUI retained their old hashes. Every member had either its old or signed-new hash.

Public recovery restored all 64 original members and preserved config/edited Skill. A subsequent ordinary full apply, deliberately locked rollback, unlocked retry and restored dev.9 identity also passed: 13 command cases, `complete`, `passed` and `interruptionPassed` all true. Final phase was `rolled-back`; config before/final SHA-256 was `a539fdd2da73a0510faf3ca25943aefa67b3d4d261077859bc4a86ff434711f2`. This qualifies interruption after committed replacement, not power loss, torn storage or graceful Ctrl+C. No application source was patched to create the interruption window.

## Evidence locations and limits

All paths below are ignored local evidence under `artifacts/validation/`:

| Record | Path | SHA-256 |
| --- | --- | --- |
| Archive/member/header inspection | `candidate-dev9-b0ed639/archive-set.json` | `717b92eb5a848e25a60fcf5fc5644d9b76b32b57356747d336add464c1f116b3` |
| Native 63 checks | `candidate-native-dev9-b0ed639/report.json` | `959d4d07206ccdb5eeee7770c2f4daa0f32b0556eebbcd7a917972f1183d2812` |
| Dependency/Skill lifecycle | `candidate-lifecycle-dev9-b0ed639/lifecycle.json` | `cac46f57d29d71a02fba282c33e7f6400c6f4b9fa3f52d4968d42219724e49bd` |
| First observer recovery | `candidate-update-midapply-dev9-b0ed639/observer-reconciliation.json` | `5a2316b1e191168fe684b9f12ba02053d5e84013c46ee440d37f755d7c84bd10` |
| Mid-apply/full-cycle recovery | `candidate-update-midapply-dev9-b0ed639-metadata/update-journey.json` | `2aeeffe5484770df0e4088daccaabd0e6e44f13c33e9d0de2ee20233c30e9e97` |

The [earlier Paint observation](2026-10-19-windows-clipboard-alpha.md) proves the clipboard repair at its exact C1 code/binary; earlier dev.7 PTY, shortcut and recovery observations retain their own source identities. They are not fresh dev.9 joined-desktop observations. File-list/text consumers, physical Explorer dragging, notifications, sensitive-clipboard refusal and remaining joined native use are still open in #36. Linux package-file evidence remains the [separate earlier package record](2026-10-17-linux-package-artifacts.md); no Linux, macOS or Windows arm64 runtime is claimed here.

This documentation may land after its observed application source without requiring a replacement build: the candidate remains bound to `b0ed6391`, and the record changes documentation only. The complete 1.0 destination, Wolf's design workshop and retained signing/feed/publication decisions are unchanged. No public release, package registration, hosted sibling, deployment or spend occurred.
