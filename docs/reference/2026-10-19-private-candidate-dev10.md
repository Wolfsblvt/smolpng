# Private candidate 0.3.0-dev.10

## Meaning

Source `a9b8a48523ce8ef5d02a91f76dbde285764e8454` produced six private platform archives and an independently restored source-history ZIP, incorporating the corrected macOS jpegoptim mapping. All archive inspections and 26 dependency/Skill lifecycle checks passed. Two full Windows native attempts reached 60/63 and 62/63, with intermittent optimizer discovery still unresolved. This record binds the exact artifacts, successes, failures and focused reproductions; it does not accept the complete candidate. [Issue #36](https://github.com/Wolfsblvt/smolpng/issues/36) retains that destination and remainder.

## Source, build and hosted evidence

The clean source includes [PR #45's catalog repair and native macOS qualification](https://github.com/Wolfsblvt/smolpng/pull/45). The unchanged maintained release builder used Go 1.27.2 on Windows, executable SHA-256 `b8cd508a0a7aac883cd750e9e75eb306d219e726bdf1cde29009490729437fbf`, explicit version `0.3.0-dev.10` and standalone ownership. All six archives completed between 21:07:41Z and 21:20:05Z on 19 October 2026.

Wolf selected one low-priority build worker while using the PC. The existing quiet helper changed only source/version/output constants; parent and sampled builder/compiler descendants retained BelowNormal priority. `GOMAXPROCS=1` and a `512MiB` Go soft memory target were inherited. The builder clears ambient `GOFLAGS`; default package parallelism follows the preserved `GOMAXPROCS=1`. This is not a hard CPU/RSS quota. Build and inspections ran sequentially, without desktop or clipboard interaction.

All eleven [Verify jobs at the exact source](https://github.com/Wolfsblvt/smolpng/actions/runs/39324332093) passed, including Linux race detection, ordinary tests/vet/module/corpus checks, six target builds and Windows SDK shell compilation/protocol checks. The separate [manual native run](https://github.com/Wolfsblvt/smolpng/actions/runs/39324382944) passed 63 PNG/JPEG checks on Linux amd64 and both Mac architectures. Those hosted native binaries used Go 1.26.6 and source-build version `0.1.0-dev`; they are related exact-source evidence, not executions of these Go 1.27.2 development archives.

## Private artifacts and inspection

The seven archives are under ignored `artifacts/candidates/0.3.0-dev.10-a9b8a48/`. They are local evidence, not public downloads. `build-index.json` and `SHA256SUMS` bind the platform packages; `SOURCE-SHA256SUMS` binds source recovery separately.

| Artifact | Bytes | SHA-256 |
| --- | ---: | --- |
| `smolpng-0.3.0-dev.10-windows-amd64.zip` | 31106228 | `1826008593e0dc56f533000faa9bc6019dd3cfaf7a17ff7bbf288aa790bce040` |
| `smolpng-0.3.0-dev.10-windows-arm64.zip` | 28526263 | `1a3336afb33007f7230d023feb3a3c8d438d430892e2b95ccaced3a631bfe90e` |
| `smolpng-0.3.0-dev.10-linux-amd64.tar.gz` | 19070401 | `2bbce121d540b52cb8a0538113b4fa0c1efeaa88567bfa0e8d598d5f1ee19dc5` |
| `smolpng-0.3.0-dev.10-linux-arm64.tar.gz` | 17699537 | `e638528aaa600f65c9accba0fca69a3a088b911dc1623f5bdca4a95565332081` |
| `smolpng-0.3.0-dev.10-darwin-amd64.tar.gz` | 19099188 | `37afdf86014d0e1f8f65e5dab0e8ff68f6c605a6c4a67cd50ae3bb45002fcfcc` |
| `smolpng-0.3.0-dev.10-darwin-arm64.tar.gz` | 18073653 | `8346a0f6d373a0249ff97d4f72e63c0e01c5ce7d81108bdcaa6f6d8cbe8e41fa` |
| `smolpng-0.3.0-dev.10-source-recovery.zip` | 9736075 | `0d0f545e418406961344604fcb26f2dd38281111ae734177c2fbd855b2d49ead` |

Independent ZIP/tar readers checked each archive and every manifest member/hash, duplicate/unsafe paths, 58 dependency notices and expected PE/ELF/Mach-O architecture headers. Windows has 158 members / 157 hashed payloads; Linux/macOS have 157 / 156. Console and qualifier binaries use the console subsystem; `smolpngw.exe` uses GUI. No external optimizer, personal settings or private agent material is bundled. Isolated Windows CLI readback confirmed exact version, commit and standalone identity.

The source ZIP contains selected `main` and its complete ancestral history, a clean repository and a second Git bundle. ZIP extraction and an independent bundle clone recovered the exact clean commit and passed full Git object checks. This establishes local isolated restoration, not disaster recovery. Earlier candidates remain unchanged.

Windows executable hashes are application `b4f10df117607fb3d9aecab834bc2cfa7316efba848193f04e6eda8391103115`, GUI launcher `cea71634019203c18638a05a192a36d6d83812fa26c08f82b8d3f07fabad8ae0`, and matching qualifier `bee746f3e1b941826e3a02919fd6b4fdd10ca3ed990ff65822f4a5f17827e424`.

## Native and lifecycle standing

The Windows amd64 host reported Core/Home 25H2, build `10.0.26200.9711`. The first native run completed all 63 checks with 60 passing. One lossless JPEG call reported jpegoptim unavailable; two PNG one-byte target calls exceeded the qualifier's two-minute invocation deadline. Original-image and final optimizer provenance checks passed. This original report remains non-green.

Direct quiet Node reproductions used the same application, inputs and tool paths, fresh output paths, and the same resource controls. Dependency readback found all tools available. JPEG orientation 6 then returned success in about 2.5 seconds; an independent Pillow decode confirmed identical RGB pixels and retained orientation 6. Target-miss and explicit-nearest calls returned their expected exit 6, with no committed output for the ordinary miss and a valid 64×64 nearest artifact still explicitly marked missed. They took about 17 and 70 seconds. A host sample during these successful rechecks reported 100% aggregate CPU and no disk queue; this supports investigating scheduling but does not establish the first failures' cause. Focused successes do not rewrite the original full run.

The qualifier's `-timeout` covers an entire CLI invocation; the application's `--timeout` covers each image. An observer deadline equal to the two-minute per-image budget cannot reserve additional startup, preflight or multi-image time. A separate full run used the packaged qualifier's existing `-timeout 5m` option while leaving the application's two-minute per-image limit and BelowNormal/GOMAXPROCS=1/512 MiB controls unchanged. It completed all 63 checks with 62 passing: both earlier PNG target checks and JPEG orientation 6 passed, but Auto JPEG high-chroma reported jpegoptim unavailable. That image passed in the first run, and its lossless case passed in the second. Final original/tool checks passed. The longer observer budget did not resolve intermittent discovery, and neither failed full report is promoted to green. No image-corruption failure was observed; the discovery cause and a clean complete native pass remain open.

The separate 26-command dependency/Skill lifecycle passed in an isolated user profile/home, with local-only processing, one image job, copying disabled and notifications disabled. It installed/reused/selected and later removed managed OxiPNG/jpegoptim, retained licenses, compressed owned PNG/JPEG copies, installed standard/custom Skills, preserved an edited Skill through sync/removal refusal, forgot only that edited installation's managed record, removed untouched copies and packaged the Skill. User config and original images remained. This is filesystem lifecycle evidence, not a real agent client's reload. The Skill ZIP is 6832 bytes with SHA-256 `0bf6c51847499f34ffbdea487fbfd3d6456113a9e3b14c45a971c8d34d5e7c82`.

Earlier [dev.9](2026-10-19-private-candidate-dev9.md) update/recovery evidence retains its own source and binary identities; it is not a fresh dev.10 run.

## Evidence and limits

Completed local inspection records under ignored `artifacts/validation/` are:

| Record | Path | SHA-256 |
| --- | --- | --- |
| All archive/member/header checks | `candidate-dev10-a9b8a48/archive-set.json` | `5715f58e7e5c7e1ee6ed3771b1f9c1a9b5dae64d510d3c68c1bcd3055b4dfad8` |
| Windows extraction/CLI readback | `candidate-dev10-a9b8a48/windows-archive.json` | `023627af8b418d4a2e1e9eb1a8b1c3305c38b47564d93ac5d00b6016a05fb4bd` |
| Source ZIP/bundle restoration | `candidate-source-dev10-a9b8a48/recovery.json` | `d5a92ff82b9231ed5a2667935d6518763093ee279c50f5ceb3441f93c0155662` |
| First native run, 60/63 | `candidate-native-dev10-a9b8a48/report.json` | `349a675d29776ab8f40339149e9f4c559d09da648694043ade06eb9402baff32` |
| Five-minute observer run, 62/63 | `candidate-native-dev10-a9b8a48-timebudget/report.json` | `102558e413ee233debfa5f8d71866ec8cad61b7d04c8f5f128389b898576cb43` |
| Independent three-call recheck inspection | `candidate-native-dev10-a9b8a48-recheck/inspection.json` | `ef18e6db7b26597be8d3175a8cd5ef32b14ffd51be63104d29e662d7628856be` |
| Dependency/Skill lifecycle, 26 checks | `candidate-lifecycle-dev10-a9b8a48/lifecycle.json` | `d83b399cd580401a3658dac0681ec57f19de437a33574042b223fcd34d30f066` |

The [earlier Paint observation](2026-10-19-windows-clipboard-alpha.md) and dev.9 lifecycle/TUI/update/cancellation observations retain their actual identities. File-list/text consumers, physical dragging, notifications, sensitive-clipboard refusal and remaining joined integration/removal remain open in #36. Related native Linux/macOS source builds do not establish these exact foreign archives' runtime, package installation or graphical behavior.

The documentation may land after the observed application commit without rebuilding the candidate to contain its own record. The complete 1.0 destination, Wolf's UI/design workshop and signing/feed/publication boundaries remain intact. No public release, package registration, deployment or spend occurred.
