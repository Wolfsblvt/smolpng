# Privacy and data

## Meaning

The current CLI has no telemetry or remote compression. This document describes actual input transport, temporary files and reports, not a promise that metadata stripping removes every private field.

Local files and stdin are processed locally. Explicit HTTP(S) inputs are downloaded from the supplied source and permitted redirects. Query parameters are sent but omitted from report labels and request-error messages. Auto never uploads and no TinyPNG credential is retrieved.

Original-derived candidates live in a private temporary directory under the application cache and are removed on ordinary completion/cancellation. A forced crash may leave a directory. Outputs persist until removed by their owner. Reports go to stdout unless a report path is explicitly selected, and may disclose local paths and optimizer versions/arguments. Safe PNG metadata removes textual/time annotations while retaining color/physical/orientation data; it is not a complete EXIF scrubber.

No persistent last-result state or native clipboard writes exist in this cut. Those features must update retention and deletion guidance when implemented. Secrets never belong in presets, reports or config files.

## Configuration and installed presets

User/project TOML files are read locally. App imports persist in the user-data `presets.toml`, with no network calls and no global environment changes. Export is explicit. `config show`/provenance can reveal local paths and user-supplied descriptive text. Shareable presets do not admit secrets or executable paths; a project/preset cannot upgrade a personal remote prohibition. Pinning and removal update only app-owned presets.

## Agent Skill state

Explicit Skill management persists local destinations, app/bundle versions and exact resource hashes under the state directory. It stores no credentials and makes no network request. `skill status` and `doctor` inspect without creating state. Directory transactions can retain a journal or verified backup after an interrupted write; changed backup bytes are preserved rather than deleted. A future human-startup observer records last-seen version/bundle and ignored-bundle preference; no deterministic command displays a synchronization prompt. The exported instruction ZIP contains guidance and a license, not a local binary or private state.


## Human and desktop state

Interactive startup explicitly reads the desktop clipboard to detect a useful source or preset. Unrelated text is not displayed as an invented filename. A copied image/file representation remains the input throughout recipe retries, even after the output is written to the clipboard. Smart copy writes PNG bytes, file references and path text; batch copy writes files/path text. Native platform behavior and persistence are qualification-dependent.

`state/setup.json` (or the platform's corresponding state directory) records only schema and first-use acknowledgement. `last-result.json` retains one human/drop operation's report, timestamp and warnings, bounded to 4MiB; it includes file paths and size/recipe information but no image bytes. A later human/drop operation replaces it. `smolpng last --clear` removes the record, never the image artifacts. Source images and outputs have their own independent filesystem lifetimes. Last-result copy verifies file availability again.

Deterministic compression does not copy by default or retain human history. Explicit `--copy` is available; Linux owners may need `--clipboard-hold` to remain alive until paste. Background `drop` defaults to at most five minutes of ownership on Linux, ending earlier if replaced or cancelled; JSON mode does not wait unless the duration was explicitly supplied. Notifications contain only a generic result count/exit status, not source paths. Output-copy failures do not delete successfully produced image files.
