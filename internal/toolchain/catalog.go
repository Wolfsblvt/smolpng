// SPDX-License-Identifier: MIT
package toolchain

import "fmt"

// Artifact identifies one immutable upstream archive, not a latest-download URL.
// Digests were read from the upstream GitHub release metadata on 2026-09-08.
// Runtime qualification remains separately recorded per actual host.
type Artifact struct {
	Tool          string `json:"tool"`
	Version       string `json:"version"`
	OS            string `json:"os"`
	Arch          string `json:"arch"`
	URL           string `json:"url"`
	SHA256        string `json:"sha256"`
	Archive       string `json:"archive"`
	Executable    string `json:"executable"`
	License       string `json:"license"`
	Source        string `json:"source"`
	Qualification string `json:"qualification"`
}

func ArtifactFor(tool, goos, arch string) (Artifact, error) {
	if tool != "oxipng" {
		return Artifact{}, fmt.Errorf("managed %s download has no qualified artifact/checksum in this build; install it separately and configure its path", tool)
	}
	targets := map[string][3]string{
		"linux/amd64":   {"x86_64-unknown-linux-musl.tar.gz", "1813750ef592c5350ca79c88f98b2c0876d05c826dc7156245256d4255c1ad17", "tar.gz"},
		"linux/arm64":   {"aarch64-unknown-linux-musl.tar.gz", "ff21aa2a005b324c8dfbdf13cd2f469fb6728b27dacded82635824ecd12e3b35", "tar.gz"},
		"windows/amd64": {"x86_64-pc-windows-msvc.zip", "7e940f83ee46874b73f53031f96a15834cb70b220af27391fb06fe7b4dd798e1", "zip"},
		"darwin/amd64":  {"x86_64-apple-darwin.tar.gz", "111883bbe42b25e01cb1ca41f39f8094e4830b173f0e42be342ecc4ca48f3131", "tar.gz"},
		"darwin/arm64":  {"aarch64-apple-darwin.tar.gz", "7039fcfc78e8aa1ed2b57d848057a0296f082e92b3e1807ac65402d10d926764", "tar.gz"},
	}
	v, ok := targets[goos+"/"+arch]
	if !ok {
		return Artifact{}, fmt.Errorf("no native OxiPNG artifact is qualified for %s/%s in this build; configure an independently installed executable", goos, arch)
	}
	exe := "oxipng"
	if goos == "windows" {
		exe += ".exe"
	}
	return Artifact{Tool: tool, Version: "10.2.1", OS: goos, Arch: arch, URL: "https://github.com/oxipng/oxipng/releases/download/v10.2.1/oxipng-10.2.1-" + v[0], SHA256: v[1], Archive: v[2], Executable: exe, License: "MIT", Source: "https://github.com/oxipng/oxipng/tree/v10.2.1", Qualification: "upstream version/archive digest pinned; installer probes exact version locally; real-host corpus qualification is separate"}, nil
}

const oxipngLicense = `The MIT License (MIT)
Copyright (c) 2016 Joshua Holmer

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`
