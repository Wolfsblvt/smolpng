# First local implementation qualification

## Meaning

This dated record preserves the first Windows implementation intake, hosted verification and real PNG optimizer observations. It proves the named source/build paths and measured outcomes, not a released or fully desktop-qualified product.

## Source and build

The delivered nested Git repository was moved intact to `D:\code\smolpng`; delivery wrappers and the recovery bundle remain in ignored `artifacts/intake/2026-10-16/`. Nyxara’s implementation history was replayed onto current GitHub main, retaining Eira’s legal-document naming corrections and repairing package/update/native-staging references. A focused regression checks that updater extraction preserves both notice files.

Go 1.27.2 on Windows amd64 built the native console application at `f70486ceff89008e98a6d79f033472c4b793c94c`, version `0.3.0-dev.1`, standalone distribution. The resulting local `artifacts/bin/smolpng.exe` SHA-256 is `563c967e0a4445f772f6e67567c823044a15deeb1a14f3b63fd36e30dc27073e`. `deps status --json` exercised that executable with an isolated home and explicit optimizer paths.

## Hosted candidate

[Verify run 38828625844](https://github.com/Wolfsblvt/smolpng/actions/runs/38828625844) passed all eleven jobs for exact head `f70486ceff89008e98a6d79f033472c4b793c94c`: root tests, vet and corpus consistency on Windows/Linux/macOS; all six CLI cross-builds; Windows GUI entry builds; Windows shell x64 and ARM64 compilation; x64 C++ protocol execution. ARM64 protocol execution was correctly skipped on the x64 runner.

The preceding run exposed macOS `/var` symlinks in temporary test fixtures. Canonicalizing fixture roots with `filepath.EvalSymlinks` repaired the fixtures without weakening production filesystem checks.

## Real Windows PNG run

The GitHub Windows amd64 artifact at `e0df49bbf7164c8d44704dd911b9d8d9eee834ec` was exercised with the delivered qualification helper. Its executable SHA-256 was `d6aa5d7111c57251521627f6b5999926bb5bb2ba185a350064e08d582aa91bec`. The helper’s local report is `artifacts/validation/png-e0df49b/report.json`; full generated outputs and logs remain local artifacts. This is a predecessor build of the same production implementation, not a native run of the later fixture-only commit.

Separate upstream binaries were pngquant 2.17.0 and OxiPNG 10.2.1. The OxiPNG archive matched the catalog’s SHA-256 `7e940f83ee46874b73f53031f96a15834cb70b220af27391fb06fe7b4dd798e1`; neither optimizer is bundled or committed. The helper used original CC0 specimens, an isolated home, disabled project lookup and remote use, and no clipboard or user integration.

25 of 29 checks passed: provenance, dependency preflight, all 11 Auto images, 8 strict lossless images, exact target success/miss, clean binary stdout and partial batch handling. Lossless alpha-gradient, alpha16 and gray16 refused before optimization with exit 4 and unavailable-optimizer diagnostics. Nearest-target execution exceeded the helper deadline. There is no evidence here of changed decoded pixels in those refused cases.

A timeout in the three-second version probe is a hypothesis, not an established cause. A later native status found both configured tools healthy, and six direct OxiPNG version probes completed in 281–1751ms. No lossless guard or optimizer timeout was changed on that evidence.

## Unobserved boundaries

No complete real-optimizer pass, native JPEG optimization, perceptual/photo/ICC calibration, graphical clipboard/paste, shortcut/Explorer/Dolphin activation, keyring login, signed package, production update-feed transaction or public binary release is claimed. Local full-suite Go 1.27 execution was cancelled during its initial compilation without a verdict; a focused domain test passed. Hosted platform suites provide the complete source-test evidence above.

## Bounded repeat on the native Go 1.27.2 build

A second run against the local `f70486ce...` executable used the existing helper flag `-timeout 10m`. Provenance and dependency preflight passed, Auto alpha-gradient failed, and Auto alpha16/flat-icon/gray16/hidden-rgb passed. The broad rerun was deliberately cancelled after those observations; no complete result or cause is claimed. The helper checkpoint in `artifacts/validation/png-f70486c/report.json` preserves all seven completed checks. The failed Auto case returned exit 4 with an empty pipeline and the same unavailable-optimizer error. Partial stdout/stderr remain in `artifacts/validation/png-f70486c.*`. Future diagnosis should focus on the failed cases and retain probe reasons/timing rather than purchase another full-suite repeat first.

## Discovery diagnostic refinement

Version-probe errors now retain cancellation, deadline, launch-error or nonzero-exit details; a failed PNG result includes the relevant discovery reasons in its warnings. The existing three-second probe deadline and compression decisions are unchanged. A focused regression failed against the original generic error and passed after this change; a batch-level test confirms those reasons reach the failed image report.

A native Go 1.27.2 diagnostic build, SHA-256 `58b56ee5219093fcef371f5ee815a15dbae365fb8d7bfb99c2c406b0b5808b15`, found both real optimizers healthy in four consecutive status calls (486–2161ms for the complete call). Strict lossless alpha-gradient then succeeded, producing 173 bytes from 16,521 bytes. An intentionally missing OxiPNG path returned exit 4, no output, and its specific path failure in the JSON warnings. Reports remain under local `artifacts/validation/probe-diagnostics-20261016/`. This diagnostic build has no embedded source commit; it does not establish that the intermittent refusal is repaired.

Local `go test ./...` passed the application, toolchain, workflow and qualification-helper packages but failed the unchanged corpus generator's byte-comparison test on Go 1.27.2 (`fixture changed: .../alpha-gradient.png`). Focused diagnostic tests and `go vet ./internal/toolchain ./internal/workflow` passed. The corpus failure requires separate investigation; neither a full local green result nor an encoder-related cause is claimed.
