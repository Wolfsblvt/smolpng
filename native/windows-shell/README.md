# Native Windows 11 menu bridge

## Meaning

This directory owns the C++ Explorer ABI bridge, its portable request protocol and the Windows SDK/package staging script. It supplies the optional Windows 11 primary-menu boundary while the Go application owns compression. Compilation, signed-package activation and observed Explorer behavior are separate qualification claims.

This is the small native COM boundary required by Windows 11's **primary** File Explorer menu. It is not another compressor, a Go language migration, a persistent daemon, an in-process Go runtime or a legacy registry hack.

`core.hpp` contains the cross-platform bounded UTF-16 menu parser, argument quoting and one-shot request serializer. `shell.cpp` implements `IExplorerCommand`, `IEnumExplorerCommand`, the COM factory and the Windows launch boundary. `AppxManifest.xml.in` defines the separately signed sparse package identity. `build.ps1` compiles and stages from an exact matching Go candidate without installing anything.

Portable verification (CMake and a C++17 compiler):

```sh
cmake -S native/windows-shell -B artifacts/native-shell
cmake --build artifacts/native-shell
ctest --test-dir artifacts/native-shell --output-on-failure
```

Native compile (Visual Studio + Windows 11 SDK):

```powershell
cmake -S native/windows-shell -B artifacts/native-shell-x64 -A x64
cmake --build artifacts/native-shell-x64 --config Release
ctest --test-dir artifacts/native-shell-x64 -C Release --output-on-failure
```

Use `-A ARM64` for the native ARM64 addon, then run its tests on an ARM64 host. No ATL, Qt or downloaded library is required; the Windows SDK COM headers and standard C++ runtime suffice. Packaging/signing/activation and safe removal are documented in [the context-menu guide](../../docs/context-menu.md).

A Linux build runs the portable core tests only. It neither compiles the Windows API bridge nor proves Explorer activation. Keep that distinction in returns.
