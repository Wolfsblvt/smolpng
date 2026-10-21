// SPDX-License-Identifier: MIT
package packaging

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/updates"
	"strconv"
	"strings"
)

type Artifact struct {
	Platform string `json:"platform"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
	SHA256   string `json:"sha256"`
	Size     int64  `json:"size"`
}

func Metadata(version string, artifacts []Artifact) (map[string][]byte, error) {
	_, pre, e := updates.VersionParts(version)
	if e != nil || pre != "" {
		return nil, fmt.Errorf("package-manager metadata requires a stable semantic version")
	}
	index := map[string]Artifact{}
	for _, a := range artifacts {
		if e = updates.HTTPS(a.URL); e != nil {
			return nil, e
		}
		digest, err := hex.DecodeString(a.SHA256)
		if err != nil || len(digest) != 32 {
			return nil, fmt.Errorf("valid archive SHA-256 is required")
		}
		switch a.Platform {
		case "windows/amd64", "windows/arm64", "linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64":
		default:
			return nil, fmt.Errorf("unsupported package platform %q", a.Platform)
		}
		if _, exists := index[a.Platform]; exists {
			return nil, fmt.Errorf("duplicate package platform %q", a.Platform)
		}
		index[a.Platform] = a
	}
	out := map[string][]byte{}
	q := strconv.Quote
	arches := map[string]any{}
	var installers strings.Builder
	for _, pair := range [][3]string{{"windows/amd64", "x64", "64bit"}, {"windows/arm64", "arm64", "arm64"}} {
		a, ok := index[pair[0]]
		if !ok {
			continue
		}
		arches[pair[2]] = map[string]any{"url": a.URL, "hash": a.SHA256}
		fmt.Fprintf(&installers, "  - Architecture: %s\n    InstallerUrl: %s\n    InstallerSha256: %s\n", pair[1], q(a.URL), strings.ToUpper(a.SHA256))
	}
	if len(arches) > 0 {
		scoop, _ := json.MarshalIndent(map[string]any{"version": version, "description": "Local-first image compression for humans, scripts, and agents.", "homepage": "https://github.com/Wolfsblvt/smolpng", "license": "MIT", "architecture": arches, "bin": []string{"smolpng.exe", "smolpngw.exe"}, "notes": "External optimizers are not bundled. Run smolpng setup or deps install; TinyPNG never uploads without permission."}, "", "  ")
		out["scoop/smolpng.json"] = append(scoop, '\n')
		common := "PackageIdentifier: Wolfsblvt.smolpng\nPackageVersion: " + q(version) + "\n"
		out["winget/Wolfsblvt.smolpng.yaml"] = []byte(common + "DefaultLocale: en-US\nManifestType: version\nManifestVersion: 1.6.0\n")
		out["winget/Wolfsblvt.smolpng.locale.en-US.yaml"] = []byte(common + "PackageLocale: en-US\nPublisher: Wolfsblvt Works\nPackageName: smolpng\nLicense: MIT\nShortDescription: Local-first image compression for humans, scripts, and agents.\nPackageUrl: https://github.com/Wolfsblvt/smolpng\nManifestType: defaultLocale\nManifestVersion: 1.6.0\n")
		out["winget/Wolfsblvt.smolpng.installer.yaml"] = []byte(common + "InstallerType: zip\nNestedInstallerType: portable\nNestedInstallerFiles:\n  - RelativeFilePath: smolpng.exe\n    PortableCommandAlias: smolpng\n  - RelativeFilePath: smolpngw.exe\n    PortableCommandAlias: smolpngw\nInstallers:\n" + installers.String() + "ManifestType: installer\nManifestVersion: 1.6.0\n")
	}
	ruby := func(s string) string { return "'" + strings.NewReplacer(`\`, `\\`, "'", `\'`).Replace(s) + "'" }
	var brew strings.Builder
	brew.WriteString("# Generated for offline review; not submitted to a tap.\nclass Smolpng < Formula\n  desc 'Local-first image compression for humans, scripts, and agents'\n  homepage 'https://github.com/Wolfsblvt/smolpng'\n  license 'MIT'\n  version " + ruby(version) + "\n")
	brewAssets := 0
	for _, os := range []string{"darwin", "linux"} {
		block := "on_macos"
		if os == "linux" {
			block = "on_linux"
		}
		fmt.Fprintf(&brew, "  %s do\n", block)
		for _, arch := range []string{"amd64", "arm64"} {
			a, ok := index[os+"/"+arch]
			if !ok {
				continue
			}
			brewAssets++
			cond := "on_intel"
			if arch == "arm64" {
				cond = "on_arm"
			}
			fmt.Fprintf(&brew, "    %s do\n      url %s\n      sha256 %s\n    end\n", cond, ruby(a.URL), ruby(a.SHA256))
		}
		brew.WriteString("  end\n")
	}
	brew.WriteString("  def install\n    bin.install 'smolpng'\n    pkgshare.install 'LICENSES', 'LICENSE', 'license-map.md', 'third-party-notices.md'\n  end\n  def caveats\n    'Optimizers are not bundled. Run smolpng setup or deps install. TinyPNG requires explicit upload permission.'\n  end\n  test do\n    assert_match 'schemaVersion', shell_output(\"#{bin}/smolpng capabilities --json\")\n  end\nend\n")
	if brewAssets > 0 {
		out["homebrew/smolpng.rb"] = []byte(brew.String())
	}
	xmlEscape := func(s string) string { var b bytes.Buffer; xml.EscapeText(&b, []byte(s)); return b.String() }
	if a, ok := index["windows/amd64"]; ok {
		out["chocolatey/smolpng.nuspec"] = []byte(`<?xml version="1.0"?><package xmlns="http://schemas.microsoft.com/packaging/2015/06/nuspec.xsd"><metadata><id>smolpng</id><version>` + xmlEscape(version) + `</version><title>smolpng</title><authors>Wolfsblvt Works</authors><projectUrl>https://github.com/Wolfsblvt/smolpng</projectUrl><requireLicenseAcceptance>false</requireLicenseAcceptance><description>Local-first image compression. External optimizers are not bundled; run smolpng setup or deps install. TinyPNG requires explicit permission.</description><tags>image png jpeg compression cli</tags></metadata><files><file src="tools\**" target="tools" /></files></package>`)
		ps := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
		out["chocolatey/tools/chocolateyinstall.ps1"] = []byte("$ErrorActionPreference='Stop'\n$tools=Split-Path -Parent $MyInvocation.MyCommand.Definition\nInstall-ChocolateyZipPackage -PackageName 'smolpng' -Url64bit " + ps(a.URL) + " -UnzipLocation $tools -Checksum64 " + ps(a.SHA256) + " -ChecksumType64 'sha256'\nNew-Item (Join-Path $tools 'smolpng-qualify.exe.ignore') -ItemType File -Force | Out-Null\nNew-Item (Join-Path $tools 'smolpngw.exe.gui') -ItemType File -Force | Out-Null\n")
	}
	for _, arch := range []string{"amd64", "arm64"} {
		if _, ok := index["linux/"+arch]; !ok {
			continue
		}
		// Windows staging exposes writable modes; do not inherit nFPM's group-writable default.
		out["nfpm/"+arch+".yaml"] = []byte(fmt.Sprintf("# Run from the generated output root after separately qualifying/installing nFPM.\nname: smolpng\narch: %s\nplatform: linux\nversion: %s\nmaintainer: Wolfsblvt Works <wolf@wolfsblvt.com>\ndescription: Local-first image compression for humans, scripts, and agents.\nlicense: MIT\nhomepage: https://github.com/Wolfsblvt/smolpng\nsection: graphics\npriority: optional\numask: 0o022\ncontents:\n  - src: stage/linux-%s/smolpng\n    dst: /usr/bin/smolpng\n    file_info:\n      mode: 0755\n  - src: stage/linux-%s/LICENSE\n    dst: /usr/share/licenses/smolpng/LICENSE\n  - src: stage/linux-%s/LICENSES/\n    dst: /usr/share/licenses/smolpng/dependencies/\n    type: tree\n  - src: stage/linux-%s/third-party-notices.md\n    dst: /usr/share/doc/smolpng/third-party-notices.md\n", arch, q(version), arch, arch, arch, arch))
	}
	out["PUBLICATION-NOT-AUTHORIZED.md"] = []byte("# Generated metadata, not a publication\n\nThese files have not been submitted, accepted, installed or verified by any package manager. Package identities and URLs require separate publication authority and platform validation. Checksums describe the generated local archives, not proof that those URLs are live. Do not add credentials, private signing keys, or invoke submission commands from this directory automatically. External optimizers are not bundled. nFPM is a separately qualified packaging tool, not a runtime dependency.\n")
	return out, nil
}
