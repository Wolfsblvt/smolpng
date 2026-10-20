# Ubuntu arm64 DEB installation and removal qualification

## Meaning

This record binds a native Ubuntu arm64 installation, real PNG/JPEG execution and package removal to one exact smolpng fixture source. The ordinary package workflow now exercises the same DEB lifecycle on amd64 and arm64; this record owns the first arm64 result, its build inputs and artifact identities. It does not qualify a public release, production feed, desktop integrations, or the joined Windows application journey.

## Exact run

- Source: [`1f00b97e356e94e05fdd7db1093ca7a4d6cb7785`](https://github.com/Wolfsblvt/smolpng/commit/1f00b97e356e94e05fdd7db1093ca7a4d6cb7785), clean checkout, native package workflow. Both the expected source and qualifier commit matched this SHA.
- [Ubuntu arm64 DEB lifecycle run](https://github.com/Wolfsblvt/smolpng/actions/runs/39349003063), job **Ubuntu arm64 DEB lifecycle**, completed successfully at `2026-10-20T00:41:14Z`.
- [Exact-source Verify run](https://github.com/Wolfsblvt/smolpng/actions/runs/39349002918): all 11 tests, builds and native-shell compile jobs passed at the same source SHA.
- Retained arm64 artifact: `linux-deb-arm64-1f00b97e356e94e05fdd7db1093ca7a4d6cb7785-39349003063`. Its `journey/report.json` is complete and passing: **63 checks, 0 failures**, both PNG and JPEG.
- Host: Ubuntu **24.04.5 LTS**, Linux **6.17.0-1022-azure**, native `aarch64`, dpkg **1.22.6**, package architecture `arm64`.
- Application and qualification compiler: Go **1.26.6**, `GOOS=linux`, `GOARCH=arm64`, `CGO_ENABLED=0`.

## Build and artifact identities

- Fixture identity: `smolpng` version `0.0.0`, `linux/arm64`, source `1f00b97e356e94e05fdd7db1093ca7a4d6cb7785`. The release builder's `.invalid` URL was fixture metadata only; it was not downloaded.
- Fixture archive `smolpng-0.0.0-linux-arm64.tar.gz`: SHA-256 `a4d099ba167086790837d4f188174ad38047f9cc169025e79d0ca734acd8a065`.
- DEB fields: package `smolpng`, version `0.0.0`, architecture `arm64`. Fixture DEB SHA-256: `237d1a2c279c05ca7acaa72c8d5d6596390338fad0b3029034c6bf733e19d35a`.
- Installed `/usr/bin/smolpng` SHA-256: `743c56e8ce2dada8d655677909378916d90b6178a78f7fdeda45e7a64b70d149`; package query named `smolpng` as its owner and read back the expected source/version/platform.
- nFPM was source-built from module `github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.48.0` with module sum `h1:Yvd07lX8YqDD+rtTwVqyeQiDX+7cnun/jFbZD501uB8=`; runner binary SHA-256 `2f807e9f9dced5e049c255a88dfba0fa74ff8e8a6557f27b9f86b6eb11de38f5`.

## Native optimizer journey

The selected jpegoptim release has no managed Linux arm64 asset. An initial run exposed that gap. Ubuntu's available `jpegoptim 1.4.7-1` package then rejected the application's `--keep-all` argument, so it could not exercise the selected application path. The accepted jpegoptim dependency remained **1.5.6**; the working route verified the upstream source archive checksum and built it on the native runner. This did not change the dependency or selected outcome.

- The [official v1.5.6 release](https://github.com/tjko/jpegoptim/releases/tag/v1.5.6) publishes source archive `jpegoptim-1.5.6.tar.gz`, SHA-256 `661a808dfffa933d78c6beb47a2937d572b9f03e94cbaaab3d4c0d72f410e9be`. The source build used GCC **13.3.0**, GNU Make **4.3**, and Ubuntu `libjpeg-dev` **8c-2ubuntu11** for arm64 (with `libjpeg-turbo8-dev` **2.1.5-2ubuntu2**). The generated native executable reported jpegoptim **1.5.6**, libjpeg-turbo **8d**, and SHA-256 `eb1dffeb3c65d053d2cde5312b11de8a1d4b9cce1762c07a18dbcd590474e26d`.
- pngquant **2.18.0**, SHA-256 `37944aac125fb3dca042ad699323ebc7c935f564be1ba02c319fe136307e4140`; package `2.18.0-1build2`.
- OxiPNG **10.2.1**, SHA-256 `80d7955f81ee1c0279116b528a1675200717568e456d0ad2cf983882ff351fab`; installed from the app's managed catalog for arm64.
- The separately downloaded Ubuntu pngquant/runtime package hashes are retained in the workflow artifact as `pngquant-packages.sha256`.

## Installed use and removal

dpkg installed the matching `arm64` DEB, and `dpkg-query -S` named `/usr/bin/smolpng` as package-owned. The installed CLI reported system-package update ownership with `canApply=false`; no feed or trusted signing keys were configured. The real qualifier completed all 63 checks with the native optimizers, including PNG/JPEG Auto and strict lossless, targets, binary stdout, partial batches, mixed-format use and original preservation.

The workflow removed `smolpng`; the post-removal package-status query was empty and `/usr/bin/smolpng` was absent. All four per-user managed OxiPNG state files present before removal passed their SHA-256 checks afterward. The fixture package and executable were not published; no RPM/APK/Arch package, graphical desktop, production update, or Windows journey was exercised by this run.

## Earlier route observations

- [First matrix attempt](https://github.com/Wolfsblvt/smolpng/actions/runs/39348115919): arm64 package build, install and removal passed; optimizer preparation stopped because no managed Linux arm64 jpegoptim archive exists.
- [Ubuntu package attempt](https://github.com/Wolfsblvt/smolpng/actions/runs/39348458621): the Ubuntu `1.4.7-1` package was available on both architectures, but its direct response was `unrecognized option '--keep-all'`; PNG checks passed and JPEG qualification failed. Package removal passed on both hosts.
- [Native source-build attempt](https://github.com/Wolfsblvt/smolpng/actions/runs/39349003063): arm64 and amd64 DEB lifecycle jobs passed with the selected jpegoptim version; 11/11 Verify jobs passed on the exact source.

The installed candidate's other distribution formats, graphical behavior and production/signing path remain separately unobserved. This fixture is not a release or a product-completion claim; the broader joined application destination and current Work remain in [Issue #36](https://github.com/Wolfsblvt/smolpng/issues/36).
