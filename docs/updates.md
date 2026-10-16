# Standalone updates

## Meaning

smolpng now implements signature-verified standalone update transactions, a short-lived replacement helper, guarded rollback, package-manager ownership detection and opt-in daily human checks. The source has no deployed release feed or production signing authority. No candidate archive is an Authenticode-signed or notarized release merely because its update manifest can be cryptographically signed.

## Commands and trust

```sh
smolpng update status --json
smolpng update check --manifest release.signed.json --trust-key release-public-keys.json --json
smolpng update apply --manifest release.signed.json --trust-key release-public-keys.json --archive update.zip --json
smolpng update recover --stage /exact/transaction/folder --trust-key release-public-keys.json
smolpng update recover --stage /exact/transaction/folder --trust-key release-public-keys.json --apply
```

Check verifies the signed manifest and reports version/platform availability without changing installed files. Apply is explicit authorization to stage and replace the selected application files. Its return says **scheduled**, not completed. A trusted copy of the old executable waits for the invoking process to exit, then performs replacement. `update status` and the exact journal/result path expose completion or recovery state.

`recover` previews the recorded stage. Explicit `recover --apply` also returns **scheduled**, with the recovery helper PID, log path and journal path; it does not restore files inside the running installed executable. A verified copy of the invoking old or new application waits for that process to exit, then performs the guarded rollback. Each recovery owns a fresh worker directory so the earlier update helper and logs remain intact. The journal moves through `rolling-back` to `rolled-back` on success; `recovery-required` plus an error records an unsuccessful restore. A prior `complete` phase describes the earlier update, not a completed rollback. Preserve the returned paths and inspect the journal or `update status` before claiming recovery finished.

The trust file contains public keys, never signing secrets:

```json
{"keys":{"release-key-id":"BASE64_ED25519_PUBLIC_KEY"}}
```

The shown key is a placeholder, not a usable trust root. A release operator must independently establish the production public key. Build variables `internal/cli.TrustedUpdateKeys` and `UpdateManifestURL` can bind reviewed public keys and a qualified HTTPS feed. With neither configured, automatic network checks do not run and unsigned update application is refused. There is no skip-verification option. Explicit public-key files are local operator trust decisions, not a substitute for authenticating a downloaded key.

## Integrity contract

The envelope signs the exact base64-decoded payload with Ed25519. The payload binds product, stable semantic version, source commit, validity interval, platforms, archive size/hash and every package-member size/hash. Unknown or expired manifests, downgrades, missing platform files, unsafe paths, duplicate members, links, unexpected files and excessive archive expansion are refused. Only credential-free HTTPS release resources are fetched; archive and manifest downloads send no API credential. A local archive undergoes the same signature, size and hash verification.

Update archives are intentionally narrower than full development packages: executables and required license notices only, at allowlisted paths. Config, presets, state, images, installed Skills and managed optimizer directories are not archive targets. The private stage stores synced new files and exact prior backups before an installed file is touched.

## Ownership and transaction

Source and unknown builds cannot self-update. Recognized Winget, Scoop, Chocolatey, Homebrew, system-package and WindowsApps paths override standalone build markers and return their owning update instructions. Standalone application names must be the standard names. An application inside a Git worktree is not an update target.

An OS lock serializes commit/rollback. Before and during publication, each installed file must match this transaction's old or new recorded hash. Complete sibling writes replace files without pre-deleting originals. The main executable is replaced last; its exact version, source commit, platform and standalone ownership are then probed. Failure restores the old files when hashes still agree. Locally modified or later-version files stop recovery rather than being deleted. Journals/backups remain available for explicit recovery instead of being silently discarded.

On Windows the helper can still encounter another running instance, locked file, antivirus scanner or denied directory. It does not kill unrelated processes or elevate. Timeout/failure leaves the journal and may need a later retry. File visibility is atomic; the transaction is recoverable, not a guarantee against power loss or a malicious same-user writer. The first native Windows experiment successfully applied an update but exposed the in-process explicit-rollback failure tracked in [#15](https://github.com/Wolfsblvt/smolpng/issues/15). Qualification of the repaired recovery path and native macOS replacement remain separate evidence.

After a successful app update, the new binary uses its own embedded Skill bundle to synchronize only untouched app-managed installations. Custom paths are offered later by the normal manual-replacement/Skill startup workflow. Modified installations are preserved. A Skill-sync failure is reported without undoing a verified app or overwriting local Skill edits.

## Optional human startup check

User configuration can select `[updates] check = "daily"`; default is `manual`. Project config and shared presets cannot enable update egress. Only the interactive root checks, only with a configured trusted feed, at most once in 24 hours, and never in offline mode. Failed checks also consume that interval. The menu offers Update now, Later, Ignore this version and Disable checks; package-manager installations get their owner instructions instead of self-replacement. Deterministic CLI, JSON output, background drop and status commands do not check implicitly.

## Evidence

Tests use fresh ephemeral test signing keys and generated temporary archives. They cover trust, validity, archive tampering, unsafe members, version ordering, package ownership, staging, successful application, failed identity rollback, interrupted resumption and modified-file protection. They execute no production update, use no production secret and claim no native installer observation. Public feed/signing operation and package publication remain separate grants.
