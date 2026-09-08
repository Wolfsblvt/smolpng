// SPDX-License-Identifier: MIT
// Package shellmenu owns a bounded, data-only desktop menu and handoff protocol.
// It does not load optimizer code into Explorer or change file associations.
package shellmenu

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/adrg/xdg"
)

const MenuFile = "smolpng-shell-menu.ini"
const MaxFiles = 512
const MaxRequest = 8 << 20
const RequestMagic = "SMOLPNG-SHELL-1\n"
const PackageName = "Wolfsblvt.smolpng.ContextMenu"
const ClassID = "90BB1F36-193A-4D1A-891B-91DB43AE63DC"

var idPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)
var generationPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)
var requestPattern = regexp.MustCompile(`^smolpng-[0-9a-f]{32}\.request$`)

type Action struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Preset      string `json:"preset"`
	Interactive bool   `json:"interactive"`
	InPlace     bool   `json:"inPlace"`
}
type Spec struct {
	Schema     int      `json:"schema"`
	Generation string   `json:"generation"`
	Mode       string   `json:"mode"`
	Executable string   `json:"executable"`
	Config     string   `json:"config"`
	Home       string   `json:"home,omitempty"`
	Actions    []Action `json:"actions"`
}
type Locations struct{ State, Requests, ServiceMenus string }

func NativeLocations(goos string) (Locations, error) {
	switch goos {
	case "windows":
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			return Locations{}, fmt.Errorf("LOCALAPPDATA unavailable")
		}
		root := filepath.Join(base, "smolpng")
		return Locations{State: filepath.Join(root, "shell-menu"), Requests: filepath.Join(root, "shell-requests")}, nil
	case "linux":
		root := filepath.Join(xdg.DataHome, "smolpng")
		return Locations{State: filepath.Join(root, "shell-menu"), Requests: filepath.Join(root, "shell-requests"), ServiceMenus: filepath.Join(xdg.DataHome, "kio", "servicemenus")}, nil
	default:
		return Locations{}, fmt.Errorf("native context menus are implemented for Windows 11 and KDE Dolphin; other desktops are not claimed")
	}
}
func New(mode, exe, config, home string, custom []Action) (Spec, error) {
	b := make([]byte, 16)
	if _, e := rand.Read(b); e != nil {
		return Spec{}, e
	}
	s := Spec{Schema: 1, Generation: hex.EncodeToString(b), Mode: mode, Executable: exe, Config: config, Home: home}
	if mode == "single" {
		s.Actions = []Action{{ID: "auto", Label: "Make smol", Preset: "auto"}}
	} else if mode == "submenu" {
		s.Actions = []Action{{ID: "replace", Label: "Make smol (replace original)", Preset: "auto", InPlace: true}, {ID: "interactive", Label: "Choose recipe...", Preset: "auto", Interactive: true}, {ID: "auto", Label: "Make smol (keep original)", Preset: "auto"}, {ID: "lossless", Label: "Lossless (keep original)", Preset: "lossless"}}
		s.Actions = append(s.Actions, custom...)
	} else {
		return s, fmt.Errorf("menu mode must be single or submenu")
	}
	return s, s.Validate()
}
func validText(s string) bool {
	for _, c := range s {
		if c < 32 || c == 127 {
			return false
		}
	}
	return utf8.ValidString(s) && !strings.ContainsAny(s, "\r\n\x00\t") && strings.TrimSpace(s) == s && len(s) < 32760
}
func (s Spec) Validate() error {
	if s.Schema != 1 || !generationPattern.MatchString(s.Generation) || !filepath.IsAbs(s.Executable) || !filepath.IsAbs(s.Config) || s.Home != "" && !filepath.IsAbs(s.Home) {
		return fmt.Errorf("invalid shell menu identity or absolute application/config path")
	}
	if !validText(s.Executable) || !validText(s.Config) || !validText(s.Home) {
		return fmt.Errorf("invalid shell menu text")
	}
	if s.Mode != "single" && s.Mode != "submenu" || len(s.Actions) < 1 || len(s.Actions) > 16 || s.Mode == "single" && len(s.Actions) != 1 {
		return fmt.Errorf("invalid shell menu mode or action count (maximum 16)")
	}
	seen := map[string]bool{}
	for _, a := range s.Actions {
		if !idPattern.MatchString(a.ID) || !idPattern.MatchString(a.Preset) || seen[a.ID] || !validText(a.Label) || len([]rune(a.Label)) < 1 || len([]rune(a.Label)) > 120 || a.InPlace && a.Interactive {
			return fmt.Errorf("invalid or duplicate shell action %q", a.ID)
		}
		if a.InPlace && a.ID != "replace" {
			return fmt.Errorf("only the explicitly labelled replace action may overwrite originals")
		}
		if a.ID == "replace" && (a.Label != "Make smol (replace original)" || !a.InPlace || a.Preset != "auto") {
			return fmt.Errorf("invalid destructive action label")
		}
		if s.Mode == "single" && (a.ID != "auto" || a.InPlace || a.Interactive || a.Preset != "auto") {
			return fmt.Errorf("single action must be nondestructive local Auto")
		}
		seen[a.ID] = true
	}
	return nil
}
func (s Spec) Action(id string) (Action, error) {
	for _, a := range s.Actions {
		if a.ID == id {
			return a, nil
		}
	}
	return Action{}, fmt.Errorf("shell action no longer exists; refresh the menu")
}

// Encode uses the Windows profile API's UTF-16LE INI shape. Values are data only;
// paths never become a PowerShell/cmd command. The exact bytes are installation-owned.
func Encode(s Spec) ([]byte, error) {
	if e := s.Validate(); e != nil {
		return nil, e
	}
	var text strings.Builder
	fmt.Fprintf(&text, "[menu]\r\nschema=1\r\ngeneration=%s\r\nmode=%s\r\nexecutable=%s\r\nconfig=%s\r\nhome=%s\r\ncount=%d\r\n", s.Generation, s.Mode, s.Executable, s.Config, s.Home, len(s.Actions))
	for i, a := range s.Actions {
		fmt.Fprintf(&text, "\r\n[action%d]\r\nid=%s\r\nlabel=%s\r\npreset=%s\r\ninteractive=%d\r\ninplace=%d\r\n", i, a.ID, a.Label, a.Preset, boolean(a.Interactive), boolean(a.InPlace))
	}
	code := utf16.Encode([]rune(text.String()))
	b := make([]byte, 2+len(code)*2)
	b[0] = 255
	b[1] = 254
	for i, c := range code {
		binary.LittleEndian.PutUint16(b[2+i*2:], c)
	}
	if len(b) > 65536 {
		return nil, fmt.Errorf("shell menu exceeds 64KiB")
	}
	return b, nil
}
func boolean(b bool) int {
	if b {
		return 1
	}
	return 0
}
func Decode(b []byte) (Spec, error) {
	var s Spec
	if len(b) < 2 || len(b) > 65536 || len(b)%2 != 0 || b[0] != 255 || b[1] != 254 {
		return s, fmt.Errorf("invalid UTF-16LE shell menu")
	}
	text, e := decodeUTF16(b[2:])
	if e != nil {
		return s, e
	}
	sections := map[string]map[string]string{}
	section := ""
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = line[1 : len(line)-1]
			if _, ok := sections[section]; ok {
				return s, fmt.Errorf("duplicate INI section")
			}
			sections[section] = map[string]string{}
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok || section == "" {
			return s, fmt.Errorf("invalid menu line")
		}
		if _, ok := sections[section][k]; ok {
			return s, fmt.Errorf("duplicate INI key")
		}
		sections[section][k] = v
	}
	m := sections["menu"]
	for _, key := range []string{"schema", "generation", "mode", "executable", "config", "home", "count"} {
		if _, ok := m[key]; !ok {
			return s, fmt.Errorf("missing menu field")
		}
	}
	if len(m) != 7 {
		return s, fmt.Errorf("unexpected menu fields")
	}
	s.Schema, _ = strconv.Atoi(m["schema"])
	s.Generation = m["generation"]
	s.Mode = m["mode"]
	s.Executable = m["executable"]
	s.Config = m["config"]
	s.Home = m["home"]
	n, e := strconv.Atoi(m["count"])
	if e != nil || n < 1 || n > 16 || len(sections) != n+1 {
		return s, fmt.Errorf("invalid action sections")
	}
	for i := 0; i < n; i++ {
		v := sections[fmt.Sprintf("action%d", i)]
		if len(v) != 5 || v["interactive"] != "0" && v["interactive"] != "1" || v["inplace"] != "0" && v["inplace"] != "1" {
			return s, fmt.Errorf("invalid action fields")
		}
		s.Actions = append(s.Actions, Action{v["id"], v["label"], v["preset"], v["interactive"] == "1", v["inplace"] == "1"})
	}
	return s, s.Validate()
}
func decodeUTF16(b []byte) (string, error) {
	if len(b)%2 != 0 {
		return "", fmt.Errorf("odd UTF-16 payload")
	}
	code := make([]uint16, len(b)/2)
	for i := range code {
		code[i] = binary.LittleEndian.Uint16(b[i*2:])
		if code[i] == 0 {
			return "", fmt.Errorf("NUL in text")
		}
	}
	for i := 0; i < len(code); i++ {
		c := code[i]
		if c >= 0xd800 && c <= 0xdbff {
			if i+1 >= len(code) || code[i+1] < 0xdc00 || code[i+1] > 0xdfff {
				return "", fmt.Errorf("invalid surrogate")
			}
			i++
		} else if c >= 0xdc00 && c <= 0xdfff {
			return "", fmt.Errorf("unpaired surrogate")
		}
	}
	return string(utf16.Decode(code)), nil
}

type Request struct {
	Generation string
	Action     string
	Files      []string
}

func DecodeRequest(b []byte) (Request, error) {
	var q Request
	if len(b) > MaxRequest {
		return q, fmt.Errorf("shell selection exceeds 8MiB")
	}
	r := bytes.NewReader(b)
	header := make([]byte, len(RequestMagic))
	if _, e := io.ReadFull(r, header); e != nil || string(header) != RequestMagic {
		return q, fmt.Errorf("invalid shell request header")
	}
	read := func(limit uint32) (string, error) {
		var n uint32
		if e := binary.Read(r, binary.LittleEndian, &n); e != nil {
			return "", e
		}
		if n == 0 || n > limit || uint64(n)*2 > uint64(r.Len()) {
			return "", fmt.Errorf("invalid request text length")
		}
		data := make([]byte, n*2)
		if _, e := io.ReadFull(r, data); e != nil {
			return "", e
		}
		return decodeUTF16(data)
	}
	var e error
	q.Generation, e = read(32)
	if e != nil {
		return q, e
	}
	q.Action, e = read(64)
	if e != nil {
		return q, e
	}
	var n uint32
	if e = binary.Read(r, binary.LittleEndian, &n); e != nil || n < 1 || n > MaxFiles {
		return q, fmt.Errorf("select 1..512 supported files")
	}
	for i := uint32(0); i < n; i++ {
		p, e := read(32767)
		if e != nil {
			return q, e
		}
		q.Files = append(q.Files, p)
	}
	if r.Len() != 0 || !generationPattern.MatchString(q.Generation) || !idPattern.MatchString(q.Action) {
		return q, fmt.Errorf("invalid/trailing shell request")
	}
	return q, nil
}
func EncodeRequest(q Request) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(RequestMagic)
	write := func(s string) {
		code := utf16.Encode([]rune(s))
		_ = binary.Write(&b, binary.LittleEndian, uint32(len(code)))
		for _, c := range code {
			_ = binary.Write(&b, binary.LittleEndian, c)
		}
	}
	write(q.Generation)
	write(q.Action)
	_ = binary.Write(&b, binary.LittleEndian, uint32(len(q.Files)))
	for _, p := range q.Files {
		write(p)
	}
	data := b.Bytes()
	if _, e := DecodeRequest(data); e != nil {
		return nil, e
	}
	return data, nil
}
func ValidateFiles(files []string) error {
	if len(files) < 1 || len(files) > MaxFiles {
		return fmt.Errorf("select 1..512 PNG/JPEG files; clipboard fallback is disabled")
	}
	for _, p := range files {
		ext := strings.ToLower(filepath.Ext(p))
		if !filepath.IsAbs(p) || strings.ContainsRune(p, 0) || ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			return fmt.Errorf("unsupported shell selection")
		}
		st, e := os.Stat(p)
		if e != nil || !st.Mode().IsRegular() {
			return fmt.Errorf("shell selection is not a readable regular file: %s", p)
		}
	}
	return nil
}

// Consume checks the exact owned request directory, prevents symlink traversal,
// bounds allocations and removes only a successfully decoded one-shot handoff.
func Consume(path, root string) (Request, error) {
	var q Request
	if filepath.Dir(path) != filepath.Clean(root) || !requestPattern.MatchString(filepath.Base(path)) {
		return q, fmt.Errorf("request is outside the native handoff directory")
	}
	for p := path; ; p = filepath.Dir(p) {
		st, e := os.Lstat(p)
		if e != nil {
			return q, e
		}
		if st.Mode()&os.ModeSymlink != 0 {
			return q, fmt.Errorf("symlink in shell handoff")
		}
		if filepath.Dir(p) == p {
			break
		}
	}
	f, e := os.Open(path)
	if e != nil {
		return q, e
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil || !st.Mode().IsRegular() || st.Size() > MaxRequest {
		return q, fmt.Errorf("invalid handoff file")
	}
	b, e := io.ReadAll(io.LimitReader(f, MaxRequest+1))
	if e != nil {
		return q, e
	}
	q, e = DecodeRequest(b)
	if e != nil {
		return q, e
	}
	if e = f.Close(); e != nil {
		return q, e
	}
	if e = os.Remove(path); e != nil {
		return q, e
	}
	return q, nil
}
