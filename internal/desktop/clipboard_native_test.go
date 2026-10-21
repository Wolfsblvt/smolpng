// SPDX-License-Identifier: MIT
//go:build nativeclipboard && (linux || darwin)

package desktop

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"image/png"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
)

// This separately selected journey replaces a disposable host's clipboard.
// Native tools/AppKit are independent consumers, not another call to our reader.
func nativeBackend(t *testing.T) string {
	t.Helper()
	backend := os.Getenv("SMOLPNG_CLIPBOARD_BACKEND")
	if runtime.GOOS == "darwin" && backend == "macos" || runtime.GOOS == "linux" && (backend == "x11" || backend == "wayland") {
		return backend
	}
	t.Fatal("select the disposable native host with SMOLPNG_CLIPBOARD_BACKEND=x11, wayland or macos")
	return ""
}

func nativeCommand(t *testing.T, stdin []byte, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = bytes.NewReader(stdin)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s %q: %v; %s", name, args, err, stderr.String())
	}
	return out
}

func nativeMIME(kind string) string {
	switch kind {
	case "image":
		return "image/png"
	case "files":
		return "text/uri-list"
	default:
		return "text/plain;charset=utf-8"
	}
}

func independentRead(t *testing.T, kind string) []byte {
	t.Helper()
	switch nativeBackend(t) {
	case "x11":
		target := nativeMIME(kind)
		if kind == "text" {
			target = "UTF8_STRING"
		}
		return nativeCommand(t, nil, "xclip", "-selection", "clipboard", "-out", "-target", target)
	case "wayland":
		return nativeCommand(t, nil, "wl-paste", "--no-newline", "--type", nativeMIME(kind))
	default:
		path := filepath.Join(t.TempDir(), "pasteboard.bin")
		nativeCommand(t, nil, "osascript", "-l", "JavaScript", "testdata/clipboard-macos.js", "read", kind, path)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
}

func independentWrite(t *testing.T, kind string, data []byte) {
	t.Helper()
	backend := nativeBackend(t)
	if backend == "macos" {
		path := filepath.Join(t.TempDir(), "pasteboard.bin")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		nativeCommand(t, nil, "osascript", "-l", "JavaScript", "testdata/clipboard-macos.js", "write", kind, path)
		return
	}
	var cmd *exec.Cmd
	if backend == "x11" {
		target := nativeMIME(kind)
		if kind == "text" {
			target = "UTF8_STRING"
		}
		cmd = exec.Command("xclip", "-selection", "clipboard", "-in", "-target", target, "-quiet")
	} else {
		cmd = exec.Command("wl-copy", "--foreground", "--type", nativeMIME(kind))
	}
	cmd.Stdin = bytes.NewReader(data)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	// Keep the exact foreign owner alive until replacement or test cleanup.
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	t.Cleanup(func() { _ = cmd.Process.Kill(); <-done })
	deadline := time.Now().Add(5 * time.Second)
	for {
		// Wait for the selected owner, rather than assuming process creation is
		// a completed selection. The consumer check also proves its bytes.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var check *exec.Cmd
		if backend == "x11" {
			target := nativeMIME(kind)
			if kind == "text" {
				target = "UTF8_STRING"
			}
			check = exec.CommandContext(ctx, "xclip", "-selection", "clipboard", "-out", "-target", target)
		} else {
			check = exec.CommandContext(ctx, "wl-paste", "--no-newline", "--type", nativeMIME(kind))
		}
		observed, err := check.Output()
		cancel()
		if err == nil && bytes.Equal(observed, data) {
			return
		}
		if time.Now().After(deadline) {
			_ = cmd.Process.Kill()
			writerErr := <-done
			done <- writerErr
			t.Fatalf("foreign %s writer never supplied its selected bytes: read %v, writer %v; %s", kind, err, writerErr, stderr.String())
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func clipboardFixture(t *testing.T, name string) (string, []byte) {
	t.Helper()
	data := testutil.Corpus()["alpha-gradient.png"]
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return path, data
}

func nativeApplication(t *testing.T) string {
	t.Helper()
	app := os.Getenv("SMOLPNG_NATIVE_APP")
	if !filepath.IsAbs(app) {
		t.Fatal("SMOLPNG_NATIVE_APP must name the exact built application")
	}
	var identity struct {
		Commit string `json:"commit"`
	}
	if err := json.Unmarshal(nativeCommand(t, nil, app, "version", "--json"), &identity); err != nil {
		t.Fatal(err)
	}
	expected := os.Getenv("SOURCE_COMMIT")
	if len(expected) != 40 || identity.Commit != expected {
		t.Fatalf("application source %q, expected exact SOURCE_COMMIT %q", identity.Commit, expected)
	}
	return app
}

func equalClipboardPixels(t *testing.T, expected, observed []byte) {
	t.Helper()
	a, err := png.Decode(bytes.NewReader(expected))
	if err != nil {
		t.Fatal(err)
	}
	b, err := png.Decode(bytes.NewReader(observed))
	if err != nil {
		t.Fatal(err)
	}
	if a.Bounds() != b.Bounds() {
		t.Fatalf("clipboard dimensions: %v, want %v", b.Bounds(), a.Bounds())
	}
	for y := a.Bounds().Min.Y; y < a.Bounds().Max.Y; y++ {
		for x := a.Bounds().Min.X; x < a.Bounds().Max.X; x++ {
			ar, ag, ab, aa := a.At(x, y).RGBA()
			br, bg, bb, ba := b.At(x, y).RGBA()
			if ar != br || ag != bg || ab != bb || aa != ba {
				t.Fatalf("clipboard pixel/alpha mismatch at %d,%d", x, y)
			}
		}
	}
	t.Logf("PNG consumer: %v pixels/alpha agree; encoded SHA-256 %x", b.Bounds(), sha256.Sum256(observed))
}

func independentFiles(t *testing.T) []string {
	t.Helper()
	data := independentRead(t, "files")
	if nativeBackend(t) == "macos" {
		var paths []string
		if err := json.Unmarshal(data, &paths); err != nil {
			t.Fatal(err)
		}
		return paths
	}
	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		u, err := url.Parse(strings.TrimSpace(line))
		if err != nil || u.Scheme != "file" || u.Host != "" || u.RawQuery != "" || u.Fragment != "" {
			t.Fatalf("invalid local file URI %q: %v", line, err)
		}
		paths = append(paths, u.Path)
	}
	return paths
}

func independentHasImage(t *testing.T) bool {
	t.Helper()
	var types []byte
	switch nativeBackend(t) {
	case "x11":
		types = nativeCommand(t, nil, "xclip", "-selection", "clipboard", "-out", "-target", "TARGETS")
	case "wayland":
		types = nativeCommand(t, nil, "wl-paste", "--list-types")
	default:
		types = independentRead(t, "types")
	}
	return bytes.Contains(types, []byte("image/png")) || bytes.Contains(types, []byte("public.png")) || bytes.Contains(types, []byte("public.tiff"))
}

func foreignFiles(t *testing.T, paths []string) []byte {
	t.Helper()
	if nativeBackend(t) == "macos" {
		data, err := json.Marshal(paths)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	var lines []string
	for _, path := range paths {
		lines = append(lines, (&url.URL{Scheme: "file", Path: path}).String())
	}
	return []byte(strings.Join(lines, "\r\n") + "\r\n")
}

func TestNativeClipboardSmartRepresentations(t *testing.T) {
	nativeBackend(t)
	a, image := clipboardFixture(t, "image space # % ü.png")
	b, _ := clipboardFixture(t, "second 雪.png")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, paths := range [][]string{{a}, {b, a}} {
		receipt, err := (NativeClipboard{}).Copy(ctx, paths, "smart")
		if err != nil {
			t.Fatal(err)
		}
		if receipt.NeedsOwner != (runtime.GOOS == "linux") {
			t.Fatal("incorrect clipboard owner lifetime", receipt)
		}
		if len(paths) == 1 {
			equalClipboardPixels(t, image, independentRead(t, "image"))
		} else if independentHasImage(t) {
			t.Fatal("batch clipboard retained a misleading single-image representation")
		}
		if actual := independentFiles(t); !reflect.DeepEqual(actual, paths) {
			t.Fatalf("native file-list %q, want %q", actual, paths)
		}
		if actual := string(independentRead(t, "text")); actual != strings.Join(paths, "\n") {
			t.Fatalf("native path text %q", actual)
		}
		t.Logf("smart output: %d file(s), native ordered file-list and exact path text agree; modes %v", len(paths), receipt.Modes)
	}
}

func TestNativeClipboardForeignSourcesReachCLI(t *testing.T) {
	nativeBackend(t)
	app := nativeApplication(t)
	a, image := clipboardFixture(t, "source # % ü.png")
	b, _ := clipboardFixture(t, "source 雪.png")
	for _, tc := range []struct {
		kind   string
		data   []byte
		inputs []string
	}{{"image", image, []string{"<clipboard image>"}}, {"files", foreignFiles(t, []string{b, a}), []string{b, a}}, {"text", []byte(a + "\n" + b), []string{a, b}}} {
		t.Run(tc.kind, func(t *testing.T) {
			independentWrite(t, tc.kind, tc.data)
			root := t.TempDir()
			out := nativeCommand(t, nil, app, "--home", filepath.Join(root, "home"), "--no-project", "compress", "--clipboard", "--target-size", "1MiB", "--strip", "keep", "--output-dir", filepath.Join(root, "outputs"), "--json")
			var report domain.Report
			if err := json.Unmarshal(out, &report); err != nil {
				t.Fatalf("CLI result was not a clean JSON report: %v; %s", err, out)
			}
			if report.ExitCode != 0 || len(report.Results) != len(tc.inputs) {
				t.Fatalf("foreign %s CLI report: %+v", tc.kind, report)
			}
			for index, result := range report.Results {
				if result.Input != tc.inputs[index] {
					t.Fatalf("CLI source order: %q, want %q", result.Input, tc.inputs[index])
				}
				data, err := os.ReadFile(result.Output)
				if err != nil {
					t.Fatal(err)
				}
				equalClipboardPixels(t, image, data)
			}
			t.Logf("foreign %s -> real CLI -> %d committed PNG artifact(s)", tc.kind, len(tc.inputs))
		})
	}
}

func TestNativeClipboardCLIOwnershipEndsOnReplacement(t *testing.T) {
	backend := nativeBackend(t)
	app := nativeApplication(t)
	source, image := clipboardFixture(t, "owner ü.png")
	root := t.TempDir()
	if backend == "macos" {
		out := nativeCommand(t, nil, app, "--home", filepath.Join(root, "home"), "--no-project", "compress", source, "--target-size", "1MiB", "--strip", "keep", "--output-dir", filepath.Join(root, "outputs"), "--copy", "smart", "--json")
		var report domain.Report
		if err := json.Unmarshal(out, &report); err != nil || report.ExitCode != 0 || len(report.Results) != 1 {
			t.Fatalf("CLI copy report: %v; %+v", err, report)
		}
		equalClipboardPixels(t, image, independentRead(t, "image"))
		if actual := independentFiles(t); !reflect.DeepEqual(actual, []string{report.Results[0].Output}) {
			t.Fatal("native file-list did not persist after CLI exit", actual)
		}
		t.Log("macOS image and native file URLs persist after the CLI writer exits")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, app, "--home", filepath.Join(root, "home"), "--no-project", "compress", source, "--target-size", "1MiB", "--strip", "keep", "--output-dir", filepath.Join(root, "outputs"), "--copy", "smart", "--clipboard-hold", "1m", "--json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	waitStarted := false
	defer func() {
		cancel()
		if waitStarted {
			<-done
		} else {
			_ = cmd.Wait()
		}
	}()
	var report domain.Report
	if err := json.NewDecoder(stdout).Decode(&report); err != nil || report.ExitCode != 0 || len(report.Results) != 1 {
		t.Fatalf("CLI did not checkpoint a successful output before ownership hold: %v; %+v", err, report)
	}
	waitStarted = true
	go func() { done <- cmd.Wait() }()
	equalClipboardPixels(t, image, independentRead(t, "image"))
	if actual := independentFiles(t); !reflect.DeepEqual(actual, []string{report.Results[0].Output}) {
		t.Fatal("CLI copied the wrong saved artifact", actual)
	}
	independentWrite(t, "text", []byte("replacement from independent native owner"))
	select {
	case err := <-done:
		// Keep one completion value for the deferred owned-process cleanup.
		done <- err
		if err != nil {
			t.Fatalf("CLI ownership exit: %v; %s", err, stderr.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("CLI did not finish after clipboard replacement")
	}
	t.Logf("%s CLI served smart data while alive, then exited successfully after replacement", backend)
}
