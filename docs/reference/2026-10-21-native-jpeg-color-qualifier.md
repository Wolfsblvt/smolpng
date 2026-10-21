# Native JPEG color qualifier

## Meaning

The shipped native qualifier now repeats the four canonical grayscale, CMYK and Display P3 JPEG controls through Auto, Lossless, profile removal and resizing. It detects profile/component loss and accepts the existing tagged non-RGB resize refusal only for the correct reason with no output. The corrected tool passed 79 PNG/JPEG checks on four hosted platforms and 51 JPEG checks on Windows amd64. This is repeatable CLI/optimizer evidence, not CMS, graphical, private-candidate or release acceptance.

## Source and corrected invocation

The implementation and its real-CLI regression are at `114e7859edca4ba12771b78c43e2734c1925d248`, based on `e62c1929e40b7e1592cf81c146f855f7c788695f`. Compression behavior, quality defaults and codecs are unchanged. The normal qualifier adds sixteen cases to its previous 63; JPEG-only mode has 51 checks. Existing PNG, target, stdout, batch, original-preservation and executable-provenance checks remain intact.

The [first hosted run](https://github.com/Wolfsblvt/smolpng/actions/runs/39663261313) at `d15c212bbc8cade8488a68e8f70d71c27d9fbdd9` completed with four failures per host: the new qualifier invoked nonexistent `--metadata all` instead of supported `--strip all`. The other 75 checks passed, but each complete qualification correctly failed. An initial Windows amd64 worktree build had the same four failures. These were qualifier command errors, not observed compression failures.

The correction uses `--strip all`. The regression first reproduced the unknown-flag failure, then passed after repair. It routes the exact generated arguments through the real CLI's dry-run parser before modeling optimizer output. Focused qualifier tests/vet passed; ordinary [Verify run 39664411744](https://github.com/Wolfsblvt/smolpng/actions/runs/39664411744) passed all eleven jobs at the corrected source. Corpus check mode also passed before the invocation-only repair.

## Hosted real optimizers

All four jobs in [Native optimizer qualification run 39664451308](https://github.com/Wolfsblvt/smolpng/actions/runs/39664451308) passed. Each built the native application and qualifier from exact source `114e7859edca4ba12771b78c43e2734c1925d248` with Go 1.26.6, recorded executable/optimizer identities before testing, and finished with `complete=true`, `passed=true`, all 79 checks passing and unchanged selected optimizer hashes.

| Host | pngquant / OxiPNG / jpegoptim | Report SHA-256 |
| --- | --- | --- |
| Ubuntu 24.04.5 amd64 | 2.18.0 / 10.2.1 / 1.5.6 | `7fa3d910dd6d1956c96abeacc8da79c13da3d0d4926cd14c108d8f24f26f5258` |
| macOS 26.6.1 (25G76) amd64 | 3.0.3 / 10.2.1 / 1.5.6 | `af883993fde7145121a90a1809e8ca6eea86a9dad6269ea748ac7cf3603b72f6` |
| macOS 26.6.2 (25G83) arm64 | 3.0.3 / 10.2.1 / 1.5.6 | `4a811eeb01422df12bbc9084019588afa5fc6f244ddf17f539d5727bf583088b` |
| Windows 11 Enterprise 25H2 arm64, 26200.9711 | 2.17.0 / 10.2.1 / 1.5.6 | `62f5c1b1c92965bc953cf897d27bf7040e1beb70e2f253040080f9bfb509a0d6` |

The Windows ARM64 application and qualifier ran natively; its optimizers were the separately identified x64 executables under Windows emulation. This is not native ARM64 optimizer qualification. Managed installation and disposable-host package extraction used the existing workflow; no workstation toolchain, signing key, production feed or package registry was changed.

| Host | Application SHA-256 | Qualifier SHA-256 |
| --- | --- | --- |
| Linux amd64 | `c51b82a08d7f9e605ccdffbedf7f3b6f29d0edda591cd65da21cb108cb043186` | `0334a9d09feca3079e92a8ebff8bef052b3aa1ea0c8230ef74c6e5686a809a33` |
| macOS amd64 | `6029de866e94c47448e9d10acf72063d8b4e70086be004608dc1743dfab1e4a7` | `a0d17d25c098916b730a5a7b1f67d564805076965f4ca000259331851c7459ef` |
| macOS arm64 | `6664f6138b831ca030f106bd92d0f5f627f7c899d3a2fe51b6b8ae42e092c44b` | `a5f5f0bef9da17cad1aaf6d3edb6ec672200baaaa66d3603e2cfdae4bb728ccb` |
| Windows arm64 | `2aba63bd380e970f58781f113a173fc2a35e8b44d7fe6d991971983704b40954` | `1aa4e7584af0acf9157db2cef60e2dd1722fed20a06fece8bfe733b918709b1f` |

The four complete artifact sets are retained locally under `artifacts/validation/jpeg-qualifier-hosted-114e785/`, using their exact `native-PLATFORM-114e7859edca4ba12771b78c43e2734c1925d248-39664451308` artifact names. Each includes source, host, Go, executable and optimizer identities plus the full per-invocation report. The failed earlier artifact sets remain separately under `artifacts/validation/jpeg-qualifier-hosted-d15c212/`.

## Windows amd64 replay

The corrected helper was built from clean source `114e7859edca4ba12771b78c43e2734c1925d248` with installed Go 1.27.2. Its SHA-256 is `87afcc427518be4de5fee54bd61632614a71f6f2e0dd0b1857385ca1a713b482`. JPEG-only mode passed all 51 checks on Windows 11 Home 25H2 amd64, 26200.9711, against the existing development application from source `5aaa04001d1da4f7809b650a4c4afd1b67d1e7c9`, SHA-256 `8aa8e355f26368f10bbedd79285330495df37eaa0148de713ff2d53541b207aa`.

Existing jpegoptim 1.5.6 remained SHA-256 `0a79681b56961ec2c5f45c4f9dff01a64e7274b4f3250c22cde5f09c0bb457f3`. The helper did not require or qualify PNG optimizers in this replay. The complete report is `artifacts/validation/native-color-qualifier-114/journey/report.json`, SHA-256 `62d7409d28017e6b2dffd05b0df3446af189b0c14ccf2410b87768eb0c232c19`. The initial failed worktree-built helper and its 79-check report remain separately under `artifacts/validation/native-color-qualifier-e62/`.

## Evidence boundary

Every original was checked unchanged. Ordinary Auto/Lossless JPEG output retained source components and exact ICC segments; Lossless additionally preserved decoded channels. Strip-all removed profiles without changing component count. Supported 64×48 resizes produced RGB while retaining a source RGB profile. Tagged grayscale/CMYK resize produced the existing explicit refusal, no output, and an unchanged original.

These checks use Go decoding and exact profile bytes. They do not perform independent CMS conversion or establish appearance in a real graphical consumer. The separate [Windows Pillow/LittleCMS record](2026-10-21-windows-jpeg-color-models.md) retains that narrower independent observation. Arbitrary profiles, print workflows, physical drag/drop, paste/notification behavior, package installation, private dev.11 bytes, a new joined application candidate and publication are not qualified by these runs.
