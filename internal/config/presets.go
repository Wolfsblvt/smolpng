// SPDX-License-Identifier: MIT
package config

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/fsutil"
)

type Entry struct {
	Name       string  `json:"name"`
	Label      string  `json:"label"`
	Builtin    bool    `json:"builtin"`
	Origin     Origin  `json:"origin"`
	Definition *Preset `json:"definition,omitempty"`
}

func (w *Workspace) List() []Entry {
	entries := []Entry{}
	labels := map[string]string{"auto": "Auto", "web": "Web balanced", "smallest": "Maximum squeeze", "lossless": "Lossless only"}
	for _, n := range []string{"auto", "web", "smallest", "lossless"} {
		entries = append(entries, Entry{Name: n, Label: labels[n], Builtin: true, Origin: Origin{File: "builtin:" + n}})
	}
	names := []string{}
	for n := range w.presets {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		p := w.presets[n]
		label := p.Label
		if label == "" {
			label = n
		}
		entries = append(entries, Entry{Name: n, Label: label, Origin: w.origins[n], Definition: &p})
	}
	return entries
}

// MenuEntries always keeps Auto first. User positions are stable 1-based slots
// after Auto, with name ordering as a tie-breaker. Unpinned recipes remain saved.
func (w *Workspace) MenuEntries() []Entry {
	all := w.List()
	pinned := []Entry{}
	base := slices.Clone(all[:4])
	for _, v := range all[4:] {
		m := v.Definition.Menu
		if m.Pinned != nil && *m.Pinned && (m.Hidden == nil || !*m.Hidden) {
			pinned = append(pinned, v)
		}
	}
	pos := func(e Entry) int {
		if e.Definition.Menu.Position == nil {
			return 1000
		}
		return *e.Definition.Menu.Position
	}
	sort.SliceStable(pinned, func(i, j int) bool {
		if pos(pinned[i]) == pos(pinned[j]) {
			return pinned[i].Name < pinned[j].Name
		}
		return pos(pinned[i]) < pos(pinned[j])
	})
	lastPos, offset := -1, 0
	for _, v := range pinned {
		p := pos(v)
		if p == lastPos {
			offset++
		} else {
			offset = 0
		}
		lastPos = p
		slot := min(max(1, p-1+offset), len(base))
		base = append(base, Entry{})
		copy(base[slot+1:], base[slot:])
		base[slot] = v
	}
	return base
}
func (w *Workspace) Definition(name string) (Preset, bool) { p, ok := w.presets[name]; return p, ok }
func (w *Workspace) Export(name string, resolved bool) ([]byte, error) {
	if !resolved {
		if domain.IsBuiltin(name) {
			return nil, fmt.Errorf("built-ins are immutable; export --resolved --name NEW_NAME to create an editable copy")
		}
		_, chain, e := w.chain(name)
		if e != nil {
			return nil, e
		}
		bundle := Bundle{Presets: map[string]Preset{}}
		for _, n := range chain {
			bundle.Presets[n] = w.presets[n]
		}
		return MarshalBundle(bundle)
	}
	eff, e := w.Resolve(name, Fields{})
	if e != nil {
		return nil, e
	}
	if name == "" {
		name = eff.Recipe.Name
	}
	if domain.IsBuiltin(name) {
		name += "-copy"
	}
	f, e := resolvedFields(eff, w.CWD)
	if e != nil {
		return nil, e
	}
	preset := Preset{Schema: 1, Label: name, Formats: slices.Clone(eff.Formats), Fields: f}
	if original, ok := w.presets[name]; ok {
		preset.Label = original.Label
		preset.Description = original.Description
		preset.Author = original.Author
		preset.Source = original.Source
		preset.Menu = original.Menu
	}
	return MarshalBundle(Bundle{Presets: map[string]Preset{name: preset}})
}
func resolvedFields(e Effective, cwd string) (Fields, error) {
	r := e.Recipe
	size := ""
	if r.TargetBytes > 0 {
		size = fmt.Sprintf("%dB", r.TargetBytes)
	}
	dim := func(d domain.Dimensions) *string {
		if d.Width == 0 {
			return Ptr("")
		}
		return Ptr(d.String())
	}
	dir := ""
	if e.Output.Directory != "" {
		var er error
		dir, er = filepath.Rel(cwd, e.Output.Directory)
		if er != nil || !portableRelative(dir) {
			return Fields{}, fmt.Errorf("resolved output directory is outside this project; use normal export or select a portable relative destination")
		}
		dir = filepath.ToSlash(dir)
	}
	return Fields{Background: Ptr(r.Background), Quality: Ptr(fmt.Sprintf("%d-%d", r.QualityMin, r.QualityMax)), Colors: Ptr(r.Colors), Speed: Ptr(r.Speed), Dither: Ptr(r.Dither), OxiLevel: Ptr(r.OxiLevel), Metadata: Ptr(r.Metadata), Quantize: Ptr(!r.Lossless), OptimizeAlpha: Ptr(r.OptimizeAlpha), Target: Ptr(size), AllowResize: Ptr(r.AllowResize), Resize: dim(r.Resize), MaxDimensions: dim(r.MaxDimensions), MinDimensions: dim(r.MinDimensions), AcceptNearest: Ptr(r.AcceptNearest), ForceOutput: Ptr(r.ForceOutput), Engine: Ptr(r.Engine), Remote: Ptr(r.RemotePolicy), Template: Ptr(e.Output.Template), Directory: Ptr(dir), Clipboard: Ptr(e.Clipboard), Recursive: Ptr(e.Recursive), Jobs: Ptr(e.Jobs), Timeout: Ptr(r.Timeout.String())}, nil
}

type ImportItem struct {
	SourceName string   `json:"sourceName"`
	Name       string   `json:"name"`
	Replaces   bool     `json:"replaces"`
	Before     *Preset  `json:"before,omitempty"`
	After      Preset   `json:"after"`
	Warnings   []string `json:"warnings"`
}
type ImportPlan struct {
	SchemaVersion int          `json:"schemaVersion"`
	Destination   string       `json:"destination"`
	PreviousHash  string       `json:"previousHash"`
	Items         []ImportItem `json:"items"`
	Applied       bool         `json:"applied"`
	next          Bundle
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func (w *Workspace) PlanImport(b Bundle, replace bool) (ImportPlan, error) {
	plan := ImportPlan{SchemaVersion: 1, Destination: w.StorePath, PreviousHash: digest(w.StoreBytes), Items: []ImportItem{}}
	if e := validatePresets(b.Presets); e != nil {
		return plan, e
	}
	if len(b.Presets) == 0 {
		return plan, fmt.Errorf("no presets to import")
	}
	names := []string{}
	for n := range b.Presets {
		names = append(names, n)
	}
	sort.Strings(names)
	reserved := map[string]bool{}
	for n := range w.presets {
		reserved[n] = true
	}
	for _, n := range names {
		reserved[n] = true
	}
	renamed := map[string]string{}
	for _, n := range names {
		target := n
		if _, exists := w.presets[n]; exists {
			if replace {
				if w.origins[n].File != w.StorePath {
					return plan, fmt.Errorf("preset %s is owned by %s; edit that file or import with a new name", n, w.origins[n].File)
				}
			} else {
				for i := 2; ; i++ {
					suffix := fmt.Sprintf("-%d", i)
					stem := n
					if len(stem)+len(suffix) > 64 {
						stem = stem[:64-len(suffix)]
					}
					target = stem + suffix
					if !reserved[target] {
						break
					}
				}
			}
		}
		renamed[n] = target
		reserved[target] = true
	}
	next := Bundle{Presets: map[string]Preset{}}
	for n, p := range w.Store.Presets {
		next.Presets[n] = p
	}
	for _, n := range names {
		p := b.Presets[n]
		if renamed[p.Extends] != "" {
			p.Extends = renamed[p.Extends]
		}
		next.Presets[renamed[n]] = p
	}
	// Persistent user presets cannot depend on a project-only recipe that vanishes
	// on the next launch elsewhere. Dependencies included in the import are remapped.
	userOnly := *w
	userOnly.Project = Layer{}
	userOnly.Store = next
	userOnly.rebuild()
	for _, n := range names {
		dest := renamed[n]
		effective, e := userOnly.Resolve(dest, Fields{})
		if e != nil {
			return plan, e
		}
		item := ImportItem{SourceName: n, Name: dest, After: next.Presets[dest], Warnings: effective.Warnings}
		if old, ok := w.Store.Presets[dest]; ok {
			item.Before = &old
			item.Replaces = true
		}
		plan.Items = append(plan.Items, item)
	}
	candidate := *w
	candidate.Store = next
	candidate.rebuild()
	if e := candidate.validateAll(); e != nil {
		return plan, e
	}
	plan.next = next
	return plan, nil
}
func (w *Workspace) validateAll() error {
	for n := range w.presets {
		if _, e := w.Resolve(n, Fields{}); e != nil {
			return e
		}
	}
	if _, e := w.Resolve("", Fields{}); e != nil {
		return e
	}
	return nil
}
func (w *Workspace) ApplyImport(ctx context.Context, plan *ImportPlan) error {
	if plan.Destination != w.StorePath || plan.PreviousHash != digest(w.StoreBytes) || plan.next.Presets == nil {
		return fmt.Errorf("import preview does not belong to this configuration snapshot")
	}
	if e := w.save(ctx, plan.next); e != nil {
		return e
	}
	plan.Applied = true
	return nil
}
func (w *Workspace) ChangeMenu(ctx context.Context, name string, pinned *bool, position *int) error {
	p, ok := w.Store.Presets[name]
	if !ok || w.origins[name].File != w.StorePath {
		return fmt.Errorf("only app-imported presets can be edited here; %s is built-in or file-owned", name)
	}
	if pinned != nil {
		p.Menu.Pinned = pinned
	}
	if position != nil {
		p.Menu.Position = position
	}
	plan, e := w.PlanImport(Bundle{Presets: map[string]Preset{name: p}}, true)
	if e != nil {
		return e
	}
	return w.ApplyImport(ctx, &plan)
}
func (w *Workspace) Remove(ctx context.Context, name string) error {
	if _, ok := w.Store.Presets[name]; !ok || w.origins[name].File != w.StorePath {
		return fmt.Errorf("preset %s is not app-managed", name)
	}
	next := Bundle{Presets: map[string]Preset{}}
	for n, p := range w.Store.Presets {
		if n != name {
			next.Presets[n] = p
		}
	}
	copy := *w
	copy.Store = next
	copy.rebuild()
	if e := copy.validateAll(); e != nil {
		return fmt.Errorf("removal would break a selected or inherited recipe: %w", e)
	}
	return w.save(ctx, next)
}
func (w *Workspace) save(ctx context.Context, b Bundle) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	if e := os.MkdirAll(filepath.Dir(w.StorePath), 0700); e != nil {
		return e
	}
	lock := w.StorePath + ".lock"
	f, e := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return fmt.Errorf("preset store is busy or locked (%s): %w", lock, e)
	}
	f.Close()
	defer os.Remove(lock)
	for _, s := range []struct {
		path     string
		expected []byte
	}{{w.StorePath, w.StoreBytes}, {w.User.Path, w.User.Bytes}, {w.Project.Path, w.Project.Bytes}} {
		if s.path == "" {
			continue
		}
		st, er := os.Lstat(s.path)
		if er == nil && !st.Mode().IsRegular() {
			return fmt.Errorf("refusing to manage nonregular config path %s", s.path)
		}
		data, er := os.ReadFile(s.path)
		if er != nil && !os.IsNotExist(er) {
			return er
		}
		if !bytes.Equal(data, s.expected) {
			return fmt.Errorf("configuration changed since preview: %s; retry with a fresh preview", s.path)
		}
	}
	data, e := MarshalBundle(b)
	if e != nil {
		return e
	}
	if len(b.Presets) == 0 {
		data = []byte("# smolpng app-managed presets; no installed definitions.\n")
	}
	if e = fsutil.AtomicWrite(ctx, w.StorePath, data, 0600, true); e != nil {
		return e
	}
	w.Store = b
	w.StoreBytes = data
	w.rebuild()
	return nil
}

// RenameExport changes a single exported leaf while retaining included parents.
func RenameExport(data []byte, oldName, newName string) ([]byte, error) {
	if !validName(newName) || domain.IsBuiltin(newName) {
		return nil, fmt.Errorf("export name is invalid or reserved")
	}
	b, e := ParseBundle(data)
	if e != nil {
		return nil, e
	}
	p, ok := b.Presets[oldName]
	if !ok {
		return nil, fmt.Errorf("exported preset %s not found", oldName)
	}
	if _, exists := b.Presets[newName]; exists && newName != oldName {
		return nil, fmt.Errorf("export name collides with an included dependency")
	}
	delete(b.Presets, oldName)
	b.Presets[newName] = p
	for n, p := range b.Presets {
		if p.Extends == oldName {
			p.Extends = newName
			b.Presets[n] = p
		}
	}
	return MarshalBundle(b)
}
func IsEmptyStore(data []byte) bool {
	return strings.TrimSpace(string(data)) == "" || strings.TrimSpace(string(data)) == "# smolpng app-managed presets; no installed definitions."
}
