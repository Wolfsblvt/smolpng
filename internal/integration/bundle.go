// SPDX-License-Identifier: MIT
package integration

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/fsutil"
)

// BundleRenderer writes a native app into a new, caller-owned staging path.
// File launchers retain Renderer; only directory-based native apps use this seam.
type BundleRenderer interface {
	RenderBundle(context.Context, Entry, string) error
}

func recordFor(s Registry, x Entry) Record {
	for _, r := range s.Records {
		if r.Entry.Path == x.Path {
			return r
		}
	}
	return Record{Entry: x}
}

func validateMembers(r Record) error {
	if filepath.Ext(r.Entry.Path) != ".app" {
		if len(r.Members) != 0 {
			return fmt.Errorf("file launcher cannot contain bundle members")
		}
		return nil
	}
	if len(r.Members) == 0 {
		return fmt.Errorf("app bundle has no recorded members")
	}
	seen := map[string]bool{}
	for _, name := range r.Members {
		if !fs.ValidPath(name) || name == "." || strings.ContainsAny(name, "\\\x00") || seen[name] {
			return fmt.Errorf("invalid app bundle member")
		}
		seen[name] = true
	}
	return nil
}

func bundleMembers(root string) ([]string, error) {
	names := []string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			if !entry.IsDir() {
				return fmt.Errorf("app bundle is not a directory")
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("app bundle contains a special file; preserving it")
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		names = append(names, filepath.ToSlash(name))
		return nil
	})
	return names, err
}

func hashRecord(r Record) (string, error) {
	if filepath.Ext(r.Entry.Path) != ".app" {
		return hashFile(r.Entry.Path)
	}
	if err := noLinks(r.Entry.Path); err != nil {
		return "", err
	}
	if _, err := os.Lstat(r.Entry.Path); os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	names, err := bundleMembers(r.Entry.Path)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	for _, name := range names {
		path := filepath.Join(r.Entry.Path, filepath.FromSlash(name))
		info, err := os.Lstat(path)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(hash, "%s\x00%s\x00%o\x00", name, info.Mode().Type(), info.Mode().Perm())
		if !info.IsDir() {
			if !slices.Contains(r.Members, name) {
				// Extra user files make the bundle modified without reading their contents.
				fmt.Fprint(hash, "untracked\x00")
				continue
			}
			file, err := os.Open(path)
			if err != nil {
				return "", err
			}
			contentHash := sha256.New()
			_, readErr := io.Copy(contentHash, file)
			closeErr := file.Close()
			if readErr != nil {
				return "", readErr
			}
			if closeErr != nil {
				return "", closeErr
			}
			fmt.Fprintf(hash, "%x\x00", contentHash.Sum(nil))
		}
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func validateBundlePending(p *Pending) error {
	if p.Stage == "" {
		if p.Backup != "" || p.Previous != nil || p.Removing {
			return fmt.Errorf("invalid pending launcher workspace")
		}
		return nil
	}
	if filepath.Ext(p.Next.Entry.Path) != ".app" || len(p.Next.SHA256) != 64 || validateMembers(p.Next) != nil {
		return fmt.Errorf("invalid pending app bundle")
	}
	parent := filepath.Dir(p.Stage)
	if filepath.Dir(parent) != filepath.Dir(p.Next.Entry.Path) || !strings.HasPrefix(filepath.Base(parent), ".smolpng-launcher-") ||
		filepath.Base(p.Stage) != filepath.Base(p.Next.Entry.Path) || p.Backup != filepath.Join(parent, "previous.app") {
		return fmt.Errorf("invalid pending app workspace")
	}
	if p.Previous != nil && (p.Previous.Entry.Path != p.Next.Entry.Path || p.Previous.SHA256 != p.Before || validateMembers(*p.Previous) != nil) {
		return fmt.Errorf("invalid previous app bundle")
	}
	if p.Removing && (p.Previous == nil || p.Previous.SHA256 != p.Next.SHA256) {
		return fmt.Errorf("invalid pending app removal")
	}
	for _, path := range []string{p.Stage, p.Backup} {
		if err := noLinks(path); err != nil {
			return err
		}
	}
	return nil
}

func atPath(r Record, path string) Record {
	r.Entry.Path = path
	return r
}

func (m Manager) moveApp(from, to string) error {
	if m.moveBundle != nil {
		return m.moveBundle(from, to)
	}
	return fsutil.MoveNoReplace(from, to)
}

// quarantineBundle transfers the named object before proving ownership. A prior
// live-path hash is only a preflight; it cannot authorize deleting a later object.
func (m Manager) quarantineBundle(ctx context.Context, s *Registry) error {
	p := s.Pending
	if err := m.moveApp(p.Next.Entry.Path, p.Backup); err != nil {
		return err
	}
	hash, err := hashRecord(atPath(*p.Previous, p.Backup))
	if err == nil && hash == p.Previous.SHA256 {
		return nil
	}
	return m.restoreChangedBundle(ctx, s)
}

func (m Manager) restoreChangedBundle(ctx context.Context, s *Registry) error {
	p := s.Pending
	// No-replace also preserves an object that appeared at the visible path after
	// quarantine. In that case both it and the workspace remain for recovery.
	if err := m.moveApp(p.Backup, p.Next.Entry.Path); err != nil {
		return fmt.Errorf("changed app bundle; preserving destination and workspace: %w", err)
	}
	if err := m.finishBundle(ctx, s); err != nil {
		return err
	}
	return fmt.Errorf("app bundle changed; restored and preserved it")
}

func (m Manager) removeBundle(ctx context.Context, s *Registry, old Record) error {
	if _, err := os.Lstat(old.Entry.Path); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	workspace, err := os.MkdirTemp(filepath.Dir(old.Entry.Path), ".smolpng-launcher-")
	if err != nil {
		return err
	}
	p := &Pending{Next: old, Previous: &old, Before: old.SHA256, Removing: true,
		Stage: filepath.Join(workspace, filepath.Base(old.Entry.Path)), Backup: filepath.Join(workspace, "previous.app")}
	s.Pending = p
	if err = m.save(ctx, *s); err != nil {
		s.Pending = nil
		_ = os.Remove(workspace)
		return err
	}
	if err = m.quarantineBundle(ctx, s); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return m.recoverBundle(ctx, s)
}

func (m Manager) installBundle(ctx context.Context, s *Registry, x Entry) error {
	renderer, ok := m.Renderer.(BundleRenderer)
	if !ok {
		return fmt.Errorf("native app renderer unavailable")
	}
	if err := os.MkdirAll(filepath.Dir(x.Path), 0755); err != nil {
		return err
	}
	workspace, err := os.MkdirTemp(filepath.Dir(x.Path), ".smolpng-launcher-")
	if err != nil {
		return err
	}
	journaled := false
	defer func() {
		if !journaled {
			_ = os.RemoveAll(workspace)
		}
	}()
	stage := filepath.Join(workspace, filepath.Base(x.Path))
	if err = renderer.RenderBundle(ctx, x, stage); err != nil {
		return err
	}
	members, err := bundleMembers(stage)
	if err != nil {
		return err
	}
	next := Record{Entry: x, Members: members, Version: m.Version}
	if err = validateMembers(next); err != nil {
		return err
	}
	if next.SHA256, err = hashRecord(atPath(next, stage)); err != nil {
		return err
	}
	old := recordFor(*s, x)
	before, err := hashRecord(old)
	if err != nil {
		return err
	}
	if before != "" && before != old.SHA256 {
		return fmt.Errorf("preserving changed app bundle: %s", x.Path)
	}
	pending := &Pending{Next: next, Before: before, Stage: stage, Backup: filepath.Join(workspace, "previous.app")}
	if before != "" {
		pending.Previous = &old
	}
	s.Pending = pending
	if err = m.save(ctx, *s); err != nil {
		s.Pending = nil
		return err
	}
	journaled = true
	if before != "" {
		if err = m.quarantineBundle(ctx, s); err != nil {
			return err
		}
	}
	if err = m.moveApp(stage, x.Path); err != nil {
		return err
	}
	return m.recoverBundle(ctx, s)
}

func (m Manager) recoverBundle(ctx context.Context, s *Registry) error {
	p := s.Pending
	if err := validateBundlePending(p); err != nil {
		return err
	}
	current, err := hashRecord(p.Next)
	if err != nil {
		return err
	}
	if p.Removing {
		backup, err := hashRecord(atPath(*p.Previous, p.Backup))
		if err != nil {
			return err
		}
		if current != "" {
			if current != p.Before || backup != "" {
				return fmt.Errorf("app removal found an occupied destination; preserving both objects")
			}
		} else if backup != "" && backup != p.Before {
			return m.restoreChangedBundle(ctx, s)
		} else {
			s.Records = slices.DeleteFunc(s.Records, func(r Record) bool { return r.Entry.Path == p.Next.Entry.Path })
		}
		return m.finishBundle(ctx, s)
	}
	if current == p.Next.SHA256 {
		replace(s, p.Next)
	} else if p.Previous != nil {
		old, err := hashRecord(*p.Previous)
		if err != nil {
			return err
		}
		backup, err := hashRecord(atPath(*p.Previous, p.Backup))
		if err != nil {
			return err
		}
		if current == "" && backup != "" && backup != p.Previous.SHA256 {
			return m.restoreChangedBundle(ctx, s)
		}
		if current == "" && backup == p.Previous.SHA256 {
			if err = m.moveApp(p.Backup, p.Next.Entry.Path); err != nil {
				return err
			}
		} else if old != p.Previous.SHA256 || backup != "" {
			return fmt.Errorf("app recovery found changed contents; preserving the workspace")
		}
	} else if current != "" {
		return fmt.Errorf("app destination changed during installation; preserving it")
	}
	return m.finishBundle(ctx, s)
}

func (m Manager) finishBundle(ctx context.Context, s *Registry) error {
	p := s.Pending
	for _, r := range []*Record{&p.Next, p.Previous} {
		if r == nil {
			continue
		}
		path := p.Stage
		if r == p.Previous {
			path = p.Backup
		}
		hash, err := hashRecord(atPath(*r, path))
		if err != nil {
			return err
		}
		if hash == "" {
			continue
		}
		if hash != r.SHA256 {
			return fmt.Errorf("app workspace changed; preserving it")
		}
		if err = os.RemoveAll(path); err != nil {
			return err
		}
	}
	if err := os.Remove(filepath.Dir(p.Stage)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	s.Pending = nil
	return m.save(ctx, *s)
}
