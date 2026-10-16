// SPDX-License-Identifier: MIT
package updates

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Previous struct {
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	Mode   uint32 `json:"mode"`
}
type Journal struct {
	Schema       int                 `json:"schema"`
	Phase        string              `json:"phase"`
	Directory    string              `json:"directory"`
	Platform     string              `json:"platform"`
	FromVersion  string              `json:"fromVersion"`
	Distribution string              `json:"distribution"`
	Created      time.Time           `json:"createdAt"`
	Envelope     []byte              `json:"envelope"`
	Previous     map[string]Previous `json:"previous"`
	SkillState   string              `json:"skillState,omitempty"`
	Error        string              `json:"error,omitempty"`
	SkillResult  string              `json:"skillResult,omitempty"`
}
type Plan struct {
	Stage       string `json:"stage"`
	FromVersion string `json:"fromVersion"`
	ToVersion   string `json:"toVersion"`
	Commit      string `json:"commit"`
	Files       []File `json:"files"`
	ResultPath  string `json:"resultPath"`
}
type Manager struct {
	Executable, Version, Distribution, Platform, SkillState string
	Trust                                                   Trust
	HTTP                                                    *http.Client
	Runner                                                  process.Runner
	Now                                                     func() time.Time
}

func (m Manager) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now().UTC()
}
func noLinks(p string) error {
	for d := filepath.Clean(p); ; d = filepath.Dir(d) {
		s, e := os.Lstat(d)
		if e != nil && !os.IsNotExist(e) {
			return e
		}
		if e == nil && s.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("update path contains a symlink: %s", d)
		}
		if filepath.Dir(d) == d {
			break
		}
	}
	return nil
}
func mainName(platform string) string {
	if strings.HasPrefix(platform, "windows/") {
		return "smolpng.exe"
	}
	return "smolpng"
}
func (m Manager) Prepare(ctx context.Context, envelope []byte, localArchive string) (Plan, error) {
	var plan Plan
	if !Installation(m.Executable, m.Distribution).CanApply {
		return plan, fmt.Errorf("self-update disabled: %s", Installation(m.Executable, m.Distribution).Instruction)
	}
	v, e := Verify(envelope, m.Trust, m.now())
	if e != nil {
		return plan, e
	}
	if !Newer(v.Manifest.Version, m.Version) {
		return plan, fmt.Errorf("release is not newer than the installed version; downgrades/equal-version replacement refused")
	}
	a, ok := v.Manifest.Assets[m.Platform]
	if !ok {
		return plan, fmt.Errorf("release has no archive for %s", m.Platform)
	}
	exe, e := filepath.Abs(m.Executable)
	if e != nil {
		return plan, e
	}
	if filepath.Base(exe) != mainName(m.Platform) {
		return plan, fmt.Errorf("self-update requires the standard executable name")
	}
	dir, e := filepath.EvalSymlinks(filepath.Dir(exe))
	if e != nil {
		return plan, e
	}
	exe = filepath.Join(dir, filepath.Base(exe))
	if e = noLinks(exe); e != nil {
		return plan, e
	}
	// Resolved physical location must also not belong to a package manager.
	if !Installation(exe, m.Distribution).CanApply {
		return plan, fmt.Errorf("resolved installation belongs to a package manager")
	}
	if s, e := os.Stat(filepath.Join(dir, ".git")); e == nil && s.IsDir() {
		return plan, fmt.Errorf("refusing update within a Git worktree")
	}
	lock, e := fsutil.TryLock(filepath.Join(dir, ".smolpng-update.lock"))
	if e != nil {
		return plan, fmt.Errorf("another update operation is active: %w", e)
	}
	defer lock.Close()
	var data []byte
	if localArchive != "" {
		data, e = ReadFile(localArchive, MaxArchive)
	} else {
		data, e = Fetch(ctx, a.URL, m.HTTP, MaxArchive)
	}
	if e != nil {
		return plan, e
	}
	if int64(len(data)) != a.Size || SHA(data) != a.SHA256 {
		return plan, fmt.Errorf("update archive integrity mismatch")
	}
	stage, e := os.MkdirTemp(dir, ".smolpng-update-stage-")
	if e != nil {
		return plan, e
	}
	// A failed preparation changes no installed file. The private staging folder
	// can be discarded before its durable journal is returned to the caller.
	retain := false
	defer func() {
		if !retain {
			_ = os.RemoveAll(stage)
		}
	}()
	if e = os.Mkdir(filepath.Join(stage, "new"), 0700); e != nil {
		return plan, e
	}
	if e = Extract(data, a, m.Platform, filepath.Join(stage, "new")); e != nil {
		return plan, e
	}
	j := Journal{Schema: 1, Phase: "staged", Directory: dir, Platform: m.Platform, FromVersion: m.Version, Distribution: m.Distribution, Created: m.now(), Envelope: envelope, Previous: map[string]Previous{}, SkillState: m.SkillState}
	for _, f := range a.Files {
		path := filepath.Join(dir, filepath.FromSlash(f.Name))
		h, size, e := fileHash(path)
		if e != nil {
			return plan, e
		}
		old := Previous{SHA256: h, Size: size, Mode: 0644}
		if h != "" {
			s, e := os.Stat(path)
			if e != nil {
				return plan, e
			}
			old.Mode = uint32(s.Mode().Perm())
			b, e := ReadFile(path, MaxFile)
			if e != nil || SHA(b) != h {
				return plan, fmt.Errorf("installed file changed while snapshotting")
			}
			if e = fsutil.AtomicWrite(ctx, filepath.Join(stage, "old", filepath.FromSlash(f.Name)), b, 0600, false); e != nil {
				return plan, e
			}
		}
		j.Previous[f.Name] = old
	}
	if j.Previous[mainName(m.Platform)].SHA256 == "" {
		return plan, fmt.Errorf("installed main executable not found")
	}
	if e = writeJournal(ctx, stage, j); e != nil {
		return plan, e
	}
	retain = true
	return Plan{Stage: stage, FromVersion: m.Version, ToVersion: v.Manifest.Version, Commit: v.Manifest.Commit, Files: a.Files, ResultPath: filepath.Join(stage, "journal.json")}, nil
}
func writeJournal(ctx context.Context, stage string, j Journal) error {
	b, e := json.MarshalIndent(j, "", "  ")
	if e != nil {
		return e
	}
	return fsutil.AtomicWrite(ctx, filepath.Join(stage, "journal.json"), append(b, '\n'), 0600, true)
}
func ReadJournal(stage string) (Journal, error) {
	var j Journal
	if e := noLinks(stage); e != nil {
		return j, e
	}
	b, e := ReadFile(filepath.Join(stage, "journal.json"), MaxManifest)
	if e != nil {
		return j, e
	}
	e = decode(b, &j)
	return j, e
}
func (m Manager) load(stage string, rollback bool) (Journal, Verified, Asset, error) {
	var v Verified
	var a Asset
	stage, e := filepath.Abs(stage)
	if e != nil {
		return Journal{}, v, a, e
	}
	j, e := ReadJournal(stage)
	if e != nil {
		return j, v, a, e
	}
	if j.Schema != 1 || filepath.Dir(stage) != j.Directory || !strings.HasPrefix(filepath.Base(stage), ".smolpng-update-stage-") || j.Platform != m.Platform {
		return j, v, a, fmt.Errorf("update journal does not match this platform/directory")
	}
	if !Installation(filepath.Join(j.Directory, mainName(j.Platform)), j.Distribution).CanApply {
		return j, v, a, fmt.Errorf("update destination is package-manager owned")
	}
	now := m.now()
	if rollback {
		now = j.Created
	}
	v, e = Verify(j.Envelope, m.Trust, now)
	if e != nil {
		return j, v, a, e
	}
	a, ok := v.Manifest.Assets[j.Platform]
	if !ok {
		return j, v, a, fmt.Errorf("journal platform absent from signed release")
	}
	if len(j.Previous) != len(a.Files) || !Newer(v.Manifest.Version, j.FromVersion) {
		return j, v, a, fmt.Errorf("invalid update snapshot")
	}
	for _, f := range a.Files {
		p, ok := j.Previous[f.Name]
		if !ok || p.SHA256 != "" && !hashPattern.MatchString(p.SHA256) || p.Size < 0 || p.Size > MaxFile {
			return j, v, a, fmt.Errorf("invalid prior file record")
		}
		if e = noLinks(filepath.Join(j.Directory, filepath.FromSlash(f.Name))); e != nil {
			return j, v, a, e
		}
		if p.SHA256 != "" {
			h, size, e := fileHash(filepath.Join(stage, "old", filepath.FromSlash(f.Name)))
			if e != nil || h != p.SHA256 || size != p.Size {
				return j, v, a, fmt.Errorf("rollback snapshot corrupted: %s", f.Name)
			}
		}
		if !rollback {
			h, size, e := fileHash(filepath.Join(stage, "new", filepath.FromSlash(f.Name)))
			if e != nil || h != f.SHA256 || size != f.Size {
				return j, v, a, fmt.Errorf("staged update corrupted: %s", f.Name)
			}
		}
	}
	return j, v, a, nil
}
func (m Manager) Apply(ctx context.Context, stage string) error {
	j, v, a, e := m.load(stage, false)
	if e != nil {
		return e
	}
	if j.Phase == "complete" {
		return nil
	}
	if j.Phase != "staged" && j.Phase != "applying" && j.Phase != "rolled-back" {
		return fmt.Errorf("update stage phase requires manual recovery: %s", j.Phase)
	}
	lock, e := fsutil.TryLock(filepath.Join(j.Directory, ".smolpng-update.lock"))
	if e != nil {
		return e
	}
	defer lock.Close()
	// Every current file must still match this transaction's prior or new bytes.
	for _, f := range a.Files {
		h, _, e := fileHash(filepath.Join(j.Directory, filepath.FromSlash(f.Name)))
		if e != nil || h != j.Previous[f.Name].SHA256 && h != f.SHA256 {
			return fmt.Errorf("installed file changed; refusing update: %s", f.Name)
		}
	}
	j.Phase = "applying"
	if e = writeJournal(ctx, stage, j); e != nil {
		return e
	}
	fail := func(cause error) error {
		recovery, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		rollbackErr := rollbackFiles(recovery, stage, &j, a)
		j.Error = cause.Error()
		if rollbackErr != nil {
			j.Phase = "recovery-required"
			j.Error += "; rollback: " + rollbackErr.Error()
		} else {
			j.Phase = "rolled-back"
		}
		_ = writeJournal(recovery, stage, j)
		return fmt.Errorf("update failed: %s", j.Error)
	}
	files := append([]File{}, a.Files...)
	sort.SliceStable(files, func(i, k int) bool {
		if files[i].Name == mainName(j.Platform) {
			return false
		}
		if files[k].Name == mainName(j.Platform) {
			return true
		}
		return files[i].Name < files[k].Name
	})
	for _, f := range files {
		dest := filepath.Join(j.Directory, filepath.FromSlash(f.Name))
		h, _, e := fileHash(dest)
		if e != nil || h != j.Previous[f.Name].SHA256 && h != f.SHA256 {
			return fail(fmt.Errorf("file changed during update: %s", f.Name))
		}
		if h == f.SHA256 {
			continue
		}
		b, e := ReadFile(filepath.Join(stage, "new", filepath.FromSlash(f.Name)), MaxFile)
		if e != nil || SHA(b) != f.SHA256 {
			return fail(fmt.Errorf("staged bytes changed"))
		}
		mode := os.FileMode(0644)
		if !strings.Contains(f.Name, "/") && strings.HasPrefix(f.Name, "smolpng") {
			mode = 0755
		}
		if e = fsutil.AtomicWrite(ctx, dest, b, mode, h != ""); e != nil {
			return fail(e)
		}
	}
	// Only signed, checksum-verified bytes are executed. Confirm their self-reported
	// source/version before allowing post-update Skill synchronization.
	runner := m.Runner
	if runner == nil {
		runner = process.OSRunner{}
	}
	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	exe := filepath.Join(j.Directory, mainName(j.Platform))
	result, e := runner.Run(probeCtx, process.Invocation{Executable: exe, Args: []string{"version", "--json"}})
	cancel()
	var build struct {
		Version      string `json:"version"`
		Commit       string `json:"commit"`
		Platform     string `json:"platform"`
		Distribution string `json:"distribution"`
	}
	if e != nil || result.ExitCode != 0 || json.Unmarshal([]byte(result.Stdout), &build) != nil || build.Version != v.Manifest.Version || build.Commit != v.Manifest.Commit || build.Platform != j.Platform || !Installation(exe, build.Distribution).CanApply {
		return fail(fmt.Errorf("new executable identity probe failed"))
	}
	j.Phase = "complete"
	j.Error = ""
	if e = writeJournal(ctx, stage, j); e != nil {
		return e
	}
	if j.SkillState != "" {
		syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		result, e := runner.Run(syncCtx, process.Invocation{Executable: exe, Args: []string{"__update-skill-sync", "--registry", j.SkillState}})
		cancel()
		if e != nil || result.ExitCode != 0 {
			j.SkillResult = "automatic Skill synchronization needs attention; modified/custom installations are preserved"
		} else {
			j.SkillResult = result.Stdout
		}
		// Skill failure does not roll back a verified executable or overwrite user edits.
		if e = writeJournal(context.Background(), stage, j); e != nil {
			return e
		}
	}
	return nil
}
func checkRollbackFiles(j Journal, a Asset) error {
	for _, f := range a.Files {
		h, _, e := fileHash(filepath.Join(j.Directory, filepath.FromSlash(f.Name)))
		if e != nil || h != j.Previous[f.Name].SHA256 && h != f.SHA256 {
			return fmt.Errorf("cannot restore changed file %s", f.Name)
		}
	}
	return nil
}

func rollbackFiles(ctx context.Context, stage string, j *Journal, a Asset) error {
	if e := checkRollbackFiles(*j, a); e != nil {
		return e
	}
	for _, f := range a.Files {
		old := j.Previous[f.Name]
		dest := filepath.Join(j.Directory, filepath.FromSlash(f.Name))
		h, _, e := fileHash(dest)
		if e != nil {
			return e
		}
		if h == old.SHA256 {
			continue
		}
		if old.SHA256 == "" {
			if e = os.Remove(dest); e != nil && !errors.Is(e, os.ErrNotExist) {
				return e
			}
			continue
		}
		b, e := ReadFile(filepath.Join(stage, "old", filepath.FromSlash(f.Name)), MaxFile)
		if e != nil || SHA(b) != old.SHA256 {
			return fmt.Errorf("rollback bytes changed")
		}
		if e = fsutil.AtomicWrite(ctx, dest, b, os.FileMode(old.Mode).Perm(), true); e != nil {
			return e
		}
	}
	return nil
}
func (m Manager) Rollback(ctx context.Context, stage string) error {
	j, _, a, e := m.load(stage, true)
	if e != nil {
		return e
	}
	lock, e := fsutil.TryLock(filepath.Join(j.Directory, ".smolpng-update.lock"))
	if e != nil {
		return e
	}
	defer lock.Close()
	j.Phase = "rolling-back"
	j.Error = ""
	if e = writeJournal(ctx, stage, j); e != nil {
		return e
	}
	if e = rollbackFiles(ctx, stage, &j, a); e != nil {
		j.Phase = "recovery-required"
		j.Error = e.Error()
		return errors.Join(e, writeJournal(ctx, stage, j))
	}
	j.Phase = "rolled-back"
	j.Error = ""
	return writeJournal(ctx, stage, j)
}
