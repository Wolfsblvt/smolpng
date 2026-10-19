# Private candidate 0.3.0-dev.7

## Meaning

Source `9b460e768b908b1efa2c79560903dd7084c2da5d` produced six private standalone platform archives and a separately restored source-recovery ZIP. Independent archive inspection passed; the extracted Windows amd64 application and its packaged helper passed 61 native PNG/JPEG checks, 26 dependency/Skill lifecycle checks and an 11-command signed-fixture update/rollback journey. This record preserves those exact artifacts and component observations. It does not accept the complete private candidate: physical desktop/clipboard/TUI use and other remaining observations stay open in [Issue #36](https://github.com/Wolfsblvt/smolpng/issues/36).

## Source and producer

The clean `main` source above includes the [release-environment correction](https://github.com/Wolfsblvt/smolpng/pull/37). The maintained `tools/release` builder completed normally for Windows, Linux and macOS, each at amd64 and arm64, with explicit version `0.3.0-dev.7` and standalone ownership. It used Windows Go 1.27.2; the selected `go.exe` SHA-256 is `b8cd508a0a7aac883cd750e9e75eb306d219e726bdf1cde29009490729437fbf`. `GOMAXPROCS=2` bounded compiler concurrency. The corrected builder fixes workspace/user configuration, toolchain choice, build flags, experiments, CGO and portable ISA baselines for its build children.

The existing module cache was used, with observed `GOPROXY=https://proxy.golang.org,direct` and `GOSUMDB=sum.golang.org`; local `go mod verify` passed. This is not a clean-cache or hermetic-build claim. All eleven [hosted Verify jobs](https://github.com/Wolfsblvt/smolpng/actions/runs/39022755699) passed on the exact source, including tests, vet, module/corpus verification, six target builds and Windows shell compilation. Native race checking was not run: neither gcc nor clang was found on this host's command path.

The preliminary `0.3.0-dev.6` build at `d2c4f46` was explicitly cancelled after four platform archives when PR #37 changed the producer contract. Its completed outputs and related lifecycle observations remain separate; they were not relabelled as dev.7 results.

## Actual artifacts

These are retained local files, not public downloads. The platform/source archives are under `artifacts/candidates/0.3.0-dev.7-9b460e7/`; the separately generated Skill ZIP is retained in `artifacts/validation/candidate-lifecycle-dev7-9b460e7/`. `SHA256SUMS` and the completed `build-index.json` bind the six platform archives; `SOURCE-SHA256SUMS` separately binds source recovery.

| Artifact | Bytes | SHA-256 |
| --- | ---: | --- |
| `smolpng-0.3.0-dev.7-windows-amd64.zip` | 31083923 | `857f153c70d12957ab20b3aad715d72138829997662ca30e76a77c1f36658746` |
| `smolpng-0.3.0-dev.7-windows-arm64.zip` | 28494426 | `df3b23db81e4887a463d82611d84b393c6e97e6ee2aa269a2d89801800320850` |
| `smolpng-0.3.0-dev.7-linux-amd64.tar.gz` | 19045274 | `9575048e2b64fbeed5e20ee871dba670c96ebe39c6966acb095e569326e4290c` |
| `smolpng-0.3.0-dev.7-linux-arm64.tar.gz` | 17646868 | `f45ee138bea97df538570e88db72b221e7490f8b7aab9e3cbfc5ec1de3d7a713` |
| `smolpng-0.3.0-dev.7-darwin-amd64.tar.gz` | 19083265 | `c3261a1c09def8c2531e73914e899c53ab334af85f40f3f3e8168fa3318ab843` |
| `smolpng-0.3.0-dev.7-darwin-arm64.tar.gz` | 18046357 | `7e4b87c1f263661097209aec543d8f62463cef1dba0e36f4a504b48a4703e3a4` |
| `smolpng-0.3.0-dev.7-source-recovery.zip` | 9612066 | `50df5bd1f7f2da35131c6e85bb7095952735f1f2e1776a2deebef5644a708445` |
| `smolpng-skill.zip` | 6832 | `0bf6c51847499f34ffbdea487fbfd3d6456113a9e3b14c45a971c8d34d5e7c82` |

Independent ZIP/tar readers checked every manifest member and hash, unique safe paths, all 58 dependency notices, and the expected platform executables. Windows archives have 153 members and 152 hashed payload files; Linux/macOS archives have 152 members and 151 hashed payload files. PE, ELF and Mach-O headers matched each architecture; Windows console/qualifier executables use the console subsystem and `smolpngw.exe` uses the GUI subsystem. Foreign executable headers are build evidence, not runtime evidence. External optimizers, personal settings and private agent material are absent.

Source recovery carries the selected main branch and its complete ancestral Git history, excluding other local refs and working files. ZIP integrity passed. Both the extracted Git repository and an independent clone from the contained bundle recovered the exact clean source HEAD and passed full Git object verification. This establishes local restore capability, not disaster qualification or a published backup service.

## Native Windows observations

The extracted application reported the exact candidate version, source and standalone ownership. Application SHA-256 is `42f2b8d96150bf9e0c93f58623bae093a215376e4cdfc541fc2d83fda7160b2e`; windowless launcher SHA-256 is `8335bd717f0e0c40a52a0d0d4052dec5a5090bc781ae45e3d11cd9b41d946bb1`; packaged qualifier SHA-256 is `4da027dcb03bf8da61b1a2ef6b36df96a18764fa0250d55b7bac20ca3751ece6`.

**Provenance correction, 19 October 2026:** All three suites below ran on the same local Windows amd64 workstation in this implementation session. They did not checkpoint its Windows edition/build or the optimizer executable hashes at execution time. A recovery-time measurement identifies the current host as Microsoft Windows 11 Home, 25H2, build `26200.9711`, 64-bit; its last boot was after those suites. This is current host evidence, not proof of the OS build used on 17 October. Exact historical OS-build provenance remains uncaptured under #36.

The qualifier's retained dependency report records the following explicitly selected paths and versions. Rehashing those retained paths on 19 October produced these identities; the report itself did not capture optimizer hashes. File timestamps are not proof that bytes remained unchanged, so these recovery measurements do not retroactively bind the 61-check execution to exact optimizer bytes.

| Optimizer | Selected path relative to repository root | SHA-256 measured on recovery |
| --- | --- | --- |
| pngquant 2.17.0 | `artifacts/toolchain/pngquant/pngquant/pngquant.exe` | `7aadad6a50aa5c40cfcb20f9478c92f6030360a3424bc262ce383dc1e97fb86a` |
| OxiPNG 10.2.1 | `artifacts/toolchain/oxipng-10.2.1/oxipng-10.2.1-x86_64-pc-windows-msvc/oxipng.exe` | `35ae3980ab831af64f4aecc98f69b81bfa146fa750a61d17ebbb12520128cbc8` |
| jpegoptim 1.5.6 | `artifacts/toolchain/jpegoptim-1.5.6/upstream/jpegoptim.exe` | `0a79681b56961ec2c5f45c4f9dff01a64e7274b4f3250c22cde5f09c0bb457f3` |

`candidate-native-dev7-9b460e7/recovery-identities-20261019.json` retains the dated host/hash measurement separately from the original report. #36 retains the exact-provenance obligation for the joined candidate journey: checkpoint host and executable identities before further qualification. The historical component results below remain useful, with the omissions above; they are not exact historical host/optimizer-byte attestations.

The packaged qualifier passed all 61 checks with separately installed pngquant 2.17.0, OxiPNG 10.2.1 and jpegoptim 1.5.6. These cover structural PNG/JPEG Auto and lossless behavior, orientation, target success/miss/nearest, binary stdout, partial batches and a mixed batch. The helper's report binds both source identities and actual executable hash. It does not use the clipboard or install optimizers.

The same extracted application then passed 26 lifecycle commands in a disposable portable home/profile. Both managed optimizers were previewed, downloaded, verified, installed and reinstalled without changing PATH. Managed selection and real PNG/JPEG compression passed; original images remained unchanged and remote use stayed false. Skill preview/install/status/sync, custom-path handling, modified-content refusal, forget and unchanged uninstall behaved as documented. The application generated an instruction-only ChatGPT Skill ZIP, inspected locally; no upload or remote-execution capability was claimed. Optimizer removal preserved settings, original images and modified Skill bytes.

An earlier dev.6 helper expected exit 2 for modified-Skill removal; the application correctly returned its existing `skillError` code 5. That incomplete helper attempt and its bounded resume remain captured. The dev.7 lifecycle used the corrected expectation from its start and passed without that helper failure.

## Signed-fixture update and recovery

The installed original was an exact copy of the dev.7 archive, sharing the already exercised portable home/profile. The maintained builder separately produced a test-only same-source Windows `0.3.0` fixture because the stable update-manifest contract requires a stable version. This fixture is not a selected public version or release. Its narrow update archive SHA-256 is `e58df2bc79dd0468d21de5fc12d718e0b93813a27136572b093d3e5a19e1366c`.

An ephemeral Ed25519 key remained only in the qualification process; retained files contain public trust and signed payloads. Untrusted-key and tampered-archive application were refused without changing installed files. Valid application returned a scheduling receipt; after that exact helper exited, its journal reached `complete`, all 64 updated member hashes matched, and the installed executable reported the fixture version/source.

A read handle then denied replacement of the disposable installed main executable. Rollback scheduling succeeded; the completed worker reported `recovery-required` with an access-denied error. Every file remained at either its exact previous or signed-new hash. After releasing that handle, ordinary recovery on the same journal reached `rolled-back`, restored every original candidate-member hash and returned version `0.3.0-dev.7`. Edited tracked Skill bytes survived. The config's durably checkpointed pre-apply and final hashes both equal `a9d7f1478919239c8a887c4c56e0faac20b2f59ddae02834cb5c82b561aa309a`. All eleven command cases passed.

The initial update helper incorrectly assumed that a failed rollback leaves every file new: the qualifier executable had already been restored before the main-file lock stopped progress. Current rollback is a guarded, resumable per-file operation; it admits exact old/new hashes, not arbitrary bytes. That first helper also retained its pre-apply config hash only in memory. Its partial-state evidence and normal recovery were preserved; the corrected full cycle checkpointed config/Skill proof before application and exercised the complete apply/locked-rollback/retry sequence afresh. No product source changed to satisfy either helper assumption.

## Retained evidence and remaining journey

Local reports live under `artifacts/validation/`: `candidate-dev7-9b460e7/`, `candidate-native-dev7-9b460e7/`, `candidate-lifecycle-dev7-9b460e7/`, `candidate-source-dev7-9b460e7/` and `candidate-update-qualified-dev7-9b460e7/`. The first update attempt and recovered partial state remain separately in `candidate-update-journey-dev7-9b460e7/`. Producer/helper stdout, stderr and receipts are retained with project-scoped local working data.

Physical dragging, graphical clipboard representations and paste, sensitive-clipboard refusal through a real graphical surface, rendered notifications/TUI, fresh candidate-specific desktop integration/removal, and forced worker-interruption recovery are not established here. The older [native Linux package record](2026-10-17-linux-package-artifacts.md) remains package-file evidence for its recorded sources; no fresh prerelease native-package or Linux installation claim is made. Windows arm64, Linux and macOS runtime remain unobserved. Signing, production feeds, registry submission, release upload, Store, website and public version choices retain their existing boundaries.

Katja retains the remaining candidate journey in #36, including the real graphical/ordinary-use seam; Nyxara retains product/review judgment, and Wolf retains the design workshop and public-release choices. This evidence record neither closes #36 nor the complete product Goal.
