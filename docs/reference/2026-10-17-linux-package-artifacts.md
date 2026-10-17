# Linux package artifact qualification

## Meaning

Eight unsigned Linux development packages were produced on Windows with nFPM 2.48.0: DEB, RPM, APK and Arch packages for amd64 and arm64. Independent readers verified installer metadata, executable architecture, all 61 payload files and their hashes, complete license notices, and file/directory permissions. This record binds that result to application source `e2b5f73cbb12b487e2d51f6b4e1c92601f923671` and repaired metadata producer `28b4c3a65a0004371512add5ed14570738869289`. It establishes package-file production and inspection, not Linux installation, runtime, signing or publication.

## Source and tools

The ordinary release builder produced standalone application archives with development version `0.0.0`, Go 1.27.2 and `CGO_ENABLED=0`. The completed amd64 artifact survived the interrupted original two-platform build; a subsequent one-platform arm64 build completed normally with two compiler workers. Only those completed, hash-verified archives supplied the final package inputs. The original overall interrupted build is not counted as complete.

| Input archive | Bytes | SHA-256 |
| --- | ---: | --- |
| `smolpng-0.0.0-linux-amd64.tar.gz` | 19043078 | `ecf52d11c7fecbde9fa27ef40051050c798695acd11702f5ad7fb80a9a5c7191` |
| `smolpng-0.0.0-linux-arm64.tar.gz` | 17650195 | `5e1b95aa045dce1b1232b1a2acb83afce0abc2375d2987f77e6b750292d8d59d` |

The actual `packaging.Metadata` implementation at the repaired producer commit generated both configurations through a one-use local caller. Application bytes were reused unchanged from the archives above. This is a qualification of the repaired producer against those exact inputs, not a claim that the application was rebuilt at the producer commit.

| Producer output | SHA-256 |
| --- | --- |
| `nfpm/amd64.yaml` | `8e1ed584b61c17fc550ba7979c8af084703f7feaa097da5f17d454d0564d059b` |
| `nfpm/arm64.yaml` | `016820e53126361a36806afc7a5406a03b6f9cdd1eeb6a5365221b5bebbeb3ff` |

The [official nFPM 2.48.0 Windows amd64 archive](https://github.com/goreleaser/nfpm/releases/tag/v2.48.0) was checked against its published digest `8ce12620212f5db557e370e35e2279ac7b106ea51e6a546a2be37a1a5d5bd8de` before execution. The executable reported version 2.48.0 and has SHA-256 `f5dee769551b5947c0591cb22fff78596427a1d6fa6526de24704cc6029d94a8`; its MIT license was retained locally.

Windows bsdtar/libarchive 3.8.9 independently read the package payloads, with executable SHA-256 `a3ac62fd4cdf619a7c53f71ca3fe5fb42f72020975670ff29afe3603f6416a8e`. [rpmfile 2.2.1](https://pypi.org/project/rpmfile/2.2.1/) read RPM installer headers from the verified MIT-licensed wheel, SHA-256 `bf2a4c488702638a95b85c4852891390bb8e54d174569c217fb21f591ead19c3`. That 18002-byte inspection tool was imported locally without a global installation or an application dependency. Its optional payload-decompression dependency was not installed; libarchive handled payloads.

## Reproduced defect and repair

The original generated definition produced an amd64 DEB whose dependency-license directory was `0775` and whose legal text files were `0664`. Windows staging exposed writable modes, and nFPM's default mask retained group-write permission. The CLI's explicit `0755` mode was already correct.

The producer now declares `umask: 0o022`. The focused regression failed on the old definition for both architectures and passed after the change. `go test ./internal/packaging -count=1`, `go vet ./internal/packaging` and `git diff --check` passed on the repaired source. No runtime engine, package payload selection or application dependency changed.

## Actual package results

Every row passed the same checks. Installer metadata says `smolpng` and version `0.0.0`; Arch adds packaging revision `-1`. Debian uses `amd64`/`arm64`, while RPM, APK and Arch use `x86_64`/`aarch64`. The embedded ELF machine is 62 for amd64 and 183 for arm64. Go build information independently identifies Linux and the matching architecture; neither ELF has a `PT_INTERP` program segment. These observations do not establish runtime compatibility on a target host.

| Architecture | Format | Bytes | SHA-256 |
| --- | --- | ---: | --- |
| amd64 | DEB | 11334358 | `488bc1c2b4c0fd4c8504903f49b5d27fbdd620156f40f12e150967b54d5ec3e6` |
| amd64 | RPM | 11493515 | `d07520849689aad248ca0682181d67f67033e65a2fe026fdc86f2028f819c924` |
| amd64 | APK | 11565286 | `1985917fc94e67b91b2184d2f5d4042dffe69f28f95d8831e1d4c8aaa640cdf3` |
| amd64 | Arch | 11337098 | `4c81b5274d9b999a643ba9bdebcfaefe13b22f00085ca48acdf57652eed4e1b6` |
| arm64 | DEB | 10361806 | `929761e360a06a723cfac4c1609d872f5283622f29e6afafa273b1d9ca358de5` |
| arm64 | RPM | 10500517 | `45935c69b3f61b6133d0f6a1d1fe9f28b4c279be2d1e2a2bf9e1147c1df8ed1e` |
| arm64 | APK | 10576558 | `8782376637eac6481d24fabb2e5a7475094c849276d158350aea939833702c60` |
| arm64 | Arch | 10340485 | `49017bba191ef1b8d14358aae9e8dad568d73adad65c580f04e1fe7b4ebaff9f` |

The 61 regular payload files comprise `/usr/bin/smolpng`, the primary license, 58 dependency-license files and third-party notices. Paths, types and the complete file set were checked before extraction. All payload directories were `0755`, the CLI was `0755`, and regular legal text was `0644`. Extracted file hashes matched both staging and the hash-bound source archive's builder receipt. No optimizer executable, qualification helper, personal configuration or image was included in these native packages.

The complete aggregate reports `complete=true`, `passed=true`, with eight passing package entries. Exact artifacts, per-command captures, installer metadata, payload inventories and extracted bytes were retained locally under `artifacts/validation/linux-nfpm-28b4c3a-final/`; `qualification.json` and `artifact-summary.json` carry the detailed account. These are local development outputs, not public download URLs.

## Earlier attempts and evidence boundary

Earlier CI-derived binaries were format fixtures with source-distribution identity; they did not supply the final standalone packages. Initial inspection attempts exposed an unsupported short bsdtar option and a Windows result-file replacement failure. They remain incomplete attempts. The supported `--ignore-zeros` option and separated live result-file reads were used for final inspection.

A later harness assertion incorrectly expected custom linker values in `go version -m` output. The observed trimmed binaries expose platform/toolchain information there, but not those values. The final check instead verified source, version, standalone ownership and every selected staged file against the source archive's hash-bound `manifest.json`. It did not execute a Linux binary or reinterpret missing build-info fields as a failed application build.

Native installation, package-manager update ownership, target-host execution, external optimizers, graphical integration and ARM64 hardware behavior remain unobserved by this record. The synthetic asset URLs are not live, the outputs are unsigned, and no registry submission, release upload or publication occurred. [Packaging](../packaging.md) owns the maintained production commands; the existing Vision and Direction retain the complete product destination.
