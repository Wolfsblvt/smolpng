// SPDX-License-Identifier: MIT
package main

import "strings"

var isolatedReleaseGoEnvironment = map[string]struct{}{
	"CGO_ENABLED":    {},
	"GO111MODULE":    {},
	"GOAMD64":        {},
	"GOARCH":         {},
	"GOARM64":        {},
	"GOENV":          {},
	"GOEXPERIMENT":   {},
	"GOFIPS140":      {},
	"GOFLAGS":        {},
	"GOOS":           {},
	"GOROOT":         {},
	"GOTOOLCHAIN":    {},
	"GOWORK":         {},
	"TINIFY_API_KEY": {},
}

// releaseBuildEnvironment preserves ordinary process discovery and module
// transport settings while refusing ambient Go workspace, user-config, build
// flags, experiments, toolchain switching and architecture tuning that could
// make a clean repository produce bytes from a different source or target
// contract.
func releaseBuildEnvironment(current []string, goos, arch string) []string {
	env := make([]string, 0, len(current)+11)
	for _, item := range current {
		key, _, ok := strings.Cut(item, "=")
		if !ok {
			env = append(env, item)
			continue
		}
		if _, isolated := isolatedReleaseGoEnvironment[strings.ToUpper(key)]; isolated {
			continue
		}
		env = append(env, item)
	}
	env = append(env,
		"GOOS="+goos,
		"GOARCH="+arch,
		"CGO_ENABLED=0",
		"GO111MODULE=on",
		"GOWORK=off",
		"GOENV=off",
		"GOTOOLCHAIN=local",
		"GOFLAGS=",
		"GOEXPERIMENT=",
		"GOFIPS140=off",
	)
	switch arch {
	case "amd64":
		env = append(env, "GOAMD64=v1")
	case "arm64":
		env = append(env, "GOARM64=v8.0")
	}
	return env
}
