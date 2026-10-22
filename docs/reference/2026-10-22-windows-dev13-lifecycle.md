# Windows dev.13 dependency, Skill and update lifecycle

## Meaning

The extracted private `0.3.0-dev.13` candidate passed 28 real dependency/Agent Skill lifecycle calls and 12 local signed-fixture update/recovery calls on Windows amd64. Independent readback confirmed the instruction ZIP's four resources and exact restoration of all 197 original package files, configuration and an edited Skill. This record preserves the original reports after the [composition record](2026-10-22-private-candidate-dev13.md); it does not rewrite earlier evidence, complete the graphical journey, qualify a production feed or accept the whole product.

## Candidate and host

The immutable application source remains `319d5e87cef6d772020732107139ebf21fb9b0db`, version `0.3.0-dev.13`, standalone Windows amd64. Both journeys hash-checked the actual extracted CLI against `41ebbba74f01fbf61175345dd17af531ac04f8e0fbfe7aed118bcf264be25e78`. The update journey independently extracted the complete Windows ZIP after checking its SHA-256, `bf7b73c9ea1b69d5c57b02177bc550e51a8bde9f32b8b6f2843dcc9bae264fd3`.

The existing host is Windows 11 Home/Core 25H2 amd64, build 10.0.26200.9711. Direct Node children used hidden process creation, and the application used its existing quiet native child constructor. No clipboard or graphical entry point was invoked. Disposable USERPROFILE/HOME/AppData locations and portable app homes kept installations out of the actual user profile. `TINIFY_API_KEY` was explicitly empty in the children; neither journey requested a provider credential or uploaded an image.

## Managed dependencies and Agent Skill

The owned run completed **2026-10-22 13:32:35.987Z–13:32:59.723Z**, exit 0. All 28 CLI calls passed with their complete arguments, output and expected refusal retained.

- OxiPNG and jpegoptim previewed missing installations under the disposable home, downloaded their existing pinned upstream artifacts, passed version probes and installed without PATH changes. Repeated installs retained exactly one version each.
- Explicit managed preference selected those actual binaries. A hash-checked separate pngquant path was configured. Real offline PNG and JPEG Auto jobs produced outputs and preserved original bytes.
- Agents, Claude and custom Skill targets were previewed before writing. Every resolved destination was inside the fixture root; every installed resource matched the candidate manifest's embedded Skill hashes.
- An edited Claude Skill survived applied sync. Uninstall refused with exit 5; explicit forget left the edited bytes intact. Unchanged agents/custom installations were removed.
- Final Skill status has no installations, pending recovery or cleanup. Both managed optimizer trees were removed through the application. The deliberately edited, forgotten fixture Skill remains preserved.

| Tool | Version | Executable SHA-256 | Installed license SHA-256 |
| --- | --- | --- | --- |
| OxiPNG | 10.2.1 | `35ae3980ab831af64f4aecc98f69b81bfa146fa750a61d17ebbb12520128cbc8` | `1bc9688d785fa50345f803824ddde000bdabedda9d1803dc41a1fedd81f784fc` |
| jpegoptim | 1.5.6 | `0a79681b56961ec2c5f45c4f9dff01a64e7274b4f3250c22cde5f09c0bb457f3` | `2d52c5a99b196c919a73fb7510c75154f84ea676115eef48af68d86ded6954a8` |

The downloaded archive hashes were `7e940f83ee46874b73f53031f96a15834cb70b220af27391fb06fe7b4dd798e1` and `db2d8caee88f2665b772c4591cba223b8b729500b88266dc06d4bf79d28fe11a`, respectively. Separately selected pngquant 2.17.0 has SHA-256 `7aadad6a50aa5c40cfcb20f9478c92f6030360a3424bc262ce383dc1e97fb86a`.

Independent ZIP inspection found exactly `smolpng/LICENSE`, `smolpng/SKILL.md`, `smolpng/references/recipes.md` and `smolpng/references/results.md`, with actual member bytes matching the candidate's resource hashes. The instruction-only ZIP SHA-256 is `22c49ec6ae5ea97e85fd9f76c06cd1aaf0cb400c6a83719ab92b754a0a5e60b2`. It is byte-identical to the earlier package because the embedded resource content did not change; that fact does not claim a client reload or ChatGPT upload acceptance.

The complete lifecycle report is **56,111 bytes**, SHA-256 `97835c96f93b1c00515298dce374444af44a0e9f9ebe4a657a5641e678d5deec`. Its original local root is `artifacts/validation/candidate-lifecycle-dev13-319d5e8/`. The disposable helper's SHA-256 is `b6197179a67bbb10a5ec76634e46f18982ad5b30e137ad6b8ab805dfa57514bd`.

## Actual update and rollback

The separate owned run completed **2026-10-22 13:37:45.606Z–13:37:59.353Z**, exit 0, with all 12 CLI calls passing. The existing local stable `0.3.0` fixture deliberately names source `b0ed6391c4d8c80fcb9d78c611aa42d8debb798f`; it is not a dev.13 build or a public release. No private signing key was read, generated or exported.

| Existing fixture input | SHA-256 |
| --- | --- |
| Narrow update archive | `427db73c1027db48f52604171f67a342d34b254148317860dcc14f9f41764cdf` |
| Signed envelope | `9e63f49535c0208a722a7829c729b47dffd15c520b20e9ee7cfc86eab391d427` |
| Test public trust file | `401bdc10cb2d4cdf9a23670cea5e88ea2d15e942acf059b4672cee9b819e1961` |

The fixture's signed validity interval is 19 October 2026–17 January 2027. All inputs were local; neither a feed request nor a signing operation was needed.

The candidate first refused the wrong public trust and a tampered archive with exit 5. All original package bytes still matched. A project Skill was then installed and deliberately edited. Applying the admitted fixture returned `scheduled` and `completed: false`; completion was established separately by `staged → applying → complete`, actual helper exit and the installed fixture's exact version/source report.

After a recovery preview, explicit rollback also returned only a scheduled acknowledgement. The observed journal moved through `rolling-back → rolled-back`, and that recovery helper exited. The installed CLI again reported dev.13's original source. An independent final pass matched all **197** original package files, including the original manifest and files outside the narrower update archive. Configuration and the edited project Skill retained their exact hashes. Forget removed Skill tracking while preserving the edit. Both exact helper PIDs were absent during independent readback.

The complete update report is **289,781 bytes**, SHA-256 `b07d3e406077edd90529b00cc25e70b704b9069d54c4c4f5f6c9a2ef0e6a017c`, at original local root `artifacts/validation/candidate-update-dev13-319d5e8/`, alongside the independent inspection receipt. The helper SHA-256 is `87f386668da0d467d193353d7149d03b401abf0331656acb4ce9ae40e0299491`.

This run did not inject a file lock or force a mid-apply interruption. Those [dev.12 observations](2026-10-22-private-candidate-dev12.md) retain their original source identity. A focused Git comparison between dev.12's `61b3ea59a3dbf21275d2bc8ba3d29fdae566776e` and dev.13's selected source found no changed bytes in `internal/updates`, `internal/skill`, `internal/fsutil`, `internal/process` or `internal/cli/update.go`. That supports reuse at those unchanged boundaries; it does not relabel the old tests as new runs or prove every caller and host unchanged.

## Retained evidence and remaining boundary

The [authored evidence payload](2026-10-22-windows-dev13-lifecycle/report.json) retains four original UTF-8 inputs with their actual hashes: both complete reports, independent instruction-ZIP inspection and independent update restoration inspection. It is **370,384 bytes**, SHA-256 `2c97b9c78e78d4b64d658c505d1ec6debadab18f4ac0f2b4d86941bc43e169a7`. Every original input round-trips byte-exactly. The wrapper is an authored join, not itself an original capture.

These observations add actual dev.13 lifecycle evidence to the composition/native results without changing the candidate archives. Katja retains the joined qualification in [Issue #36](https://github.com/Wolfsblvt/smolpng/issues/36). Physical drop/Send To, rendered notification/focus and smart clipboard/file-picker consumption remain held for a newly authorized desktop window. Live TinyPNG, trusted Explorer signing/registration, broader physical host use, the final public-experience workshop and publication retain their separate custody and authority. This record does not complete that product destination.
