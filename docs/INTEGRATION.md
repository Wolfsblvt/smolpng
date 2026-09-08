# Desktop integration

## Meaning

smolpng can create and remove its own per-user Windows and Linux launchers. Every file is previewed before explicit application and tracked by exact SHA-256. This is a functional installer surface, not proof that a particular file manager supports every drop or clipboard representation.

```sh
smolpng integrate install --desktop
smolpng integrate install --desktop --apply
smolpng integrate install --desktop --send-to --apply
smolpng integrate shortcut create mcp-icon --preset mcp-icon --apply
smolpng integrate status --json
smolpng integrate remove --id preset-mcp-icon --apply
```

`--send-to` is Windows-specific. Ordinary installation creates a menu entry opening the TUI and an Auto entry running the background workflow. Optional desktop targets provide both variants. A custom preset shortcut can use `--interactive` to start the recipe picker instead. The preset must exist when created; it is resolved by name when the launcher runs. No preset definition or credential is embedded.

## Native behavior

Windows uses Known Folders for user Programs, Desktop and Send To. Windows Script Host produces a `.lnk` in a private temporary directory through a fixed, encoded PowerShell program. Fields are quoted, never evaluated as user script. The completed link is installed atomically. Background links target the adjacent `smolpngw.exe`, whose entry point explicitly supplies `drop`; interactive links target `smolpng.exe`. Missing companion binaries stop installation. No administrator elevation, PATH mutation, registry context menu or default-file association occurs.

Linux uses user XDG applications and Desktop locations. Desktop files quote executable and argument values according to the desktop-entry specification, append a standalone `%F` for file lists, and select `Terminal=true` only for interactive operation. No shell wrapper or MIME default is installed. Some desktops still require a user to mark a desktop launcher trusted, or choose their preferred terminal. That is a native qualification step, not silently automated security-policy changes.

The working directory is the user's home, not the binary folder. Explicit portable installation records its data root through `--home`; relocating the application requires reinstalling its app-owned launchers. macOS Finder wrappers are not installed by this command.

## Modification and interruption safety

`integrations.json` lives under the smolpng state directory. The registry records destinations, arguments, app version and installed hashes. Reinstall only replaces an unchanged app-owned file. Removal never deletes modified or unrelated files. An OS lock prevents concurrent integration operations.

A pending journal precedes publication. `integrate repair --apply` recognizes only exact old/new hashes from that journal; ambiguous or user-modified bytes stop automatic recovery. A partially completed multi-file operation can be inspected and resumed without deleting unowned files. Moving native user folders may require manual disposition of old recorded destinations rather than authorizing deletion in an unexpected path.

Sources: Microsoft Windows Script Host shortcut documentation and the freedesktop Desktop Entry Specification, consulted during implementation. Tests exercise generated command lines, quoting, preview, ownership protection, interruption recovery and actual temporary filesystem writes. Native `.lnk` creation and launch/drop execution remain unobserved in this headless session.
