# Project map

## Meaning

This map locates smolpng's maintained source and explains responsibility boundaries as they become real. The repository is a single Go application, not a multi-product framework. The future self-hosted server is a sibling and is not implemented here.

## Current map

`AGENTS.md` is the development front door. `README.md` is the user front door. `docs/VISION.md` owns the complete destination; `DIRECTION.md` owns the current frontier; `DECISIONS.md` owns durable rationale; `ARCHITECTURE.md` owns actual system shape; `DEVELOPMENT.md` owns exact commands. Root license files separate rights.

`.github/workflows/bootstrap-environment.yml` is temporary build-transport recovery, not product CI or permanent infrastructure. It will be removed after the isolated dependency/toolchain artifact is consumed. Product source and generated corpus paths will be added to this map with their first usable implementation.
