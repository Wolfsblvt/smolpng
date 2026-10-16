# Complete Windows PNG qualification

## Meaning

The repaired native Windows amd64 executable passed all 29 local PNG qualification checks with real pngquant and OxiPNG. This record preserves the exact build, measured contracts and provenance so the earlier partial runs remain historical evidence rather than the current result. It does not qualify graphical desktop behavior, JPEG optimization or distribution.

## Exact executable and tools

Go 1.27.2 built source `2abae9390167d35dbca942c6a69ec83bca90cfd9` with `CGO_ENABLED=0`, `-trimpath`, version `0.3.0-dev.1`, standalone distribution and the exact source commit embedded in both the application and qualification helper. The executable SHA-256 is `d4fa081535f61c27191c49885452d0a7800429963c04ad40d47748854488b687`.

Separate pngquant 2.17.0 and OxiPNG 10.2.1 were selected through explicit paths. Neither was bundled or installed into machine PATH. The helper used a new isolated home, disabled project lookup and forced offline/remote-never compression. It exercised generated CC0 specimens without clipboard, shell integration or provider quota.

The complete local receipt and per-check reports are `artifacts/validation/png-2abae93/report.json`; the executable is `artifacts/bin/2abae93/smolpng.exe`. The helper returned exit 0 and `Passed 29 local checks`.

## Observed contracts

| Checks | Result |
| --- | --- |
| Build provenance and compatible-tool preflight | 2/2 passed |
| Auto across all eleven structural PNG specimens | 11/11 passed |
| Strict lossless across the same specimens | 11/11 passed; decoded channels, including hidden RGB, preserved |
| Exact target success, refusal and explicit nearest acceptance | 3/3 passed |
| Clean binary stdout and partial batch behavior | 2/2 passed |

The target specimen produced 97 bytes. It met the 10KiB target. The impossible one-byte target returned exit 6; refusal produced no output, while explicit nearest acceptance preserved the miss and wrote the 97-byte artifact. Every original remained byte-identical, and generated outputs passed the helper's dimension, alpha, format and reported-byte checks.

## Source evidence and replay

Native `go test ./...` and `go vet ./...` passed on Go 1.27.2; local logs and exit receipts are under `artifacts/validation/probe-startup-fix/`. [Verify run 38936539830](https://github.com/Wolfsblvt/smolpng/actions/runs/38936539830) passed all eleven hosted jobs at `2abae939...`, including Windows/Linux/macOS tests, six cross-builds and Windows shell compilation.

The contribution was then replayed onto `83b1df2f679fdc0018220a48ca2f48c0e373408c` as `181dd613d9b471cc031fdc906e08dc90f5847628`. A tree comparison against the tested source showed only the already-landed upload-artifact action update in `.github/workflows/verify.yml`; application, helper and test bytes were unchanged. The native claim remains bound to the executable above. Later documentation and hosted checks do not rename its embedded commit.

## Limits

The [captured discovery deadline](2026-10-16-optimizer-discovery.md) supplied a demonstrated repair target. This successful complete run supports that repair without establishing that every historical refusal had the same cause or that optimizer startup can never exceed the new budget. Compression deadlines were unchanged; the helper's external per-invocation deadline was ten minutes.

jpegoptim was absent. Native JPEG optimization, real photographs/ICC calibration, graphical clipboard/paste, drag/drop, shortcuts, Explorer/Dolphin activation, keyring use, signing, updates against a production feed and public binary publication remain unobserved by this run.
