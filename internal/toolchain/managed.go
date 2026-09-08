// SPDX-License-Identifier: MIT
package toolchain

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"github.com/Wolfsblvt/smolpng/internal/process"
)

type Installed struct {
	Artifact    Artifact          `json:"artifact"`
	Directory   string            `json:"directory"`
	Files       map[string]string `json:"files"`
	InstalledAt time.Time         `json:"installedAt"`
}
type managedIndex struct {
	SchemaVersion int                  `json:"schemaVersion"`
	Active        string               `json:"active"`
	Versions      map[string]Installed `json:"versions"`
}
type ManagedStatus struct {
	Name     string `json:"name"`
	State    string `json:"state"`
	Path     string `json:"path,omitempty"`
	Version  string `json:"version,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Versions int    `json:"retainedVersions"`
}
type Managed struct {
	Root   string
	Runner process.Runner
	Client *http.Client
}

func digest(b []byte) string    { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func toolName(name string) bool { return name == "oxipng" || name == "pngquant" }
func noManagedLinks(path string) error {
	abs, e := filepath.Abs(path)
	if e != nil {
		return e
	}
	for p := abs; ; p = filepath.Dir(p) {
		st, e := os.Lstat(p)
		if e != nil && !os.IsNotExist(e) {
			return e
		}
		if e == nil && st.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("managed path contains a symlink: %s", p)
		}
		if filepath.Dir(p) == p {
			break
		}
	}
	return nil
}
func (m Managed) index(name string) (managedIndex, error) {
	if !toolName(name) || m.Root == "" {
		return managedIndex{}, fmt.Errorf("invalid managed dependency root/name")
	}
	path := filepath.Join(m.Root, name, "index.json")
	if e := noManagedLinks(path); e != nil {
		return managedIndex{}, e
	}
	f, e := os.Open(path)
	if e != nil {
		return managedIndex{}, e
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil || !st.Mode().IsRegular() {
		return managedIndex{}, fmt.Errorf("managed index is not a regular file")
	}
	b, e := io.ReadAll(io.LimitReader(f, (64<<10)+1))
	if e != nil {
		return managedIndex{}, e
	}
	if len(b) > 64<<10 {
		return managedIndex{}, fmt.Errorf("managed index is too large")
	}
	var index managedIndex
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if e = dec.Decode(&index); e != nil {
		return index, e
	}
	if dec.Decode(new(any)) != io.EOF {
		return index, fmt.Errorf("trailing managed index data")
	}
	if index.SchemaVersion != 1 || index.Active == "" || len(index.Versions) > 32 {
		return index, fmt.Errorf("invalid managed index")
	}
	active, ok := index.Versions[index.Active]
	if !ok {
		return index, fmt.Errorf("managed index has no active version")
	}
	_ = active
	for key, v := range index.Versions {
		if key == "" || filepath.Base(key) != key || strings.ContainsAny(key, "\\/:\x00") || v.Directory != "versions/"+key || v.Artifact.Tool != name || len(v.Files) != 2 || v.Files[v.Artifact.Executable] == "" || v.Files["LICENSE"] == "" {
			return index, fmt.Errorf("invalid installed version record")
		}
		if e := validateArtifact(v.Artifact); e != nil {
			return index, e
		}
		for _, h := range v.Files {
			if b, e := hex.DecodeString(h); e != nil || len(b) != 32 {
				return index, fmt.Errorf("invalid installed hash")
			}
		}
	}
	return index, nil
}

var managedVersion = regexp.MustCompile(`^(0|[1-9][0-9]{0,8})\.(0|[1-9][0-9]{0,8})\.(0|[1-9][0-9]{0,8})$`)

func compareManagedVersion(a, b string) int {
	left, right := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		x, _ := strconv.Atoi(left[i])
		y, _ := strconv.Atoi(right[i])
		if x < y {
			return -1
		}
		if x > y {
			return 1
		}
	}
	return 0
}
func validateArtifact(a Artifact) error {
	u, e := url.Parse(a.URL)
	h, he := hex.DecodeString(a.SHA256)
	if a.Tool != "oxipng" || !managedVersion.MatchString(a.Version) || a.Executable != "oxipng" && a.Executable != "oxipng.exe" || a.Archive != "zip" && a.Archive != "tar.gz" || e != nil || u.Scheme != "https" || u.Host != "github.com" || u.User != nil || !strings.HasPrefix(u.Path, "/oxipng/oxipng/releases/download/v"+a.Version+"/") || u.RawQuery != "" || u.Fragment != "" || he != nil || len(h) != 32 {
		return fmt.Errorf("unqualified artifact metadata")
	}
	return nil
}
func installedMatch(root string, v Installed) error {
	dir := filepath.Join(root, filepath.FromSlash(v.Directory))
	if e := noManagedLinks(dir); e != nil {
		return e
	}
	entries, e := os.ReadDir(dir)
	if e != nil {
		return e
	}
	if len(entries) != len(v.Files) {
		return fmt.Errorf("installed directory contains missing or extra files")
	}
	for _, entry := range entries {
		want, ok := v.Files[entry.Name()]
		if !ok || entry.Type()&os.ModeSymlink != 0 || entry.IsDir() {
			return fmt.Errorf("unexpected installed content")
		}
		f, e := os.Open(filepath.Join(dir, entry.Name()))
		if e != nil {
			return e
		}
		st, e := f.Stat()
		if e != nil || !st.Mode().IsRegular() || st.Size() > maxArchiveBytes {
			f.Close()
			return fmt.Errorf("invalid installed file")
		}
		b, e := io.ReadAll(io.LimitReader(f, maxArchiveBytes+1))
		f.Close()
		if e != nil || digest(b) != want {
			return fmt.Errorf("installed bytes changed: %s", entry.Name())
		}
	}
	return nil
}
func (m Managed) Inspect(name string) ManagedStatus {
	status := ManagedStatus{Name: name, State: "missing"}
	index, e := m.index(name)
	if e != nil {
		if os.IsNotExist(e) {
			if _, st := os.Lstat(filepath.Join(m.Root, name)); st == nil {
				status.State = "untracked"
				status.Reason = "existing directory has no valid management index"
			}
		} else {
			status.State = "unknown"
			status.Reason = e.Error()
		}
		return status
	}
	v := index.Versions[index.Active]
	if v.Artifact.OS != runtime.GOOS || v.Artifact.Arch != runtime.GOARCH {
		status.State = "incompatible"
		status.Reason = "recorded managed optimizer belongs to another OS/architecture"
		return status
	}
	status.Version = v.Artifact.Version
	status.Versions = len(index.Versions)
	status.Path = filepath.Join(m.Root, name, filepath.FromSlash(v.Directory), v.Artifact.Executable)
	if e = installedMatch(filepath.Join(m.Root, name), v); e != nil {
		status.State = "modified"
		status.Reason = e.Error()
		return status
	}
	status.State = "current"
	return status
}
func (m Managed) lock(name string) (*os.File, error) {
	if !toolName(name) || m.Root == "" {
		return nil, fmt.Errorf("invalid tool")
	}
	if e := noManagedLinks(m.Root); e != nil {
		return nil, e
	}
	if e := os.MkdirAll(m.Root, 0700); e != nil {
		return nil, e
	}
	if e := noManagedLinks(filepath.Join(m.Root, name+".lock")); e != nil {
		return nil, e
	}
	return fsutil.TryLock(filepath.Join(m.Root, name+".lock"))
}
func (m Managed) download(ctx context.Context, a Artifact) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	c := m.Client
	if c == nil {
		c = &http.Client{}
	}
	client := *c
	client.CheckRedirect = func(r *http.Request, via []*http.Request) error {
		if len(via) >= 5 || r.URL.Scheme != "https" || r.URL.User != nil {
			return fmt.Errorf("unsupported download redirect")
		}
		switch r.URL.Hostname() {
		case "github.com", "release-assets.githubusercontent.com", "objects.githubusercontent.com", "github-releases.githubusercontent.com":
			return nil
		}
		return fmt.Errorf("download redirect left the approved upstream hosts")
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, a.URL, nil)
	if e != nil {
		return nil, e
	}
	req.Header.Set("User-Agent", "smolpng")
	res, e := client.Do(req)
	if e != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("upstream download failed; signed redirect/query data omitted")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("upstream archive returned HTTP %d", res.StatusCode)
	}
	if res.ContentLength > maxArchiveBytes {
		return nil, fmt.Errorf("archive exceeds size bound")
	}
	b, e := io.ReadAll(io.LimitReader(res.Body, maxArchiveBytes+1))
	if e != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("archive download was incomplete")
	}
	if int64(len(b)) > maxArchiveBytes {
		return nil, fmt.Errorf("archive exceeds size bound")
	}
	if digest(b) != a.SHA256 {
		return nil, fmt.Errorf("archive SHA-256 mismatch; no downloaded executable was run")
	}
	return b, nil
}

// Install verifies the complete upstream archive before extracting or executing
// any bytes. A new immutable version directory is staged; activation is a small
// atomic index update. Previously active files are never overwritten.
func (m Managed) Install(ctx context.Context, a Artifact) (ManagedStatus, error) {
	if e := ctx.Err(); e != nil {
		return ManagedStatus{}, e
	}
	if e := validateArtifact(a); e != nil {
		return ManagedStatus{}, e
	}
	if a.OS != runtime.GOOS || a.Arch != runtime.GOARCH {
		return ManagedStatus{}, fmt.Errorf("managed installation must match the running OS/architecture")
	}
	lock, e := m.lock(a.Tool)
	if e != nil {
		return ManagedStatus{}, e
	}
	defer lock.Close()
	toolRoot := filepath.Join(m.Root, a.Tool)
	index, e := m.index(a.Tool)
	if os.IsNotExist(e) {
		if _, er := os.Lstat(toolRoot); er == nil {
			return ManagedStatus{}, fmt.Errorf("refusing to adopt an existing untracked tool directory")
		}
		index = managedIndex{SchemaVersion: 1, Versions: map[string]Installed{}}
	} else if e != nil {
		return ManagedStatus{}, e
	}
	if len(index.Versions) > 0 {
		current := index.Versions[index.Active]
		if e = installedMatch(toolRoot, current); e != nil {
			return ManagedStatus{}, fmt.Errorf("active managed installation was modified: %w", e)
		}
		if compareManagedVersion(a.Version, current.Artifact.Version) < 0 {
			return ManagedStatus{}, fmt.Errorf("refusing managed downgrade from %s to %s; update smolpng or explicitly remove the unchanged managed installation first", current.Artifact.Version, a.Version)
		}
		if current.Artifact.SHA256 == a.SHA256 && current.Artifact.Version == a.Version {
			return m.Inspect(a.Tool), nil
		}
	}
	if len(index.Versions) >= 32 {
		return ManagedStatus{}, fmt.Errorf("retained version limit reached; remove or inspect the managed installation first")
	}
	archive, e := m.download(ctx, a)
	if e != nil {
		return ManagedStatus{}, e
	}
	binary, e := ArchiveExecutable(archive, a.Archive, a.Executable)
	if e != nil {
		return ManagedStatus{}, e
	}
	fresh := len(index.Versions) == 0
	finalRoot := toolRoot
	if fresh {
		toolRoot, e = os.MkdirTemp(m.Root, ".prepare-"+a.Tool+"-")
		if e != nil {
			return ManagedStatus{}, e
		}
		defer os.RemoveAll(toolRoot)
	}
	if e = os.MkdirAll(filepath.Join(toolRoot, "versions"), 0700); e != nil {
		return ManagedStatus{}, e
	}
	stage, e := os.MkdirTemp(filepath.Join(toolRoot, "versions"), ".install-")
	if e != nil {
		return ManagedStatus{}, e
	}
	defer os.RemoveAll(stage)
	if e = os.WriteFile(filepath.Join(stage, a.Executable), binary, 0700); e != nil {
		return ManagedStatus{}, e
	}
	if e = os.WriteFile(filepath.Join(stage, "LICENSE"), []byte(oxipngLicense), 0600); e != nil {
		return ManagedStatus{}, e
	}
	runner := m.Runner
	if runner == nil {
		runner = process.OSRunner{}
	}
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	v, e := runner.Run(probeCtx, process.Invocation{Executable: filepath.Join(stage, a.Executable), Args: []string{"--version"}})
	cancel()
	if e != nil || v.ExitCode != 0 || version.FindString(v.Stdout+" "+v.Stderr) != a.Version {
		return ManagedStatus{}, fmt.Errorf("downloaded optimizer failed the exact version probe; previous active version unchanged")
	}
	if e = ctx.Err(); e != nil {
		return ManagedStatus{}, e
	}
	key := a.Version + "-" + a.OS + "-" + a.Arch + "-" + a.SHA256[:12]
	record := Installed{Artifact: a, Directory: "versions/" + key, Files: map[string]string{a.Executable: digest(binary), "LICENSE": digest([]byte(oxipngLicense))}, InstalledAt: time.Now().UTC()}
	dest := filepath.Join(toolRoot, "versions", key)
	if _, e = os.Lstat(dest); e == nil {
		// A crash before index activation may have left the exact verified version.
		if e = installedMatch(toolRoot, record); e != nil {
			return ManagedStatus{}, fmt.Errorf("existing version directory has unexpected content: %w", e)
		}
	} else if !os.IsNotExist(e) {
		return ManagedStatus{}, e
	} else if e = fsutil.MoveNoReplace(stage, dest); e != nil {
		return ManagedStatus{}, e
	}
	index.Active = key
	index.Versions[key] = record
	b, e := json.MarshalIndent(index, "", "  ")
	if e != nil {
		return ManagedStatus{}, e
	}
	if e = fsutil.AtomicWrite(ctx, filepath.Join(toolRoot, "index.json"), append(b, '\n'), 0600, true); e != nil {
		return ManagedStatus{}, fmt.Errorf("verified version is staged but not activated: %w", e)
	}
	if fresh {
		if e = fsutil.MoveNoReplace(toolRoot, finalRoot); e != nil {
			return ManagedStatus{}, e
		}
	}
	return m.Inspect(a.Tool), nil
}

// Remove refuses modified or untracked content, including retired version files.
// It renames the owned tree out of discovery before deleting it, so locked-file
// failures cannot expose a half-removed active program.
func (m Managed) Remove(ctx context.Context, name string) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	lock, e := m.lock(name)
	if e != nil {
		return e
	}
	defer lock.Close()
	index, e := m.index(name)
	if e != nil {
		return e
	}
	root := filepath.Join(m.Root, name)
	expected := map[string]bool{"index.json": true, "versions": true}
	for key, v := range index.Versions {
		if e = installedMatch(root, v); e != nil {
			return e
		}
		expected["versions/"+key] = true
		for f := range v.Files {
			expected["versions/"+key+"/"+f] = true
		}
	}
	e = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == root {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		if !expected[filepath.ToSlash(rel)] || d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("untracked managed content: %s", rel)
		}
		return nil
	})
	if e != nil {
		return e
	}
	tomb, e := os.MkdirTemp(m.Root, ".remove-"+name+"-")
	if e != nil {
		return e
	}
	if e = os.Remove(tomb); e != nil {
		return e
	}
	if e = fsutil.MoveNoReplace(root, tomb); e != nil {
		return e
	}
	if e = os.RemoveAll(tomb); e != nil {
		return fmt.Errorf("dependency removed from discovery; cleanup remains at %s: %w", tomb, e)
	}
	return nil
}
func (m Managed) All() []ManagedStatus {
	names := []string{"pngquant", "oxipng"}
	sort.Strings(names)
	out := []ManagedStatus{}
	for _, n := range names {
		out = append(out, m.Inspect(n))
	}
	return out
}
