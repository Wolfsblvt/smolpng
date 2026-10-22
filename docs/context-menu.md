# Image right-click menus

## Meaning

The optional right-click surface is deliberately one action by default: **Make smol**, creating a smaller sibling while preserving the selected originals. An alternative `smolpng` submenu includes **Make smol (replace original)** first, **Choose recipe...**, preserve-original Auto, lossless, and selected installed user presets. Replacement is named explicitly; it is not the default single action.

The same settings are reachable from `setup` or `settings` through desktop integrations. This document owns menu behavior, installation, protection and the native qualification boundary. It does not claim that authored Windows bridge code has already been compiled or activated on a desktop.

## Configure and remove

```sh
smolpng integrate shell install
smolpng integrate shell install --apply
smolpng integrate shell install --mode submenu --preset mcp-icon --preset pixel-art
smolpng integrate shell install --mode submenu --preset mcp-icon --preset pixel-art --apply
smolpng integrate shell status --json
smolpng integrate shell remove --apply
```

Without `--apply`, install/remove is a preview. Custom entries must resolve from user storage; project-only presets cannot become durable launchers. At most sixteen actions are admitted. Shell actions are local-only, irrespective of a global TinyPNG preference, and never implicitly consume clipboard input. Use the ordinary app for a consciously selected remote operation.

Each action freezes its action ID and menu generation, not image paths. It resolves the latest installed user preset at invocation while keeping the action's explicit local/output boundary. After editing menu entries, stale open menus refuse rather than run a different action. Custom presets preserve original files and may still specify their explicit resize/matte/target behavior. Their action names use preset identifiers so the selected durable recipe stays visible.

Menu settings and service files use the existing journalled integration manager, with exact installed hashes. A later install removes only obsolete untouched app-created entries. Local edits, symlinks and unrelated files are preserved. `integrate shell repair --apply` reconciles an interrupted file operation only when recorded hashes agree.

## Windows 11: the primary menu, not a registry shortcut

Microsoft's modern menu requires native `IExplorerCommand` code and package identity. Static registry verbs would merely populate **Show more options**; smolpng does not substitute that legacy route or disable Windows 11's menu globally.

The original small C++ bridge is in `native/windows-shell/`. It implements the COM factory, single/submenu commands and enumeration, selection filtering and launch handoff. No Go runtime or optimizer is loaded into Explorer. Menu construction only reads a bounded local data file and selection properties. Compression, keyring, network, config resolution, notifications and output all remain in the existing Go executable after `Invoke`.

The bridge accepts only PNG/JPEG filesystem selections (one to 512 files), not directories, virtual items, URLs or clipboard fallback. Multi-selection passes through a bounded one-shot UTF-16 handoff instead of a potentially truncated 32K command line. The Go receiver validates the current installation, action/generation and real selected files before using the shared compression engine. Interactive commands launch the console executable; other actions launch `smolpngw.exe`.

### Build and activate the optional native component

The ordinary cross-built Go ZIP does **not** contain a qualified native DLL or signed menu package. A Windows SDK/Visual Studio developer host must build the matching x64 or ARM64 bridge from the exact same clean source as the application candidate. The compiler is needed only for this small addon, not to rescue or rebuild the Go application.

```powershell
# From source root, in a Visual Studio developer PowerShell:
.\native\windows-shell\build.ps1 `
  -AppDirectory C:\Candidates\smolpng-windows-amd64 `
  -Architecture amd64 `
  -Publisher 'CN=YOUR-EXISTING-CERTIFICATE-SUBJECT' `
  -PackageVersion 0.3.0.1 `
  -Out C:\Candidates\smolpng-shell
```

This creates a new `payload` directory containing the candidate EXEs, compiled DLL, notices and hash receipt, plus an **unsigned** sparse MSIX. `-SigningThumbprint` may select an already provisioned certificate; the script never creates keys, installs certificate trust, registers the package or restarts Explorer. Its solid placeholder package logos are not commissioned branding. Real release signing remains separately authorized.

After signing through the intended trusted route, use the **staged payload's** executable:

```powershell
C:\Candidates\smolpng-shell\payload\smolpng.exe integrate shell install --apply
C:\Candidates\smolpng-shell\payload\smolpng.exe integrate shell activate `
  --package C:\Candidates\smolpng-shell\smolpng-shell.msix
C:\Candidates\smolpng-shell\payload\smolpng.exe integrate shell activate `
  --package C:\Candidates\smolpng-shell\smolpng-shell.msix --apply
```

Activation checks the sparse manifest purpose/CLSID/file types, exact source-bound payload hashes and PE architecture. Windows `Add-AppxPackage` performs signature trust validation. The app never adds a trusted root or bypasses a failed signature. Installed package readback is saved separately from configured menu state. `status` distinguishes settings from native package registration; it does not claim the Explorer UI has been observed.

`integrate shell deactivate --apply` unregisters only the exact package identity previously recorded by this app. `remove --apply` removes menu settings, hiding its actions, but deliberately does not pretend it also removed package identity. Updating/replacing the executable may require rebuilding the matching native addon before a later activation. Existing protocol-compatible registrations remain a separate host check.

Menu source/portable protocol tests, a Windows DLL compile, package signing/registration, and observed Explorer behavior are four different proofs. The source includes a Windows build-only CI job, with no signing or installation. The current authoring host only ran the portable C++ protocol tests; DLL build, sparse-package validation and Explorer activation remain native qualification items.

## Linux: KDE Dolphin

On Linux, `integrate shell install --apply` writes executable service files under the native XDG data location's `kio/servicemenus/`. MIME types are PNG and JPEG, URLs are restricted to `file`, and `%F` passes each selected filename independently. No shell string, default MIME association, system PATH or administrator privilege is involved.

Dolphin inherits `Terminal` from the service file, not an individual action. Submenu mode therefore uses one background service and one terminal-enabled interactive service grouped under the same submenu. Single mode removes the obsolete interactive service after checking ownership. The terminal is selected by KDE, not hard-coded as Konsole. User settings or desktop policy may still hide a service until enabled/refreshed in Dolphin.

This is **Dolphin integration**, not a claim of universal Linux context-menu support. Nautilus/Nemo/Thunar have separate mechanisms and are not installed by this command. [Native Ubuntu X11 qualification](reference/2026-10-22-native-linux-desktop.md) exercised real Dolphin single/submenu actions and terminal completion. Physical desktop feel, other hosts/protocols, clipboard paste and rendered notifications remain separate.

## Data and lifecycle

Menu data lives in a predictable native per-user folder so the shell bridge can find it even when the app uses portable config. It contains installation/config paths, labels, preset IDs and a random generation, not secrets or image bytes. Windows selection handoffs under `shell-requests/` contain selected absolute filenames, are bounded to 8MiB, and are removed after successful decoding. A crash or invalid handoff can leave a request file; it is not history and may be manually removed once no menu invocation is active. Normal image retention and last-result behavior remain unchanged.

Neither bridge nor parser is a security sandbox against the same user deliberately modifying their installation. The ownership/hash checks prevent accidental overwrites and stale or malformed instructions; they do not elevate privileges.

## Primary integration references

- Microsoft, Windows 11 context menus and sparse package identity: https://learn.microsoft.com/en-us/windows/apps/desktop/modernize/integrate-packaged-app-with-file-explorer
- KDE, per-user Dolphin service menus: https://develop.kde.org/docs/apps/dolphin/service-menus/
- KDE KIO, action launch uses the service's terminal setting: https://invent.kde.org/frameworks/kio/-/blob/master/src/gui/applicationlauncherjob.cpp

Read during implementation on 2026-09-08. Published release behavior still depends on the exact target Windows/KDE versions and native observation.
