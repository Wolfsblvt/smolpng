// SPDX-License-Identifier: MIT
package integration

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"os"
	"path/filepath"
	"slices"
)

type Record struct {
	Entry   Entry  `json:"entry"`
	SHA256  string `json:"sha256"`
	Version string `json:"version"`
}
type Pending struct {
	Next   Record `json:"next"`
	Before string `json:"before,omitempty"`
}
type Registry struct {
	Schema  int      `json:"schema"`
	Records []Record `json:"records"`
	Pending *Pending `json:"pending,omitempty"`
}
type Status struct {
	Record
	State string `json:"state"`
}
type Manager struct {
	StatePath, Version string
	Allowed            Locations
	Renderer           Renderer
}

func digest(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func noLinks(path string) error {
	for p := filepath.Clean(path); ; p = filepath.Dir(p) {
		s, e := os.Lstat(p)
		if e != nil && !os.IsNotExist(e) {
			return e
		}
		if e == nil && s.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink not allowed: %s", p)
		}
		if filepath.Dir(p) == p {
			break
		}
	}
	return nil
}
func hashFile(path string) (string, error) {
	if e := noLinks(path); e != nil {
		return "", e
	}
	s, e := os.Lstat(path)
	if os.IsNotExist(e) {
		return "", nil
	}
	if e != nil {
		return "", e
	}
	if !s.Mode().IsRegular() || s.Size() > 1<<20 {
		return "", fmt.Errorf("unexpected launcher file")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	return digest(b), nil
}
func (m Manager) validate(x Entry) error {
	if !identifier.MatchString(x.ID) || !filepath.IsAbs(x.Path) {
		return fmt.Errorf("invalid managed launcher record")
	}
	extension := filepath.Ext(x.Path)
	if extension != ".lnk" && extension != ".desktop" && !(extension == ".ini" && x.ID == "shell-menu" && filepath.Dir(x.Path) == m.Allowed.ShellState) {
		return fmt.Errorf("invalid launcher extension")
	}
	if filepath.Base(x.Path) != "smolpng-"+x.ID+extension {
		return fmt.Errorf("invalid managed launcher filename")
	}
	dir := filepath.Dir(x.Path)
	if !slices.Contains([]string{m.Allowed.Menu, m.Allowed.Desktop, m.Allowed.SendTo, m.Allowed.ShellState, m.Allowed.ServiceMenus}, dir) {
		return fmt.Errorf("launcher location no longer matches native user folders")
	}
	return noLinks(x.Path)
}
func (m Manager) load() (Registry, error) {
	s := Registry{Schema: 1, Records: []Record{}}
	if e := noLinks(m.StatePath); e != nil {
		return s, e
	}
	b, e := os.ReadFile(m.StatePath)
	if os.IsNotExist(e) {
		return s, nil
	}
	if e != nil {
		return s, e
	}
	if len(b) > 1<<20 {
		return s, fmt.Errorf("integration registry too large")
	}
	if e = json.Unmarshal(b, &s); e != nil || s.Schema != 1 {
		return s, fmt.Errorf("invalid integration registry")
	}
	seen := map[string]bool{}
	for _, rec := range s.Records {
		if e = m.validate(rec.Entry); e != nil {
			return s, e
		}
		if seen[rec.Entry.Path] || len(rec.SHA256) != 64 {
			return s, fmt.Errorf("invalid duplicate/hash in launcher registry")
		}
		seen[rec.Entry.Path] = true
	}
	if s.Pending != nil {
		if e = m.validate(s.Pending.Next.Entry); e != nil {
			return s, e
		}
	}
	return s, nil
}
func (m Manager) save(ctx context.Context, s Registry) error {
	b, e := json.MarshalIndent(s, "", "  ")
	if e != nil {
		return e
	}
	return fsutil.AtomicWrite(ctx, m.StatePath, append(b, '\n'), 0600, true)
}
func (m Manager) Status() ([]Status, bool, error) {
	s, e := m.load()
	if e != nil {
		return nil, false, e
	}
	out := []Status{}
	for _, r := range s.Records {
		h, e := hashFile(r.Entry.Path)
		state := "current"
		if e != nil {
			state = "unreadable"
		} else if h == "" {
			state = "missing"
		} else if h != r.SHA256 {
			state = "modified"
		}
		out = append(out, Status{Record: r, State: state})
	}
	return out, s.Pending != nil, nil
}
func (m Manager) mutate(ctx context.Context, fn func(*Registry) error) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	if e := noLinks(m.StatePath); e != nil {
		return e
	}
	if e := os.MkdirAll(filepath.Dir(m.StatePath), 0700); e != nil {
		return e
	}
	lock, e := fsutil.TryLock(m.StatePath + ".lock")
	if e != nil {
		return fmt.Errorf("integration operation already active: %w", e)
	}
	defer lock.Close()
	s, e := m.load()
	if e != nil {
		return e
	}
	if s.Pending != nil {
		p := s.Pending
		h, e := hashFile(p.Next.Entry.Path)
		if e != nil {
			return e
		}
		if h == p.Next.SHA256 {
			replace(&s, p.Next)
		} else if h != p.Before {
			return fmt.Errorf("pending launcher changed; preserve it and resolve manually")
		}
		s.Pending = nil
		if e = m.save(ctx, s); e != nil {
			return e
		}
	}
	return fn(&s)
}
func replace(s *Registry, r Record) {
	for i, old := range s.Records {
		if old.Entry.Path == r.Entry.Path {
			s.Records[i] = r
			return
		}
	}
	s.Records = append(s.Records, r)
}
func (m Manager) Install(ctx context.Context, entries []Entry) error {
	return m.mutate(ctx, func(s *Registry) error {
		// Validate the whole plan before installing its first file.
		for _, x := range entries {
			if e := m.validate(x); e != nil {
				return e
			}
			before, e := hashFile(x.Path)
			if e != nil {
				return e
			}
			owned := false
			for _, rec := range s.Records {
				if rec.Entry.Path == x.Path {
					owned = true
					if before != "" && before != rec.SHA256 {
						return fmt.Errorf("preserving modified shortcut: %s", x.Path)
					}
				}
			}
			if before != "" && !owned {
				return fmt.Errorf("preserving unowned shortcut: %s", x.Path)
			}
			st, e := os.Stat(x.Executable)
			if e != nil || !st.Mode().IsRegular() {
				return fmt.Errorf("launcher executable missing: %s", x.Executable)
			}
		}
		for _, x := range entries {
			if m.Renderer == nil {
				return fmt.Errorf("launcher renderer unavailable")
			}
			b, e := m.Renderer.Render(ctx, x)
			if e != nil {
				return e
			}
			if len(b) == 0 || len(b) > 1<<20 {
				return fmt.Errorf("invalid rendered launcher")
			}
			before, e := hashFile(x.Path)
			if e != nil {
				return e
			}
			rec := Record{Entry: x, SHA256: digest(b), Version: m.Version}
			s.Pending = &Pending{Next: rec, Before: before}
			if e = m.save(ctx, *s); e != nil {
				return e
			}
			// Recheck after rendering and journaling. Never overwrite locally changed bytes.
			h, e := hashFile(x.Path)
			if e != nil || h != before {
				return fmt.Errorf("launcher changed during installation: %s", x.Path)
			}
			if e = fsutil.AtomicWrite(ctx, x.Path, b, 0755, before != ""); e != nil {
				return e
			}
			replace(s, rec)
			s.Pending = nil
			if e = m.save(ctx, *s); e != nil {
				return e
			}
		}
		return nil
	})
}
func (m Manager) Remove(ctx context.Context, id string) error {
	return m.mutate(ctx, func(s *Registry) error {
		kept := []Record{}
		for _, rec := range s.Records {
			if id != "" && rec.Entry.ID != id {
				kept = append(kept, rec)
				continue
			}
			h, e := hashFile(rec.Entry.Path)
			if e != nil {
				return e
			}
			if h != "" && h != rec.SHA256 {
				return fmt.Errorf("preserving modified shortcut: %s", rec.Entry.Path)
			}
		}
		for _, rec := range s.Records {
			if id != "" && rec.Entry.ID != id {
				continue
			}
			if e := os.Remove(rec.Entry.Path); e != nil && !errors.Is(e, os.ErrNotExist) {
				return e
			}
		}
		s.Records = kept
		return m.save(ctx, *s)
	})
}
func (m Manager) Repair(ctx context.Context) error {
	return m.mutate(ctx, func(*Registry) error { return nil })
}
func RegistryName(root string) string { return filepath.Join(root, "integrations.json") }
