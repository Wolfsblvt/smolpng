# SPDX-License-Identifier: MIT
# Run in a Visual Studio developer PowerShell with the Windows 11 SDK available.
# Builds/stages only. Never registers a package, creates keys, installs trust,
# restarts Explorer, downloads tools or modifies the supplied application directory.
[CmdletBinding()]
param(
 [Parameter(Mandatory=$true)][string]$AppDirectory,
 [Parameter(Mandatory=$true)][ValidateSet('amd64','arm64')][string]$Architecture,
 [Parameter(Mandatory=$true)][string]$Publisher,
 [Parameter(Mandatory=$true)][string]$PackageVersion,
 [Parameter(Mandatory=$true)][string]$Out,
 [string]$SigningThumbprint = ''
)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$Root = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$AppDirectory = (Resolve-Path -LiteralPath $AppDirectory).Path
$Out = [IO.Path]::GetFullPath($Out)
if (Test-Path -LiteralPath $Out) { throw 'Output directory must be new.' }
if ($PackageVersion -notmatch '^\d+\.\d+\.\d+\.\d+$') { throw 'PackageVersion must contain four numeric components.' }
foreach($part in $PackageVersion.Split('.')) { if ([uint64]$part -gt 65535) { throw 'PackageVersion components must be at most 65535.' } }
if ($SigningThumbprint -and $SigningThumbprint -notmatch '^[0-9a-fA-F]{40}$') { throw 'Select an existing SHA-1 certificate thumbprint; never pass a private key or password.' }
function Run-Native([string]$Program,[string[]]$Arguments) {
 & $Program @Arguments
 if ($LASTEXITCODE -ne 0) { throw "$Program failed with exit $LASTEXITCODE" }
}
Push-Location $Root
try {
 $Commit = (& git rev-parse HEAD).Trim()
 if ($LASTEXITCODE -ne 0 -or $Commit -notmatch '^[0-9a-f]{40}$') { throw 'Source commit unavailable.' }
 if (& git status --porcelain --untracked-files=normal) { throw 'Build from clean source, not an uncommitted bridge.' }
 $CLI = Join-Path $AppDirectory 'smolpng.exe'
 $Metadata = (& $CLI version --json | ConvertFrom-Json)
 if ($LASTEXITCODE -ne 0 -or $Metadata.commit -ne $Commit -or $Metadata.platform -ne "windows/$Architecture") { throw 'Application candidate must match this exact source commit and architecture.' }
 foreach($name in @('smolpng.exe','smolpngw.exe','LICENSE','THIRD_PARTY_NOTICES.md','LICENSES')) { if (!(Test-Path -LiteralPath (Join-Path $AppDirectory $name))) { throw "Candidate is missing $name" } }
 New-Item -ItemType Directory -Path $Out | Out-Null
 $Build = Join-Path $Out 'build'
 $Payload = Join-Path $Out 'payload'
 $Package = Join-Path $Out 'package'
 New-Item -ItemType Directory -Path $Payload,$Package,(Join-Path $Package 'Assets') | Out-Null
 $TargetArch = if ($Architecture -eq 'amd64') { 'x64' } else { 'ARM64' }
 Run-Native 'cmake' @('-S',$PSScriptRoot,'-B',$Build,'-A',$TargetArch)
 Run-Native 'cmake' @('--build',$Build,'--config','Release')
 # x64 contract executable can be executed on x64 hosts. ARM64 binaries are
 # compiled only here; the target host runs its matching test executable.
 if ($Architecture -eq 'amd64') { Run-Native 'ctest' @('--test-dir',$Build,'-C','Release','--output-on-failure') }
 foreach($name in @('smolpng.exe','smolpngw.exe','LICENSE','THIRD_PARTY_NOTICES.md')) { Copy-Item -LiteralPath (Join-Path $AppDirectory $name) -Destination $Payload }
 Copy-Item -LiteralPath (Join-Path $AppDirectory 'LICENSES') -Destination $Payload -Recurse
 Copy-Item -LiteralPath (Join-Path $Build 'Release\SmolpngShell.dll') -Destination $Payload
 $Hashes = [ordered]@{}
 foreach($name in @('smolpng.exe','smolpngw.exe','SmolpngShell.dll')) { $Hashes[$name] = (Get-FileHash -LiteralPath (Join-Path $Payload $name) -Algorithm SHA256).Hash.ToLowerInvariant() }
 $Receipt = [ordered]@{schema=1;commit=$Commit;architecture=$Architecture;files=$Hashes}
 [IO.File]::WriteAllText((Join-Path $Payload 'shell-bridge.json'),($Receipt|ConvertTo-Json -Depth 5),[Text.UTF8Encoding]::new($false))
 $Manifest = [IO.File]::ReadAllText((Join-Path $PSScriptRoot 'AppxManifest.xml.in'))
 $Manifest = $Manifest.Replace('@PUBLISHER@',[Security.SecurityElement]::Escape($Publisher)).Replace('@VERSION@',$PackageVersion).Replace('@ARCH@',$Architecture.Replace('amd64','x64'))
 [IO.File]::WriteAllText((Join-Path $Package 'AppxManifest.xml'),$Manifest,[Text.UTF8Encoding]::new($false))
 # Neutral, project-generated solid placeholders. These are package resources,
 # not final commissioned brand assets, and contain no third-party icon artwork.
 Add-Type -AssemblyName System.Drawing
 foreach($item in @(@('StoreLogo.png',50),@('Square44x44Logo.png',44),@('Square150x150Logo.png',150))) {
  $Image = [Drawing.Bitmap]::new([int]$item[1],[int]$item[1])
  $Graphics = [Drawing.Graphics]::FromImage($Image)
  try { $Graphics.Clear([Drawing.Color]::FromArgb(255,48,61,80));$Image.Save((Join-Path (Join-Path $Package 'Assets') $item[0]),[Drawing.Imaging.ImageFormat]::Png) } finally { $Graphics.Dispose();$Image.Dispose() }
 }
 $MSIX = Join-Path $Out 'smolpng-shell.msix'
 # Sparse packages intentionally refer to external EXE/DLL files. MakeAppx /nv
 # skips its full-package file-presence check; activation still validates schema,
 # signatures, architecture and the app's receipt on the real Windows host.
 Run-Native 'makeappx.exe' @('pack','/d',$Package,'/p',$MSIX,'/nv')
 if ($SigningThumbprint) { Run-Native 'signtool.exe' @('sign','/sha1',$SigningThumbprint,'/fd','SHA256',$MSIX) }
 Write-Host "Staged native payload: $Payload"
 Write-Host "Package: $MSIX"
 if (!$SigningThumbprint) { Write-Warning 'UNSIGNED: cannot activate yet. Use your existing trusted certificate/signing route. No key or trust was created.' }
 Write-Host 'Use the staged payload smolpng.exe to install menu settings and activate this package. No installation was performed.'
} finally { Pop-Location }
