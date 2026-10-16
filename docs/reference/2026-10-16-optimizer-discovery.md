# Windows optimizer discovery deadline

## Meaning

This record preserves a captured version-probe deadline failure and the resulting startup-budget repair. It separates that demonstrated cause of false optimizer unavailability from earlier unexplained native refusals and from complete PNG or graphical qualification.

## Captured failure

The native Go 1.27.2 Windows amd64 build at `c2067062dfb91f5146155fd1c41f251d89714262` used separately installed pngquant 2.17.0 and OxiPNG 10.2.1. A bounded repeat of `deps status --json` stopped on its first attempt: pngquant was unavailable with `version probe failed: context deadline exceeded`, while OxiPNG remained available. The process returned normally after 12,308ms. The report and receipt are retained locally under `artifacts/validation/native-c206706-discovery/`.

Before that capture, the same build passed strict lossless alpha-gradient, alpha16 and gray16 with real OxiPNG, preserving originals and producing 173, 133 and 117 bytes respectively. A separate nearest-target invocation returned the expected exit 6 and a 97-byte output for an impossible one-byte target in 66,119ms. These focused observations are retained under `artifacts/validation/native-c206706-focused/`; they do not explain every earlier refusal or establish a complete native pass.

## Repair and regression

Discovery allows up to fifteen seconds per version probe instead of three. An earlier caller deadline still wins, and hung tools remain bounded. This trades a longer wait for an actually stalled tool against the observed false refusal of a configured optimizer; it adds no retries, cache or background work.

A regression using Go's standard-library virtual-time test facility refused a compatible five-second startup against the original implementation and passed after the change. Separate cases preserve an earlier caller deadline and bounded termination of a hung tool without real-time sleeps. The focused toolchain tests and static check passed on Go 1.27.2.

## Scope

The captured deadline establishes one concrete discovery failure. A native rerun is required to judge the repaired executable, and cannot guarantee that process startup will always finish within its budget. This repair changes neither image validation nor compression deadlines. JPEG optimizer behavior, graphical clipboard and shell integration, signing and publication remain separate qualifications. [Issue 4](https://github.com/Wolfsblvt/smolpng/issues/4) owns the active native investigation.
