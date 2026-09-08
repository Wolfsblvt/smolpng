// SPDX-License-Identifier: MIT
package integration

import "golang.org/x/sys/windows"

func NativeLocations() (Locations, error) {
	var l Locations
	var e error
	if l.Menu, e = windows.KnownFolderPath(windows.FOLDERID_Programs, 0); e != nil {
		return l, e
	}
	if l.Desktop, e = windows.KnownFolderPath(windows.FOLDERID_Desktop, 0); e != nil {
		return l, e
	}
	l.SendTo, e = windows.KnownFolderPath(windows.FOLDERID_SendTo, 0)
	return l, e
}
