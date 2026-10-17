# Human, clipboard and drop workflows

## Meaning

The keyboard UI, native clipboard adapter, settings, background/drop command and last-result recovery are implemented. This guide owns their usage and explicit native-host limitations. [Desktop integration](integration.md) owns opt-in shortcut installation; a signed installer and browser-paste behavior on every desktop remain unqualified.

## Ordinary path

Run `smolpng icon.png`, or copy an image/file/path and run `smolpng`. The first human launch checks local optimizers and lets you use available tools, configure executable paths, explicitly download/check a managed OxiPNG copy, open an upstream download page or explicitly install the personal Agent Skill. Acknowledging setup is not a claim that both tools exist. Missing dependencies continue to produce ordinary engine errors. TinyPNG keys are never requested during local setup.

The recipe menu offers Auto, a strict size target, built-ins, pinned presets, saved recipes and a custom form. Arrow keys and Enter select; S changes source, comma opens settings, D opens setup and Q quits. Menus scroll when necessary. The tiny neutral header is replaceable and can be disabled in settings or with `--no-header`; it is not final mascot artwork. `--plain` reduces decoration. `--accessible` or a dumb terminal uses typed option names, blank to keep a value, `/clear` to empty a field and `/back` to return.

After output, copy image/path/files or open its folder. Trying another recipe uses the same original clipboard snapshot, not the newly compressed clipboard data. An invalid/unsupported input is reported without modifying it. A target miss without explicit nearest acceptance produces no artifact, and remains visible in the last-result report.

## Presets and preferences

Copied `[presets.name]` TOML with `schema=1` is previewed before import. When an image/file is also present it stays primary and the import is a nonblocking menu hint. Imports use the same safe parser, collision remapping and byte-snapshot checks as the deterministic preset commands.

`settings` edits user defaults, output, clipboard, drop recipe and header. Only changed scalar fields are written; comments and unrelated TOML remain intact. `config set` is the noninteractive route, for example:

```sh
smolpng config set clipboard.single path
smolpng config set drop.mode background
smolpng config set drop.notify false
smolpng config set ui.header compact
```

Project/preset overrides can still supersede user defaults. Inspect `config show --effective --provenance` instead of guessing. The editor refuses to set a user default that exists only inside one project.

## Clipboard and background behavior

Smart single-result copy offers PNG bytes, file references and absolute path text in one write. A batch offers file references and newline-separated paths. Missing/deleted outputs are not copied as if they still existed. Native clipboard access may fail on headless hosts; the saved image remains usable and a warning explains what was unavailable.

On Linux an owning process may need to stay alive. Keep the result menu open until paste. `smolpng drop` retains ownership for up to five minutes, ending early if replaced or cancelled. `--clipboard-hold` can choose a shorter/longer bounded duration, including zero; JSON drop defaults to zero unless explicitly selected. It is a transient owner, not an installed daemon or background service. Windows/macOS clipboard persistence remains native API behavior to qualify.

`drop` applies the configured preset and tries the copy/notification without asking questions. A project deliberately configured for interactive drop opens the UI outside JSON mode. Windows `smolpngw.exe` selects drop with no console when built with `-H=windowsgui`; selecting interactive drop allocates a console deliberately. Ordinary dragging onto `smolpng.exe` opens the interactive root. Use the [Windows Auto-shortcut guide](integration.md#windows-drop-files-onto-an-auto-shortcut) for a background drop target, or `integrate install` for opt-in Start-menu/Desktop/Send To entries.

A missing notification is recoverable with `smolpng last`. `last --clear` removes the metadata only. Report images remain normal files, not temporary hidden downloads.

## Evidence and local qualification

Unit/integration tests exercise keyboard events, source priority, original preservation through repeat recipes, safe preset imports, exact smart-copy payloads, notification argument quoting, settings patches and state recovery. A real Linux pseudo-terminal has exercised Enter/Done and a target-already-satisfied image without an optimizer. Cross-builds cover the native API compilation paths only.

Katja must still exercise actual PNG/file-list/path clipboard representations, browser file-picker paste, X11/Wayland ownership, Windows/macOS clipboard persistence, native notifications, no-console/console transition, drag/drop, locked outputs, cancellation and visual TUI rendering. Use the exact checkpoint and qualification carrier accompanying the candidate. Do not rebuild the product from scratch merely to run those checks.

Setup offers the app-pinned OxiPNG archive only when a native catalog entry exists. It displays version, source, license, checksum and per-user destination before a separate download confirmation. Healthy current system/configured choices are not replaced. There is no catalog entry for pngquant or native Windows arm64 OxiPNG in this build; use existing installations or explicit paths. Preference changes and protected removal are in [Dependency setup](dependencies.md).

Candidate packages include `smolpng-qualify` for the real optimizer/file contract. Run it against the included `smolpng` binary and a new output directory; it refuses the wrong source commit and records per-case evidence. Follow with manual graphical checks. A passing command suite cannot substitute for browser-paste or notification observations.
