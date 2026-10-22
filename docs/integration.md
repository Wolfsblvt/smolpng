# Desktop integration

## Meaning

smolpng can create and remove its own per-user Windows and Linux launchers and unsigned macOS development applets. Every destination is previewed before explicit application and tracked by exact SHA-256; app bundles include all owned members. This is a functional installer surface, not proof that a particular file manager supports every drop or clipboard representation. [Native macOS file-open and Terminal journeys](reference/2026-10-22-macos-finder.md) passed on the exact Intel/ARM64 hosts recorded there.

```sh
smolpng integrate install --desktop
smolpng integrate install --desktop --apply
smolpng integrate install --desktop --send-to --apply
smolpng integrate shortcut create mcp-icon --preset mcp-icon --apply
smolpng integrate status --json
smolpng integrate remove --id preset-mcp-icon --apply
```

## Windows: drop files onto an Auto shortcut

Extract a development candidate into the folder you intend to keep. Keep `smolpng.exe` and `smolpngw.exe` together; source builds need both [Windows entry points](DEVELOPMENT.md#keyboard-and-desktop-verification). Candidate archives are currently local builds, not a signed public installer. See [Packaging](packaging.md#candidate-builds).

Open PowerShell in that folder and inspect your separate local optimizers:

```powershell
.\smolpng.exe deps status --json
```

PNG Auto uses pngquant; OxiPNG can improve its result. JPEG Auto uses jpegoptim. If a needed program is missing, run `.\smolpng.exe setup` to select existing programs or follow [Dependency setup](dependencies.md). The standard archive contains none of those optimizers. A TinyPNG key is unnecessary for this local workflow.

Create just one desktop drop target, first previewing its destination and arguments:

```powershell
.\smolpng.exe integrate shortcut create auto --preset auto
.\smolpng.exe integrate shortcut create auto --preset auto --apply
.\smolpng.exe integrate status --json
```

The desktop file is `smolpng-preset-auto.lnk` (Explorer may hide `.lnk`). Drop one or more PNG/JPEG files onto it. It runs Auto through the background executable. By default, each result appears beside its original as `name.smol.png` or `name.smol.jpg`; another run creates `name.smol-2.png` or `name.smol-2.jpg`. Originals stay in place. Explicit user/preset output and clipboard preferences still apply. Launchers ignore project settings, and Auto stays local unless user settings explicitly permit remote processing.

If the notification disappears or nothing obvious happens, inspect the saved result before trying again:

```powershell
.\smolpng.exe last --json
```

That report preserves per-file outcomes, warnings and output paths; it can expose a missing optimizer or a partial batch. Drag the original files again when trying another recipe, rather than compressing the previous outputs. To remove this shortcut, preview `.\smolpng.exe integrate remove --id preset-auto`, then repeat with `--apply`. Move the application by removing its old launchers and recreating them from its new folder.

Native shortcut creation and Windows activation with file arguments have passed. Physical Explorer dragging, rendered notifications and clipboard/paste still need ordinary desktop observation; this guide does not promote the activation check into those claims.

## Launcher variants

`--send-to` is Windows-specific. Ordinary installation creates a menu entry opening the TUI and an Auto entry running the background workflow. Optional desktop targets provide both variants. A custom preset shortcut can use `--interactive` to start the recipe picker instead. The preset must exist in built-ins or installed user settings when created; project-only definitions must first be imported into user storage. It is resolved by name when the launcher runs. No preset definition or credential is embedded.

## macOS development applets

On macOS, `integrate install --desktop` previews `.app` destinations in `~/Applications` and `~/Desktop`; repeat with `--apply` to create them. The built-in AppleScript compiler creates native file-open/drop handlers around the installed `smolpng` executable. Keep that executable at its recorded location. The applets use the operating system's provisional appearance and carry no production signing, notarization or accepted product icon.

The Auto applet passes dropped files to `smolpng drop`. The interactive applet opens the existing recipe picker in Terminal. Auto also opens Terminal when the current configuration selects `drop.mode=interactive`; configuration is resolved by Go rather than a second TOML parser. The Terminal command removes its temporary script before starting the CLI. File paths remain separate arguments, including spaces, quotes and Unicode. No optimizer is bundled or executed by the applet itself.

Use `smolpng last --json` to recover background outcomes, and `integrate status --json` to inspect ownership. Remove unchanged applets with `integrate remove --apply`. Native compilation, one/many-file events, both Terminal variants and owned lifecycle passed on macOS 15.7.9 Intel and macOS 26.6.2 ARM64. The macOS 26 Intel hosted image remained unqualified after an independently reported IconServices failure. This source addition is not a public macOS package or a claim of physical Finder dragging.

## Native behavior

Windows uses Known Folders for user Programs, Desktop and Send To. Windows Script Host produces a `.lnk` in a private temporary directory through a fixed, encoded PowerShell program. Fields are quoted, never evaluated as user script. The completed link is installed atomically. Background links target the adjacent `smolpngw.exe`, whose entry point explicitly supplies `drop`; interactive links target `smolpng.exe`. Missing companion binaries stop installation. No administrator elevation, PATH mutation, registry context menu or default-file association occurs.

Linux uses user XDG applications and Desktop locations. Desktop files quote executable and argument values according to the desktop-entry specification, append a standalone `%F` for file lists, and select `Terminal=true` only for interactive operation. No shell wrapper or MIME default is installed. Some desktops still require a user to mark a desktop launcher trusted, or choose their preferred terminal. That is a native qualification step, not silently automated security-policy changes.

If an Auto entry is already running without a terminal and the current configuration selects `drop.mode=interactive`, smolpng hands the same invocation to GLib's `gio launch` through a private temporary `Terminal=true` entry. The desktop chooses the terminal; the temporary entry is removed after the launch request. This configured foreground route requires `gio` and a configured terminal. Background mode creates neither. A successful launch request is not a completed compression; the opened workflow and `smolpng last` carry that result.

The working directory is the user's home, not the binary folder. Launchers ignore transient project settings through `--no-project` and retain an explicitly selected config file. Explicit portable installation records its data root through `--home`; relocating the application requires reinstalling its app-owned launchers.

## Modification and interruption safety

`integrations.json` lives under the smolpng state directory. The registry records destinations, arguments, app version and installed hashes. Reinstall only replaces an unchanged app-owned file. Removal never deletes modified or unrelated files. An OS lock prevents concurrent integration operations.

A pending journal precedes publication. `integrate repair --apply` recognizes only exact old/new hashes from that journal; ambiguous or user-modified bytes stop automatic recovery. A partially completed multi-file operation can be inspected and resumed without deleting unowned files. Moving native user folders may require manual disposition of old recorded destinations rather than authorizing deletion in an unexpected path.

Directory-based macOS applets record the complete member list and content/permission digest. Added user files make a bundle modified without reading those files' contents. Removal and replacement journal the operation, then atomically move the existing applet into a private sibling workspace before verifying its ownership. A late edit is restored when the destination is absent. If another object has appeared there, both objects are preserved and the operation refuses. Replacement publishes only after that verification; removal deletes only the verified quarantined bundle. Interrupted recovery reconciles the destination before cleanup, and an already-cleaned workspace does not leave the operation stuck. Symlinks, changed transaction contents and ambiguous destinations still stop automatic replacement or deletion. This protects ordinary named-path changes, not a hostile same-user process deliberately mutating retained open handles inside the private workspace.

Sources: Microsoft Windows Script Host shortcut documentation and the freedesktop Desktop Entry Specification, consulted during implementation. Tests exercise generated command lines, quoting, preview, ownership protection, interruption recovery and actual temporary filesystem writes. Native Windows `.lnk` creation and ShellExecute activation with appended PNG/JPEG file arguments have now passed, including collision handling and last-result recovery; see the [dated native record](reference/2026-10-16-windows-shortcut-qualification.md). The GUI entry point discards terminal reports and retains results for `smolpng last`. [PNG clipboard paste in Paint](reference/2026-10-19-windows-clipboard-alpha.md) now has a real Windows consumer observation, including correct alpha compositing after writer exit. Physical Explorer dragging, rendered notifications, file-list/text paste and other-host clipboard behavior remain unobserved.
