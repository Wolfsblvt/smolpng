# smolpng

[![Verify](https://github.com/Wolfsblvt/smolpng/actions/workflows/verify.yml/badge.svg)](https://github.com/Wolfsblvt/smolpng/actions/workflows/verify.yml)
[![Inputs: PNG + JPEG](https://img.shields.io/badge/inputs-PNG_%2B_JPEG-566C86)](docs/cli.md)
[![Targets: exact bytes](https://img.shields.io/badge/targets-exact_bytes-39756A)](docs/cli.md)
[![Recipes: shareable TOML](https://img.shields.io/badge/recipes-shareable_TOML-8A668C)](docs/presets.md)

**Make images smol.** Local-first image compression for humans, scripts, and agents.

An icon, a screenshot, a photo. A very unreasonable upload limit. smolpng helps the image fit, keeps the original, and reports a missed target instead of pretending it fits. One engine serves the keyboard interface, background drop workflow, and deterministic CLI.

**Source candidate, not a published release.** Build this source branch to try it. Windows, Linux, and macOS are build targets; native desktop behavior still needs host-specific qualification.

**[Build and try it](#make-one-image-smol)** · [User guides](docs/README.md) · [Full product vision](docs/VISION.md)

## Make one image smol

Use **Go 1.26.6 or a compatible newer toolchain**. From a checkout of this source branch:

```sh
go mod download
go build -o artifacts/bin/ ./cmd/smolpng
./artifacts/bin/smolpng deps status
```

On Windows, use `.\artifacts\bin\smolpng.exe` for the executable commands. Go is needed to build, not to run the binary.

**Optimizers are separate, not bundled.** Install pngquant and OxiPNG for the full local PNG pipeline, or jpegoptim for JPEG. The [dependency guide](docs/dependencies.md) covers compatible versions, explicit paths, preview-first managed OxiPNG/jpegoptim installation on supported host targets, and separate setup where no catalog artifact exists.

With those tools available, replace `icon.png` with your image:

```sh
./artifacts/bin/smolpng compress icon.png --target-size 10KiB --json
```

Success gives you `icon.smol.png` beside the original and `target.met: true` in the report. Collisions increment to `icon.smol-2.png`. An already-fitting image can return unchanged; a missed target exits **6** without writing an over-limit result by default.

Below, `smolpng` means that built executable, or the same binary placed on your PATH.

## A little utility, not a control panel

The actual terminal header, still a replaceable placeholder rather than final mascot art:

```text
  ( . . )~
   smolpng
```

```sh
smolpng icon.png                    # Keyboard menu, Auto selected
smolpng                            # Detect clipboard image/file/path
smolpng drop icon.png               # Configured recipe, copy and notify
smolpng last --json                 # Recover the previous human/drop result
```

Use arrows and Enter, or `--accessible` for typed choices; `--plain` and `--no-header` remove decoration. Clipboard and notification behavior depends on the host. On Linux, keep the result screen open until you paste. [Desktop guide and limitations](docs/desktop.md).

Scripts use `compress`, which never prompts or infers clipboard input. Files, directories, globs, HTTP(S) URLs, stdin (`-`), and explicit `--clipboard` input are supported. Preview a batch before running it:

```sh
smolpng compress images --recursive --output-dir optimized --dry-run --json
smolpng compress images --recursive --output-dir optimized --jobs 2 --jsonl
```

Directories are shallow by default. Local outputs sit beside originals; detached inputs use OS Downloads/smolpng. Binary stdout and separate report files are also supported. [Inputs, outputs, and exit codes](docs/cli.md).

## Smaller files, explicit trade-offs

| Input → output | Local processing |
| --- | --- |
| Static PNG → PNG | pngquant reduces colors; OxiPNG optimizes losslessly. Transparency is preserved by default. |
| JPEG (`.jpg`, `.jpeg`) → JPEG | jpegoptim supplies lossless/lossy candidates; explicit resizing encodes from the original upright image. |

Animated PNG compression is refused, not flattened. GIF, WebP, AVIF, and SVG are unsupported; this is not a format converter.

Built-ins are `auto`, `web`, `smallest`, and `lossless`. Auto may be lossy. Choose `--preset lossless` for identical decoded pixels, including hidden RGB in transparent PNGs. Every lossy attempt starts from the original. Quality numbers are encoder-specific, not percentages of visual fidelity. [PNG invariants](docs/ARCHITECTURE.md#image-invariants) · [JPEG details](docs/jpeg.md).

Targets mean **bytes ≤ requested bytes**: `10KB` is 10,000; `10KiB` is 10,240. Resizing requires explicit permission. `--accept-nearest` saves a best over-limit result but still exits 6; size growth requires `--force-output`.

**Lossless pixels do not mean untouched metadata.** PNG `safe` removes text/time chunks but retains other metadata, including EXIF: it is not a complete privacy scrub. JPEG `safe` removes GPS/private EXIF while retaining ICC and display orientation. `--strip keep|safe|all` selects the policy; incompatible preservation/transform combinations are refused. [PNG metadata implementation](internal/pngimage/png.go) · [JPEG metadata](docs/jpeg.md#metadata-and-orientation).

Judge pixels as well as bytes with `smolpng compare images --recursive --output-dir artifacts/review --json`, which exports candidates and an offline visual report. [Comparison guide](docs/comparison.md).

## Share a recipe, not a script

Add this to your project's `.smolpng.toml` for a local PNG target with explicit resize permission:

```toml
[presets.mcp-icon]
schema = 1
extends = "auto"
formats = ["png"]
target_size = "10KiB"
min_quality = 55
allow_resize = true
min_dimensions = "64x64"
remote = "never"
```

Run `smolpng compress icon.png --preset mcp-icon --json`. Recipes inherit and export as data only, never scripts, credentials, executable paths, or permission to bypass the user's upload policy. [Configuration and presets](docs/presets.md) · [CC0 examples](presets/examples/).

Local agents use the same CLI with a version-bound [Agent Skill](docs/agent-skill.md). Instructions do not supply a runtime; a ChatGPT Skill package is not an MCP server or hosted compressor.

## Local by default

Auto starts with remote policy `never`; a TinyPNG key is **not upload permission**. With local tools installed, local-file compression works offline. `--offline` rejects URLs and remote processing. Explicit URLs download from their source; [optional TinyPNG](docs/remote.md) requires a permitting policy and reports attempted uploads even when local output wins. Dependency downloads and configured update checks are separate network operations.

There is no telemetry or analytics. Human/drop workflows retain local last-result metadata; ordinary `compress` does not silently create history. Reports and clipboard history can expose paths, and forced termination can leave temporary cache files. [Privacy and data](docs/PRIVACY-AND-DATA.md).

Originals stay intact unless you choose `--in-place`; completed writes are atomic. Optimizers run as your user, not in a sandbox. Do not expose this trusted-user CLI directly as a public upload service. [Security boundaries](docs/SECURITY.md).

## Development and the fuller destination

```sh
go mod verify
go test ./...
go vet ./...
go run ./tools/corpus --check
```

Tests substitute optimizer/HTTP/desktop boundaries. The [34-image CC0 corpus](testdata/corpus/README.md) includes 32 project-generated structural specimens, with complete RGB, grayscale, CMYK and Display P3 profile controls, plus two photographic originals; it does not establish general photographic quality or arbitrary color-profile fidelity. Cross-builds and native bridge compilation do not prove graphical behavior or signed Explorer activation. [Development and qualification](docs/DEVELOPMENT.md) · [Architecture](docs/ARCHITECTURE.md) · [Source map](docs/PROJECT-MAP.md).

Desktop integration and standalone-update source exist; production signing/feed operation, installers, and package publication remain separate work. The website, final mascot art, self-hosted `smolpng-web`, authenticated MCP integration, and hosting are future surfaces, not deployable services here. [Current direction](docs/DIRECTION.md) distinguishes implemented source from the [complete vision](docs/VISION.md).

An open-source project from Wolfsblvt Works.

## License

The application and ordinary documentation use [MIT](LICENSE), allowing use, modification, and redistribution, with CC0 examples/corpus, third-party terms, and separate future website/brand rights identified in the [license map](license-map.md).
