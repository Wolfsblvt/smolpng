// SPDX-License-Identifier: MIT
package packaging

import (
	"strings"
	"testing"

	"github.com/tc-hib/winres/version"
)

func TestWindowsVersionKeepsFullReleaseAndSourceIdentity(t *testing.T) {
	for _, release := range []string{"0.3.0-dev.12", "1.2.3", "65535.65535.65535"} {
		t.Run(release, func(t *testing.T) {
			info, err := windowsVersionInfo("smolpng.exe", release, "1234567890123456789012345678901234567890")
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := version.FromBytes(info.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			fields := decoded.Table().GetMainTranslation()
			if fields[version.FileVersion] != release || fields[version.ProductVersion] != release {
				t.Fatalf("full release identity lost: %+v", fields)
			}
			if fields[version.ProductName] != "smolpng" || fields[version.CompanyName] != "Wolfsblvt Works" ||
				fields[version.Comments] != "Source commit: 1234567890123456789012345678901234567890; distribution: standalone" {
				t.Fatalf("product/source identity lost: %+v", fields)
			}
			if decoded.Flags.Prerelease != strings.Contains(release, "-") {
				t.Fatalf("prerelease flag differs for %q", release)
			}
			if decoded.FileVersion != decoded.ProductVersion || decoded.FileVersion[3] != 0 {
				t.Fatalf("numeric version = %v / %v", decoded.FileVersion, decoded.ProductVersion)
			}
		})
	}
}

func TestWindowsVersionMapsNumericComponentsWithoutWrapping(t *testing.T) {
	info, err := windowsVersionInfo("smolpng.exe", "1.2.345-dev.9", "source")
	if err != nil {
		t.Fatal(err)
	}
	if info.FileVersion != [4]uint16{1, 2, 345, 0} {
		t.Fatalf("numeric version = %v", info.FileVersion)
	}
	for _, invalid := range []string{"65536.0.0", "0.65536.0", "0.0.65536", "latest", "1.2", "1.2.3+metadata"} {
		if _, err := windowsVersionInfo("smolpng.exe", invalid, "source"); err == nil {
			t.Fatalf("accepted unrepresentable/unsupported version %q", invalid)
		}
	}
}

func TestWindowsVersionNamesEachReleasedExecutable(t *testing.T) {
	for name, description := range map[string]string{
		"smolpng.exe":         "smolpng image compressor",
		"smolpngw.exe":        "smolpng windowless launcher",
		"smolpng-qualify.exe": "smolpng native qualification helper",
	} {
		info, err := windowsVersionInfo(name, "0.3.0-dev.12", "source")
		if err != nil {
			t.Fatal(err)
		}
		fields := info.Table().GetMainTranslation()
		if fields[version.OriginalFilename] != name || fields[version.InternalName] != strings.TrimSuffix(name, ".exe") ||
			fields[version.FileDescription] != description {
			t.Fatalf("wrong executable identity: %+v", fields)
		}
	}
}
