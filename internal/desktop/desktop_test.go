package desktop

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/Wolfsblvt/smolpng/internal/process"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
	native "golang.design/x/clipboard"
)

func output(t *testing.T, name string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if e := os.WriteFile(p, testutil.Corpus()["flat-icon.png"], 0600); e != nil {
		t.Fatal(e)
	}
	return p
}
func items(t *testing.T, opts []native.Option) map[native.Format][]byte {
	t.Helper()
	out := map[native.Format][]byte{}
	for _, o := range opts {
		v, ok := o.(native.Item)
		if !ok {
			t.Fatalf("unexpected option %T", o)
		}
		if _, ok = out[v.Format]; ok {
			t.Fatal("format repeated")
		}
		out[v.Format] = v.Bytes
	}
	return out
}
func TestSmartPayloadPublishesImageFilesAndPathTogether(t *testing.T) {
	p := output(t, "image space # % ü.png")
	opts, modes, e := CopyPayload([]string{p}, "smart")
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(modes, []string{"image", "files", "path-text"}) {
		t.Fatal(modes)
	}
	parts := items(t, opts)
	b, _ := os.ReadFile(p)
	if !bytes.Equal(parts[native.FmtImage], b) || string(parts[native.FmtText]) != p {
		t.Fatal("smart representation mismatch")
	}
	u, e := url.Parse(strings.TrimSpace(string(parts[native.FmtFiles])))
	if e != nil || u.Scheme != "file" || u.RawQuery != "" || u.Fragment != "" {
		t.Fatal(u, e)
	}
	expected := filepath.ToSlash(p)
	if runtime.GOOS == "windows" {
		expected = "/" + expected
	}
	if u.Path != expected {
		t.Fatalf("URI path %q want %q", u.Path, expected)
	}
}
func TestBatchPayloadDoesNotPretendMultipleImagesAreOne(t *testing.T) {
	a, b := output(t, "a.png"), output(t, "b.png")
	opts, modes, e := CopyPayload([]string{a, b}, "smart")
	if e != nil || !reflect.DeepEqual(modes, []string{"files", "path-text"}) {
		t.Fatal(modes, e)
	}
	parts := items(t, opts)
	if _, ok := parts[native.FmtImage]; ok {
		t.Fatal("batch had single-image bytes")
	}
	if string(parts[native.FmtText]) != a+"\n"+b {
		t.Fatal("batch path ordering changed")
	}
	if _, _, e = CopyPayload([]string{a, b}, "image"); e == nil {
		t.Fatal("batch image copy allowed")
	}
}
func TestCopyModesAndMissingArtifacts(t *testing.T) {
	p := output(t, "x.png")
	for _, mode := range []string{"path", "file", "files", "image", "directory", "none"} {
		opts, _, e := CopyPayload([]string{p}, mode)
		if e != nil {
			t.Fatal(mode, e)
		}
		if mode == "none" && len(opts) != 0 {
			t.Fatal("none copied data")
		}
	}
	for _, tt := range []struct {
		paths []string
		mode  string
	}{{nil, "smart"}, {[]string{filepath.Join(t.TempDir(), "missing")}, "path"}, {[]string{t.TempDir()}, "files"}, {[]string{p}, "invented"}} {
		if _, _, e := CopyPayload(tt.paths, tt.mode); e == nil {
			t.Fatalf("accepted %+v", tt)
		}
	}
	os.WriteFile(p, []byte("bad PNG"), 0600)
	if _, _, e := CopyPayload([]string{p}, "image"); e == nil {
		t.Fatal("invalid PNG offered")
	}
}
func TestOwnerHoldStopsOnReplacementCancellationOrBound(t *testing.T) {
	changed := make(chan struct{})
	close(changed)
	start := time.Now()
	Hold(context.Background(), CopyReceipt{NeedsOwner: true, Changed: changed}, time.Hour)
	if time.Since(start) > time.Second {
		t.Fatal("replacement not observed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	Hold(ctx, CopyReceipt{NeedsOwner: true, Changed: make(chan struct{})}, time.Hour)
	start = time.Now()
	Hold(context.Background(), CopyReceipt{NeedsOwner: true, Changed: make(chan struct{})}, time.Millisecond)
	if time.Since(start) > time.Second {
		t.Fatal("unbounded owner lifetime")
	}
	Hold(context.Background(), CopyReceipt{NeedsOwner: false}, time.Hour)
}

type recorder struct {
	calls []process.Invocation
	exit  int
	err   error
}

func (r *recorder) Run(_ context.Context, i process.Invocation) (process.Result, error) {
	r.calls = append(r.calls, i)
	return process.Result{ExitCode: r.exit}, r.err
}
func TestDesktopCommandsKeepArgumentsSeparateAndRejectCredentialURLs(t *testing.T) {
	r := &recorder{}
	n := NativeActions{Runner: r}
	for _, u := range []string{"file:///etc/passwd", "http://example.test/x", "https://name:secret@example.test/", "invalid"} {
		if e := n.OpenURL(context.Background(), u); e == nil {
			t.Fatal(u)
		}
	}
	if len(r.calls) != 0 {
		t.Fatal("invalid URLs launched a process")
	}
	p := filepath.Join(t.TempDir(), "a; ' $(touch BAD)")
	if e := n.OpenFolder(context.Background(), p); e != nil {
		t.Fatal(e)
	}
	call := r.calls[0]
	if call.Args[len(call.Args)-1] != p {
		t.Fatal(call)
	}
	if e := n.OpenURL(context.Background(), "https://example.test/a?q=x%20y"); e != nil {
		t.Fatal(e)
	}
	r.exit = 7
	if e := n.Notify(context.Background(), "title", "body"); e == nil {
		t.Fatal("native failure hidden")
	}
	r.exit = 0
	r.err = errors.New("no desktop")
	if e := n.OpenFolder(context.Background(), p); e == nil {
		t.Fatal("process error hidden")
	}
}
func TestWindowsNotificationEscapesDataInsideFixedScript(t *testing.T) {
	title := "Wolf's '; Exit 7; #"
	body := "line\n$env:SECRET `$(danger)"
	raw, e := base64.StdEncoding.DecodeString(notificationScript(title, body))
	if e != nil {
		t.Fatal(e)
	}
	u := make([]uint16, len(raw)/2)
	for i := range u {
		u[i] = binary.LittleEndian.Uint16(raw[i*2:])
	}
	script := string(utf16.Decode(u))
	if !strings.Contains(script, "$n.BalloonTipTitle='Wolf''s ''; Exit 7; #'") || !strings.Contains(script, "$n.BalloonTipText='"+body+"'") {
		t.Fatal("data escaped outside literal", script)
	}
}
