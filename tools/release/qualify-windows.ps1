# SPDX-License-Identifier: MIT
param(
    [Parameter(Mandatory)][string]$BuildPath,
    [Parameter(Mandatory)][string]$ExpectedVersion,
    [Parameter(Mandatory)][string]$ExpectedCommit,
    [Parameter(Mandatory)][string]$ReportPath
)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

if (Test-Path -LiteralPath $ReportPath) { throw 'Report path must be new.' }
[IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName([IO.Path]::GetFullPath($ReportPath))) | Out-Null
$report = [ordered]@{
    complete = $false
    version = $ExpectedVersion
    source = $ExpectedCommit
    windows = [Environment]::OSVersion.VersionString
    powershell = $PSVersionTable.PSVersion.ToString()
    files = [Collections.Generic.List[object]]::new()
    archives = [Collections.Generic.List[object]]::new()
    meaning = 'Native version-resource reads and archive identity; no graphical or native ARM64 execution claim.'
}
function Save-Report {
    [IO.File]::WriteAllText($ReportPath, ($report | ConvertTo-Json -Depth 12) + "`n", [Text.UTF8Encoding]::new($false))
}
Save-Report
try {
    $index = Get-Content -LiteralPath (Join-Path $BuildPath 'build-index.json') -Raw | ConvertFrom-Json -AsHashtable
    if (-not $index['complete'] -or $index['source'] -cne $ExpectedCommit -or $index['version'] -cne $ExpectedVersion) {
        throw 'Build index does not establish the selected complete source/version.'
    }
    $expectedNumbers = @($ExpectedVersion.Split('-')[0].Split('.') | ForEach-Object { [int]$_ }) + @(0)
    foreach ($architecture in @('amd64', 'arm64')) {
        $platform = "windows-$architecture"
        $archivePath = Join-Path $BuildPath "smolpng-$ExpectedVersion-$platform.zip"
        $archive = [IO.Compression.ZipFile]::OpenRead($archivePath)
        try {
            $reader = [IO.StreamReader]::new($archive.GetEntry('manifest.json').Open())
            try { $manifest = $reader.ReadToEnd() | ConvertFrom-Json -AsHashtable } finally { $reader.Dispose() }
            if ($manifest['commit'] -cne $ExpectedCommit -or $manifest['version'] -cne $ExpectedVersion -or
                $manifest['platform'] -cne "windows/$architecture") { throw "Wrong archive identity: $platform" }
            foreach ($name in @('smolpng.exe', 'smolpngw.exe', 'smolpng-qualify.exe')) {
                $path = Join-Path $BuildPath "stage/$platform/$name"
                $data = [IO.File]::ReadAllBytes($path)
                $peOffset = [BitConverter]::ToInt32($data, 0x3c)
                if ([BitConverter]::ToUInt32($data, $peOffset) -ne 0x4550) { throw "Missing PE signature: $path" }
                $machine = [BitConverter]::ToUInt16($data, $peOffset + 4)
                $optional = $peOffset + 24
                if ([BitConverter]::ToUInt16($data, $optional) -ne 0x20b) { throw "Expected PE32+ image: $path" }
                $subsystem = [BitConverter]::ToUInt16($data, $optional + 68)
                $certificateOffset = [BitConverter]::ToUInt32($data, $optional + 112 + 4 * 8)
                $certificateSize = [BitConverter]::ToUInt32($data, $optional + 112 + 4 * 8 + 4)
                $expectedMachine = if ($architecture -eq 'amd64') { 0x8664 } else { 0xaa64 }
                $expectedSubsystem = if ($name -eq 'smolpngw.exe') { 2 } else { 3 }
                if ($machine -ne $expectedMachine -or $subsystem -ne $expectedSubsystem -or
                    $certificateOffset -ne 0 -or $certificateSize -ne 0) { throw "Wrong machine/subsystem or unexpected certificate: $path" }
                $info = [Diagnostics.FileVersionInfo]::GetVersionInfo($path)
                $hash = (Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash.ToLowerInvariant()
                $row = [ordered]@{
                    platform = $platform; name = $name; sha256 = $hash; bytes = $data.Length
                    machine = $machine; subsystem = $subsystem; embeddedCertificatePresent = $false
                    productName = $info.ProductName; companyName = $info.CompanyName
                    fileVersion = $info.FileVersion; productVersion = $info.ProductVersion
                    originalFilename = $info.OriginalFilename; description = $info.FileDescription
                    comments = $info.Comments; prerelease = $info.IsPreRelease
                    numericVersion = @($info.FileMajorPart, $info.FileMinorPart, $info.FileBuildPart, $info.FilePrivatePart)
                }
                $report.files.Add($row)
                Save-Report
                if ($info.ProductName -cne 'smolpng' -or $info.CompanyName -cne 'Wolfsblvt Works' -or
                    $info.FileVersion -cne $ExpectedVersion -or $info.ProductVersion -cne $ExpectedVersion -or
                    $info.OriginalFilename -cne $name -or [string]::IsNullOrEmpty($info.FileDescription) -or
                    $info.Comments -cne "Source commit: $ExpectedCommit; distribution: standalone" -or
                    $info.IsPreRelease -ne $ExpectedVersion.Contains('-') -or
                    ($row.numericVersion -join '.') -cne ($expectedNumbers -join '.')) { throw "Version resource mismatch: $path" }
                $member = $archive.GetEntry($name).Open()
                try { $memberHash = [Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($member)).ToLowerInvariant() }
                finally { $member.Dispose() }
                if ($manifest['files'][$name] -cne $hash -or $memberHash -cne $hash) { throw "Stage/archive/member receipt disagree: $path" }
            }
        } finally { $archive.Dispose() }
        $archiveHash = (Get-FileHash -LiteralPath $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
        $artifact = @($index['artifacts'] | Where-Object { $_['platform'] -ceq "windows/$architecture" })
        if ($artifact.Count -ne 1 -or $artifact[0]['sha256'] -cne $archiveHash) { throw "Archive/index digest mismatch: $platform" }
        $report.archives.Add([ordered]@{ platform = $platform; sha256 = $archiveHash; name = [IO.Path]::GetFileName($archivePath) })
        Save-Report
    }
    $report.complete = $true
} catch {
    $report['error'] = $_.Exception.Message
    throw
} finally { Save-Report }
