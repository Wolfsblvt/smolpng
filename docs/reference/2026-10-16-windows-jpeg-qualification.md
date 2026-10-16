# Native Windows JPEG qualification

## Meaning

The native Windows amd64 build at `b90a5165c7816c7cee3fc8c397872898c058807d` passed 43 JPEG qualification checks after repairing discovery of jpegoptim's actual version banner. This record preserves the failure, exact tool/build identity and observed structural contracts. Synthetic corpus success does not select final photographic quality defaults or qualify real ICC profiles and graphical desktop behavior.

## Discovery failure and repair

The separately downloaded upstream executable reports `jpegoptim v1.5.6  Win64 (Sep 15 2025)`. The previous parser's numeric word boundary skipped the major version after `v`, read `5.6`, and refused the supported tool. The native preflight at `2abae9390167d35dbca942c6a69ec83bca90cfd9` captured this exact refusal before image processing; local evidence is `artifacts/validation/jpeg-2abae93/`.

Discovery now recognizes the prefix and reports normalized `1.5.6`. The existing supported-family rule is unchanged. The actual-banner regression failed before the repair and passed afterward; unprefixed 1.4.7 remains accepted and prefixed 2.0.1 remains rejected. Native Go 1.27.2 toolchain tests and vet passed.

## Exact executable and optimizer

Go 1.27.2 built the standalone Windows amd64 executable with `CGO_ENABLED=0`, `-trimpath`, version `0.3.0-dev.1` and embedded commit `b90a5165c7816c7cee3fc8c397872898c058807d`. Its SHA-256 is `d390e276c475b3b17d47d512fcdcaa932f3b2244a68fba8b1aca4c5006558e9f`.

The [upstream v1.5.6 release](https://github.com/tjko/jpegoptim/releases/tag/v1.5.6) supplied `jpegoptim-1.5.6-x64-windows.zip`, verified against its published SHA-256 `db2d8caee88f2665b772c4591cba223b8b729500b88266dc06d4bf79d28fe11a`. The extracted executable hashes to `0a79681b56961ec2c5f45c4f9dff01a64e7274b4f3250c22cde5f09c0bb457f3`; its accompanying copyright notice identifies GPL-3.0-or-later. The original notices remain beside it locally. Nothing was bundled, added to machine PATH or installed system-wide.

## Observed contracts

A disposable local driver invoked the executable directly with a new isolated home, `--no-project`, explicit tool paths, `--offline` and `--remote-policy never`. It used the thirteen checked-in generated CC0 JPEG specimens. Pillow 12.3.0 independently decoded outputs and checked dimensions, EXIF orientation, safe EXIF field removal and strict RGB pixel equality. Per-check arguments, timings and outputs were saved before moving on; `report.json` has an explicit completed marker. The process exited 0.

| Checks | Result |
| --- | --- |
| Embedded build identity and real optimizer preflight | 2/2 passed |
| Auto across all thirteen JPEG specimens | 13/13 passed |
| Strict lossless across the same specimens | 13/13 passed; independently decoded pixels identical |
| Explicit 12x8 resize across all eight EXIF orientations | 8/8 passed; output orientation normalized |
| Target success, refusal and explicit nearest acceptance | 3/3 passed |
| Strict-lossless orientation normalization and keep-metadata resize refusals | 2/2 passed; no output written |
| Clean JPEG binary stdout and partial batch behavior | 2/2 passed |

jpegoptim appeared in 34 selected output pipelines; the compression reports had no warnings. Thirty-five checked file outputs were progressive JPEGs. Ordinary compression never grew the source and every checked original remained unchanged. Reported output byte counts matched actual artifacts. The 10KiB target passed with 7,763 bytes; a one-byte target refused without output, while explicit nearest acceptance wrote 2,812 bytes and retained exit 6.

The exact local report is `artifacts/validation/jpeg-b90a516/report.json`, with native stdout/stderr and images beside it; the driver hash is recorded there. The final receipt marks all 43 checks passed and complete. [Verify run 38939489484](https://github.com/Wolfsblvt/smolpng/actions/runs/38939489484) passed all eleven hosted jobs on the same source.

## Limits

This run covers structural generated media, not natural photographs, perceptual quality, complete valid ICC profiles, CMYK rendering, every jpegoptim version or other operating systems. Source tests and these direct native runs do not establish graphical clipboard/paste, drop/shortcut execution, shell activation, signing, production update operation or binary publication. The accepted PNG/JPEG product destination remains unchanged.
