// SPDX-License-Identifier: MIT
package main

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func currentHostIdentity() hostIdentity {
	host := baseHostIdentity()
	version := windows.RtlGetVersion()
	host.KernelVersion = fmt.Sprintf("%d.%d.%d", version.MajorVersion, version.MinorVersion, version.BuildNumber)
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		host.DetailsUnavailable = "Windows edition/release/revision could not be read"
		return host
	}
	defer key.Close()
	missing := []string{}
	// ProductName can retain a Windows 10 label on Windows 11; use the native
	// kernel build and explicit edition/release values instead of that label.
	for _, field := range []struct {
		name   string
		target *string
	}{{"EditionID", &host.Edition}, {"DisplayVersion", &host.Release}} {
		value, _, err := key.GetStringValue(field.name)
		if err != nil {
			missing = append(missing, field.name)
		} else {
			*field.target = value
		}
	}
	revision, _, err := key.GetIntegerValue("UBR")
	if err != nil {
		missing = append(missing, "UBR")
	} else {
		host.Revision = &revision
	}
	if len(missing) > 0 {
		host.DetailsUnavailable = "Windows values unavailable: " + strings.Join(missing, ", ")
	}
	return host
}
