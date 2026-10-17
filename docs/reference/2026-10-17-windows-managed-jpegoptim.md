# Windows managed jpegoptim qualification, 17 October 2026

## Meaning

This record qualifies the real Windows x64 per-user jpegoptim lifecycle added by source `94811aa5222b9afed0b15f29eefb81349ac6455c`: preview, upstream download, hash verification, retained license, managed preference/discovery, local JPEG compression and removal. It uses a disposable application home and a CC0 corpus photo. It is not keyboard-interface, graphical, other-host, package-publication or production-signing evidence.

## Exact specimen

Go 1.27.2 built the Windows CLI from that clean commit. The executable SHA-256 is `040c2efdb5ae0371752c4d6de23b4eb47ec8203c4475eb7e071f65e634d07144`; Go build metadata reports the source revision above and `vcs.modified=false`.

The pinned [upstream jpegoptim 1.5.6 release](https://github.com/tjko/jpegoptim/releases/tag/v1.5.6) supplies the Windows x64 ZIP, SHA-256 `db2d8caee88f2665b772c4591cba223b8b729500b88266dc06d4bf79d28fe11a`. Its executable hashes to `0a79681b56961ec2c5f45c4f9dff01a64e7274b4f3250c22cde5f09c0bb457f3`. The complete upstream `COPYRIGHT`, two separating newlines and `LICENSE` are retained as a 35,871-byte installed notice, SHA-256 `2d52c5a99b196c919a73fb7510c75154f84ea676115eef48af68d86ded6954a8`. The selected license is GPL-3.0-or-later, not OxiPNG's MIT license.

## Real lifecycle

Every invocation selected the disposable home explicitly and suppressed project configuration. Each command's stdout, stderr and exit status was captured before dependent checks. Noninteractive processes used Windows no-console creation.

1. `deps install jpegoptim --json` returned a preview and left the displayed destination absent.
2. `deps install jpegoptim --apply --json` downloaded the real upstream archive, passed the pinned digest and exact `v1.5.6` version checks, and reported the installation current.
3. The retained notice contained the upstream copyright and full GPL version 3 terms.
4. `deps use-managed --apply --json`, followed by `deps status --json`, reported the actual installed path, version 1.5.6 and source `managed`.
5. `compress original.jpg --preset auto --copy none --json` used jpegoptim 1.5.6, reported `remoteUsed=false`, preserved the original hash and committed a smaller output.
6. `deps remove jpegoptim --json` left the installed executable present. Repeating with `--apply` removed it while preserving both image files.

The image is Aammell's CC0 landscape from the [corpus manifest](../../testdata/corpus/manifest.toml). The original is 399,226 bytes, SHA-256 `d266b5bfa1179a6298f1c084e8c539ef81698643f7c45d9adb34d43a30ed2758`. Auto produced 276,012 bytes, SHA-256 `fa37021bb3c7e1df719120575255f491e30fa04183d08fa8c315f3fdd9397e7e`: a 30.9% saving for this particular photo. This is not a universal saving or pixel-identity claim.

## Supporting source evidence and limits

The new regression first failed because jpegoptim had no managed artifact. It then passed across installation, retained notices, managed discovery and removal. Missing copyright/license cases refuse before executing or activating a downloaded program. Catalog tests cover the three x64 host archives, wrong-upstream rejection and absent native arm64 artifacts. The full root Go test suite and `go vet ./...` passed for the candidate source bytes.

The Linux and macOS x64 ZIPs were independently downloaded, matched to their published SHA-256 values and inspected for the selected executable plus `COPYRIGHT` and `LICENSE`. They were not executed on those hosts. The keyboard setup menu is unchanged and still offers its existing managed OxiPNG route; managed JPEG setup is exposed through the CLI. No personal configuration, system tool, PATH, clipboard, website/UI source, credential or remote quota was changed by this qualification.
