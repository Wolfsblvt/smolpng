# Windows PNG clipboard transparency, 19 October 2026

## Meaning

A real Paint paste exposed lost transparency in smolpng's Windows bitmap representation. Preferring encoded PNG while retaining the native bitmap fallback repaired that consumer path. The saved repaired paste matches the copied PNG composited over Paint's white canvas exactly. This record qualifies one Windows image consumer, not file-list/text paste or other hosts.

## Failing candidate and control

Private `0.3.0-dev.8`, source `d03ce5b8568d50d8ccad1a9cb429f4efbfd1efb6`, passed all 63 packaged native compression/provenance checks before this graphical test. Its Windows application SHA-256 is `493f5804a953f219e3947e147e5c038e67d0e42e39a6fd7f75e10dabe6b687a0`.

The host was Windows 11 Home/Core, 25H2, kernel `10.0.26200`, revision `9711`; Paint reports file/product version `11.2605.81.0`. smolpng compressed a disposable copy of the generated CC0 64×64 alpha-gradient PNG in a portable home, with remote policy `never`, notifications disabled and no project discovery. Original/config hashes stayed unchanged. The retained PNG SHA-256 is `308a21f5503e67d8fc1382526ecaa90368121899106eccd629fb8f27b564a378`.

Wolf authorized fixture-only clipboard replacement. Public `last --copy smart --json` published image, files and path-text together, reported `needsOwner=false` and exited. Paint's Edit → Paste consumed the image afterward. Saving its 407×181 canvas produced `paint-paste.png`, SHA-256 `f1552b04f8b2597bbfb8d73d236a73c5771b92841d97e4efafed082671264b51`.

Pillow 12.3.0 analysis of the top-left 64×64 region found RGB values identical to the copied PNG, but transparent pixels had become opaque. Directly opening the exact PNG in Paint displayed its checkerboard transparency correctly. The compressed file was intact; this isolated a clipboard representation/consumer discrepancy.

## Repair and repeated consumer journey

The existing clipboard library registers `image/png` as Windows' native `PNG` format. Image and Smart copy now offer that encoded PNG first and retain the existing `FmtImage` bitmap second, within one `WriteAll` transaction. File-list/path formats and public copy modes remain intact. No dependency or native API implementation was added.

The regression failed for Image and Smart on the original source, then passed with the repair. Desktop/CLI tests and focused vet passed. Exact fix source `25db365efff1fb0ca656c0959f4db44d42fa8950` was built as private test identity `0.3.0-dev.8-clipboard-alpha`; executable SHA-256 is `2ecbd6c7c7cc745ca50e56b2dcc073d8d89e1e48976d0cd5d8d4edd1557184b7`.

That executable copied the same retained PNG and exited. Paint received it through Edit → Paste onto a new white 407×181 canvas. The saved `paint-paste-fixed.png` has SHA-256 `f8f9d07d2720d5121a9b6e8a81f6b40a07929f2fc44a4900b7a43b909ba20efc`.

| Comparison of the 64×64 pasted region | Original bitmap-only payload | PNG-first repaired payload |
| --- | --- | --- |
| Raw source RGB | Copied unchanged; alpha ignored | Alpha correctly applied |
| Expected source composited over white | Differences up to 255 | Exact; all four channel differences zero |
| Writer exited before paste | Yes | Yes |

This proves alpha compositing in the observed Paint journey and clipboard availability after writer exit. Paint's white canvas naturally makes the saved composite opaque; this is not a claim of a transparent document layer.

## Evidence and remaining boundaries

Original writer/consumer receipts and saved canvas remain under ignored `artifacts/validation/candidate-clipboard-dev8-d03ce5b/`. Fix build, version/copy receipts, saved canvas and independent analysis remain under `artifacts/validation/clipboard-png-alpha-25db365/`. No retained candidate archive was patched.

Computer Use's automatic Paint launch failed by packaged ID, installed executable alias and one private-connection refresh with `accessibility window-opened handler did not become ready`. Wolf opened the empty app once. Existing-window discovery, screenshots, menu actions and saving then worked. An offscreen indexed Save button was refused; the verified Save As destination was committed through its keyboard shortcut. The owner-opened window was left open with the saved test artifact. Automatic app launch and absence of focus effects are not qualified.

No unrelated clipboard was read and no personal image was used. File-dialog listings are excluded from public evidence. The copied file and ordinary last-result checkpoint survive clipboard replacement or a vanished notification. File-list/text consumers, sensitive-clipboard refusal, physical Explorer dragging, notifications and Linux/macOS clipboard remain separate observations in [the candidate journey](https://github.com/Wolfsblvt/smolpng/issues/36). Public release, production signing/feed operation and the complete 1.0 destination remain open.
