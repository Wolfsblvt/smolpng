# smolpng

[![Inputs: PNG + JPEG](https://img.shields.io/badge/inputs-PNG_%2B_JPEG-566C86)](docs/cli.md)
[![Targets: exact bytes](https://img.shields.io/badge/targets-exact_bytes-39756A)](docs/cli.md)
[![Recipes: shareable TOML](https://img.shields.io/badge/recipes-shareable_TOML-8A668C)](docs/presets.md)

**Make images smol.** Local-first image compression for humans, scripts, and agents.

The current private source candidate provides a keyboard-driven PNG/JPEG utility and deterministic CLI over separately installed pngquant, OxiPNG and jpegoptim. It preserves originals, compares original-derived candidates and treats upload limits as exact byte targets. The full PNG/JPEG, desktop and agent destination is retained in [Vision](docs/VISION.md); unimplemented pieces are not advertised as working features.

## Quick start

Use Go 1.26.6 or a compatible newer toolchain. From the repository root:

```sh
go mod download
go build -o artifacts/bin/ ./cmd/smolpng
./artifacts/bin/smolpng deps status
```

Install pngquant separately. For OxiPNG, `./artifacts/bin/smolpng deps install oxipng` previews a pinned upstream archive; add `--apply` to download, verify and install it per-user. Or use existing tools with `--pngquant-path` and `--oxipng-path`. JPEG uses separately installed jpegoptim. See [dependency setup](docs/dependencies.md).

With the PNG tools available:

```sh
./artifacts/bin/smolpng compress icon.png --target-size 10KiB --json
```

The result defaults to `icon.smol.png` beside `icon.png`; existing outputs get a collision number. A target is `outputBytes <= requestedBytes`, so `10KB` is 10,000 bytes and `10KiB` is 10,240. Missing a target exits `6`; `--accept-nearest` may write the best result but does not pretend the target was met. The examples below use `smolpng` as shorthand for the built executable.

```sh
smolpng compress icon.png --preset lossless
smolpng compress images --recursive --output-dir optimized --dry-run --json
smolpng compress images --recursive --output-dir optimized --jobs 2 --jsonl
smolpng inspect icon.png --json
```

Detached clipboard/URL/stdin output uses the operating-system Downloads folder under `smolpng/`. `--in-place` is the explicit destructive exception. `--output -` streams clean image bytes; use `--report-file` instead of JSON on that same stream.

Built-ins: `auto`, `web`, `smallest`, `lossless`. Auto is local under the initial `never` policy. The TUI, native clipboard adapter, background/drop workflow, last-result recovery and tracked Agent Skill are implemented. Actual graphical clipboard/notification behavior still needs real-host qualification; TinyPNG is available only under explicit upload permission; native keyring login is separate from permission. Managed OxiPNG uses pinned 10.2.1 archives on Linux/macOS x64/arm64 and Windows x64; pngquant and native Windows arm64 OxiPNG remain manual. [Dependency setup](docs/dependencies.md). [CLI reference](docs/cli.md) · [Development](docs/DEVELOPMENT.md) · [Current boundary](docs/DIRECTION.md).

## One action for a human

```sh
smolpng icon.png                   # Keyboard menu, Auto selected
smolpng                            # Detect image/file/path in clipboard
smolpng drop icon.png              # Configured background recipe, copy and notify
smolpng last --json                # Recover a dismissed notification's result
smolpng settings                   # Edit user defaults without erasing comments
```

Use arrows/Enter, `S` for source, `,` for settings, `D` for setup, and result shortcuts for copying or opening the folder. First use checks local tools and offers explicit paths, a separately confirmed managed OxiPNG download, or the optional Agent Skill. It neither downloads before that confirmation nor requires an API key. Copy a valid preset block before launch to preview/import it. The header is a two-line replaceable placeholder, not final commissioned art.

`--plain` removes decorative styling; `--accessible` uses typed choice names rather than arrow keys or numbered menus. `--no-header` removes the header. Scripts should use `compress`; a redirected interactive invocation refuses rather than hanging for input.

Windows builds also include an optional `smolpngw.exe` drop launcher compiled without a console. Linux clipboard ownership can need a living process: keep the result screen open until paste, or let `drop` serve it for a bounded period. Read the [desktop workflow guide](docs/desktop.md) for exact limits and qualification status.

## JPEG, optional TinyPNG, and native launchers

```sh
smolpng compress photo.jpg --quality 65-85 --json
smolpng compress photo.jpg --preset lossless --json
smolpng providers login tinypng
smolpng compress icon.png --engine tinypng --remote-policy always --json
smolpng integrate install --desktop              # preview native launchers
smolpng integrate install --desktop --apply      # explicitly install
smolpng update status --json
```

[JPEG](docs/jpeg.md) preserves display orientation and ICC under the safe default while removing GPS/private metadata. Geometry changes come directly from the original decode, not repeated lossy output. jpegoptim is installed separately; use `--jpegoptim-path` or `deps configure jpegoptim PATH --apply`.

[TinyPNG](docs/remote.md) reads the process environment before the native keyring, never from a secret command-line argument. Its reports disclose upload attempts and service counts even when local output wins. Ordinary tests never call the service. A key does not enable Auto upload.

[Desktop integrations](docs/integration.md) create per-user menu, optional desktop and Windows Send To shortcuts. Modified or unrelated files are preserved. [Standalone updates](docs/updates.md) require an independently trusted signed manifest, use guarded backups and rollback, and synchronize only unmodified managed Skills. Package-manager installations are never self-overwritten. This candidate has no production signing key/feed, package registration, or public deployment.

## Share a recipe, not a script

Place this in your project's `.smolpng.toml`, or copy it to a file and preview/import it with `smolpng preset import recipe.toml` and then `--apply`:

```toml
[presets.mcp-icon]
schema = 1
label = "MCP icon · ≤10 KiB"
extends = "auto"
formats = ["png"]
target_size = "10KiB"
min_quality = 55
allow_resize = true
min_dimensions = "64x64"
remote = "never"
clipboard = "smart"

[presets.mcp-icon.menu]
pinned = true
position = 2
```

```sh
smolpng compress icon.png --preset mcp-icon --json
smolpng config show --effective --preset mcp-icon --provenance
smolpng preset export mcp-icon --resolved
```

User settings, the nearest project config, inherited presets and explicit flags combine predictably. Imports cannot contain commands, secrets, executable paths or a global upload permission. A preview shows collision renames or replacements before an import is committed. Export preserves parents, or freezes effective recipe values. Clipboard preference and pinned placement drive the keyboard/drop workflows. Deterministic compression copies only when `--copy` was explicitly requested, and never writes last-result history implicitly.

[Preset and configuration reference](docs/presets.md) · [Copyable CC0 examples](presets/examples/)

## Regression corpus

The [classified corpus](testdata/corpus/README.md) contains 24 CC0 structural PNG/JPEG specimens, including 16-bit alpha, all EXIF orientations, real synthetic GPS metadata, and a true progressive JPEG. Run `go run ./tools/corpus --check` to verify byte and provenance integrity. Use `--recursive` to compare the whole corpus; photographs and full ICC color-managed evaluation remain explicit admission gaps, not quality claims derived from synthetic patterns.

## Verify

```sh
go test ./...
go vet ./...
```

Tests exercise real filesystem/process boundaries and fake optimizer/HTTP boundaries using original CC0 specimens. Cross-compilation and fake processes do not qualify a real desktop or optimizer version. Standard packages do not bundle optimizer executables.

An open-source project from Wolfsblvt Works.

## Local agent integration

```sh
smolpng skill install --target agents       # preview only
smolpng skill install --target agents --apply
smolpng skill status --json
smolpng skill package --target chatgpt --output smolpng-skill.zip
```

The embedded instructions follow the installed CLI. Destinations and exact hashes are tracked; synchronization never silently overwrites local edits. Custom paths and instruction-only ChatGPT packaging are supported. See the [Agent Skill guide](docs/agent-skill.md) for version replacement, safe updates and the local-runtime boundary.

## Compare before changing defaults

`smolpng compare images --recursive --output-dir artifacts/review --json` measures the same original through local presets and exports candidates with an offline visual report. See [comparison](docs/comparison.md) for tool isolation, repeats, targets and metric limits; [explicit background](docs/background.md) explains optional PNG alpha flattening.

### Right-click images

`smolpng integrate shell install --apply` selects a single preserve-original **Make smol** action; `--mode submenu --preset mcp-icon` exposes labelled replacement, interactive and custom-user recipes. Windows 11 uses an original native COM bridge and signed sparse package, **not** a legacy registry verb. Settings can be authored here; the native SDK build and signed activation remain a separate host qualification. Linux installation targets KDE Dolphin only. See [context-menu setup and boundaries](docs/context-menu.md).

## License

The application and ordinary documentation use the [MIT license](LICENSE), allowing use, modification and redistribution; corpus, dependency and future website/brand rights are identified separately in the [license map](license-map.md).
