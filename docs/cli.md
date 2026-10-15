# CLI contract

## Meaning

This guide owns the implemented PNG/JPEG command interface, including deterministic compression, discovery, configuration, human workflows and integrations. The canonical noninteractive path is `smolpng compress INPUT...`; it never prompts. Root `--auto` is the one-shot convenience form; bare `smolpng [INPUT...]` opens the keyboard interface.

## Input/output

INPUT supports files, directories, explicit globs, HTTP(S) URLs and stdin `-`. `--clipboard` explicitly selects the clipboard abstraction. Directory traversal is shallow unless `--recursive`; use `--hidden`, `--follow-symlinks`, `--include`, `--exclude` and `--no-ignore` deliberately. `.gitignore` support is scoped to discovered files, not Git's index or global excludes. Generated `.smol.*` and the chosen output subtree are skipped during discovery.

Default output is `{stem}.smol{ext}` beside a local original, or OS Downloads/smolpng for detached input. `--output` is exact for one input, `--output-dir` mirrors relative batch paths, and templates support stem/ext/timestamp. `--overwrite` cannot replace originals implicitly. `--in-place` refuses symlinks and checks the source again before commit. Racing no-clobber publication refuses rather than overwriting.

`--max-input-size` defaults to 64MiB; `--max-pixels` defaults to 64 million. Upper accepted limits are 1GiB and 256 million pixels. HTTP requests have redirect/downgrade guards and bounded reads; `--offline` rejects URLs. A trusted CLI may intentionally fetch loopback/internal URLs and must not be exposed as an untrusted web service.

## Recipes

`--preset auto|web|smallest|lossless`, `--quality MIN-MAX`, `--min-quality`, `--colors`, `--speed`, `--dither`, `--oxipng-level`, `--strip keep|safe|all`, `--quantize`/`--no-quantize`, and explicit hidden-alpha controls shape local processing. Each lossy attempt is original-derived.

`--target-size` is inclusive bytes with exact SI/IEC units. No resize occurs unless `--allow-resize`, `--resize` or `--max-dimensions` is explicit. Target fallback respects `--min-dimensions`. `--accept-nearest` writes on a miss while retaining exit 6; `--force-output` permits size growth. `--jobs` is 1..32; fail-fast uses serial work. `--timeout` defaults to 2m including input and optimization; zero disables it.

Local optimization is the default. TinyPNG is available only under an explicit permitting remote policy; default `never` does not look up credentials or upload. See [optional TinyPNG](REMOTE.md) for policy and reporting. Safe extra optimizer arguments are restricted; `exec TOOL -- ARGS` is the deliberately unrestricted escape hatch and returns upstream exit codes.

## Reports and discovery

Human reports use stdout; diagnostics use stderr. `--json` emits an envelope with schemaVersion, results, summary and exitCode. `--jsonl` emits one result per image in planned order. A result includes source/output sizes, dimension changes, recipe, pipeline/version, target, resize and remote use, warnings and exit code. Only committed outputs have artifact paths; dry runs explicitly label paths planned. `--verbose` adds argument arrays.

Binary `--stdout`/`--output -` is incompatible with JSON stdout. Use `--report-file` for a sidecar report; reports are not stored implicitly. `inspect`, `capabilities`, `doctor`, `deps status`, `version` and shell completions are implemented. Source builds and cross-builds do not qualify graphical clipboard behavior.

| Exit | Meaning |
| ---: | --- |
| 0 | Success, unchanged, or planned |
| 2 | Invalid arguments/policy |
| 3 | Invalid/unavailable input |
| 4 | Dependency unavailable |
| 5 | Optimizer/output failure or timeout |
| 6 | Strict target missed |
| 7 | HTTP/remote operation failure |
| 8 | Partial or failed batch |
| 130 | Cancellation |

## Configuration and presets

The implemented `config` and `preset` families are documented in [Configuration and presets](presets.md). `--config FILE` selects user configuration, `--no-project` suppresses project discovery, and `compress --preset NAME` now accepts inherited saved recipes. `config path` remains usable when parsing fails. These additions do not change binary stdout, target exit codes, or local-first permission.


## Human and background commands

Bare `smolpng [INPUT...]` opens the keyboard interface; `--auto` retains deterministic one-shot behavior. `setup` checks local dependencies and offers explicit paths or optional Agent Skill installation, with no automatic network. `settings` edits supported user defaults while preserving existing TOML comments. `config set KEY VALUE` performs the same scalar edit without prompts. `--plain`, `--accessible` and `--no-header` adjust interactive presentation. The accessible path uses displayed names, never numeric choices.

`drop [INPUT...]` runs the configured drop preset; no input explicitly selects the clipboard. It records the last operation and attempts the configured copy/notification. `drop.mode='interactive'` selects the human path only outside JSON/JSONL mode. `smolpngw.exe` is the Windows no-console entry point for the same command. It is built with the GUI subsystem; `integrate install` separately creates tracked shortcuts; neither a build nor shortcut creation proves graphical drop behavior.

`--copy smart|path|image|file|files|directory|none` explicitly selects copied output for `compress`, or overrides the configured human/drop preference. Binary stdout has no artifact file to copy and conflicts with non-none copy. A requested deterministic copy failure returns nonzero while retaining the saved artifact. Human/background copying is best effort with visible/retained warnings. `--clipboard-hold 0..30m` optionally retains Linux ownership; ordinary compression and JSON drop default to no wait. Human menus retain ownership while the result screen remains open; non-JSON drop defaults to five minutes or until replacement/cancellation.

`last [--json]` reads the previous human/drop result and reports current artifact existence. `last --copy MODE [--clipboard-hold 5m]` reoffers existing outputs; `last --open` opens their folder; `last --clear` deletes metadata only. Dry-run never updates the clipboard, last result or notification. There is no implicit history for ordinary automation.

## Dependency management

`deps install/update [oxipng]` previews the installed app's fixed catalog without network or writes; `--apply` downloads and verifies before activation. `deps configure TOOL PATH` probes and saves an explicit local path. `deps use-system` or `use-managed` previews clearing saved overrides and changing discovery preference; `--apply` performs that setting change. `deps remove TOOL --apply` only removes unchanged recorded per-user versions. `deps licenses` and `doctor` expose archive/native-availability and managed current/modified/missing/untracked state. No command changes PATH, updates package-manager-owned binaries or follows an unpinned latest release. See [Dependencies](dependencies.md).

The machine-readable [compression schema](../schemas/compression-v1.schema.json) covers the JSON report envelope and standalone JSONL result records. It permits additive fields while enforcing the current required fields, byte/count types, target shape and stable exit codes. Other command families (configuration, Skills, dependencies, inspection) are not claimed to match the compression schema. Repository tests align the declared properties and required fields with the actual Go JSON tags; candidate validation additionally checks real emitted JSON against the schema.

## JPEG, providers, launchers and updates

JPEG shares `compress`, `inspect`, presets, output naming, strict target and JSON contracts. `--jpegoptim-path` selects a separate JPEG optimizer. See [JPEG](JPEG.md) for orientation, metadata and geometry bounds.

`providers login/status/logout tinypng` manages native credentials without changing upload policy. Login supports masked terminal input or `--stdin`; no key argument exists. `--remote-policy always|compare|fallback` is explicit permission; initial `never` and deterministic `ask` do not upload. See [Remote](REMOTE.md).

`integrate install/status/remove/repair` and `integrate shortcut create NAME` expose exact tracked destinations, preview by default and `--apply` for mutation. `update status/check/apply/recover` respects installation ownership and signed manifests. Apply reports scheduling rather than falsely claiming the later worker completed. See [Integration](INTEGRATION.md) and [Updates](UPDATES.md).
