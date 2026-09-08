# Architecture

## Meaning

smolpng is a local image-optimization orchestrator with one engine behind human and machine interfaces. This initial record describes the accepted load-bearing seams and distinguishes them from current implementation: application source is not yet present.

## Accepted shape

Inputs resolve into format-neutral jobs. Configuration and recipes produce explicit requests. The PNG planner creates original-derived candidates through separately discovered pngquant and OxiPNG processes. Validation and selection enforce pixels/dimensions/target policy. The output boundary atomically writes the artifact; reporting exposes a stable structured result. The TUI and background launcher call this same path.

External boundaries are the process runner, HTTP transport, clipboard, native credential store, and desktop integration. Pure policy, naming, size parsing, and candidate ranking remain directly testable. No dependency-injection framework, dynamic registry, local MCP service, or second engine is selected.

## State and horizon

User TOML, project TOML, and named presets are separate from state such as last result and Skill installation hashes. Credentials never enter these files. Installed Skills have a manifest and explicit management state. Updates respect distribution ownership and only synchronize unmodified files.

JPEG is accepted but its backend remains subject to corpus and packaging evidence. The future self-hosted sibling consumes a deliberately stable CLI/JSON artifact contract. Neither future consumer earns empty provider packages or hosted infrastructure in this repository.
