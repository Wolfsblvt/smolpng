# Native clipboard interoperability, 22 October 2026

## Meaning

The existing clipboard adapters now have direct protocol/API consumer evidence on X11, Wayland and native macOS amd64/arm64. Qualification exposed and repaired Linux text/file input being mistaken for PNG. All selected journeys passed with a typed native macOS producer; its earlier JavaScript producer failures did not establish an application defect, and the speculative macOS fallback was withdrawn. This record preserves those distinctions and exact host/source identities. It does not qualify rendered paste, file pickers, drag/drop, notifications or the joined private candidate.

## Exact source and native hosts

[Run 39690082395](https://github.com/Wolfsblvt/smolpng/actions/runs/39690082395) passed all four jobs at application/journey source `283a76aeee8df5f69b9f2fd70bc3a0855f671cbb`, for [Issue #61](https://github.com/Wolfsblvt/smolpng/issues/61) and [PR #62](https://github.com/Wolfsblvt/smolpng/pull/62). Each host built the source-bound application and explicitly tagged Go test executable with Go 1.26.6 and CGO disabled. The native CLI checked its expected commit before input work. The preserving replay onto current main changed only Decisions/Direction/Releasing, retaining the owner-selected release sequence and JPEG disposition; executable and journey source bytes stayed unchanged.

| Native route | Host observation | Independent consumer/producer | Application SHA-256 |
| --- | --- | --- | --- |
| X11 amd64 | Ubuntu 24.04.5; Linux 6.17.0-1022-azure | Xvfb 21.1.12, xclip 0.13-3 | `c34ea488223a94fc7ff51507c9dd987c57bd818c61172a84075e88930e2eb95a` |
| Wayland amd64 | Ubuntu 24.04.5; Linux 7.0.0-1017-azure | Sway 1.9-1build2, wl-clipboard 2.2.1-1build1 | `c34ea488223a94fc7ff51507c9dd987c57bd818c61172a84075e88930e2eb95a` |
| macOS amd64 | macOS 26.6.1, build 25G76; Darwin 25.6.0 | AppKit, Apple Swift 6.3.3 | `7ed9df69385d7ecf049b2fc482dbedd386db1a94cbf99ca6316913f6f3ee37a6` |
| macOS arm64 | macOS 26.6.2, build 25G83; Darwin 25.6.0 | AppKit, Apple Swift 6.3.3 | `24c50e5e823845c324413d6ba471357f787b184a07b1de1464a6815e69f65b99` |

X11 ran in its own offscreen X server with Wayland unset. Wayland used a headless compositor, disabled XWayland and left DISPLAY unset, so a successful clipboard read could not silently use the X11 bridge. Their kernels were different; matching Linux application/test hashes establish identical built bytes, not identical hosts. macOS used the actual native architecture on each host. Its 62-line Swift/AppKit consumer is qualification-only and compiled with the existing host toolchain; no Swift application runtime or new Go dependency is selected.

The [retained evidence](2026-10-22-native-clipboard/report.json) joins captured reports, complete test logs, source/toolchain/kernel/OS/transport observations and application/test/consumer hashes. Original UTF-8 payload values and their SHA-256 digests are retained inside the authored JSON account. Its SHA-256 is `88c4037a46c13564883fedfc6be9847b21c7b3a450b6f561e756c88c23442f2c`. This outlives the CI artifacts' 14-day transport lifetime.

## Observed behavior

Every host passed smart single-image PNG/file-list/path output and batch file-list/path output without a misleading image representation; ordered local files and exact path text survived spaces, Unicode, percent and hash characters. PNG consumers preserved the fixture's 64×64 RGBA/alpha observations. Foreign native PNG, file-list and path-text offers each reached the real CLI and committed the expected ordered outputs. Linux kept the CLI alive to serve smart data until an independent owner replaced it, then exited successfully. macOS retained image and file URL data after the CLI writer exited.

The generated PNG already fit the explicit 1 MiB target. These are clipboard, source resolution and output-commit observations, not optimizer or JPEG qualification. Their PNG hash is `330d3e88ec6ae93d58fc5bfc616cc5866ba3ddda66c9662f969f261e8780d66b`. Native file-URL reads on macOS may return decomposed Unicode filesystem spelling; the test checks the same file and order there while retaining a separate exact path-text check.

## Reproduced defect and corrected hypotheses

At `76c0fa3` / [39687136462](https://github.com/Wolfsblvt/smolpng/actions/runs/39687136462), X11 smart output and owner lifetime passed, but xclip file/text input was treated as a clipboard image and refused. Diagnostic `0df901b` / [39687953677](https://github.com/Wolfsblvt/smolpng/actions/runs/39687953677) retained the CLI's JSON: non-image payloads of 193/154 bytes reached image decoding and failed as “not a PNG image.” Some X11 owners answer unadvertised targets with their payload; the reader had requested image bytes without checking offered formats. Linux now consults the existing format API before its image request, retaining the old fallback when enumeration is unavailable. Image preference when offered, sensitivity refusal and Windows/macOS input behavior are preserved. Both X11 and Wayland passed at first repair source `28a6901` and in the final four-host run.

The first macOS failure compared canonically equivalent path spellings too strictly. After that test correction, an apparent modern-file input failure prompted a native fallback, but the added independent producer readback demonstrated that the JavaScript writer published zero readable URL items. That disproved the diagnosis. The fallback and direct-dependency change were withdrawn before acceptance. Explicit JavaScript item publication at `f372296` also failed its own receipt. A typed Swift/AppKit producer then passed its immediate receipt, an independent process readback and the real CLI with the original macOS reader. No macOS application repair survives. Retained attempt logs preserve the false inference and correction rather than rewriting earlier red runs as successful evidence.

## Remaining boundary

Ordinary root tests exclude the `nativeclipboard` build tag. [Development](../DEVELOPMENT.md#hosted-verification-source) owns the selected command and disposable-host setup. Root verification and final-head hosted results remain their own evidence; this record binds the native run above. There was no Wolf workstation clipboard, desktop control, personal content, remote image service, signing, installer activation, publication or new private candidate. A headless/API consumer is not a rendered application paste or ordinary human use. Those accepted remainders stay with Issue #36 and standing product direction.
