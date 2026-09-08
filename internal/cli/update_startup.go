// SPDX-License-Identifier: MIT
package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"github.com/Wolfsblvt/smolpng/internal/state"
	"github.com/Wolfsblvt/smolpng/internal/ui"
	"github.com/Wolfsblvt/smolpng/internal/updates"
	"os"
	"path/filepath"
	"time"
)

type updateCheckState struct {
	SchemaVersion int       `json:"schemaVersion"`
	CheckedAt     time.Time `json:"checkedAt"`
	Ignored       string    `json:"ignored,omitempty"`
}

func updateCheckDue(policy string, offline, configured bool, last, now time.Time) bool {
	return policy == "daily" && !offline && configured && (last.IsZero() || now.Sub(last) >= 24*time.Hour)
}

// Only interactive root calls this function. Neither compress, drop, JSON nor
// update status performs a startup check. A project cannot opt a user into egress.
func (a *app) updateStartup(ctx context.Context, p ui.Prompter, offline bool) (bool, error) {
	dirs, e := a.dirs()
	if e != nil {
		return false, e
	}
	w, e := a.loadConfig(dirs)
	if e != nil {
		return false, e
	}
	policy := "manual"
	if w.User.Document.Updates.Check != nil {
		policy = *w.User.Document.Updates.Check
	}
	if !updateCheckDue(policy, offline, UpdateManifestURL != "" && TrustedUpdateKeys != "", time.Time{}, time.Now()) {
		return false, nil
	}
	file := filepath.Join(dirs.State, "update-check.json")
	var s updateCheckState
	// Claim the interval under an OS lock before fetching. A failed check does not
	// trigger another request on every attempted compression that day.
	if e = os.MkdirAll(dirs.State, 0700); e != nil {
		return false, e
	}
	lock, e := fsutil.TryLock(file + ".lock")
	if e != nil {
		return false, nil
	}
	e = state.Read(file, &s)
	if e != nil && !os.IsNotExist(e) {
		lock.Close()
		return false, a.notice(ctx, p, "Update check state needs attention", e.Error())
	}
	if s.SchemaVersion != 0 && s.SchemaVersion != 1 {
		lock.Close()
		return false, fmt.Errorf("unsupported update-check state")
	}
	now := time.Now().UTC()
	if !updateCheckDue(policy, offline, true, s.CheckedAt, now) {
		lock.Close()
		return false, nil
	}
	s.SchemaVersion = 1
	s.CheckedAt = now
	e = state.Write(ctx, file, s)
	lock.Close()
	if e != nil {
		return false, e
	}
	m, e := a.updateManager("")
	if e != nil {
		return false, a.notice(ctx, p, "Update check unavailable", e.Error())
	}
	checkCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	data, e := a.updateEnvelope(checkCtx, "")
	cancel()
	if e != nil {
		return false, a.notice(ctx, p, "Update check could not complete", e.Error())
	}
	v, e := updates.Verify(data, m.Trust, now)
	if e != nil {
		return false, a.notice(ctx, p, "Update was not trusted", e.Error())
	}
	if !updates.Newer(v.Manifest.Version, Version) || s.Ignored == v.Manifest.Version {
		return false, nil
	}
	if _, ok := v.Manifest.Assets[m.Platform]; !ok {
		return false, nil
	}
	channel := updates.Installation(m.Executable, m.Distribution)
	choices := []ui.Option{{Label: "Later", Value: "later"}, {Label: "Ignore this version", Value: "ignore"}, {Label: "Disable automatic checks", Value: "disable"}}
	if channel.CanApply {
		choices = append([]ui.Option{{Label: "Update now", Value: "apply"}}, choices...)
	} else {
		choices = append([]ui.Option{{Label: "Show owning package-manager instructions", Value: "owner"}}, choices...)
	}
	action, e := p.Choose(ctx, "smolpng "+v.Manifest.Version+" is available", channel.Instruction+"\nVerified by key "+v.KeyID+". Config, dependencies and custom Skills are not replaced.", choices, "later")
	if errors.Is(e, ui.ErrBack) {
		return false, nil
	}
	if e != nil {
		return false, e
	}
	switch action {
	case "owner":
		return false, a.notice(ctx, p, "Update through the installation owner", channel.Instruction)
	case "ignore":
		s.Ignored = v.Manifest.Version
		return false, state.Write(ctx, file, s)
	case "disable":
		return false, w.SaveUserSettings(ctx, map[string]any{"updates.check": "manual"})
	case "apply":
		plan, e := m.Prepare(ctx, data, "")
		if e != nil {
			return false, a.notice(ctx, p, "Update was not applied", e.Error())
		}
		_, e = updates.Launch(ctx, m.Executable, plan.Stage, "", filepath.Join(dirs.State, "update-last.txt"))
		if e != nil {
			return false, a.notice(ctx, p, "Update needs attention", e.Error())
		}
		fmt.Fprintln(a.env.Out, "Update scheduled after this process exits. Result journal:", plan.ResultPath)
		return true, nil
	}
	return false, nil
}
