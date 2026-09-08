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
