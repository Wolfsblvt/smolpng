//go:build !windows

// SPDX-License-Identifier: MIT
package main

func currentHostIdentity() hostIdentity {
	host := baseHostIdentity()
	host.DetailsUnavailable = "Detailed OS identity is currently collected only on Windows; record this host's OS release separately"
	return host
}
