# CLI contract

## Meaning

This is the current static-PNG machine interface. The canonical noninteractive path is `smolpng compress INPUT...`; it never prompts. Root `--auto` is the one-shot convenience form. Bare-root interactive use is not yet implemented in this cut.

## Input/output

INPUT supports files, directories, explicit globs, HTTP(S) URLs and stdin `-`. `--clipboard` explicitly selects the clipboard abstraction. Directory traversal is shallow unless `--recursive`; use `--hidden`, `--follow-symlinks`, `--include`, `--exclude` and `--no-ignore` deliberately. `.gitignore` support is scoped to discovered files, not Git's index or global excludes. Generated `.smol.*` and the chosen output subtree are skipped during discovery.

Default output is `{stem}.smol{ext}` beside a local original, or OS Downloads/smolpng for detached input. `--output` is exact for one input, `--output-dir` mirrors relative batch paths, and templates support stem/ext/timestamp. `--overwrite` cannot replace originals implicitly. `--in-place` refuses symlinks and checks the source again before commit. Racing no-clobber publication refuses rather than overwriting.

`--max-input-size` defaults to 64MiB; `--max-pixels` defaults to 64 million. Upper accepted limits are 1GiB and 256 million pixels. HTTP requests have redirect/downgrade guards and bounded reads; `--offline` rejects URLs. A trusted CLI may intentionally fetch loopback/internal URLs and must not be exposed as an untrusted web service.

## Recipes

`--preset auto|web|smallest|lossless`, `--quality MIN-MAX`, `--min-quality`, `--colors`, `--speed`, `--dither`, `--oxipng-level`, `--strip keep|safe|all`, `--quantize`/`--no-quantize`, and explicit hidden-alpha controls shape local processing. Each lossy attempt is original-derived.

`--target-size` is inclusive bytes with exact SI/IEC units. No resize occurs unless `--allow-resize`, `--resize` or `--max-dimensions` is explicit. Target fallback respects `--min-dimensions`. `--accept-nearest` writes on a miss while retaining exit 6; `--force-output` permits size growth. `--jobs` is 1..32; fail-fast uses serial work. `--timeout` defaults to 2m including input and optimization; zero disables it.

Only local optimization exists. Remote policy defaults to never. A TinyPNG key is not read; upload-requiring options fail explicitly, not silently or with a fake success. Safe extra optimizer arguments are restricted; `exec TOOL -- ARGS` is the deliberately unrestricted escape hatch and returns upstream exit codes.

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
