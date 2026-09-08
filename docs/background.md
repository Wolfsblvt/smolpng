# Explicit PNG background

## Meaning

`--background '#ffffff'` flattens transparency onto a chosen opaque RGB colour. It is an explicit pixel-changing preparation before local compression, shared by the CLI, Custom form and data-only presets. It does **not** detect a subject, remove an existing painted background, erase a colour key, or run a segmentation model. Those editing operations are not selected 1.0 work.

```powershell
smolpng compress transparent.png --background '#ffffff'
smolpng compress transparent.png --background black --target-size 10KiB
```

Accepted values are `#RRGGBB`, `white`, `black`; an empty value clears an inherited choice. Defaults preserve alpha. The result remains PNG and ordinary source/collision/target policies apply. The pipeline includes `smolpng-background` and the colour even without verbose arguments. Every resize/quantization starts from this one original-derived preparation; an unflattened original cannot win the comparison.

Partial alpha uses source-over compositing in encoded channels with 16-bit intermediate precision and round-to-nearest. An 8-bit input does not become 16-bit merely because it was flattened. This is not a colour-managed linear-light renderer. Non-sRGB ICC/cHRM/cICP interpretation requires a qualified colour conversion, so the safe policy refuses those images rather than mixing colour spaces. An explicit `--strip all` accepts discarding that interpretation. Ordinary gamma/sRGB labels and orientation are preserved by the existing metadata policy.

Strict lossless rejects flattening. TinyPNG combinations reject it before upload, because a remote comparison must not silently compare the unflattened source. JPEG has no alpha and rejects this PNG-only operation. A requested preparation can run without an external optimizer; further shrinking still benefits from pngquant and OxiPNG. A larger prepared result needs `--force-output`, like other explicit transformations.

Example shared preset:

```toml
[presets.white-card]
schema = 1
extends = "web"
formats = ["png"]
background = "#ffffff"
```
