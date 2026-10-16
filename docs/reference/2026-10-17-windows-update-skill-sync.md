# Native Windows update Skill synchronization, 17 October 2026

## Meaning

A real signed local Windows update refreshed an unchanged managed Agent Skill, preserved a locally edited managed installation byte-for-byte, and left an unchanged custom installation awaiting consent. This record qualifies the automatic post-update synchronization path with actual tracked directories and a changed embedded bundle. It complements the separate [executable recovery record](2026-10-17-windows-update-recovery.md); neither record claims a production release or a running agent client's reload.

## Exact specimen

Windows 11 amd64, Go 1.27.2, CGO disabled, standalone builds with `-trimpath -buildvcs=false`. Old fixture 1.0.0 was built from `1997672577166a3a1ce9cc67fa5d3f9549d6072c`; new fixture 1.0.1 reused the verified executables at `3453a38e470d75d5d3ac78102c86eac708c3e32d` from the recovery qualification. The new executable hashes are recorded there. Their embedded Skill bundle hashes differ:

- Old: `6b3f6e7c64a0648c283c9d4ec57a12d6b4389ec112f371b65eb99910e9ba7b82`.
- New: `b34a5fcecdf978e8fb6b2b647cc0c1feec2a58580474ebc0a1f12f7b182a477a`.

The installation, application state and user profile were new disposable directories. `USERPROFILE`, `HOME`, `APPDATA`, `LOCALAPPDATA` and `SMOLPNG_HOME` were explicitly redirected for application processes. The ordinary `skill install --apply` command created three tracked destinations: shared personal, Claude personal, and an explicit custom parent. All destination and registry paths were checked to remain inside that fixture. An extra comment was then appended to the Claude copy's `SKILL.md`.

The local archive contained only the new application executables and license. An ephemeral test Ed25519 key remained in process memory; only public trust and the signed envelope were retained. No production key, feed, upload, real user installation or actual harness directory was used.

## Observed result

The ordinary old executable's `update apply` returned `scheduled` and `completed: false`. Its original helper was observed to exit, and the journal reached `complete`. The new executable's ordinary `skill status` and independent directory hashing established:

| Tracked installation | Journal action | Observed bytes |
| --- | --- | --- |
| Untouched managed shared-personal copy | `synced` | Every file matched the new embedded bundle's resource hashes. |
| Edited managed Claude-personal copy | `preserved-modified` | Every file matched its pre-update hash, including the local edit. |
| Untouched custom copy | `needs-consent` | Every file matched its pre-update hash. |

Both installed executable hashes matched the signed new members. Config, preset and independent optimizer-marker sentinels remained byte-exact. The complete receipt and command outputs are retained locally under `artifacts/validation/windows-update-skills-3453a38/`.

The first observer requested a nonexistent `bundle.hashes` field; the actual JSON contract uses `bundle.files`. The transaction had already completed and its journal/status were captured. Correcting that reader and rehashing the same existing files passed all assertions without repeating the update. The receipt retains the original observer failure and correction.

This qualifies actual filesystem synchronization by the updater, not a real agent loading the updated Skill, custom-sync consent acceptance, locked Skill directories, other operating systems or production distribution. It does not broaden the earlier empty-registry recovery fixture into a Skill-preservation claim.
