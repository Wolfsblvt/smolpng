# Optional TinyPNG

## Meaning

This guide owns the optional TinyPNG upload policy, credential lookup and remote-result contract. TinyPNG is implemented alongside local compression, but a stored credential never grants permission to upload. Ordinary tests substitute HTTP and do not contact the paid service.

TinyPNG is now an implemented remote adapter, not a requirement for local compression. No real service call is part of ordinary tests. Public API contract consulted 2026-09-08: https://tinify.com/developers/reference/http . It uses HTTPS Basic authentication, a binary POST to `/shrink`, a provider output URL, and the `Compression-Count` response header.

```
smolpng providers login tinypng
smolpng providers status tinypng --json
smolpng compress icon.png --engine tinypng --remote-policy always --json
smolpng compress icon.png --target-size 10KiB --remote-policy fallback --json
smolpng providers logout tinypng
```

Login is masked and uses the native keyring. Automation can use `providers login tinypng --stdin`, or supply `TINIFY_API_KEY` to the process. There is no API-key argument, plaintext config storage, or persistent environment mutation. A present but invalid environment value fails rather than silently using the keyring. Status does not contact the service. Logout removes only the stored key, not an environment variable owned by the parent process.

Default `never` and deterministic `ask` run locally without looking up credentials. `always` or an explicitly permitted TinyPNG engine uploads once. `compare` evaluates a remote candidate as well as local results. `fallback` uploads only after local target miss or optimizer unavailability/failure, not after invalid input. Imported/project presets cannot grant upload permission. Remote failures return explicit errors rather than secretly retrying billable requests.

Uploaded input is metadata-filtered according to the selected policy. `keep` is rejected before upload because the service cannot promise arbitrary metadata preservation. Results must preserve format, dimensions and transparency, and satisfy the same no-growth and target rules. TinyPNG's quality is service-controlled; local numeric quality floors are not a service guarantee. Automatic remote resize search is deliberately absent to avoid multiplying quota consumption; explicit local resize and remote use are not combined in this cut.

`remoteUsed` means an upload was attempted, even when a local candidate wins or a response is lost. The optional `remote` object reports attempted/selected state, credential source and reported monthly compression count. It never contains the credential or provider output URL. No automatic retry occurs after a potentially accepted upload. Only exact HTTPS `api.tinify.com/output/...` locations are accepted, and HTTP redirects cannot forward the Authorization header.

The native adapter is go-keyring v0.2.8 (MIT, with retained file-specific Apache notice). Windows uses Credential Manager, Linux Secret Service, and macOS `/usr/bin/security`; the pinned macOS setter passes secret material through stdin, not process arguments. Their actual unlocked-session behavior remains a native qualification task. Headless or another-user callers may require process environment credentials.
