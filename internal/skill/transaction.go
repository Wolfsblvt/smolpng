// SPDX-License-Identifier: MIT
package skill

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/fsutil"
)

func validateTransaction(t *transaction) error {
	if !filepath.IsAbs(t.Destination) || filepath.Base(t.Destination) != "smolpng" {
		return fmt.Errorf("invalid pending Skill destination")
	}
	for _, p := range []string{t.Stage, t.Backup} {
		if p != "" && (!filepath.IsAbs(p) || filepath.Dir(p) != filepath.Dir(t.Destination) || !strings.HasPrefix(filepath.Base(p), ".smolpng-skill-")) {
			return fmt.Errorf("invalid pending Skill workspace")
		}
	}
	for _, r := range []*Installation{t.Previous, t.Next} {
		if r != nil && r.Destination != t.Destination {
			return fmt.Errorf("pending Skill destination mismatch")
		}
	}
	return nil
}
func (m Manager) swap(ctx context.Context, s *registry, old, next *Installation) error {
	dest := ""
	if next != nil {
		dest = next.Destination
	} else {
		dest = old.Destination
	}
	if e := noSymlinks(dest); e != nil {
		return e
	}
	if e := os.MkdirAll(filepath.Dir(dest), 0755); e != nil {
		return e
	}
	t := &transaction{Destination: dest, Previous: old, Next: next}
	if next != nil {
		stage, e := os.MkdirTemp(filepath.Dir(dest), ".smolpng-skill-stage-")
		if e != nil {
			return e
		}
		t.Stage = stage
		// No journal yet: only this newly created, private staging directory is ours.
		defer func() {
			if s.Pending == nil && t.Stage != "" && exact(t.Stage, next.Files) {
				_ = os.RemoveAll(t.Stage)
			}
		}()
		for _, name := range sortedKeys(m.Bundle.Files) {
			if e = fsutil.AtomicWrite(ctx, filepath.Join(stage, filepath.FromSlash(name)), m.Bundle.Files[name], 0644, false); e != nil {
				_ = os.RemoveAll(stage)
				return e
			}
		}
		if !exact(stage, next.Files) {
			return fmt.Errorf("staged Skill does not match embedded manifest")
		}
	}
	if old != nil {
		p, e := os.MkdirTemp(filepath.Dir(dest), ".smolpng-skill-backup-")
		if e != nil {
			return e
		}
		if e = os.Remove(p); e != nil {
			return e
		}
		t.Backup = p
	}
	s.Pending = t
	if e := m.save(ctx, s); e != nil {
		s.Pending = nil
		return e
	}
	// From this point a recoverable journal owns both directory moves.
	fail := func(cause error) error {
		// The next explicit operation can recover even when the caller cancelled.
		if e := m.recover(context.WithoutCancel(ctx), s); e != nil {
			return fmt.Errorf("%w; Skill recovery also refused: %v", cause, e)
		}
		return cause
	}
	if old != nil {
		if !exact(dest, old.Files) {
			return fail(fmt.Errorf("installed Skill changed after preview; preserving it"))
		}
		if e := fsutil.MoveNoReplace(dest, t.Backup); e != nil {
			return fail(e)
		}
	} else if !missing(dest) {
		return fail(fmt.Errorf("destination appeared during install; preserving it"))
	}
	if next != nil {
		if e := fsutil.MoveNoReplace(t.Stage, dest); e != nil {
			return fail(e)
		}
	}
	replace(s, dest, next)
	s.Pending = nil
	if old != nil {
		s.Cleanup = append(s.Cleanup, residue{t.Backup, old.Files})
	}
	s.LastVersion = m.AppVersion
	s.LastBundle = m.Bundle.Hash
	if e := m.save(ctx, s); e != nil {
		// The on-disk journal still describes the exact old/new bytes. Retain backups.
		s.Pending = t
		return fmt.Errorf("Skill moved but registry commit failed; next write will recover: %w", e)
	}
	m.clean(s)
	return m.save(ctx, s)
}

// Recovery completes an already-installed new bundle, or restores the exact old
// bundle. Unexpected bytes stop recovery; no unknown directory is ever removed.
func (m Manager) recover(ctx context.Context, s *registry) error {
	t := s.Pending
	if t == nil {
		m.clean(s)
		return nil
	}
	if e := validateTransaction(t); e != nil {
		return e
	}
	if e := noSymlinks(t.Destination); e != nil {
		return e
	}
	switch {
	case t.Next != nil && exact(t.Destination, t.Next.Files):
		replace(s, t.Destination, t.Next)
		if t.Previous != nil {
			s.Cleanup = append(s.Cleanup, residue{t.Backup, t.Previous.Files})
		}
	case t.Previous != nil && exact(t.Destination, t.Previous.Files) && missing(t.Backup):
		// The old directory was never moved; abandon the staged replacement.
	case t.Previous != nil && missing(t.Destination) && exact(t.Backup, t.Previous.Files):
		if e := fsutil.MoveNoReplace(t.Backup, t.Destination); e != nil {
			return e
		}
	case t.Previous == nil && missing(t.Destination):
		// An install was interrupted before committing the directory.
	default:
		return fmt.Errorf("Skill transaction has unexpected user bytes; inspect registry %s before repair", m.StatePath)
	}
	if t.Next != nil && t.Stage != "" {
		s.Cleanup = append(s.Cleanup, residue{t.Stage, t.Next.Files})
	}
	s.Pending = nil
	if e := m.save(ctx, s); e != nil {
		return e
	}
	m.clean(s)
	return m.save(ctx, s)
}
func (m Manager) clean(s *registry) {
	pending := []residue{}
	for _, c := range s.Cleanup {
		if missing(c.Path) {
			continue
		}
		if exact(c.Path, c.Files) {
			if e := os.RemoveAll(c.Path); e == nil {
				continue
			}
		}
		pending = append(pending, c)
	}
	s.Cleanup = pending
}
