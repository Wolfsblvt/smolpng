# Windows ARM64 native compression qualification

## Meaning

This record preserves the exact hosted Windows ARM64 CLI compression result for source `9fe72e514f9dab48325d9baea72ed929a2b6701c`. It qualifies the native ARM64 application and matching qualifier across the 63 PNG/JPEG checks with real Windows optimizer executables, while keeping graphical desktop and distribution behavior separate.

## Result

The GitHub Actions run [39349334496](https://github.com/Wolfsblvt/smolpng/actions/runs/39349334496) completed successfully on the `windows-11-vs2026-arm64` hosted runner, image `20261013.183.4`. The host reported Windows 11 Enterprise 25H2, ARM64, kernel `10.0.26200`, revision `9457`. Go `1.26.6` built both the application and native qualifier with `GOOS=windows`, `GOARCH=arm64`, and `CGO_ENABLED=0`; the workflow verified both PE machine types as ARM64.

The qualifier bound both `expectedCommit` and `qualifierCommit` to `9fe72e514f9dab48325d9baea72ed929a2b6701c`. It completed and passed all 63 checks against both PNG and JPEG fixtures, using isolated temporary application state and explicit optimizer paths. The app executable SHA-256 was `9ad2c12fee9fe4d50b18c99961ec06af3376bafd6b8268315b0d142865ebc84b`; the qualifier executable SHA-256 was `7aec0b7ae2dca0e7bc65dc9f7c1dd3056ce3b3992506b1a596c261884f33e028`.

| Optimizer | Version | Executable architecture | Executable SHA-256 | Pinned archive SHA-256 |
| --- | --- | --- | --- | --- |
| pngquant | 2.17.0 | AMD64 (x64) | `7aadad6a50aa5c40cfcb20f9478c92f6030360a3424bc262ce383dc1e97fb86a` | `bd0257aeeccfe446a4cd764927e26f8af6051796f28abed104307284107b120d` |
| OxiPNG | 10.2.1 | AMD64 (x64) | `35ae3980ab831af64f4aecc98f69b81bfa146fa750a61d17ebbb12520128cbc8` | `7e940f83ee46874b73f53031f96a15834cb70b220af27391fb06fe7b4dd798e1` |
| jpegoptim | 1.5.6 | AMD64 (x64) | `0a79681b56961ec2c5f45c4f9dff01a64e7274b4f3250c22cde5f09c0bb457f3` | `db2d8caee88f2665b772c4591cba223b8b729500b88266dc06d4bf79d28fe11a` |

The optimizer executables were x64 and ran under Windows on ARM emulation. This does not misstate their architecture as ARM64. The identical-head hosted `Verify` run [39349334495](https://github.com/Wolfsblvt/smolpng/actions/runs/39349334495) also completed successfully; that suite is separate from the native optimizer journey.

The uploaded artifact is `native-windows-arm64-9fe72e514f9dab48325d9baea72ed929a2b6701c-39349334496`, artifact ID `12232517699`, SHA-256 digest `3eb25915685bbcce1f7eb09f70821fd9fe944971c0d535eb550003d0fc7c84df`, expiring `2026-11-03T01:45:50Z`. Its `journey/report.json` SHA-256 is `f89a8c0e0868976622ba33920b1c3fdcf5fc338f9c330143828074f8b9fe8994`. It contains reports and hashes, not executable or fixture payloads.

## Route correction

An earlier run, [39348993281](https://github.com/Wolfsblvt/smolpng/actions/runs/39348993281), built the native ARM64 executables but stopped before compression: PowerShell passed the hashtable `.Values` collection as one joined path to `Get-FileHash`. The workflow now hashes each optimizer path in an explicit loop. The successful run above exercises that correction. No local build or test was run for this qualification.

## Boundary

This result qualifies source-built CLI/filesystem PNG and JPEG compression on one hosted Windows ARM64 image with the listed x64 optimizer versions. It does not qualify graphical Windows behavior, clipboard or paste, physical drag/drop, notifications, desktop launchers, updater or package installation, other Windows ARM64 hosts or optimizer releases, user acceptance, the complete private candidate, signing/feed operation, or publication.
