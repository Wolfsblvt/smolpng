# smolpng

**Make images smol.** Local-first image compression for humans, scripts, and agents.

The current private source candidate provides a keyboard-driven PNG utility and deterministic CLI over separately installed pngquant and OxiPNG. It preserves originals, compares original-derived candidates and treats upload limits as exact byte targets. The full PNG/JPEG, desktop and agent destination is retained in [Vision](docs/VISION.md); unimplemented pieces are not advertised as working features.

## Quick start

Use Go 1.26.6 or a compatible newer toolchain. From the repository root:

```sh
go mod download
go build -o artifacts/bin/ ./cmd/smolpng
go run ./cmd/smolpng deps status
go run ./cmd/smolpng compress icon.png --target-size 10KiB --json
```

Install pngquant/OxiPNG separately or use `--pngquant-path` and `--oxipng-path`. The result defaults to `icon.smol.png` beside `icon.png`; existing outputs get a collision number. A target is `outputBytes <= requestedBytes`, so `10KB` is 10,000 bytes and `10KiB` is 10,240. Missing a target exits `6`; `--accept-nearest` may write the best result but does not pretend the target was met.

```sh
smolpng compress icon.png --preset lossless
smolpng compress images --recursive --output-dir optimized --dry-run --json
smolpng compress images --recursive --output-dir optimized --jobs 2 --jsonl
smolpng inspect icon.png --json
```

Detached clipboard/URL/stdin output uses the operating-system Downloads folder under `smolpng/`. `--in-place` is the explicit destructive exception. `--output -` streams clean PNG bytes; use `--report-file` instead of JSON on that same stream.

Built-ins: `auto`, `web`, `smallest`, `lossless`. Auto never uploads. The TUI, native clipboard adapter, background/drop workflow, last-result recovery and tracked Agent Skill are implemented. Actual graphical clipboard/notification behavior still needs real-host qualification; TinyPNG and managed optimizer downloads are not implemented yet. [CLI reference](docs/cli.md) · [Development](docs/DEVELOPMENT.md) · [Current boundary](docs/DIRECTION.md).

## One action for a human

```sh
smolpng icon.png                   # Keyboard menu, Auto selected
smolpng                            # Detect image/file/path in clipboard
smolpng drop icon.png              # Configured background recipe, copy and notify
smolpng last --json                # Recover a dismissed notification's result
smolpng settings                   # Edit user defaults without erasing comments
```

Use arrows/Enter, `S` for source, `,` for settings, `D` for setup, and result shortcuts for copying or opening the folder. First use checks local tools and offers explicit paths or the optional Agent Skill; it does not download anything or require an API key. Copy a valid preset block before launch to preview/import it. The header is a two-line replaceable placeholder, not final commissioned art.

`--plain` removes decorative styling; `--accessible` uses typed choice names rather than arrow keys or numbered menus. `--no-header` removes the header. Scripts should use `compress`; a redirected interactive invocation refuses rather than hanging for input.

Windows builds also include an optional `smolpngw.exe` drop launcher compiled without a console. Linux clipboard ownership can need a living process: keep the result screen open until paste, or let `drop` serve it for a bounded period. Read the [desktop workflow guide](docs/desktop.md) for exact limits and qualification status.

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

## Verify

```sh
go test ./...
go vet ./...
```

Tests exercise real filesystem/process boundaries and fake optimizer/HTTP boundaries using original CC0 specimens. Cross-compilation and fake processes do not qualify a real desktop or optimizer version. Standard packages do not bundle either optimizer.

An open-source project from Wolfsblvt Works.

## Local agent integration

```sh
smolpng skill install --target agents       # preview only
smolpng skill install --target agents --apply
smolpng skill status --json
smolpng skill package --target chatgpt --output smolpng-skill.zip
```

The embedded instructions follow the installed CLI. Destinations and exact hashes are tracked; synchronization never silently overwrites local edits. Custom paths and instruction-only ChatGPT packaging are supported. See the [Agent Skill guide](docs/agent-skill.md) for version replacement, safe updates and the local-runtime boundary.

## License

Application and ordinary docs: [MIT](LICENSE). Generated corpus: CC0. Dependencies and future website/brand terms remain separate in the [license map](license-map.md).
