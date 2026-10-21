// SPDX-License-Identifier: MIT
package packaging

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/updates"
	"github.com/tc-hib/winres"
	"github.com/tc-hib/winres/version"
)

// VersionWindowsExecutable adds release identity to a newly built, unsigned PE.
// It preserves other resources and returns new bytes before archive hashes exist.
func VersionWindowsExecutable(executable []byte, name, releaseVersion, commit string) ([]byte, error) {
	info, err := windowsVersionInfo(name, releaseVersion, commit)
	if err != nil {
		return nil, err
	}
	resources, err := winres.LoadFromEXE(bytes.NewReader(executable))
	if err != nil && !errors.Is(err, winres.ErrNoResources) {
		return nil, fmt.Errorf("read Windows resources: %w", err)
	}
	if resources == nil {
		resources = &winres.ResourceSet{}
	}
	resources.SetVersionInfo(info)
	var output bytes.Buffer
	// The default refuses signed input. No signature-removal or preservation
	// override belongs in the ordinary local build path.
	if err := resources.WriteToEXE(&output, bytes.NewReader(executable)); err != nil {
		return nil, fmt.Errorf("write Windows version resources: %w", err)
	}
	return output.Bytes(), nil
}

func windowsVersionInfo(name, releaseVersion, commit string) (version.Info, error) {
	parts, prerelease, err := updates.VersionParts(releaseVersion)
	if err != nil {
		return version.Info{}, err
	}
	var number [4]uint16
	for index, part := range parts {
		if part > 65535 {
			return version.Info{}, fmt.Errorf("Windows version components must fit 16 bits: %s", releaseVersion)
		}
		number[index] = uint16(part)
	}
	var description string
	switch name {
	case "smolpng.exe":
		description = "smolpng image compressor"
	case "smolpngw.exe":
		description = "smolpng windowless launcher"
	case "smolpng-qualify.exe":
		description = "smolpng native qualification helper"
	default:
		return version.Info{}, fmt.Errorf("unknown Windows release executable %q", name)
	}
	info := version.Info{FileVersion: number, ProductVersion: number, Type: version.App}
	info.Flags.Prerelease = prerelease != ""
	for key, value := range map[string]string{
		version.CompanyName:      "Wolfsblvt Works",
		version.ProductName:      "smolpng",
		version.FileDescription:  description,
		version.FileVersion:      releaseVersion,
		version.ProductVersion:   releaseVersion,
		version.InternalName:     strings.TrimSuffix(name, ".exe"),
		version.OriginalFilename: name,
		version.Comments:         "Source commit: " + commit + "; distribution: standalone",
	} {
		if err := info.Set(version.LangDefault, key, value); err != nil {
			return version.Info{}, fmt.Errorf("Windows version field %s: %w", key, err)
		}
	}
	return info, nil
}
