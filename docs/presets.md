# Configuration and presets

## Meaning

This is the implemented TOML contract for user/project settings, data-only inherited presets, provenance, preview/import and export. It owns file placement and persistence semantics. Menu and clipboard preferences are stored for the human workflow; they do not cause implicit clipboard writes or desktop integration from `compress`. PNG and JPEG format filters are enforced against the actual input format.

## Configuration sources and precedence

`config path --json` shows the actual paths. User `config.toml` follows OS-standard/portable paths. `--config FILE` overrides it; `SMOLPNG_CONFIG` is the environment equivalent. An explicitly selected missing file is an error. `--home DIR`, `SMOLPNG_HOME`, `--portable`, or the `smolpng.portable` marker select portable storage without altering the OS Downloads directory.

The nearest `.smolpng.toml` at or above the invocation directory is the project config; search stops at the current Git worktree root. `--no-project` disables it. Paths in a global `[output]` table are relative to that table's file, while a preset's relative output directory is relative to the invocation directory.

Resolution starts with the selected root built-in, applies user defaults then project defaults, then the inherited named-preset chain, then CLI overrides. Built-in flavor selection does not erase user defaults for unspecified properties. `[png.auto]` supplies Auto-specific defaults. An explicit false/zero is an override, not absence. A lossless-derived recipe cannot turn quantization back on; extend Auto instead.

`config show` displays loaded source documents. `config show --effective --preset NAME --provenance` shows resolved fields and each owning file and TOML key. `config validate` validates all definitions and the selected default. `config init` creates a minimal user file; `config init --project` creates a project file. Neither replaces an existing file. Version and path diagnostics still work when configuration is invalid.

## User settings

```toml
config_version = 1

[defaults]
preset = "auto"
remote_policy = "never"
quality = "70-95"
recursive = false
jobs = 2

[output]
local_files = "source"
detached = "{downloads}/smolpng"
template = "{stem}.smol{ext}"
collision = "increment"

[clipboard]
single = "smart"
batch = "files"

[drop]
mode = "background"
preset = "auto"
notify = true

[ui]
header = "cute"
animations = true

[updates]
check = "manual"
channel = "stable"
```

`[tools]` may contain `pngquant` and `oxipng` executable paths in user configuration only. Automatically discovered project files cannot choose executables. Explicit CLI tool paths remain the highest-precedence selection. `output.directory` selects a directory for all sources; omit it for source siblings and detached Downloads. Explicit `--output`, stdout or in-place replaces an inherited directory, never becomes ambiguous with it.

Unknown fields and TOML type mismatches fail rather than being ignored. This candidate accepts the implemented fields, not every illustrative spelling from the historical design conversation. Integer palette limits and OxiPNG levels use integers; omit a palette field for its inherited value. User `[updates] check = "daily"` enables rate-limited checks only in human startup and only with a configured trusted feed. Default is `manual`; project configs and imported presets cannot enable update network behavior.

## Preset fields

Each `[presets.NAME]` requires `schema = 1`. Names use lowercase letters, digits, underscores and hyphens, begin with a letter, and are at most 64 characters. `auto`, `web`, `smallest` and `lossless` are immutable/reserved.

| Field | Meaning |
| --- | --- |
| `label`, `description`, `author`, `source` | Display/provenance text, retained during normal export |
| `extends` | One built-in or named parent; absent means Auto |
| `formats` | `['png']` or an accepted future JPEG restriction |
| `quality`, `min_quality` | MIN-MAX and optional minimum override |
| `colors`, `speed`, `dither`, `oxipng_level` | Palette 2..256, speed 1..11, auto/none/0..1, level 0..6 |
| `metadata`, `quantize`, `optimize_alpha` | keep/safe/all, lossy permission, hidden-RGB tail cleanup |
| `target_size`, `accept_nearest` | Exact byte target; empty target string clears inheritance |
| `allow_resize`, `resize`, `max_dimensions`, `min_dimensions` | Explicit transform/fallback controls; empty dimension string clears a transform |
| `engine`, `remote` | Requested provider behavior, constrained by user permission |
| `template`, `output_directory` | Output filename and portable relative directory |
| `clipboard` | none, smart, path, image, file/files, directory preference |
| `recursive`, `jobs`, `timeout` | Batch traversal, 1..32 workers, Go duration (`0s` disables) |
| `[presets.NAME.menu]` | `pinned`, `position` 1..1000, and `hidden` |

Cycles, unknown parents and chains deeper than 64 are rejected. Project definitions replace same-named user definitions as a whole; they do not silently merge two independently authored recipes. Parent menu/display metadata is not inherited. Pinned entries use deterministic name tie-breaking; Auto always stays first. Unpinned recipes remain in the saved-recipe list.

The examples directory contains [MCP icons](../presets/examples/mcp-icon.toml), [pixel art](../presets/examples/pixel-art.toml), [screenshots](../presets/examples/screenshot.toml) and [archival cleanup](../presets/examples/archive.toml). They are CC0 and parsed/resolved by the tests.

## Safe clipboard or file import

```sh
smolpng preset import recipe.toml
smolpng preset import recipe.toml --apply
smolpng preset import --clipboard
smolpng preset import --clipboard --apply
smolpng preset import - --apply
```

Without `--apply` this is a read-only preview. A single fenced TOML block is also accepted. Actual image/file clipboard representations are left untouched when reading the text. Native clipboard availability is host-dependent and not supplied by this cut's injected interface.

Imports only admit a `presets` table. Commands/hooks, credentials, environment, raw optimizer arguments, executable paths, absolute output paths and traversal cannot appear in the schema. Limits are 1MiB per document, 256 presets per document and bounded inheritance. A remote recipe can be imported and inspected but cannot upload under `remote_policy = 'never'`. A project cannot increase the user's remote grant. An explicit CLI remote-policy choice is the deliberate invocation grant.

App imports live in a separate user-data `presets.toml`. This avoids reformatting comments or hand-authored settings in `config.toml`. Name collisions default to `name-2`; intra-bundle parent references are remapped together. `--replace` previews before/after definitions and may replace only app-managed entries. File-owned definitions are edited in their own source. Persisted imports cannot depend on a project-only parent that would disappear when the user changes directories.

A store lock and snapshot comparison prevent replacing a changed preview. A stale lock is reported with its exact path, not automatically deleted. Commit uses an atomic write. No partial import occurs. Hand-edited store content can be loaded, but changed bytes during preview/apply cause a retry instead of lost edits.

## Export and management

```sh
smolpng preset list --json
smolpng preset show mcp-icon --json
smolpng preset export mcp-icon
smolpng preset export mcp-icon --resolved --output frozen.toml
smolpng preset export auto --resolved --name my-auto
smolpng preset pin mcp-icon
smolpng preset move mcp-icon --position 2
smolpng preset unpin mcp-icon
smolpng preset remove mcp-icon
```

Normal export includes the non-built-in parent closure and preserves inheritance. Resolved export freezes effective compression settings and clears `extends`. It never embeds a credential/tool path; a destination outside the invocation project cannot be frozen into a portable preset and produces a clear error. `{downloads}` placement remains installation policy, not a portable absolute machine path. Exported output files must not already exist.

Only app-owned entries are mutated by pin/move/remove. Removing a required parent or selected default is refused. No export/import relicenses user-authored material; optional author/source fields are preserved.
