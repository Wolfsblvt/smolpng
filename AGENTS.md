# smolpng repository instructions

## Meaning

This repository owns smolpng, the local-first image compressor for humans, scripts, and agents. Wolf authorized this private implementation and its ordinary Git delivery. The application is PNG-first; PNG and JPEG, desktop convenience, shareable presets, Agent Skills, and distribution remain the accepted 1.0 destination. The current source implements the deterministic PNG CLI and focused tests; desktop and broader 1.0 behavior remain separately qualified.

## Source and boundaries

Read the current `docs/VISION.md`, `docs/DIRECTION.md`, `docs/DECISIONS.md`, and `docs/ARCHITECTURE.md` when present. During bootstrap, the owner-provided smolpng product and licensing briefs carry that meaning. Keep public publication, releases, package submission, Pages, credentials, signing, spend, and domain changes outside this grant. No additional mutable repository is declared.

The selected language is Go. All interfaces share one compression engine. Auto remains local unless explicitly permitted otherwise; a TinyPNG key is capability, not upload permission. Every lossy candidate starts from the original. Preserve originals, use atomic writes, and never invoke optimizers through a shell. Do not bundle pngquant or OxiPNG executables. Imported presets are data only. Never silently overwrite modified installed Skills. Mark real desktop behavior unqualified until exercised on the corresponding host.

## Durable branch

```text
main:
  maintainer integration: PR preferred
  external contributions: not currently accepted
  required pre-integration evidence: none
  required approval/review: none
  resolved conversations: no
  automatic CI: none at bootstrap
  automatic retained branch effects: none
  other pre-update evidence: inspect the effective diff and exact applicable test evidence
```

Bootstrap may commit directly. Later coherent work uses contribution-named branches and PRs where the connector permits. If remote writes fail, the owner explicitly authorized local Git delivery with full history in a downloadable repository ZIP. Preserve useful linear commits; no merge commits, public effects, or force update of durable branches.

## Commands

Bootstrap: `go mod download`. Build: `go build -o artifacts/bin/ ./cmd/smolpng`. Root verification: `go test ./...`; static check: `go vet ./...`. Smoke: `go run ./cmd/smolpng capabilities --json`. `docs/DEVELOPMENT.md` owns platform, fixture and real-host details.
