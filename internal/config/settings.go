// SPDX-License-Identifier: MIT
package config

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

// UserSettingKinds is intentionally limited to preferences the human workflow
// edits. Arbitrary executable commands, secrets and hooks are not a setting API.
var UserSettingKinds = map[string]string{
	"defaults.jpeg_quality": "string",
	"defaults.preset":       "string", "defaults.remote_policy": "string", "defaults.quality": "string", "defaults.jobs": "int", "defaults.recursive": "bool",
	"output.template": "string", "output.directory": "string", "output.detached": "string", "output.collision": "string",
	"clipboard.single": "string", "clipboard.batch": "string", "drop.mode": "string", "drop.preset": "string", "drop.copy": "string", "drop.notify": "bool",
	"updates.check": "string", "updates.channel": "string", "ui.header": "string", "ui.animations": "bool", "tools.pngquant": "string", "tools.oxipng": "string", "tools.jpegoptim": "string", "tools.preference": "string",
}

func SettingValue(key, text string) (any, error) {
	switch UserSettingKinds[key] {
	case "string":
		return text, nil
	case "int":
		n, e := strconv.Atoi(text)
		return n, e
	case "bool":
		b, e := strconv.ParseBool(text)
		return b, e
	}
	return nil, fmt.Errorf("setting %q is not supported by the settings editor", key)
}

type edit struct {
	start, end int
	text       string
}

// PatchSettings uses the pinned TOML parser's byte ranges to preserve comments,
// formatting and untouched values. A shape it cannot update safely is rejected;
// it never falls back to rewriting an entire handwritten configuration.
func PatchSettings(original []byte, values map[string]any) ([]byte, error) {
	if _, e := Parse(original); e != nil {
		return nil, e
	}
	encoded := map[string]string{}
	for key, value := range values {
		kind, ok := UserSettingKinds[key]
		if !ok {
			return nil, fmt.Errorf("unsupported setting %q", key)
		}
		switch kind {
		case "string":
			if _, ok := value.(string); !ok {
				return nil, fmt.Errorf("%s needs a string", key)
			}
		case "int":
			if _, ok := value.(int); !ok {
				return nil, fmt.Errorf("%s needs an integer", key)
			}
		case "bool":
			if _, ok := value.(bool); !ok {
				return nil, fmt.Errorf("%s needs a boolean", key)
			}
		}
		b, e := toml.Marshal(map[string]any{"value": value})
		if e != nil {
			return nil, e
		}
		_, raw, ok := strings.Cut(string(b), "=")
		if !ok {
			return nil, fmt.Errorf("cannot encode setting value")
		}
		encoded[key] = strings.TrimSpace(raw)
	}
	parser := unstable.Parser{}
	parser.Reset(original)
	table := []string{}
	edits := []edit{}
	found := map[string]bool{}
	headers := map[string]int{}
	parts := func(n *unstable.Node) []string {
		out := []string{}
		keys := n.Key()
		for keys.Next() {
			out = append(out, string(keys.Node().Data))
		}
		return out
	}
	var valueNodes func(*unstable.Node, []string) error
	valueNodes = func(n *unstable.Node, base []string) error {
		keyParts := append(append([]string{}, base...), parts(n)...)
		key := strings.Join(keyParts, ".")
		v := n.Value()
		if text, ok := encoded[key]; ok {
			if v.Kind != unstable.String && v.Kind != unstable.Integer && v.Kind != unstable.Bool {
				return fmt.Errorf("cannot replace non-scalar setting %q", key)
			}
			start, end := int(v.Raw.Offset), int(v.Raw.Offset+v.Raw.Length)
			edits = append(edits, edit{start, end, text})
			found[key] = true
		}
		if v.Kind == unstable.InlineTable {
			children := v.Children()
			for children.Next() {
				if children.Node().Kind == unstable.KeyValue {
					if e := valueNodes(children.Node(), keyParts); e != nil {
						return e
					}
				}
			}
		}
		return nil
	}
	for parser.NextExpression() {
		n := parser.Expression()
		switch n.Kind {
		case unstable.Table:
			table = parts(n)
			end := 0
			it := n.Key()
			for it.Next() {
				k := it.Node()
				end = int(k.Raw.Offset + k.Raw.Length)
			}
			if end > len(original) {
				return nil, fmt.Errorf("invalid parser range")
			}
			lineEnd := bytes.IndexByte(original[end:], '\n')
			if lineEnd < 0 {
				end = len(original)
			} else {
				end += lineEnd + 1
			}
			headers[strings.Join(table, ".")] = end
		case unstable.ArrayTable:
			return nil, fmt.Errorf("settings editor does not rewrite array-table configuration; edit the file directly")
		case unstable.KeyValue:
			if e := valueNodes(n, table); e != nil {
				return nil, e
			}
		}
	}
	if e := parser.Error(); e != nil {
		return nil, e
	}
	inserts := map[int][]string{}
	keys := sortedSettingKeys(encoded)
	for _, key := range keys {
		if found[key] {
			continue
		}
		section, leaf, _ := strings.Cut(key, ".")
		offset := 0
		assignment := key
		if at, ok := headers[section]; ok {
			offset = at
			assignment = leaf
		}
		inserts[offset] = append(inserts[offset], assignment+" = "+encoded[key]+"\n")
	}
	for offset, lines := range inserts {
		text := strings.Join(lines, "")
		if offset > 0 && original[offset-1] != '\n' {
			text = "\n" + text
		}
		edits = append(edits, edit{offset, offset, text})
	}
	sort.SliceStable(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	out := bytes.Clone(original)
	for _, e := range edits {
		out = append(append(append([]byte{}, out[:e.start]...), []byte(e.text)...), out[e.end:]...)
	}
	if _, e := Parse(out); e != nil {
		return nil, fmt.Errorf("settings update would conflict with this TOML layout; no file changed: %w", e)
	}
	return out, nil
}
func sortedSettingKeys[V any](m map[string]V) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func (w *Workspace) SaveUserSettings(ctx context.Context, values map[string]any) error {
	if len(values) == 0 {
		return nil
	}
	data, e := PatchSettings(w.User.Bytes, values)
	if e != nil {
		return e
	}
	doc, e := Parse(data)
	if e != nil {
		return e
	}
	candidate := *w
	candidate.User.Document = doc
	candidate.User.Exists = true
	if e = candidate.rebuild(); e != nil {
		return e
	}
	if e = candidate.validateAll(); e != nil {
		return e
	}
	// Ordinary user defaults must also work outside this particular project.
	global := candidate
	global.Project = Layer{}
	if e = global.rebuild(); e != nil {
		return e
	}
	if e = global.validateAll(); e != nil {
		return fmt.Errorf("user settings would not work outside this project: %w", e)
	}
	if e = os.MkdirAll(filepath.Dir(w.User.Path), 0700); e != nil {
		return e
	}
	lock, e := fsutil.TryLock(w.User.Path + ".settings.lock")
	if e != nil {
		return fmt.Errorf("settings are being edited elsewhere: %w", e)
	}
	defer lock.Close()
	current, e := os.ReadFile(w.User.Path)
	if e != nil && !os.IsNotExist(e) {
		return e
	}
	if !bytes.Equal(current, w.User.Bytes) {
		return fmt.Errorf("configuration changed since this settings screen opened; reload before saving")
	}
	if info, e := os.Lstat(w.User.Path); e == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("settings editor refuses nonregular/symlink config paths")
	}
	if e = fsutil.AtomicWrite(ctx, w.User.Path, data, 0600, w.User.Exists); e != nil {
		return e
	}
	w.User.Bytes = data
	w.User.Document = doc
	w.User.Exists = true
	w.rebuild()
	return nil
}
