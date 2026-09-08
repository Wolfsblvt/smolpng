// SPDX-License-Identifier: MIT
// Package ui owns terminal interaction; it contains no compression implementation.
package ui

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
)

var ErrBack = errors.New("back or cancelled")

type Option struct{ Label, Value, Key string }
type Field struct {
	Key, Label, Value string
	Options           []Option
	Validate          func(string) error
}
type Prompter interface {
	Choose(context.Context, string, string, []Option, string) (string, error)
	Edit(context.Context, string, []Field) (map[string]string, error)
}
type Terminal struct {
	In                io.Reader
	Out               io.Writer
	Plain, Accessible bool
	reader            *bufio.Reader
}

func IsTerminal(in io.Reader, out io.Writer) bool {
	i, ok := in.(*os.File)
	if !ok {
		return false
	}
	o, ok := out.(*os.File)
	return ok && term.IsTerminal(int(i.Fd())) && term.IsTerminal(int(o.Fd()))
}
func Width(out io.Writer) int {
	if f, ok := out.(*os.File); ok {
		w, _, e := term.GetSize(int(f.Fd()))
		if e == nil {
			return w
		}
	}
	return 80
}
func Safe(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return ' '
		}
		return r
	}, ansi.Strip(s))
}
func line(s string) string { return strings.ReplaceAll(Safe(s), "\n", " ") }

// Header is a replaceable neutral placeholder, not final commissioned branding.
func Header(kind string, width int) string {
	if kind == "none" {
		return ""
	}
	if kind == "compact" || width < 48 {
		return "smolpng\n"
	}
	return "  ( . . )~\n   smolpng\n"
}
func (t *Terminal) Choose(ctx context.Context, title, description string, options []Option, initial string) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no available choices")
	}
	if t.Accessible {
		return t.chooseAccessible(ctx, title, description, options, initial)
	}
	m := menu{title: title, description: description, options: options, height: 24, plain: t.Plain}
	for i, o := range options {
		if o.Value == initial {
			m.index = i
		}
	}
	result, e := tea.NewProgram(m, tea.WithInput(t.In), tea.WithOutput(t.Out), tea.WithContext(ctx)).Run()
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if e != nil {
		return "", e
	}
	done, ok := result.(menu)
	if !ok || done.cancel {
		return "", ErrBack
	}
	return done.chosen, nil
}
func (t *Terminal) Edit(ctx context.Context, title string, fields []Field) (map[string]string, error) {
	if t.Accessible {
		return t.editAccessible(ctx, title, fields)
	}
	values := make([]string, len(fields))
	formFields := []huh.Field{}
	for i, f := range fields {
		values[i] = f.Value
		if len(f.Options) > 0 {
			opts := []huh.Option[string]{}
			for _, o := range f.Options {
				opts = append(opts, huh.NewOption(line(o.Label), o.Value))
			}
			formFields = append(formFields, huh.NewSelect[string]().Title(line(f.Label)).Options(opts...).Value(&values[i]))
		} else {
			field := huh.NewInput().Title(line(f.Label)).Value(&values[i])
			if f.Validate != nil {
				field.Validate(f.Validate)
			}
			formFields = append(formFields, field)
		}
	}
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))
	form := huh.NewForm(huh.NewGroup(formFields...).Title(line(title))).WithInput(t.In).WithOutput(t.Out).WithKeyMap(keymap)
	if t.Plain {
		form.WithTheme(huh.ThemeFunc(huh.ThemeBase))
	}
	e := form.RunWithContext(ctx)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if errors.Is(e, huh.ErrUserAborted) {
		return nil, ErrBack
	}
	if e != nil {
		return nil, e
	}
	out := map[string]string{}
	for i, f := range fields {
		if f.Validate != nil {
			if e = f.Validate(values[i]); e != nil {
				return nil, e
			}
		}
		out[f.Key] = values[i]
	}
	return out, nil
}

type menu struct {
	title, description string
	options            []Option
	index, height      int
	chosen             string
	cancel, plain      bool
}

func (m menu) Init() tea.Cmd { return nil }
func (m menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k", "shift+tab":
			m.index = (m.index + len(m.options) - 1) % len(m.options)
		case "down", "j", "tab":
			m.index = (m.index + 1) % len(m.options)
		case "home":
			m.index = 0
		case "end":
			m.index = len(m.options) - 1
		case "enter":
			m.chosen = m.options[m.index].Value
			return m, tea.Quit
		case "esc", "ctrl+c":
			m.cancel = true
			return m, tea.Quit
		default:
			for _, o := range m.options {
				if o.Key != "" && msg.String() == o.Key {
					m.chosen = o.Value
					return m, tea.Quit
				}
			}
		}
	}
	return m, nil
}
func (m menu) View() tea.View {
	if m.chosen != "" || m.cancel {
		return tea.NewView("")
	}
	var b strings.Builder
	fmt.Fprintln(&b, Safe(m.title))
	if m.description != "" {
		fmt.Fprintln(&b, Safe(m.description))
	}
	b.WriteByte('\n')
	rows := max(4, min(12, m.height-8))
	start := max(0, m.index-rows+1)
	end := min(len(m.options), start+rows)
	if start > 0 {
		b.WriteString("  ...\n")
	}
	for i := start; i < end; i++ {
		marker := "  "
		if i == m.index {
			marker = "❯ "
			if m.plain {
				marker = "> "
			}
		}
		label := line(m.options[i].Label)
		if m.options[i].Key != "" {
			label += "  [" + m.options[i].Key + "]"
		}
		fmt.Fprintln(&b, marker+label)
	}
	if end < len(m.options) {
		b.WriteString("  ...\n")
	}
	b.WriteString("\n↑/↓ navigate · Enter select · Esc back\n")
	if m.plain {
		return tea.NewView(strings.ReplaceAll(b.String(), "↑/↓", "Up/Down"))
	}
	return tea.NewView(b.String())
}
func (t *Terminal) readLine(ctx context.Context) (string, error) {
	if e := ctx.Err(); e != nil {
		return "", e
	}
	if t.reader == nil {
		t.reader = bufio.NewReader(t.In)
	}
	s, e := t.reader.ReadString('\n')
	if e != nil {
		return "", e
	}
	s = strings.TrimSpace(s)
	if s == "/back" {
		return "", ErrBack
	}
	return s, nil
}
func (t *Terminal) chooseAccessible(ctx context.Context, title, description string, options []Option, initial string) (string, error) {
	fmt.Fprintln(t.Out, Safe(title))
	fmt.Fprintln(t.Out, Safe(description))
	for _, o := range options {
		fmt.Fprintf(t.Out, "%s: %s\n", line(o.Value), line(o.Label))
	}
	for {
		fmt.Fprintf(t.Out, "Type a choice name (Enter: %s; /back cancels): ", line(initial))
		s, e := t.readLine(ctx)
		if e != nil {
			return "", e
		}
		if s == "" {
			s = initial
		}
		for _, o := range options {
			if strings.EqualFold(s, o.Value) || strings.EqualFold(s, o.Label) {
				return o.Value, nil
			}
		}
		fmt.Fprintln(t.Out, "Choose one of the displayed names.")
	}
}
func (t *Terminal) editAccessible(ctx context.Context, title string, fields []Field) (map[string]string, error) {
	fmt.Fprintln(t.Out, Safe(title))
	out := map[string]string{}
	for _, f := range fields {
		if len(f.Options) > 0 {
			v, e := t.chooseAccessible(ctx, f.Label, "", f.Options, f.Value)
			if e != nil {
				return nil, e
			}
			out[f.Key] = v
			continue
		}
		for {
			fmt.Fprintf(t.Out, "%s [%s] (/clear empties, /back cancels): ", line(f.Label), line(f.Value))
			s, e := t.readLine(ctx)
			if e != nil {
				return nil, e
			}
			if s == "" {
				s = f.Value
			} else if s == "/clear" {
				s = ""
			}
			if f.Validate != nil {
				if e = f.Validate(s); e != nil {
					fmt.Fprintln(t.Out, Safe(e.Error()))
					continue
				}
			}
			out[f.Key] = s
			break
		}
	}
	return out, nil
}
