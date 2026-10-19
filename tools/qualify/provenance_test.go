// SPDX-License-Identifier: MIT
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestOptimizerIdentityUsesSelectedExecutableBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jpegoptim")
	if err := os.WriteFile(path, []byte("abc"), 0600); err != nil {
		t.Fatal(err)
	}
	report, err := json.Marshal(map[string]any{"tools": map[string]any{"jpegoptim": map[string]any{"path": path, "version": "same-version"}}})
	if err != nil {
		t.Fatal(err)
	}
	hashes, err := optimizerHashes(report, []string{"jpeg"})
	if err != nil || len(hashes) != 1 || hashes["jpegoptim"] != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatal(hashes, err)
	}
	if err := os.WriteFile(path, []byte("different bytes, same version"), 0600); err != nil {
		t.Fatal(err)
	}
	changed, err := optimizerHashes(report, []string{"jpeg"})
	if err != nil || changed["jpegoptim"] == hashes["jpegoptim"] {
		t.Fatal("changed executable retained its old identity", changed, err)
	}
	if _, err := optimizerHashes(report, []string{"png"}); err == nil {
		t.Fatal("missing selected PNG executables accepted")
	}
}

func TestOptimizerIdentityRefusesUnreadableOrNonFileInputs(t *testing.T) {
	for _, path := range []string{"", filepath.Join(t.TempDir(), "missing"), t.TempDir()} {
		report, err := json.Marshal(map[string]any{"tools": map[string]any{"jpegoptim": map[string]any{"path": path}}})
		if err != nil {
			t.Fatal(err)
		}
		if hashes, err := optimizerHashes(report, []string{"jpeg"}); err == nil || len(hashes) != 0 {
			t.Fatal("unreadable executable received an identity", hashes, err)
		}
	}
	if _, err := optimizerHashes([]byte("not JSON"), []string{"jpeg"}); err == nil {
		t.Fatal("malformed dependency report accepted")
	}
}

func TestChangedOptimizerCannotProducePassingEvidence(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "jpegoptim")
	if err := os.WriteFile(path, []byte("changed since preflight"), 0600); err != nil {
		t.Fatal(err)
	}
	tools, err := json.Marshal(map[string]any{"tools": map[string]any{"jpegoptim": map[string]any{"path": path}}})
	if err != nil {
		t.Fatal(err)
	}
	s := suite{root: root, data: evidence{Passed: true, Formats: []string{"jpeg"}, Tools: tools,
		OptimizerSHA256: map[string]string{"jpegoptim": "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"}}}
	s.verifyOptimizerIdentity()
	s.data.Complete = true
	if err := s.save(); err != nil {
		t.Fatal(err)
	}
	stored, err := os.ReadFile(filepath.Join(root, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report evidence
	if err := json.Unmarshal(stored, &report); err != nil {
		t.Fatal(err)
	}
	if report.Passed || !report.Complete || len(report.Checks) != 1 || len(report.Checks[0].Errors) == 0 {
		t.Fatal("changed optimizer retained passing qualification", report)
	}
}

func TestHostIdentityNamesItsActualPlatformAndCoverage(t *testing.T) {
	host := currentHostIdentity()
	if host.OS != runtime.GOOS || host.Architecture != runtime.GOARCH {
		t.Fatal(host)
	}
	if runtime.GOOS == "windows" {
		if host.KernelVersion == "" {
			t.Fatal("Windows kernel build omitted", host)
		}
	} else if host.DetailsUnavailable == "" {
		t.Fatal("uncollected OS details were not declared", host)
	}
}
