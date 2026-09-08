# Third-party notices

## Meaning

This document identifies third-party software used by smolpng and separates compiled Go libraries from externally installed optimizers. The initial repository contains no application dependencies or optimizer binaries. Candidate dependency downloads used to qualify the build environment are not application adoption or redistribution.

## Selected optimizer relationship

The PNG engine will invoke pngquant and OxiPNG as separate programs. Neither executable is bundled in a standard smolpng archive. Their exact versions, source licenses, and any managed-download checksums will be recorded before adoption.

## Dependency evidence

The implementation will pin Go module versions in `go.mod` and `go.sum`, retain applicable license notices, and keep development-only toolchain artifacts outside product distributions. No security-audit or live-optimizer claim follows from merely downloading a library.
