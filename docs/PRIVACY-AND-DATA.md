# Privacy and data

## Meaning

smolpng has no telemetry, analytics or operated image-hosting service. Local Auto remains local under the initial remote policy. Explicit URL input downloads from its source, explicit TinyPNG policies upload an image, and explicitly enabled update/dependency operations contact their selected servers. These are separate permissions.

## Image data

Local inputs, stdin and clipboard images become private temporary original-derived copies. Temporary candidates are cleaned after normal completion and cancellation; forced process/machine termination can leave private cache residue. Outputs persist at the selected paths. Only explicit in-place mode replaces an original. PNG/JPEG safe metadata policies retain rendering interpretation; JPEG safe removes private EXIF/GPS by reconstructing an orientation-only record.

On desktop backends that expose a sensitive-content marker, smolpng checks that marker before reading image, file or text representations. Marked content is refused, so password-manager copies do not enter automatic source discovery, preset import or explicit clipboard compression. Platforms without a readable marker retain ordinary clipboard behavior without claiming this protection.

Safe/all also remove opaque data appended after a JPEG's primary still image, with a warning that attached depth/refocus, motion or other auxiliary content does not survive. The attachment is removed before local optimization or an explicitly permitted remote upload. Keep preserves these compound files unchanged. See [JPEG metadata and orientation](jpeg.md#metadata-and-orientation) for target and geometry consequences.

TinyPNG receives metadata-filtered input only after the selected user/invocation policy permits it. Keep-all metadata, unsupported color/orientation and remote geometry combinations are refused rather than silently discarding preservation promises. `remoteUsed` reports attempted uploads even if the result is not selected or the response is lost. Retention after upload is controlled by the service's current terms, not by smolpng. No billable request is automatically retried. API keys are read from process environment or native keyring and never included in reports or optimizer process environments.

## Reports, state and installed integrations

Reports may include source/artifact paths, dimensions, sizes, versions, non-secret tool argument arrays and warnings. URL query parameters are used for explicit downloads but omitted from source labels. A report file is written only when selected. Human/drop workflows retain last-result metadata; `last --clear` deletes that metadata, not output images.

Preset, dependency, Skill and shortcut registries live in per-user state/data locations. They retain destinations and exact file hashes needed to protect edits. They are not synchronized to a cloud service. Clipboard smart publishing can expose file/path/image representations to the desktop and its clipboard history manager. A tagged JPEG may expose only file/path representations rather than a miscolored conversion.

## Update data

Updates are manual by default. Only a user-enabled daily human-startup policy can perform a periodic request, and only when a trusted feed/key is configured. Projects, copied presets, deterministic CLI, JSON and background drop cannot opt users into those checks. A check records the last attempt time and ignored version. No device identifier or API key is sent.

Standalone updates retain signed payloads, old/new file hashes, private backups, completion/rollback status and a bounded worker log in a transaction directory beside the application. That evidence is intentionally retained for recovery. Config, user images, managed optimizers and custom Skills are not package members. Public keys are not secrets; private signing keys never enter normal application state or release archives.

## Right-click menus

Menu settings contain application/config paths and preset identifiers, not credentials or images. Actions are deliberately local-only. Windows hands at most 512 selected absolute filenames through a bounded native-user `shell-requests` file; the Go receiver deletes a decoded request. A crash or invalid request can leave filenames there until manually cleared. Menu receipts and package-registration receipts persist until removed. Removing menu settings hides actions without silently uninstalling the separate signed Windows package. The context-menu guide distinguishes configured files, registered package identity and actually observed desktop behavior.
