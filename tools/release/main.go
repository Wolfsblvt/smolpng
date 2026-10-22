// SPDX-License-Identifier: MIT
// Command release builds local private candidates and optional offline manifests.
// It has no publishing, network installation or signing effect.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/packaging"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"github.com/Wolfsblvt/smolpng/internal/updates"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, "release:", e)
		os.Exit(1)
	}
}
func run() error {
	version := flag.String("version", "", "Explicit candidate or stable semantic version")
	platforms := flag.String("platforms", "windows/amd64,windows/arm64,linux/amd64,linux/arm64,darwin/amd64,darwin/arm64", "Comma-separated build platforms")
	output := flag.String("out", "artifacts/releases", "New output directory; existing output is never replaced")
	base := flag.String("base-url", "", "Future HTTPS asset directory for offline package/update metadata; no publication occurs")
	flag.Parse()
	if flag.NArg() != 0 {
		return fmt.Errorf("unexpected argument")
	}
	_, pre, e := updates.VersionParts(*version)
	if e != nil {
		return e
	}
	if *base != "" {
		u, err := url.Parse(*base)
		if err != nil || u.RawQuery != "" || u.Fragment != "" {
			return fmt.Errorf("base-url must not contain a query or fragment")
		}
		if e = updates.HTTPS(*base); e != nil {
			return e
		}
		if pre != "" {
			return fmt.Errorf("package/update metadata needs a stable version; omit base-url for private prerelease candidates")
		}
	}
	git := func(args ...string) (string, error) {
		c := process.CommandContext(context.Background(), "git", args...)
		b, e := c.Output()
		return strings.TrimSpace(string(b)), e
	}
	root, e := git("rev-parse", "--show-toplevel")
	if e != nil {
		return e
	}
	cwd, e := os.Getwd()
	if e != nil || filepath.Clean(cwd) != filepath.Clean(root) {
		return fmt.Errorf("run from repository root")
	}
	dirty, e := git("status", "--porcelain")
	if e != nil || dirty != "" {
		return fmt.Errorf("build a clean committed source tree")
	}
	sha, e := git("rev-parse", "HEAD")
	if e != nil {
		return e
	}
	epoch, e := git("show", "-s", "--format=%ct", "HEAD")
	if e != nil {
		return e
	}
	seconds, e := strconv.ParseInt(epoch, 10, 64)
	if e != nil {
		return e
	}
	stamp := time.Unix(seconds, 0).UTC()
	out, e := filepath.Abs(*output)
	if e != nil {
		return e
	}
	if _, e = os.Lstat(out); !os.IsNotExist(e) {
		return fmt.Errorf("output directory must not already exist")
	}
	if e = os.MkdirAll(out, 0755); e != nil {
		return e
	}
	tree := process.CommandContext(context.Background(), "git", "ls-tree", "-r", "-z", "--full-tree", sha)
	tracked, e := tree.Output()
	if e != nil {
		return e
	}
	committed, e := committedFiles(tracked, func(object string) ([]byte, error) {
		cmd := process.CommandContext(context.Background(), "git", "cat-file", "blob", object)
		data, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("read committed Git blob %s: %w", object, err)
		}
		return data, nil
	})
	if e != nil {
		return e
	}
	sourceRoot, e := os.MkdirTemp("", "smolpng-release-source-*")
	if e != nil {
		return e
	}
	defer os.RemoveAll(sourceRoot)
	if e = materializeCommittedTree(sourceRoot, committed); e != nil {
		return e
	}
	common := releaseInputs(committed)
	artifacts := []packaging.Artifact{}
	manifest := updates.Manifest{Schema: 1, Product: "smolpng", Version: *version, Commit: sha, Channel: "stable", Issued: time.Now().UTC(), Expires: time.Now().UTC().Add(90 * 24 * time.Hour), Assets: map[string]updates.Asset{}}
	allHashes := map[string]string{}
	valid := map[string]bool{"windows/amd64": true, "windows/arm64": true, "linux/amd64": true, "linux/arm64": true, "darwin/amd64": true, "darwin/arm64": true}
	seen := map[string]bool{}
	for _, platform := range strings.Split(*platforms, ",") {
		if !valid[platform] || seen[platform] {
			return fmt.Errorf("invalid/duplicate platform %q", platform)
		}
		seen[platform] = true
		parts := strings.Split(platform, "/")
		goos, arch := parts[0], parts[1]
		id := goos + "-" + arch
		stage := filepath.Join(out, "stage", id)
		if e = os.MkdirAll(stage, 0755); e != nil {
			return e
		}
		files := map[string]packaging.Content{}
		for n, b := range common {
			files[n] = b
		}
		suffix := ""
		format := "tar.gz"
		if goos == "windows" {
			suffix = ".exe"
			format = "zip"
		}
		build := func(pkg, name, extra string) error {
			flags := "-X github.com/Wolfsblvt/smolpng/internal/cli.Version=" + *version + " -X github.com/Wolfsblvt/smolpng/internal/cli.Commit=" + sha + " -X github.com/Wolfsblvt/smolpng/internal/cli.Distribution=standalone -X main.Commit=" + sha + " " + extra
			cmd := process.CommandContext(context.Background(), "go", "build", "-trimpath", "-buildvcs=false", "-ldflags", flags, "-o", filepath.Join(stage, name), pkg)
			cmd.Dir = sourceRoot
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = releaseBuildEnvironment(os.Environ(), goos, arch)
			if e := cmd.Run(); e != nil {
				return e
			}
			b, e := os.ReadFile(filepath.Join(stage, name))
			if e != nil {
				return e
			}
			if goos == "windows" {
				b, e = packaging.VersionWindowsExecutable(b, name, *version, sha)
				if e != nil {
					return e
				}
				if e = os.WriteFile(filepath.Join(stage, name), b, 0755); e != nil {
					return e
				}
			}
			files[name] = packaging.Content{Data: b, Executable: true}
			return nil
		}
		if e = build("./cmd/smolpng", "smolpng"+suffix, ""); e != nil {
			return e
		}
		if goos == "windows" {
			if e = build("./cmd/smolpngw", "smolpngw.exe", "-H=windowsgui"); e != nil {
				return e
			}
		}
		if e = build("./tools/qualify", "smolpng-qualify"+suffix, ""); e != nil {
			return e
		}
		for name, c := range common {
			p := filepath.Join(stage, filepath.FromSlash(name))
			os.MkdirAll(filepath.Dir(p), 0755)
			if e = packaging.WriteNew(p, c.Data); e != nil {
				return e
			}
		}
		hashes := map[string]string{}
		for name, c := range files {
			hashes[name] = updates.SHA(c.Data)
		}
		receipt, _ := json.MarshalIndent(map[string]any{"schemaVersion": 1, "version": *version, "commit": sha, "platform": platform, "distribution": "standalone", "files": hashes, "nativeQualification": "not established by this build"}, "", "  ")
		files["manifest.json"] = packaging.Content{Data: append(receipt, '\n')}
		data, e := packaging.Archive(files, format, stamp)
		if e != nil {
			return e
		}
		name := "smolpng-" + *version + "-" + id + "." + format
		if e = packaging.WriteNew(filepath.Join(out, name), data); e != nil {
			return e
		}
		allHashes[name] = updates.SHA(data)
		artifact := packaging.Artifact{Platform: platform, Filename: name, SHA256: updates.SHA(data), Size: int64(len(data))}
		if *base != "" {
			artifact.URL = strings.TrimRight(*base, "/") + "/" + name
			updateName := "smolpng-update-" + *version + "-" + id + "." + format
			data, a, e := packaging.UpdateAsset(files, platform, strings.TrimRight(*base, "/")+"/"+updateName, format, stamp)
			if e != nil {
				return e
			}
			if e = packaging.WriteNew(filepath.Join(out, updateName), data); e != nil {
				return e
			}
			allHashes[updateName] = updates.SHA(data)
			manifest.Assets[platform] = a
		}
		artifacts = append(artifacts, artifact)
		b, _ := json.MarshalIndent(map[string]any{"source": sha, "version": *version, "complete": false, "artifacts": artifacts}, "", "  ")
		if e = os.WriteFile(filepath.Join(out, "build-index.json"), b, 0644); e != nil {
			return e
		}
		fmt.Println("Built and packaged", platform)
	}
	if *base != "" {
		b, _ := json.MarshalIndent(manifest, "", "  ")
		if e = packaging.WriteNew(filepath.Join(out, "release.UNSIGNED.json"), append(b, '\n')); e != nil {
			return e
		}
		generated, e := packaging.Metadata(*version, artifacts)
		if e != nil {
			return e
		}
		for name, data := range generated {
			p := filepath.Join(out, filepath.FromSlash(name))
			os.MkdirAll(filepath.Dir(p), 0755)
			if e = packaging.WriteNew(p, data); e != nil {
				return e
			}
		}
	}
	hashNames := []string{}
	for n := range allHashes {
		hashNames = append(hashNames, n)
	}
	sort.Strings(hashNames)
	var sums strings.Builder
	for _, n := range hashNames {
		fmt.Fprintf(&sums, "%s  %s\n", allHashes[n], n)
	}
	if e = packaging.WriteNew(filepath.Join(out, "SHA256SUMS"), []byte(sums.String())); e != nil {
		return e
	}
	b, _ := json.MarshalIndent(map[string]any{"source": sha, "version": *version, "complete": true, "artifacts": artifacts}, "", "  ")
	return os.WriteFile(filepath.Join(out, "build-index.json"), append(b, '\n'), 0644)
}
