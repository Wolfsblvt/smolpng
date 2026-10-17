# Native Winget manifest validation, 17 October 2026

## Meaning

This record qualifies the generated Winget manifest shape against Windows Package Manager 1.29.380. Native validation exposed an unsupported portable-installer scope field, repaired in source `73407aa4624a47651fc74ba49d7bcdd3fee1843b`. The fixture is synthetic metadata, not a release, real downloadable package or installation qualification.

## Failure and repair

The actual `packaging.Metadata` generator at main `1df768014ebc32c41d3f7f72fc8a2a02a4a9cce9` emitted version, default-locale and installer YAML from its existing unit-test inputs: version `0.2.0`, reserved `example.test` URLs and dummy archive hashes. No generated string was manually reconstructed for validation.

`winget validate --manifest <fixture-directory> --disable-interactivity` reported `Scope is not supported for InstallerType portable` and exited with code `2316632104`. A second warning came from the test fixture's deliberately identical hashes across architectures; that was fixture data, not another generator repair.

The repair removes only `Scope: user` from the ZIP/portable installer manifest. The schema version, package identity, two executable aliases, architectures, URLs and exact supplied hashes remain unchanged. Scope at installation remains a Winget operation; this validation does not qualify installation behavior. Microsoft's [validate contract](https://learn.microsoft.com/en-us/windows/package-manager/winget/validate) and [portable workflow source](https://github.com/microsoft/winget-cli/blob/master/src/AppInstallerCLICore/Workflows/PortableFlow.cpp) supplied the native context.

## Passing specimen

The corrected generator used the same version and reserved URL pattern, with one deterministic SHA-256 per platform instead of duplicate dummy hashes. Native validation returned `Manifest validation succeeded.` and exit **0**, with no warnings. Processes used Windows no-console creation, captured stdout/stderr and a fixed timeout. No package was installed or submitted.

| Generated file | SHA-256 |
| --- | --- |
| `Wolfsblvt.smolpng.installer.yaml` | `2fb1230476a00594840e6935ea6c7fa942cb9ec06841e271907a7273d0c871b8` |
| `Wolfsblvt.smolpng.locale.en-US.yaml` | `ea3d4726d9bb06aade54a58dc83e08ae06f06ac148b845aef5a79fc8e0b23b59` |
| `Wolfsblvt.smolpng.yaml` | `e90953bb71109a16bd7d2173d4206744fbe12a23629d401dbf29b0759fae9187` |

The added packaging assertion was observed failing for the unsupported field and passing after its removal. `go test ./internal/packaging` and `go vet ./internal/packaging` passed. This supports metadata generation and native manifest acceptance; actual release assets, clean installation, registry acceptance and publication remain separate.
