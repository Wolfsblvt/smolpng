// SPDX-License-Identifier: MIT
package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
)

type hostIdentity struct {
	OS                 string  `json:"os"`
	Architecture       string  `json:"architecture"`
	KernelVersion      string  `json:"kernelVersion,omitempty"`
	Edition            string  `json:"edition,omitempty"`
	Release            string  `json:"release,omitempty"`
	Revision           *uint64 `json:"revision,omitempty"`
	DetailsUnavailable string  `json:"detailsUnavailable,omitempty"`
}

func baseHostIdentity() hostIdentity {
	return hostIdentity{OS: runtime.GOOS, Architecture: runtime.GOARCH}
}

func executableHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("executable is not a regular file")
	}
	hash := sha256.New()
	if _, err = io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func optimizerHashes(report []byte, formats []string) (map[string]string, error) {
	var health struct {
		Tools map[string]struct {
			Path string `json:"path"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(report, &health); err != nil {
		return nil, fmt.Errorf("cannot identify optimizer executables: %w", err)
	}
	hashes := map[string]string{}
	for _, format := range formats {
		for _, name := range toolsForFormat(format) {
			hash, err := executableHash(health.Tools[name].Path)
			if err != nil {
				return nil, fmt.Errorf("cannot hash selected optimizer %s: %w", name, err)
			}
			hashes[name] = hash
		}
	}
	return hashes, nil
}

func (s *suite) verifyOptimizerIdentity() {
	identity := check{Name: "final-optimizer-provenance"}
	current, err := optimizerHashes(s.data.Tools, s.data.Formats)
	if err != nil {
		identity.Errors = append(identity.Errors, err.Error())
	} else {
		for name, expected := range s.data.OptimizerSHA256 {
			if current[name] != expected {
				identity.Errors = append(identity.Errors, "optimizer executable changed during qualification: "+name)
			}
		}
	}
	s.add(identity)
}
