# Machine results

A schema-version-1 compression envelope has `results`, `summary` and `exitCode`. Each result reports status, input, actual output path (only after successful commit or stream), byte counts, saved percentage, format, recipe, before/after dimensions, pipeline and tool versions, target status, resize use, remote use, warnings, and item exit code.

JSONL emits one result per input. A planning result's `output` is only a proposed path. A failed target normally has no output. With `--accept-nearest`, it has an output but `target.met` remains false and exit code remains 6. Treat the JSON numbers as exact; display rounding must not turn a miss into success.

| Exit | Meaning | Agent action |
| ---: | --- | --- |
| 0 | Successful, unchanged or planned | Distinguish planned from real output |
| 2 | Invalid arguments or contradictory policy | Correct the invocation without weakening permission |
| 3 | Invalid/unavailable input | Report the input problem |
| 4 | Required local dependency missing | Request/setup a local optimizer, not a remote fallback |
| 5 | Optimizer or output failure | Report safe diagnostics and preserve originals |
| 6 | Target missed | Explain the size/quality/dimension constraint; no false success |
| 7 | Network/remote operation failed or unavailable | Do not retry with wider permission |
| 8 | One or more batch items failed | Return existing successful artifacts and enumerate failures |
| 130 | Cancelled | Do not assume a partially started item produced output |

Raw `smolpng exec` returns upstream codes instead; it is not this contract. Files referenced by a result belong to that host. If the harness must upload or copy them, verify they exist there and use its real file-transfer route, not an invented ChatGPT sandbox link.
