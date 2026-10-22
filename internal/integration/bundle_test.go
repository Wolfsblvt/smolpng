// SPDX-License-Identifier: MIT
package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
