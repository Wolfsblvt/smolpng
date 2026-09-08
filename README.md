# smolpng

**Make images smol.** Local-first image compression for humans, scripts, and agents.

The current private source candidate provides a deterministic PNG CLI over separately installed pngquant and OxiPNG. It preserves originals, compares original-derived candidates and treats upload limits as exact byte targets. The full PNG/JPEG, desktop and agent destination is retained in [Vision](docs/VISION.md); unimplemented pieces are not advertised as working features.

## Quick start

Use Go 1.26.6 or a compatible newer toolchain. From the repository root:

```sh
go mod download
go build -o artifacts/bin/ ./cmd/smolpng
go run ./cmd/smolpng deps status
go run ./cmd/smolpng compress icon.png --target-size 10KiB --json
```

Install pngquant/OxiPNG separately or use `--pngquant-path` and `--oxipng-path`. The result defaults to `icon.smol.png` beside `icon.png`; existing outputs get a collision number. A target is `outputBytes <= requestedBytes`, so `10KB` is 10,000 bytes and `10KiB` is 10,240. Missing a target exits `6`; `--accept-nearest` may write the best result but does not pretend the target was met.

```sh
smolpng compress icon.png --preset lossless
smolpng compress images --recursive --output-dir optimized --dry-run --json
smolpng compress images --recursive --output-dir optimized --jobs 2 --jsonl
smolpng inspect icon.png --json
```

Detached clipboard/URL/stdin output uses the operating-system Downloads folder under `smolpng/`. `--in-place` is the explicit destructive exception. `--output -` streams clean PNG bytes; use `--report-file` instead of JSON on that same stream.

Built-ins: `auto`, `web`, `smallest`, `lossless`. Auto never uploads. The native clipboard, TUI, configuration/preset lifecycle, TinyPNG, managed installation and Agent Skill are the next product cuts, not currently claimed here. [CLI reference](docs/cli.md) · [Development](docs/DEVELOPMENT.md) · [Current boundary](docs/DIRECTION.md).

## Verify

```sh
go test ./...
go vet ./...
```

Tests exercise real filesystem/process boundaries and fake optimizer/HTTP boundaries using original CC0 specimens. Cross-compilation and fake processes do not qualify a real desktop or optimizer version. Standard packages do not bundle either optimizer.

An open-source project from Wolfsblvt Works.

## License

Application and ordinary docs: [MIT](LICENSE). Generated corpus: CC0. Dependencies and future website/brand terms remain separate in the [license map](license-map.md).
