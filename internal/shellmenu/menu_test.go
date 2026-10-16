package shellmenu

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func canonicalTempDir(t *testing.T) string {
	t.Helper()
	// macOS's temporary root traverses /var -> /private/var. Fixtures that
	// exercise symlink-free storage use its physical location.
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func testSpec(t *testing.T, mode string) Spec {
	t.Helper()
	dir := canonicalTempDir(t)
	s, e := New(mode, filepath.Join(dir, "smolpng.exe"), filepath.Join(dir, "config.toml"), "", nil)
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func TestMenuRoundTripAndExplicitDestruction(t *testing.T) {
	for _, mode := range []string{"single", "submenu"} {
		s := testSpec(t, mode)
		b, e := Encode(s)
		if e != nil {
			t.Fatal(e)
		}
		got, e := Decode(b)
		if e != nil || !reflect.DeepEqual(s, got) {
			t.Fatal(got, e)
		}
		if mode == "single" && (len(got.Actions) != 1 || got.Actions[0].InPlace) {
			t.Fatal("single became destructive")
		}
		if mode == "submenu" && (!got.Actions[0].InPlace || got.Actions[0].Label != "Make smol (replace original)") {
			t.Fatal("replacement lost explicit label")
		}
	}
	s := testSpec(t, "submenu")
	s.Actions[0].Label = "Auto"
	if s.Validate() == nil {
		t.Fatal("unlabelled destructive action accepted")
	}
	s = testSpec(t, "single")
	s.Actions[0].InPlace = true
	if s.Validate() == nil {
		t.Fatal("single replacement accepted")
	}
}
func TestMenuRejectsInjectedUnknownAndDuplicatedFields(t *testing.T) {
	s := testSpec(t, "submenu")
	s.Actions = append(s.Actions, Action{ID: "custom", Label: "recipe\nExec=malicious", Preset: "auto"})
	if _, e := Encode(s); e == nil {
		t.Fatal("new-line injection accepted")
	}
	s = testSpec(t, "single")
	b, _ := Encode(s)
	for _, data := range [][]byte{nil, b[:3], append(append([]byte{}, b...), 0), bytes.Repeat([]byte{1}, 65538)} {
		if _, e := Decode(data); e == nil {
			t.Fatal("invalid encoding accepted")
		}
	}
	s = testSpec(t, "submenu")
	s.Actions = append(s.Actions, s.Actions[0])
	if s.Validate() == nil {
		t.Fatal("duplicate action accepted")
	}
}
func TestNativeCPlusPlusHandoffMatchesGo(t *testing.T) {
	b, e := os.ReadFile("testdata/native-request.bin")
	if e != nil {
		t.Fatal(e)
	}
	q, e := DecodeRequest(b)
	if e != nil {
		t.Fatal(e)
	}
	if q.Generation != "00112233445566778899aabbccddeeff" || q.Action != "auto" || len(q.Files) != 2 || q.Files[0] != "C:\\Pictures\\snow 雪.png" {
		t.Fatal(q)
	}
	again, e := EncodeRequest(q)
	if e != nil || !bytes.Equal(again, b) {
		t.Fatal("cross-language wire disagreement", e)
	}
	for _, data := range [][]byte{b[:len(b)-1], append(append([]byte{}, b...), 0), []byte("SMOLPNG-SHELL-2\n")} {
		if _, e := DecodeRequest(data); e == nil {
			t.Fatal("malformed request accepted")
		}
	}
	q.Files = nil
	if _, e := EncodeRequest(q); e == nil {
		t.Fatal("empty selection accepted")
	}
}
func TestHandoffConsumeValidatesLocationAndPreservesUnrelatedFiles(t *testing.T) {
	dir := canonicalTempDir(t)
	s := testSpec(t, "single")
	source := filepath.Join(dir, "photo.jpg")
	_ = os.WriteFile(source, []byte("original"), 0600)
	q := Request{s.Generation, "auto", []string{source}}
	data, e := EncodeRequest(q)
	if e != nil {
		t.Fatal(e)
	}
	p := filepath.Join(dir, "smolpng-00112233445566778899aabbccddeeff.request")
	_ = os.WriteFile(p, data, 0600)
	got, e := Consume(p, dir)
	if e != nil || len(got.Files) != 1 {
		t.Fatal(e)
	}
	if _, e = os.Stat(p); !os.IsNotExist(e) {
		t.Fatal("consumed request retained")
	}
	if _, e = Consume(source, dir); e == nil {
		t.Fatal("unrelated file accepted")
	}
	b, _ := os.ReadFile(source)
	if string(b) != "original" {
		t.Fatal("deleted unrelated file")
	}
	_ = os.WriteFile(p, data, 0600)
	if _, e = Consume(p, filepath.Join(dir, "other")); e == nil {
		t.Fatal("request root escaped")
	}
}
func TestSelectionFilesOnlyNoClipboardFallback(t *testing.T) {
	if ValidateFiles(nil) == nil {
		t.Fatal("empty selection accepted")
	}
	d := t.TempDir()
	p := filepath.Join(d, "a.PNG")
	_ = os.WriteFile(p, nil, 0600)
	if e := ValidateFiles([]string{p}); e != nil {
		t.Fatal(e)
	}
	if ValidateFiles([]string{d}) == nil || ValidateFiles([]string{"https://host/a.png"}) == nil || ValidateFiles([]string{"relative.png"}) == nil {
		t.Fatal("directory/URL/relative input accepted")
	}
}
func TestDolphinFormatIsSpecificAndUsesIndependentFiles(t *testing.T) {
	s := testSpec(t, "submenu")
	files, e := DolphinFiles(s, func(args []string) string { return strings.Join(args, "|") + " %F" })
	if e != nil {
		t.Fatal(e)
	}
	text := string(files["dolphin-menu"]) + string(files["dolphin-interactive"])
	for _, value := range []string{"Type=Service", "MimeType=image/png;image/jpeg;", "X-KDE-Submenu=smolpng", "Terminal=true", "--generation|" + s.Generation, " %F"} {
		if !strings.Contains(text, value) {
			t.Fatal("missing", value)
		}
	}
}
