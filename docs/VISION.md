# smolpng Vision

## Meaning

smolpng takes an image from wherever a person or agent already has it, makes it substantially smaller with sensible quality, and places the result where the next action needs it. This Vision preserves Wolf's ratified product destination and character, including the complete 1.0 horizon and later self-hosted family. It is larger than the first PNG implementation and does not claim that planned behavior already exists.

## One action, not a smaller ambition

The ordinary journey is an icon rejected by a hostile upload limit: copy or drop the image, choose Auto or an MCP-icon recipe, then paste or select the smaller result without visiting an upload service. smolpng should be tiny in interaction cost, not tiny in product ambition. It is cute without visual noise, transparent when inspected, powerful without becoming a control panel, and deliberately boring to automation.

The name remains **smolpng** when JPEG arrives. The primary line is **Make images smol.** The descriptive line is **Local-first image compression for humans, scripts, and agents.** The secondary maker attribution is **An open-source project from Wolfsblvt Works.**

## People and interfaces

Everyday users want one action and no codec vocabulary. Power users want tunable quality, metadata, target sizes, output behavior, resizing, batching, and recipes. Developers want a dependable local command, project settings, exit codes, streams, and structured results. Agents want discoverable capabilities and a matching installed Skill, not a terminal interface they must reverse-engineer.

One engine serves three first-class modes: interactive `smolpng [INPUT...]`, deterministic `smolpng compress INPUT...`, and configured background `smolpng drop [INPUT...]`. The TUI uses arrow keys, Enter, Escape, and useful shortcuts—never a numbered-choice interrogation. Auto is first. Ordinary use usually takes one selection; target mode adds the requested size. Custom controls belong in a form, not eleven consecutive questions.

## Trust in the result

Auto is local by default. A configured TinyPNG key grants capability, not permission to upload. Originals are preserved unless in-place replacement is explicitly selected. Final files are committed atomically. Every lossy candidate derives from the original or an equivalent lossless normalized source, never a previously degraded candidate.

A target is strict: output bytes must be less than or equal to requested bytes. Search preserves dimensions by default and reduces them only with explicit permission. A missed target is a distinct outcome, not disguised success. Strict lossless preserves decoded pixels, including hidden RGB under transparent pixels. Results expose the actual pipeline, tool versions, target, transformations, remote use, and artifact path.

## Complete 1.0 destination

PNG and JPEG are accepted formats. PNG uses external pngquant and OxiPNG. JPEG uses externally installed jpegoptim for ordinary local optimization; the bounded JPEGli comparison did not earn another runtime, build, and distribution surface. Orientation must remain correct, and safe metadata handling preserves ICC while removing GPS/private camera metadata without accidental rotation.

Inputs include one or many files, directories, explicit globs, bounded HTTP(S) downloads, stdin, clipboard images, native file lists/URIs, text paths/URLs, and desktop drag/drop. Interactive mode may detect clipboard sources; scripts must name their input or explicitly request the clipboard.

Existing files default beside the original with `{stem}.smol{ext}`. Detached images default to the operating-system Downloads folder under `smolpng/`, never the shortcut's working directory. Collisions increment, such as `icon.smol-2.png`. Smart clipboard publishes useful image, file-list, and absolute-path representations together where the host permits; batch results copy files. A vanished notification never loses the result: `smolpng last` recovers it.

Directory input is shallow by default, recursive by explicit choice, conservative about hidden paths and directory symlinks, respects ignore rules, excludes generated outputs, and mirrors relative structure under an explicit output root. One recipe applies to the batch. Partial failure remains inspectable.

Built-ins are Auto, Web balanced, Maximum squeeze, Lossless, and strict target-size planning, with explicit TinyPNG and Custom workflows. Exact optimizer values are calibrated against legal regression images rather than made immutable by their first appearance in code.

Configuration follows built-ins, per-user TOML, project `.smolpng.toml`, selected preset, then invocation flags. Effective values show their provenance. OS-standard locations are normal; a deliberately selected portable directory is also supported. Environment variables serve secrets and deployment location, not a duplicate system for every slider.

Named presets are the safe extension surface. They inherit, export normally or as resolved recipes, have labels, descriptions, format support, output/clipboard behavior, and pinned menu placement. A copied `[presets.name]` block can be detected in the clipboard, reviewed, imported, and used without TOML surgery. Built-ins remain immutable. Imported presets are data only: no commands, hooks, credentials, arbitrary environment, executable paths, unrestricted raw arguments, or global privacy-policy override. Import collisions offer a distinct name or reviewed replacement. The README shows one useful example; the manual and website make more examples easy to copy.

First interactive use and `setup` share onboarding: discover and probe system optimizers, recommend healthy existing tools, offer per-user managed installation with sources/licenses/checksums, then offer desktop and Agent Skill convenience. No optimizer is bundled with the standard application, no managed installation silently mutates PATH, and startup does not perform an update check by default.

Windows and Linux are first-class desktop targets, including x64 and arm64 where dependencies permit. Windows convenience includes interactive and Auto Start-menu entries, quiet drop targets, optional Send To, notifications, and a no-console launcher. Linux includes desktop launchers and actual X11/Wayland clipboard qualification. macOS CLI/TUI and both architectures belong in the target matrix; polished Finder/drop packaging requires real host and signing evidence. Compilation alone is not a desktop qualification.

TinyPNG is the only initial remote provider. Policies are never, ask, fallback, compare, and always, with never the default. Credentials use `TINIFY_API_KEY` from the process environment first, then the native keyring. No command-line key is accepted. Local agents invoke the tool under the user's account without receiving the secret. Headless/keyring-unavailable environments have the explicit environment route. Ordinary CI uses fake HTTP and never consumes quota.

The version-bound Agent Skill teaches deterministic CLI use, capabilities, local Auto, target semantics, project recipes, JSON/JSONL, artifact paths, privacy, credentials, broad-directory dry runs, and stable exit codes. Harness detection and custom destinations are supported. Every app-created installation records destination and exact resource hashes. App updates synchronize only unmodified managed installations automatically; custom paths are offered sync and modified bytes are preserved. Manual executable replacement compares embedded content with last-seen state and prompts interactively only when Skill content changed. Diagnostics remain machine-readable and noninteractive stdout remains clean.

Distribution includes portable archives, signed installer/MSIX where qualified, Winget, Scoop, Chocolatey submission, Linux packages, and Homebrew. Standalone updates verify release integrity and replace safely without touching settings, dependencies, or modified Skills. Package-manager-owned installations receive the appropriate update command instead of self-overwrite. Optional startup checks are opt-in and human-only. A Store supporter/convenience edition is a later equal-feature route, subject to licensing and policy qualification.

## Product family and public face

The public static website lives in this repository, explains rather than accepts uploads, and includes installation, presets, agents, local-first/privacy facts, artist credit, and support links. A lightweight accessible terminal animation has a static/reduced-motion fallback and does not execute a shell. Actual deployment state, DNS, and credentials belong in private operations sources.

A commissioned original mascot supplies related but distinct full illustration, app icon, tiny glyph, and marketing assets. The TUI has a small replaceable default-on ASCII seam; final artwork and selection remain Wolf's taste work, not an implementation default accidentally canonized. Software, site implementation, copyable presets/corpus, protected editorial copy, and brand assets have separate license boundaries.

Future `smolpng-web` is an AGPL self-hosted browser/agent sibling for private or trusted deployments, consuming the CLI/JSON contract rather than duplicating compression. It includes file drop/picker, batch ZIP, configurable retention, history off by default, optional persistence, and container deployment. Wolfsblvt Works does not thereby become a public image-processing service.

The future ChatGPT plugin combines a workflow Skill with an authenticated MCP-backed app in that sibling. It must qualify actual file input/output, OAuth, and plan/surface behavior. The local CLI does not need an MCP server. A binary embedded in a ChatGPT Skill is an explicit future experiment, not a supported execution claim. WebP, AVIF, and further providers remain possibilities when they earn ordinary value without creating a generic media platform.

## Evidence and continuing judgment

Generated, CC0, or unambiguously licensed images exercise structure, pixels, alpha, dimensions, metadata, target outcomes, original preservation, and candidate selection. Lossy bytes are not globally snapshotted across optimizer versions. Source, fake-boundary tests, real optimizer runs, cross-builds, and real desktop journeys remain distinct evidence.

This destination was ratified through Wolf's supplied product and licensing briefs and the implementation assignment. Nyxara may make reversible improvements within it and should propose useful expansion rather than merely transcribe a prompt. A first coherent tranche sequences this whole destination; it never quietly replaces it with a tiny wrapper.
