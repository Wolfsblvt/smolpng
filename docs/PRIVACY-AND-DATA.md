# Privacy and data

## Meaning

The current CLI has no telemetry or remote compression. This document describes actual input transport, temporary files and reports, not a promise that metadata stripping removes every private field.

Local files and stdin are processed locally. Explicit HTTP(S) inputs are downloaded from the supplied source and permitted redirects. Query parameters are sent but omitted from report labels and request-error messages. Auto never uploads and no TinyPNG credential is retrieved.

Original-derived candidates live in a private temporary directory under the application cache and are removed on ordinary completion/cancellation. A forced crash may leave a directory. Outputs persist until removed by their owner. Reports go to stdout unless a report path is explicitly selected, and may disclose local paths and optimizer versions/arguments. Safe PNG metadata removes textual/time annotations while retaining color/physical/orientation data; it is not a complete EXIF scrubber.

No persistent last-result state or native clipboard writes exist in this cut. Those features must update retention and deletion guidance when implemented. Secrets never belong in presets, reports or config files.

## Configuration and installed presets

User/project TOML files are read locally. App imports persist in the user-data `presets.toml`, with no network calls and no global environment changes. Export is explicit. `config show`/provenance can reveal local paths and user-supplied descriptive text. Shareable presets do not admit secrets or executable paths; a project/preset cannot upgrade a personal remote prohibition. Pinning and removal update only app-owned presets.
