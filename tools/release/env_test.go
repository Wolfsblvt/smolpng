// SPDX-License-Identifier: MIT
package main

import (
	"strings"
	"testing"
)

func TestReleaseBuildEnvironmentRefusesAmbientSourceAndTargetOverrides(t *testing.T) {
	input := []string{
		"PATH=/tools",
		"HOME=/home/wolf",
		"gOwOrK=/outside/go.work",
		"GOENV=/outside/goenv",
		"GO111MODULE=off",
		"GOFLAGS=-tags=outside",
		"GOEXPERIMENT=arenas",
		"GOFIPS140=latest",
		"GOTOOLCHAIN=auto",
		"GOROOT=/outside/go",
		"GOOS=plan9",
		"GOARCH=386",
		"GOAMD64=v4",
		"GOARM64=v9.5",
		"CGO_ENABLED=1",
		"TINIFY_API_KEY=secret",
	}

	got := environmentByUpperKey(t, releaseBuildEnvironment(input, "linux", "amd64"))

	for key, want := range map[string]string{
		"PATH":         "/tools",
		"HOME":         "/home/wolf",
		"GOOS":         "linux",
		"GOARCH":       "amd64",
		"CGO_ENABLED":  "0",
		"GO111MODULE":  "on",
		"GOWORK":       "off",
		"GOENV":        "off",
		"GOTOOLCHAIN":  "local",
		"GOFLAGS":      "",
		"GOEXPERIMENT": "",
		"GOFIPS140":    "off",
		"GOAMD64":      "v1",
	} {
		if got[key] != want {
			t.Fatalf("%s = %q, want %q", key, got[key], want)
		}
	}
	for _, key := range []string{"GOARM64", "GOROOT", "TINIFY_API_KEY"} {
		if _, exists := got[key]; exists {
			t.Fatalf("%s must not survive the isolated release environment", key)
		}
	}
}

func TestReleaseBuildEnvironmentUsesArm64BaselineWithoutAmd64Tuning(t *testing.T) {
	got := environmentByUpperKey(t, releaseBuildEnvironment(
		[]string{"GOAMD64=v3", "GOARM64=v9.4", "PATH=/tools"},
		"windows",
		"arm64",
	))

	if got["GOARM64"] != "v8.0" {
		t.Fatalf("GOARM64 = %q, want v8.0", got["GOARM64"])
	}
	if _, exists := got["GOAMD64"]; exists {
		t.Fatal("GOAMD64 must not be present for an arm64 release build")
	}
}

func environmentByUpperKey(t *testing.T, values []string) map[string]string {
	t.Helper()
	result := map[string]string{}
	for _, value := range values {
		key, item, ok := strings.Cut(value, "=")
		if !ok {
			t.Fatalf("invalid environment entry %q", value)
		}
		key = strings.ToUpper(key)
		if _, duplicate := result[key]; duplicate {
			t.Fatalf("duplicate environment key %s", key)
		}
		result[key] = item
	}
	return result
}
