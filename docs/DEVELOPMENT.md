# Development

## Meaning

This guide owns the actual prerequisites, root commands, and qualification boundaries for developing smolpng. The initial repository is an orientation foundation, not yet an executable candidate. Commands are added with runnable source rather than presented as working before they exist.

## Toolchain and authoring environment

Go is selected. The current authoring sandbox has Go 1.23.2 but cannot resolve the public module proxy; that installed version is not the product target. The temporary build-environment workflow selects Go 1.26.6 and obtains candidate modules through the normal Go checksum-verified route. It exports only the toolchain, isolated public module cache, and repository Git bundle for local authoring. Those transport artifacts are not release assets and are not included in application distributions.

## Verification boundaries

The forthcoming ordinary root suite uses Go's testing package with focused deterministic cases and fake process/HTTP boundaries. Separate real-optimizer and real-desktop qualification will say exactly which host and versions ran. Cross-compilation never proves clipboard, keyring, shortcut, notification, or browser paste behavior.

Generated build and test output belongs under untracked `artifacts/`. Required regression fixtures remain source with provenance. No CI job publishes a release, package, website, or repository mutation.
