# Recipes and configuration

Built-ins: `auto` (recommended local default), `web` (balanced), `smallest` (stronger quantization), `lossless` (pixel-identical, including hidden RGB beneath transparency). Exact lossy bytes depend on upstream versions. Recipe ranking is not a measured visual-quality score.

JPEG quality ceilings are 85 / 80 / 70 for Auto / Web balanced / Maximum squeeze; PNG settings are independent. `--quality MIN-MAX` overrides both formats; `--jpeg-quality MIN-MAX` or TOML `jpeg_quality` specializes JPEG. Within a layer the JPEG-specific range wins; later shared overrides still win over earlier layers. Inspect `recipe.jpegQuality` and provenance, and preserve both ranges in resolved presets. Use Lossless for pixel identity rather than assuming an ordinary lossy preset promises it.

Precedence is built-in defaults, per-user configuration, nearest project `.smolpng.toml`, inherited named preset, then explicit CLI fields. Inspect `config show --effective --provenance` instead of guessing why an output path or option differs.

A reusable recipe:

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

[presets.mcp-icon.menu]
pinned = true
position = 2
```

Review the resizing permission with the user before importing this example. `preset import FILE` previews definitions, changes, warnings and collision renames. `--apply` commits. `--replace` only replaces app-managed collisions and does not silently rewrite a project's or user's handwritten config. Normal export includes inheritance dependencies; `--resolved` freezes effective values. Imported recipes cannot contain commands, hooks, arbitrary environment variables, credential fields, executable paths or raw optimizer arguments.

A recipe can request a remote provider but cannot independently enable uploads. The installed build's capability and the user's explicit remote permission must both exist. CLI `--remote-policy never` makes a local-only invocation explicit; do not change it just to make a blocked recipe run.
