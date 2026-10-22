// SPDX-License-Identifier: MIT
package integration

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/fsutil"
)

type appRenderer struct {
	textRenderer
	content     string
	afterRender func()
}

func (r appRenderer) RenderBundle(_ context.Context, _ Entry, destination string) error {
	if err := os.MkdirAll(filepath.Join(destination, "Contents", "MacOS"), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(destination, "Contents", "MacOS", "applet"), []byte(r.content), 0755); err != nil {
		return err
	}
	if r.afterRender != nil {
		r.afterRender()
	}
	return nil
}

func appFixture(t *testing.T) (Manager, []Entry) {
	t.Helper()
	m, files := fixture(t)
	entries, err := Plan(m.Allowed, Request{OS: "darwin", Executable: files[0].Executable, WorkingDirectory: files[0].WorkingDirectory})
	if err != nil {
		t.Fatal(err)
	}
	m.Renderer = appRenderer{content: "native fixture"}
	return m, entries
}

func TestMacPlanKeepsNativeModesAndArgumentBoundaries(t *testing.T) {
	_, entries := appFixture(t)
	if len(entries) != 2 || !entries[0].Terminal || entries[1].Terminal || entries[1].Arguments[0] != "drop" {
		t.Fatal(entries)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Path) != ".app" || entry.Arguments[len(entry.Arguments)-1] != "--" {
			t.Fatal(entry)
		}
	}
}

func TestMacConfigurationQueryKeepsLocationFlagsAndExcludesInputs(t *testing.T) {
	want := []string{"--home", "/data ' ä", "--config", "/settings -- ä.toml", "--no-project"}
	for _, prefix := range [][]string{nil, {"drop"}} {
		args := append(slices.Clone(prefix), want...)
		args = append(args, "--preset", "auto", "--", "--home", "this-is-an-input.png")
		if got := macConfigurationArguments(args); !slices.Equal(got, want) {
			t.Fatalf("configuration query: got %q, want %q", got, want)
		}
	}
}

func TestAppInstallRepeatAndRemovalUseCompleteOwnedContents(t *testing.T) {
	m, entries := appFixture(t)
	ctx := context.Background()
	for range 2 {
		if err := m.Install(ctx, entries); err != nil {
			t.Fatal(err)
		}
		status, pending, err := m.Status()
		if err != nil || pending || len(status) != 2 || status[0].State != "current" || len(status[0].Members) == 0 {
			t.Fatal(status, pending, err)
		}
	}
	if err := m.Remove(ctx, ""); err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if _, err := os.Stat(entry.Path); !os.IsNotExist(err) {
			t.Fatal("app remains after removal", err)
		}
	}
}

func TestAppPreservesModifiedAndAdditionalUserFiles(t *testing.T) {
	for _, name := range []string{"Contents/MacOS/applet", "private-user-file"} {
		t.Run(name, func(t *testing.T) {
			m, entries := appFixture(t)
			ctx := context.Background()
			if err := m.Install(ctx, entries); err != nil {
				t.Fatal(err)
			}
			changed := filepath.Join(entries[0].Path, filepath.FromSlash(name))
			if err := os.WriteFile(changed, []byte("user-owned edit"), 0600); err != nil {
				t.Fatal(err)
			}
			status, _, err := m.Status()
			if err != nil || status[0].State != "modified" {
				t.Fatal(status, err)
			}
			if m.Remove(ctx, "") == nil || m.Install(ctx, entries) == nil {
				t.Fatal("modified bundle accepted for replacement/removal")
			}
			data, err := os.ReadFile(changed)
			if err != nil || string(data) != "user-owned edit" {
				t.Fatal("user bytes lost", err)
			}
			if _, err = os.Stat(entries[1].Path); err != nil {
				t.Fatal("another target was removed before whole-plan refusal", err)
			}
		})
	}
}

func TestAppPreservesUnownedDirectory(t *testing.T) {
	m, entries := appFixture(t)
	if err := os.MkdirAll(entries[0].Path, 0755); err != nil {
		t.Fatal(err)
	}
	if err := m.Install(context.Background(), entries); err == nil {
		t.Fatal("unowned app directory adopted")
	}
}

func TestAppChangedDuringRenderingIsPreserved(t *testing.T) {
	m, entries := appFixture(t)
	ctx := context.Background()
	if err := m.Install(ctx, entries[:1]); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(entries[0].Path, "Contents", "MacOS", "applet")
	m.Renderer = appRenderer{content: "replacement", afterRender: func() {
		if err := os.WriteFile(file, []byte("concurrent edit"), 0755); err != nil {
			t.Fatal(err)
		}
	}}
	if err := m.Install(ctx, entries[:1]); err == nil {
		t.Fatal("render-time edit overwritten")
	}
	data, _ := os.ReadFile(file)
	if string(data) != "concurrent edit" {
		t.Fatal("concurrent edit lost")
	}
}

func TestAppRepairRestoresInterruptedReplacementOrCommitsPublishedBundle(t *testing.T) {
	for _, published := range []bool{false, true} {
		t.Run(map[bool]string{false: "restore-prior", true: "accept-published"}[published], func(t *testing.T) {
			m, entries := appFixture(t)
			ctx := context.Background()
			if err := m.Install(ctx, entries[:1]); err != nil {
				t.Fatal(err)
			}
			registry, err := m.load()
			if err != nil {
				t.Fatal(err)
			}
			old := registry.Records[0]
			workspace, err := os.MkdirTemp(m.Allowed.Menu, ".smolpng-launcher-")
			if err != nil {
				t.Fatal(err)
			}
			stage := filepath.Join(workspace, filepath.Base(entries[0].Path))
			if err = (appRenderer{content: "new version"}).RenderBundle(ctx, entries[0], stage); err != nil {
				t.Fatal(err)
			}
			next := Record{Entry: entries[0], Version: "next"}
			next.Members, err = bundleMembers(stage)
			if err != nil {
				t.Fatal(err)
			}
			next.SHA256, err = hashRecord(atPath(next, stage))
			if err != nil {
				t.Fatal(err)
			}
			backup := filepath.Join(workspace, "previous.app")
			registry.Pending = &Pending{Next: next, Previous: &old, Before: old.SHA256, Stage: stage, Backup: backup}
			if err = m.save(ctx, registry); err != nil {
				t.Fatal(err)
			}
			if err = os.Rename(old.Entry.Path, backup); err != nil {
				t.Fatal(err)
			}
			if published {
				if err = os.Rename(stage, next.Entry.Path); err != nil {
					t.Fatal(err)
				}
			}
			if err = m.Repair(ctx); err != nil {
				t.Fatal(err)
			}
			status, pending, err := m.Status()
			if err != nil || pending || len(status) != 1 || status[0].State != "current" {
				t.Fatal(status, pending, err)
			}
			want := old.SHA256
			if published {
				want = next.SHA256
			}
			if status[0].SHA256 != want {
				t.Fatal("wrong bundle after recovery")
			}
			if _, err = os.Stat(workspace); !os.IsNotExist(err) {
				t.Fatal("owned transaction residue remains", err)
			}
		})
	}
}

func TestBundleJournalCannotNameOutsideWorkspace(t *testing.T) {
	m, entries := appFixture(t)
	if err := m.Install(context.Background(), entries[:1]); err != nil {
		t.Fatal(err)
	}
	registry, err := m.load()
	if err != nil {
		t.Fatal(err)
	}
	registry.Pending = &Pending{Next: registry.Records[0], Stage: filepath.Join(t.TempDir(), filepath.Base(entries[0].Path))}
	if err = m.save(context.Background(), registry); err != nil {
		t.Fatal(err)
	}
	if err = m.Repair(context.Background()); err == nil || !strings.Contains(err.Error(), "workspace") {
		t.Fatal("outside workspace accepted", err)
	}
}

func TestAppLateChangesSurviveRemovalAndReplacement(t *testing.T) {
	for _, removing := range []bool{false, true} {
		for _, replacement := range []bool{false, true} {
			name := map[bool]string{false: "install", true: "remove"}[removing] + "/" + map[bool]string{false: "edit", true: "replacement"}[replacement]
			t.Run(name, func(t *testing.T) {
				m, entries := appFixture(t)
				ctx := context.Background()
				if err := m.Install(ctx, entries[:1]); err != nil {
					t.Fatal(err)
				}
				destination := entries[0].Path
				changed := false
				m.moveBundle = func(from, to string) error {
					if from == destination && !changed {
						changed = true
						if replacement {
							if err := os.Rename(destination, filepath.Join(t.TempDir(), "prior.app")); err != nil {
								t.Fatal(err)
							}
							if err := os.Mkdir(destination, 0755); err != nil {
								t.Fatal(err)
							}
						}
						if err := os.WriteFile(filepath.Join(destination, "late-user-file"), []byte("late user bytes"), 0600); err != nil {
							t.Fatal(err)
						}
					}
					return fsutil.MoveNoReplace(from, to)
				}
				m.Renderer = appRenderer{content: "new version"}
				var err error
				if removing {
					err = m.Remove(ctx, entries[0].ID)
				} else {
					err = m.Install(ctx, entries[:1])
				}
				if err == nil || !changed {
					t.Fatal("late change did not refuse the operation", err)
				}
				data, err := os.ReadFile(filepath.Join(destination, "late-user-file"))
				if err != nil || string(data) != "late user bytes" {
					t.Fatal("late bytes were not restored", err)
				}
				status, pending, err := m.Status()
				if err != nil || pending || len(status) != 1 || status[0].State != "modified" {
					t.Fatal("refused transaction did not settle safely", status, pending, err)
				}
			})
		}
	}
}

func TestAppNewDestinationAfterQuarantinePreservesBothObjects(t *testing.T) {
	for _, removing := range []bool{false, true} {
		t.Run(map[bool]string{false: "install", true: "remove"}[removing], func(t *testing.T) {
			m, entries := appFixture(t)
			ctx := context.Background()
			if err := m.Install(ctx, entries[:1]); err != nil {
				t.Fatal(err)
			}
			destination := entries[0].Path
			m.moveBundle = func(from, to string) error {
				if err := fsutil.MoveNoReplace(from, to); err != nil {
					return err
				}
				if from == destination {
					if err := os.Mkdir(destination, 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(destination, "new-user-file"), []byte("new occupant"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				return nil
			}
			m.Renderer = appRenderer{content: "new version"}
			var err error
			if removing {
				err = m.Remove(ctx, entries[0].ID)
			} else {
				err = m.Install(ctx, entries[:1])
			}
			if err == nil || m.Repair(ctx) == nil {
				t.Fatal("new occupant was accepted for overwrite or cleanup")
			}
			data, err := os.ReadFile(filepath.Join(destination, "new-user-file"))
			if err != nil || string(data) != "new occupant" {
				t.Fatal("new occupant lost", err)
			}
			registry, err := m.load()
			if err != nil || registry.Pending == nil {
				t.Fatal("recovery workspace not retained", err)
			}
			backup, err := hashRecord(atPath(*registry.Pending.Previous, registry.Pending.Backup))
			if err != nil || backup != registry.Pending.Previous.SHA256 {
				t.Fatal("prior bundle not preserved", backup, err)
			}
		})
	}
}

func TestAppRepairWithAlreadyCleanedWorkspace(t *testing.T) {
	for _, destination := range []string{"previous", "published", "missing", "changed"} {
		t.Run(destination, func(t *testing.T) {
			m, entries := appFixture(t)
			ctx := context.Background()
			if err := m.Install(ctx, entries[:1]); err != nil {
				t.Fatal(err)
			}
			registry, err := m.load()
			if err != nil {
				t.Fatal(err)
			}
			old := registry.Records[0]
			workspace, err := os.MkdirTemp(m.Allowed.Menu, ".smolpng-launcher-")
			if err != nil {
				t.Fatal(err)
			}
			stage := filepath.Join(workspace, filepath.Base(old.Entry.Path))
			if err = (appRenderer{content: "next version"}).RenderBundle(ctx, old.Entry, stage); err != nil {
				t.Fatal(err)
			}
			next := old
			next.SHA256, err = hashRecord(atPath(next, stage))
			if err != nil {
				t.Fatal(err)
			}
			registry.Pending = &Pending{Next: next, Previous: &old, Before: old.SHA256, Stage: stage, Backup: filepath.Join(workspace, "previous.app")}
			if err = m.save(ctx, registry); err != nil {
				t.Fatal(err)
			}
			if destination != "previous" {
				if err = os.Rename(old.Entry.Path, registry.Pending.Backup); err != nil {
					t.Fatal(err)
				}
			}
			if destination == "published" || destination == "changed" {
				if err = os.Rename(stage, next.Entry.Path); err != nil {
					t.Fatal(err)
				}
				if destination == "changed" {
					if err = os.WriteFile(filepath.Join(next.Entry.Path, "user-file"), []byte("preserve"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err = os.RemoveAll(workspace); err != nil {
				t.Fatal(err)
			}
			err = m.Repair(ctx)
			if destination == "missing" || destination == "changed" {
				if err == nil {
					t.Fatal("missing workspace excused a contradictory destination")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			status, pending, err := m.Status()
			want := old.SHA256
			if destination == "published" {
				want = next.SHA256
			}
			if err != nil || pending || len(status) != 1 || status[0].SHA256 != want || status[0].State != "current" {
				t.Fatal("wrong settled record", status, pending, err)
			}
			if err = m.Repair(ctx); err != nil {
				t.Fatal("repair not idempotent", err)
			}
			if err = m.Remove(ctx, ""); err != nil {
				t.Fatal("next mutation remained blocked", err)
			}
		})
	}
}

func TestAppInterruptedRemovalReconcilesBeforeCleanup(t *testing.T) {
	for _, phase := range []string{"journaled", "quarantined", "deleted", "workspace-gone"} {
		t.Run(phase, func(t *testing.T) {
			m, entries := appFixture(t)
			ctx := context.Background()
			if err := m.Install(ctx, entries[:1]); err != nil {
				t.Fatal(err)
			}
			registry, err := m.load()
			if err != nil {
				t.Fatal(err)
			}
			old := registry.Records[0]
			workspace, err := os.MkdirTemp(m.Allowed.Menu, ".smolpng-launcher-")
			if err != nil {
				t.Fatal(err)
			}
			backup := filepath.Join(workspace, "previous.app")
			registry.Pending = &Pending{Next: old, Previous: &old, Before: old.SHA256, Removing: true,
				Stage: filepath.Join(workspace, filepath.Base(old.Entry.Path)), Backup: backup}
			if err = m.save(ctx, registry); err != nil {
				t.Fatal(err)
			}
			if phase != "journaled" {
				if err = os.Rename(old.Entry.Path, backup); err != nil {
					t.Fatal(err)
				}
			}
			if phase == "deleted" || phase == "workspace-gone" {
				if err = os.RemoveAll(backup); err != nil {
					t.Fatal(err)
				}
			}
			if phase == "workspace-gone" {
				if err = os.Remove(workspace); err != nil {
					t.Fatal(err)
				}
			}
			for range 2 {
				if err = m.Repair(ctx); err != nil {
					t.Fatal(err)
				}
			}
			status, pending, err := m.Status()
			want := 0
			if phase == "journaled" {
				want = 1
			}
			if err != nil || pending || len(status) != want || (want == 1 && status[0].State != "current") {
				t.Fatal("wrong removal recovery state", status, pending, err)
			}
			if err = m.Install(ctx, entries[:1]); err != nil {
				t.Fatal("next mutation blocked", err)
			}
			if err = m.Remove(ctx, ""); err != nil {
				t.Fatal(err)
			}
		})
	}
}
