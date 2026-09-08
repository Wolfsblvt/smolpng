package ui

import (
	"bytes"
	tea "charm.land/bubbletea/v2"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func press(m menu, code rune, text string) menu {
	next, _ := m.Update(tea.KeyPressMsg{Code: code, Text: text})
	return next.(menu)
}
func TestMenuKeyboardNavigationSelectionAndShortcuts(t *testing.T) {
	m := menu{options: []Option{{Label: "Auto", Value: "auto"}, {Label: "Target", Value: "target"}, {Label: "Settings", Value: "settings", Key: ","}}, height: 24}
	m = press(m, tea.KeyDown, "")
	if m.index != 1 {
		t.Fatal(m.index)
	}
	m = press(m, tea.KeyUp, "")
	if m.index != 0 {
		t.Fatal(m.index)
	}
	m = press(m, tea.KeyUp, "")
	if m.index != 2 {
		t.Fatal("did not wrap")
	}
	m = press(m, tea.KeyHome, "")
	m = press(m, tea.KeyEnter, "")
	if m.chosen != "auto" {
		t.Fatal(m.chosen)
	}
	m.chosen = ""
	m = press(m, ',', ",")
	if m.chosen != "settings" {
		t.Fatal("shortcut ignored")
	}
	m.chosen = ""
	m = press(m, tea.KeyEscape, "")
	if !m.cancel {
		t.Fatal("escape not back")
	}
}
func TestMenuViewportAndUntrustedLabelsAreSafe(t *testing.T) {
	opts := []Option{}
	for i := 0; i < 40; i++ {
		opts = append(opts, Option{Label: fmt.Sprintf("Recipe %d", i), Value: fmt.Sprint(i)})
	}
	opts[30].Label = "\x1b[2J\x1b]52;c;ZXZpbA==\aordinary"
	m := menu{title: "title", options: opts, height: 12, index: 30}
	text := m.View().Content
	if strings.Contains(text, "\x1b") || !strings.Contains(text, "ordinary") || strings.Contains(text, "Recipe 0\n") {
		t.Fatal(text)
	}
	if strings.Count(text, "\n") > 15 {
		t.Fatal("menu ignores viewport", text)
	}
}
func TestAccessibleSelectionUsesNamesNotNumbers(t *testing.T) {
	var out bytes.Buffer
	tui := &Terminal{In: strings.NewReader("1\nweb\n"), Out: &out, Accessible: true}
	got, e := tui.Choose(context.Background(), "Recipes", "", []Option{{"Auto", "auto", ""}, {"Web", "web", ""}}, "auto")
	if e != nil || got != "web" || !strings.Contains(out.String(), "Choose one of the displayed names") {
		t.Fatal(got, e, out.String())
	}
}
func TestAccessibleFormValidationBackAndEmptyValue(t *testing.T) {
	var out bytes.Buffer
	tui := &Terminal{In: strings.NewReader("wrong\nok\n/clear\n"), Out: &out, Accessible: true}
	values, e := tui.Edit(context.Background(), "settings", []Field{{Key: "x", Value: "old", Validate: func(v string) error {
		if v != "ok" {
			return fmt.Errorf("need ok")
		}
		return nil
	}}, {Key: "y", Value: "old"}})
	if e != nil || values["x"] != "ok" || values["y"] != "" {
		t.Fatal(values, e)
	}
	tui = &Terminal{In: strings.NewReader("/back\n"), Out: &out, Accessible: true}
	if _, e = tui.Edit(context.Background(), "", []Field{{Key: "x"}}); !errors.Is(e, ErrBack) {
		t.Fatal(e)
	}
}
func TestHeaderSeamAndNoFakeTerminal(t *testing.T) {
	if Header("none", 80) != "" || Header("cute", 30) != Header("compact", 80) || strings.Count(Header("cute", 80), "\n") > 3 {
		t.Fatal("header contract")
	}
	if IsTerminal(strings.NewReader(""), new(bytes.Buffer)) {
		t.Fatal("redirected streams treated as terminal")
	}
	if strings.Contains(Safe("\x1b[31mred\x1b[0m\x00\r"), "\x1b") {
		t.Fatal("terminal injection")
	}
}
