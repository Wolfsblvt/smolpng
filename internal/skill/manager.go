// SPDX-License-Identifier: MIT
// Package skill manages only explicitly installed, hash-tracked smolpng Skills.
package skill

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/fsutil"
)

type Installation struct {
	Destination   string            `json:"destination"`
	Target        string            `json:"target"`
	Managed       bool              `json:"managed"`
	AppVersion    string            `json:"appVersion"`
	BundleVersion string            `json:"bundleVersion"`
	BundleHash    string            `json:"bundleHash"`
	Files         map[string]string `json:"files"`
}
type InstallationStatus struct {
	Installation
	State   string   `json:"state"`
	Changes []string `json:"changes"`
}
type Startup struct {
	PreviousVersion   string `json:"previousVersion,omitempty"`
	CurrentVersion    string `json:"currentVersion"`
	ExecutableChanged bool   `json:"executableChanged"`
	ContentChanged    bool   `json:"contentChanged"`
	Prompt            bool   `json:"prompt"`
	Ignored           bool   `json:"ignored"`
}
type Snapshot struct {
	SchemaVersion   int                  `json:"schemaVersion"`
	Bundle          Bundle               `json:"bundle"`
	Installations   []InstallationStatus `json:"installations"`
	Startup         Startup              `json:"startup"`
	RecoveryPending bool                 `json:"recoveryPending"`
	CleanupPending  []string             `json:"cleanupPending"`
}
type residue struct {
	Path  string            `json:"path"`
	Files map[string]string `json:"files"`
}
type transaction struct {
	Destination string        `json:"destination"`
	Stage       string        `json:"stage,omitempty"`
	Backup      string        `json:"backup,omitempty"`
	Previous    *Installation `json:"previous,omitempty"`
	Next        *Installation `json:"next,omitempty"`
}
type registry struct {
	SchemaVersion int            `json:"schemaVersion"`
	Installations []Installation `json:"installations"`
	LastVersion   string         `json:"lastVersion,omitempty"`
	LastBundle    string         `json:"lastBundle,omitempty"`
	PendingBundle string         `json:"pendingBundle,omitempty"`
	IgnoredBundle string         `json:"ignoredBundle,omitempty"`
	Pending       *transaction   `json:"pending,omitempty"`
	Cleanup       []residue      `json:"cleanup,omitempty"`
}
type Manager struct {
	StatePath  string
	Bundle     Bundle
	AppVersion string
}

// Inspect is strictly read-only, including first use and crash-recovery diagnosis.
func (m Manager) Inspect() (Snapshot, error) {
	s, e := m.load()
	if e != nil {
		return Snapshot{}, e
	}
	out := Snapshot{SchemaVersion: 1, Bundle: m.Bundle, Installations: []InstallationStatus{}, Startup: m.startup(s), RecoveryPending: s.Pending != nil, CleanupPending: []string{}}
	for _, i := range s.Installations {
		out.Installations = append(out.Installations, m.status(i))
	}
	for _, r := range s.Cleanup {
		out.CleanupPending = append(out.CleanupPending, r.Path)
	}
	return out, nil
}
func (m Manager) startup(s *registry) Startup {
	v := Startup{PreviousVersion: s.LastVersion, CurrentVersion: m.AppVersion, ExecutableChanged: s.LastVersion != "" && s.LastVersion != m.AppVersion, ContentChanged: s.LastBundle != "" && s.LastBundle != m.Bundle.Hash || s.PendingBundle == m.Bundle.Hash, Ignored: s.IgnoredBundle == m.Bundle.Hash}
	for _, i := range s.Installations {
		if i.BundleHash != m.Bundle.Hash {
			v.Prompt = !v.Ignored
			v.ContentChanged = true
		}
	}
	return v
}

// Observe is called at interactive startup, never printed into machine output.
// An unchanged Skill does not prompt merely because the executable version changed.
func (m Manager) Observe(ctx context.Context) (Startup, error) {
	var result Startup
	e := m.mutate(ctx, func(s *registry) error {
		result = m.startup(s)
		if result.Prompt {
			s.PendingBundle = m.Bundle.Hash
		}
		s.LastVersion = m.AppVersion
		s.LastBundle = m.Bundle.Hash
		return nil
	})
	return result, e
}
func (m Manager) IgnoreCurrent(ctx context.Context) error {
	return m.mutate(ctx, func(s *registry) error { s.IgnoredBundle = m.Bundle.Hash; s.PendingBundle = ""; return nil })
}

func (m Manager) status(i Installation) InstallationStatus {
	status := InstallationStatus{Installation: i, Changes: []string{}}
	actual, e := directoryHashes(i.Destination)
	if errors.Is(e, os.ErrNotExist) {
		status.State = "missing"
		status.Changes = append(status.Changes, "directory missing")
		return status
	}
	if e != nil {
		status.State = "unknown"
		status.Changes = append(status.Changes, e.Error())
		return status
	}
	for _, name := range sortedKeys(i.Files) {
		h, ok := actual[name]
		if !ok {
			status.Changes = append(status.Changes, "missing: "+name)
		} else if h != i.Files[name] {
			status.Changes = append(status.Changes, "modified: "+name)
		}
	}
	for _, name := range sortedKeys(actual) {
		if _, ok := i.Files[name]; !ok {
			status.Changes = append(status.Changes, "untracked: "+name)
		}
	}
	if len(status.Changes) > 0 {
		status.State = "modified"
		return status
	}
	status.State = "current"
	if i.BundleHash != m.Bundle.Hash {
		status.State = "stale"
	}
	return status
}
func (m Manager) Install(ctx context.Context, dest, target string, managed bool) error {
	return m.mutate(ctx, func(s *registry) error {
		p, e := destination(dest)
		if e != nil {
			return e
		}
		for _, i := range s.Installations {
			if samePath(i.Destination, p) {
				return fmt.Errorf("destination already tracked; inspect or sync it")
			}
		}
		if _, e = os.Lstat(p); !errors.Is(e, os.ErrNotExist) {
			if e != nil {
				return e
			}
			return fmt.Errorf("destination already exists; no untracked content is overwritten")
		}
		record := m.record(p, target, managed)
		return m.swap(ctx, s, nil, &record)
	})
}
func (m Manager) Sync(ctx context.Context, dest string) error {
	return m.mutate(ctx, func(s *registry) error {
		i, e := find(s, dest)
		if e != nil {
			return e
		}
		status := m.status(*i)
		if status.State == "current" {
			return nil
		}
		if status.State != "stale" {
			return fmt.Errorf("preserving %s installation at %s; inspect, restore manually, or forget it", status.State, i.Destination)
		}
		next := m.record(i.Destination, i.Target, i.Managed)
		previous := *i
		return m.swap(ctx, s, &previous, &next)
	})
}
func (m Manager) Uninstall(ctx context.Context, dest string) error {
	return m.mutate(ctx, func(s *registry) error {
		i, e := find(s, dest)
		if e != nil {
			return e
		}
		status := m.status(*i)
		if status.State == "missing" {
			replace(s, i.Destination, nil)
			return nil
		}
		if status.State != "current" && status.State != "stale" {
			return fmt.Errorf("preserving %s installation; forget stops tracking without deleting user bytes", status.State)
		}
		previous := *i
		return m.swap(ctx, s, &previous, nil)
	})
}
func (m Manager) Forget(ctx context.Context, dest string) error {
	return m.mutate(ctx, func(s *registry) error {
		i, e := find(s, dest)
		if e != nil {
			return e
		}
		replace(s, i.Destination, nil)
		return nil
	})
}

type SyncResult struct {
	Destination string `json:"destination"`
	Action      string `json:"action"`
	Error       string `json:"error,omitempty"`
}

// SyncAfterUpdate is the integration point for an actual in-app updater. It never
// writes a custom installation or a locally modified app-managed installation.
func (m Manager) SyncAfterUpdate(ctx context.Context) ([]SyncResult, error) {
	return m.SyncAll(ctx, false)
}
func (m Manager) SyncAll(ctx context.Context, includeCustom bool) ([]SyncResult, error) {
	snapshot, e := m.Inspect()
	if e != nil {
		return nil, e
	}
	out := []SyncResult{}
	for _, i := range snapshot.Installations {
		if e = ctx.Err(); e != nil {
			return out, e
		}
		r := SyncResult{Destination: i.Destination, Action: "unchanged"}
		switch {
		case i.State == "current":
		case i.State != "stale":
			r.Action = "preserved-" + i.State
		case !i.Managed && !includeCustom:
			r.Action = "needs-consent"
		default:
			if e = m.Sync(ctx, i.Destination); e != nil {
				r.Action = "refused"
				r.Error = e.Error()
			} else {
				r.Action = "synced"
			}
		}
		out = append(out, r)
	}
	return out, nil
}
func (m Manager) Recover(ctx context.Context) error {
	return m.mutate(ctx, func(*registry) error { return nil })
}
func (m Manager) record(dest, target string, managed bool) Installation {
	return Installation{dest, target, managed, m.AppVersion, m.Bundle.Version, m.Bundle.Hash, m.Bundle.Hashes}
}

func find(s *registry, dest string) (*Installation, error) {
	p, e := filepath.Abs(dest)
	if e != nil {
		return nil, e
	}
	for n := range s.Installations {
		if samePath(p, s.Installations[n].Destination) {
			return &s.Installations[n], nil
		}
	}
	return nil, fmt.Errorf("destination is not tracked by smolpng: %s", p)
}
func replace(s *registry, dest string, next *Installation) {
	items := []Installation{}
	for _, i := range s.Installations {
		if !samePath(dest, i.Destination) {
			items = append(items, i)
		}
	}
	if next != nil {
		items = append(items, *next)
	}
	s.Installations = items
}
func samePath(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	return a == b || runtime.GOOS == "windows" && strings.EqualFold(a, b)
}
func destination(dest string) (string, error) {
	p, e := filepath.Abs(dest)
	if e != nil {
		return "", e
	}
	// The CLI always supplies a dedicated child; keep the lower-level API safe too.
	if filepath.Base(p) != "smolpng" {
		return "", fmt.Errorf("Skill destination must be a dedicated smolpng directory")
	}
	if e = noSymlinks(p); e != nil {
		return "", e
	}
	return p, nil
}
func noSymlinks(p string) error {
	p, e := filepath.Abs(p)
	if e != nil {
		return e
	}
	for {
		st, e := os.Lstat(p)
		if e == nil && st.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to manage a symlink path: %s", p)
		}
		if e != nil && !errors.Is(e, os.ErrNotExist) {
			return e
		}
		parent := filepath.Dir(p)
		if parent == p {
			return nil
		}
		p = parent
	}
}
func directoryHashes(root string) (map[string]string, error) {
	if e := noSymlinks(root); e != nil {
		return nil, e
	}
	st, e := os.Lstat(root)
	if e != nil {
		return nil, e
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("destination is not a directory")
	}
	result := map[string]string{}
	total := int64(0)
	entries := 0
	e = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == root {
			return nil
		}
		entries++
		if entries > MaxFiles*4 {
			return fmt.Errorf("too many Skill entries")
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink found in Skill: %s", rel)
		}
		if d.IsDir() {
			children, e := os.ReadDir(p)
			if e != nil {
				return e
			}
			if len(children) == 0 {
				return fmt.Errorf("untracked empty directory: %s", rel)
			}
			return nil
		}
		if !d.Type().IsRegular() || !validFile(rel) {
			return fmt.Errorf("unsupported Skill entry: %s", rel)
		}
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		data, err := io.ReadAll(io.LimitReader(f, (1<<20)+1))
		f.Close()
		if err != nil {
			return err
		}
		total += int64(len(data))
		if total > MaxBytes || len(data) > 1<<20 || len(result) >= MaxFiles {
			return fmt.Errorf("installed Skill exceeds inspection limit")
		}
		result[rel] = digest(data)
		return nil
	})
	return result, e
}
func exact(root string, expected map[string]string) bool {
	actual, e := directoryHashes(root)
	return e == nil && reflect.DeepEqual(actual, expected)
}
func missing(p string) bool { _, e := os.Lstat(p); return errors.Is(e, os.ErrNotExist) }

func (m Manager) load() (*registry, error) {
	if e := noSymlinks(m.StatePath); e != nil {
		return nil, e
	}
	f, e := os.Open(m.StatePath)
	if errors.Is(e, os.ErrNotExist) {
		return &registry{SchemaVersion: 1, Installations: []Installation{}}, nil
	}
	if e != nil {
		return nil, e
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil {
		return nil, e
	}
	if !st.Mode().IsRegular() {
		return nil, fmt.Errorf("Skill registry is not a regular file")
	}
	data, e := io.ReadAll(io.LimitReader(f, 2<<20))
	if e != nil {
		return nil, e
	}
	if len(data) >= 2<<20 {
		return nil, fmt.Errorf("Skill registry exceeds 2MiB")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var s registry
	if e = dec.Decode(&s); e != nil {
		return nil, fmt.Errorf("read Skill registry: %w", e)
	}
	if dec.Decode(new(any)) != io.EOF {
		return nil, fmt.Errorf("trailing Skill registry content")
	}
	if s.SchemaVersion != 1 {
		return nil, fmt.Errorf("unsupported Skill registry version")
	}
	for _, i := range s.Installations {
		if !filepath.IsAbs(i.Destination) || filepath.Base(i.Destination) != "smolpng" || len(i.Files) == 0 {
			return nil, fmt.Errorf("invalid tracked destination")
		}
		for n, h := range i.Files {
			if !validFile(n) || len(h) != 64 {
				return nil, fmt.Errorf("invalid tracked hash")
			}
		}
	}
	if s.Pending != nil {
		if e = validateTransaction(s.Pending); e != nil {
			return nil, e
		}
	}
	for _, c := range s.Cleanup {
		if !filepath.IsAbs(c.Path) || !strings.HasPrefix(filepath.Base(c.Path), ".smolpng-skill-") {
			return nil, fmt.Errorf("invalid cleanup path")
		}
	}
	return &s, nil
}
func (m Manager) save(ctx context.Context, s *registry) error {
	b, e := json.MarshalIndent(s, "", "  ")
	if e != nil {
		return e
	}
	return fsutil.AtomicWrite(ctx, m.StatePath, append(b, '\n'), 0600, true)
}
func (m Manager) mutate(ctx context.Context, fn func(*registry) error) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	if e := noSymlinks(m.StatePath); e != nil {
		return e
	}
	if e := os.MkdirAll(filepath.Dir(m.StatePath), 0700); e != nil {
		return e
	}
	lockPath := m.StatePath + ".lock"
	if e := noSymlinks(lockPath); e != nil {
		return e
	}
	lock, e := fsutil.TryLock(lockPath)
	if e != nil {
		return fmt.Errorf("another Skill operation is running or its lock is unavailable: %w", e)
	}
	defer lock.Close()
	s, e := m.load()
	if e != nil {
		return e
	}
	if e = m.recover(ctx, s); e != nil {
		return e
	}
	if e = fn(s); e != nil {
		return e
	}
	return m.save(ctx, s)
}
