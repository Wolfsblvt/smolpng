// SPDX-License-Identifier: MIT
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/updates"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// These are release inputs, not a deployed feed or a generated trust authority.
// Unconfigured builds never contact a made-up update URL or accept unsigned code.
var UpdateManifestURL = ""
var TrustedUpdateKeys = ""

func updateTrust(file string) (updates.Trust, error) {
	data := []byte(TrustedUpdateKeys)
	var e error
	if file != "" {
		data, e = updates.ReadFile(file, updates.MaxManifest)
		if e != nil {
			return updates.Trust{}, e
		}
	}
	if len(data) == 0 {
		return updates.Trust{}, fmt.Errorf("no trusted update key configured; an explicitly selected --trust-key public-key file is required")
	}
	return updates.ParseTrust(data)
}
func (a *app) updateManager(keyFile string) (updates.Manager, error) {
	exe, e := a.applicationExecutable()
	if e != nil {
		return updates.Manager{}, e
	}
	p, e := a.dirs()
	if e != nil {
		return updates.Manager{}, e
	}
	trust, e := updateTrust(keyFile)
	if e != nil {
		return updates.Manager{}, e
	}
	return updates.Manager{Executable: exe, Version: Version, Distribution: Distribution, Platform: runtime.GOOS + "/" + runtime.GOARCH, Trust: trust, HTTP: a.env.UpdateHTTP, Runner: a.env.Runner, SkillState: filepath.Join(p.State, "skill-installations.json")}, nil
}
func (a *app) updateEnvelope(ctx context.Context, location string) ([]byte, error) {
	if location == "" {
		location = UpdateManifestURL
	}
	if location == "" {
		return nil, fmt.Errorf("no public update feed has been configured; select a signed manifest explicitly")
	}
	if strings.HasPrefix(location, "https://") {
		return updates.Fetch(ctx, location, a.env.UpdateHTTP, updates.MaxManifest)
	}
	p, e := a.absoluteInvocationPath(location)
	if e != nil {
		return nil, e
	}
	return updates.ReadFile(p, updates.MaxManifest)
}
func (a *app) updateStatus() error {
	exe, e := a.applicationExecutable()
	if e != nil {
		return e
	}
	p, e := a.dirs()
	if e != nil {
		return e
	}
	info := map[string]any{"schemaVersion": 1, "version": Version, "commit": Commit, "channel": updates.Installation(exe, Distribution), "feedConfigured": UpdateManifestURL != "", "trustedKeysConfigured": TrustedUpdateKeys != "", "automaticChecks": "off unless explicitly enabled; no network check performed"}
	b, e := updates.ReadFile(filepath.Join(p.State, "update-last.txt"), 4096)
	if e == nil {
		stage := strings.TrimSpace(string(b))
		j, er := updates.ReadJournal(stage)
		if er == nil {
			info["lastUpdate"] = map[string]any{"stage": stage, "phase": j.Phase, "error": j.Error, "skillResult": j.SkillResult}
		} else {
			info["lastUpdateError"] = "last journal unavailable; original state pointer preserved"
		}
	} else if !os.IsNotExist(e) {
		return domain.E(3, "cannot inspect update state", e)
	}
	return a.emit(info)
}
func (a *app) updateCommand() *cobra.Command {
	root := &cobra.Command{Use: "update", Short: "Signed standalone updates; package-manager ownership is respected", Args: cobra.NoArgs, RunE: func(*cobra.Command, []string) error { return a.updateStatus() }}
	root.AddCommand(&cobra.Command{Use: "status", Args: cobra.NoArgs, RunE: func(*cobra.Command, []string) error { return a.updateStatus() }})
	for _, action := range []string{"check", "apply"} {
		var manifest, key, archive string
		cmd := &cobra.Command{Use: action, Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
			keyPath := key
			var e error
			if key != "" {
				keyPath, e = a.absoluteInvocationPath(key)
				if e != nil {
					return e
				}
			}
			m, e := a.updateManager(keyPath)
			if e != nil {
				return domain.E(2, e.Error(), e)
			}
			channel := updates.Installation(m.Executable, m.Distribution)
			if action == "apply" && !channel.CanApply {
				return domain.E(2, channel.Instruction, nil)
			}
			data, e := a.updateEnvelope(c.Context(), manifest)
			if e != nil {
				return domain.E(2, e.Error(), e)
			}
			v, e := updates.Verify(data, m.Trust, time.Now().UTC())
			if e != nil {
				return domain.E(5, e.Error(), e)
			}
			asset, available := v.Manifest.Assets[m.Platform]
			if action == "check" {
				return a.emit(map[string]any{"schemaVersion": 1, "release": v.Manifest.Version, "sourceCommit": v.Manifest.Commit, "newer": updates.Newer(v.Manifest.Version, Version), "platformAvailable": available, "channel": channel, "asset": asset, "verifiedKeyId": v.KeyID})
			}
			local := archive
			if local != "" {
				local, e = a.absoluteInvocationPath(local)
				if e != nil {
					return e
				}
			}
			plan, e := m.Prepare(c.Context(), data, local)
			if e != nil {
				return domain.E(5, e.Error(), e)
			}
			p, e := a.dirs()
			if e != nil {
				return e
			}
			pid, e := updates.Launch(c.Context(), m.Executable, plan.Stage, keyPath, filepath.Join(p.State, "update-last.txt"))
			if e != nil {
				return domain.E(5, e.Error(), e)
			}
			return a.emit(map[string]any{"schemaVersion": 1, "status": "scheduled", "helperPid": pid, "plan": plan, "completed": false, "instruction": "The trusted old helper waits for this process to exit, applies verified files, probes the new build and synchronizes only unmodified managed Skills. Inspect update status or the journal for completion."})
		}}
		cmd.Flags().StringVar(&manifest, "manifest", "", "Signed manifest file or HTTPS URL; defaults to the build's qualified feed")
		cmd.Flags().StringVar(&key, "trust-key", "", "Explicit JSON Ed25519 public-key trust file (never a secret key)")
		if action == "apply" {
			cmd.Flags().StringVar(&archive, "archive", "", "Use an already downloaded archive; signature/size/hash checks remain mandatory")
		}
		root.AddCommand(cmd)
	}
	var stage, key string
	var apply bool
	recover := &cobra.Command{Use: "recover", Short: "Preview or roll back one recorded update, preserving unrecognized edits", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		if stage == "" {
			return domain.E(2, "recover requires --stage from an update plan/status", nil)
		}
		path, e := a.absoluteInvocationPath(stage)
		if e != nil {
			return e
		}
		if apply {
			keyFile := key
			if key != "" {
				keyFile, e = a.absoluteInvocationPath(key)
				if e != nil {
					return e
				}
			}
			m, e := a.updateManager(keyFile)
			if e != nil {
				return e
			}
			if e = m.Rollback(c.Context(), path); e != nil {
				return domain.E(5, e.Error(), e)
			}
		}
		j, e := updates.ReadJournal(path)
		if e != nil {
			return e
		}
		return a.emit(map[string]any{"schemaVersion": 1, "applied": apply, "stage": path, "phase": j.Phase, "fromVersion": j.FromVersion, "error": j.Error})
	}}
	recover.Flags().StringVar(&stage, "stage", "", "Exact transaction folder")
	recover.Flags().StringVar(&key, "trust-key", "", "Same public-key trust selected for the update")
	recover.Flags().BoolVar(&apply, "apply", false, "Roll back only when installed and backup hashes agree")
	root.AddCommand(recover)
	return root
}
func (a *app) updateInternalCommands() []*cobra.Command {
	var stage, key string
	var pid int
	worker := &cobra.Command{Use: "__update-worker", Hidden: true, Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		if stage == "" {
			return fmt.Errorf("missing update stage")
		}
		trust, e := updateTrust(key)
		if e != nil {
			return e
		}
		if e = updates.Wait(c.Context(), pid); e != nil {
			return e
		}
		m := updates.Manager{Trust: trust, Platform: runtime.GOOS + "/" + runtime.GOARCH, Runner: a.env.Runner}
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Minute)
		defer cancel()
		return m.Apply(ctx, stage)
	}}
	worker.Flags().StringVar(&stage, "stage", "", "Internal stage")
	worker.Flags().StringVar(&key, "trust-key", "", "Selected public key")
	worker.Flags().IntVar(&pid, "wait-pid", 0, "Invoking process")
	var registry string
	sync := &cobra.Command{Use: "__update-skill-sync", Hidden: true, Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		if !filepath.IsAbs(registry) {
			return fmt.Errorf("Skill registry must be absolute")
		}
		m, e := a.skillManager()
		if e != nil {
			return e
		}
		m.StatePath = registry
		results, e := m.SyncAfterUpdate(c.Context())
		if e != nil {
			return e
		}
		b, e := json.Marshal(map[string]any{"schemaVersion": 1, "results": results})
		if e != nil {
			return e
		}
		_, e = a.env.Out.Write(append(b, '\n'))
		return e
	}}
	sync.Flags().StringVar(&registry, "registry", "", "App-owned Skill registry")
	return []*cobra.Command{worker, sync}
}
