# Third-party notices

## Meaning

These are the exact libraries compiled into the current smolpng executable. Their license files were read from the checksum-verified source modules and retained without relicensing. External optimizers are separate programs, not bundled binaries.

| Module | License | Retained text |
| --- | --- | --- |
| `github.com/adrg/xdg v0.5.3` | MIT | `LICENSES/xdg.txt` |
| `github.com/bmatcuk/doublestar/v4 v4.10.0` | MIT | `LICENSES/doublestar.txt` |
| `github.com/spf13/cobra v1.10.2` | Apache-2.0 | `LICENSES/cobra.txt` |
| `github.com/spf13/pflag v1.0.9` | BSD-3-Clause | `LICENSES/pflag.txt` |
| `github.com/inconshreveable/mousetrap v1.1.0` | Apache-2.0 | `LICENSES/mousetrap.txt` |
| `golang.org/x/image v0.45.0` | BSD-3-Clause | `LICENSES/x-image.txt` |
| `golang.org/x/sys v0.47.0` | BSD-3-Clause | `LICENSES/x-sys.txt` |
| Go runtime/standard library 1.26.6 | BSD-3-Clause | `LICENSES/go.txt` |

The indirect testify pin belongs to xdg's module test graph and is not used or compiled into smolpng. The project uses ordinary Go tests, not its assertion library. Patent and notice files are retained beside the license where present.

pngquant keeps its upstream GPL/commercial licensing and additional notices: https://pngquant.org/licensing.html . OxiPNG is MIT: https://github.com/oxipng/oxipng . Standard smolpng archives bundle neither executable. Managed installation, containers and Store packaging require their own exact distribution qualification. External-process separation is not a universal legal clearance.

The selected standard MIT, CC0 and AGPLv3 texts are stored in `LICENSES/`. Storing AGPL does not apply it to this executable. No font or commissioned brand assets are included.
