# Project map

## Meaning

This map locates the current executable, image contracts, tests and documentation. It describes existing ownership, not an inventory of empty future packages.

| Path | Responsibility |
| --- | --- |
| `cmd/smolpng` | Process entry and cancellation |
| `internal/cli` | Commands, flags, reporting |
| `internal/domain` | Recipes, limits, results, size parsing, exit codes |
| `internal/input` | Source planning, traversal, HTTP/stdin, clipboard interface |
| `internal/output`, `internal/fsutil` | Names, collision/original protection, atomic writes |
| `internal/pngimage` | Validation, metadata, pixels and resizing |
| `internal/workflow` | Candidate and batch execution |
| `internal/process`, `internal/toolchain` | Direct invocation and executable discovery |
| `internal/paths` | Native user and portable locations |
| `internal/testutil`, `tools/corpus`, `testdata/corpus` | Original fixtures and boundary doubles |
| `docs`, `LICENSES` | Maintained behavior/destination and rights |

Start with Development for commands, Architecture for invariants, and Vision for the complete destination. Generated output belongs under ignored `artifacts/`.

`internal/config/` owns TOML and preset resolution/import state, `docs/presets.md` owns its user contract, and `presets/examples/` contains tested CC0 copyable recipes.

`skills/smolpng/` is the embedded user-facing guidance; `internal/skill/` owns its tracked installation, synchronization and recovery. `docs/agent-skill.md` owns the lifecycle and client/runtime distinction.


`internal/ui` owns keyboard and typed-name presentation; `internal/desktop` owns native clipboard and OS actions; `internal/state` owns bounded human metadata. `cmd/smolpngw` is Windows's explicit no-console drop entry point. `internal/config/settings.go` patches supported scalar TOML values without discarding user comments. The [documentation index](README.md) routes the now-expanded maintained subject set.
