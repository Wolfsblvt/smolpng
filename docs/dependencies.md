# Local optimizers and managed installation

## Meaning

smolpng invokes separate local optimizer programs rather than bundling them. This guide owns discovery, the fixed OxiPNG download catalog and protected per-user maintenance. A catalog checksum and passing version probe do not prove native output quality or graphical desktop behavior.

## Find or select a tool

```sh
smolpng deps status --json
smolpng deps licenses
smolpng deps configure pngquant /path/to/pngquant
smolpng deps configure oxipng /path/to/oxipng
```

Use `.exe` paths on Windows. Explicit invocation flags win, then saved user paths, then compatible system tools, then managed tools. A saved broken path remains an error rather than silently changing the user's selected executable. An unsupported PATH entry no longer shadows a valid managed installation.

`deps use-managed` previews clearing the saved path overrides and preferring a verified managed copy; `--apply` saves the change. `deps use-system --apply` restores normal system-first preference. Both may fall back to the other compatible local source. Neither removes anything or edits PATH. Project configs cannot select executable paths or management preference. `config show --effective --provenance` identifies the user setting's source.

## Explicit download

```sh
smolpng deps install oxipng --json
smolpng deps install oxipng --apply --json
```

The first command is a read-only preview. The second accepts the displayed source/license and requests the exact pinned archive. Setup provides the same consequence after a separate confirmation. No package manager or administrator prompt is invoked.

This application catalog selects **OxiPNG 10.2.1** from its upstream GitHub release, with SHA-256 digests read from the release metadata on 2026-09-08. Linux amd64/arm64 use musl archives, macOS amd64/arm64 use the respective native archives, and Windows amd64 uses the MSVC archive. No native Windows arm64 artifact is selected. There is **no qualified automatic pngquant download** in this build; use its official separate installation and license route.

Before any downloaded executable runs, the complete archive must match the pinned hash. Extraction validates every member's name/type, enforces size and entry bounds, and selects exactly one executable. The executable is probed for the exact pinned version; the installer writes its MIT license beside it. Failed checks leave the previous active version unchanged. Downloads are not universally runtime-qualified just because their hashes match.

## Where bytes live

```text
data/tools/oxipng/
  index.json
  versions/<version>-<os>-<arch>-<archive-hash-prefix>/
    oxipng[.exe]
    LICENSE
```

The index records the selected artifact, hashes of installed files, retained versions and active selection. System installs are outside this tree. Each installed version's executable/license must still match before managed discovery uses it; modified, unknown, incompatible and untracked states are reported. An explicit saved or invocation path is instead trusted as the user's own selection.

An app-owned advisory lock is kept beside the tool root. No PATH entries are changed. Portable placement follows the app's selected data root. Moving a portable folder to another OS/architecture does not silently execute the old managed binary.

## Updates and removal

```sh
smolpng deps update oxipng
smolpng deps update oxipng --apply
smolpng deps remove oxipng
smolpng deps remove oxipng --apply
```

Update means the catalog shipped by the installed smolpng build, not whatever upstream published most recently. An already identical installation is a local no-op. A newer app may install a new immutable version and atomically switch the index; previous versions remain. An older catalog cannot downgrade a newer recorded version. At 32 retained versions, inspect/remove before adding more.

Removal checks every recorded version and refuses user edits, extra files or untracked content. It removes only the managed tool tree, never a system program, config, presets, Skill or image. Removal renames the verified tree out of discovery before cleanup. Locked files can cause a refusal or a clearly reported cleanup residue, not a falsely successful half-active executable.

A hard crash can leave a staging or removal directory. Inspect its bytes before manual cleanup; this cut does not sweep unknown directories or overwrite an installation merely to make a green check appear. The local owner is trusted, and these checks do not defend against a malicious simultaneous writer with the same account permissions.
