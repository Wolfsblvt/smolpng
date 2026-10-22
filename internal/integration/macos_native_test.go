//go:build darwin && nativefinder

// SPDX-License-Identifier: MIT
package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/state"
)

// This opt-in journey opens native applications only on an explicitly disposable
// hosted Mac. The ordinary go test ./... surface never includes it.
func TestNativeFinder(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "true" || os.Getenv("SMOLPNG_DISPOSABLE_MAC") != "yes" {
		t.Fatal("native Finder journey requires an explicitly disposable hosted Mac")
	}
	evidence := os.Getenv("EVIDENCE")
	app := os.Getenv("SMOLPNG_NATIVE_APP")
	if !filepath.IsAbs(evidence) || !filepath.IsAbs(app) {
		t.Fatal("absolute evidence and application paths required")
	}
	root := filepath.Join(evidence, "journey")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	user := filepath.Join(root, "user ' ä")
	home := filepath.Join(root, "data ' ä")
	for _, dir := range []string{user, filepath.Join(home, "state")} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HOME", user)
	t.Setenv("TINIFY_API_KEY", "")
	t.Setenv("SMOLPNG_HOME", "")
	t.Setenv("SMOLPNG_CONFIG", "")
	// The application path itself exercises shell-sensitive argument boundaries.
	appBytes, err := os.ReadFile(app)
	if err != nil {
		t.Fatal(err)
	}
	app = filepath.Join(root, "smolpng ' ä")
	if err := os.WriteFile(app, appBytes, 0755); err != nil {
		t.Fatal(err)
	}
	t.Logf("application SHA-256: %x", sha256.Sum256(appBytes))
	config := fmt.Sprintf("config_version=1\n[defaults]\nremote_policy='never'\n[drop]\nnotify=false\ncopy='none'\n[clipboard]\nsingle='none'\nbatch='none'\n[ui]\nheader='none'\nanimations=false\n[tools]\npngquant=%s\noxipng=%s\njpegoptim=%s\n", strconv.Quote(os.Getenv("PNGQUANT")), strconv.Quote(os.Getenv("OXIPNG")), strconv.Quote(os.Getenv("JPEGOPTIM")))
	configPath := filepath.Join(home, "config.toml")
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	if err := state.Write(context.Background(), filepath.Join(home, "state", "setup.json"), state.Setup{SchemaVersion: 1, Acknowledged: true}); err != nil {
		t.Fatal(err)
	}
	sequence := 0
	run := func(program string, args ...string) ([]byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, program, args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		sequence++
		payload, _ := json.MarshalIndent(map[string]any{"program": program, "arguments": args, "output": string(out), "error": fmt.Sprint(err)}, "", "  ")
		if writeErr := os.WriteFile(filepath.Join(root, fmt.Sprintf("command-%02d.json", sequence)), payload, 0600); writeErr != nil {
			t.Fatal(writeErr)
		}
		t.Logf("command %02d %s: %v", sequence, filepath.Base(program), err)
		return out, err
	}
	cli := func(args ...string) []byte {
		out, err := run(app, append([]string{"--home", home, "--no-project"}, args...)...)
		if err != nil {
			t.Fatalf("CLI %v: %v\n%s", args, err, out)
		}
		return out
	}
	cli("config", "validate", "--json")
	var installed struct {
		Entries []Entry `json:"entries"`
	}
	if err := json.Unmarshal(cli("integrate", "install", "--desktop", "--apply", "--json"), &installed); err != nil || len(installed.Entries) != 4 {
		t.Fatalf("four native launchers: %v %+v", err, installed)
	}
	cli("integrate", "install", "--desktop", "--apply", "--json")
	status := func(want string) {
		var result struct {
			Entries []Status `json:"entries"`
			Pending bool     `json:"recoveryPending"`
		}
		if err := json.Unmarshal(cli("integrate", "status", "--json"), &result); err != nil || result.Pending || len(result.Entries) != 4 {
			t.Fatalf("native status: %v %+v", err, result)
		}
		for _, entry := range result.Entries {
			if entry.State != want {
				t.Fatalf("%s: want %s, got %s", entry.Entry.ID, want, entry.State)
			}
		}
	}
	status("current")
	findEntry := func(id string) Entry {
		for _, entry := range installed.Entries {
			if entry.ID == id {
				return entry
			}
		}
		t.Fatalf("entry %s missing", id)
		return Entry{}
	}
	inputs := []string{}
	originals := map[string][32]byte{}
	for i, name := range []string{"one ' ; $(touch injected) ä.png", "二 --image.jpg"} {
		path := filepath.Join(root, name)
		img := image.NewNRGBA(image.Rect(0, 0, 48, 32))
		for y := 0; y < 32; y++ {
			for x := 0; x < 48; x++ {
				img.SetNRGBA(x, y, color.NRGBA{uint8(x * 5), uint8(y * 7), 120, 255})
			}
		}
		var data bytes.Buffer
		if i == 0 {
			err = png.Encode(&data, img)
		} else {
			err = jpeg.Encode(&data, img, &jpeg.Options{Quality: 95})
		}
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, path)
		originals[path] = sha256.Sum256(data.Bytes())
	}
	lastPath := filepath.Join(home, "state", "last-result.json")
	verifyLast := func(paths []string) {
		last, err := state.ReadLast(lastPath)
		if err != nil || last.Report.ExitCode != 0 || len(last.Report.Results) != len(paths) {
			t.Fatalf("last-result: %v %+v", err, last)
		}
		for _, result := range last.Report.Results {
			if !slices.Contains(paths, result.Input) || result.Output == "" {
				t.Fatalf("file handoff: %+v", result)
			}
			out, err := os.ReadFile(result.Output)
			if err != nil || len(out) == 0 {
				t.Fatalf("output: %v", err)
			}
			var decodeErr error
			if filepath.Ext(result.Output) == ".png" {
				_, decodeErr = png.Decode(bytes.NewReader(out))
			} else {
				_, decodeErr = jpeg.Decode(bytes.NewReader(out))
			}
			if decodeErr != nil {
				t.Fatal(decodeErr)
			}
		}
		for path, before := range originals {
			data, err := os.ReadFile(path)
			if err != nil || sha256.Sum256(data) != before {
				t.Fatalf("original changed: %s %v", path, err)
			}
		}
		if _, err := os.Stat(filepath.Join(root, "injected")); !os.IsNotExist(err) {
			t.Fatal("filename shell fragment executed")
		}
		var public struct {
			Last state.Last `json:"last"`
		}
		if err := json.Unmarshal(cli("last", "--json"), &public); err != nil || public.Last.Report.ExitCode != 0 || len(public.Last.Report.Results) != len(paths) {
			t.Fatalf("public last: %v %+v", err, public)
		}
	}
	background := findEntry("auto")
	for _, files := range [][]string{inputs[:1], inputs} {
		_ = os.Remove(lastPath)
		out, err := run("/usr/bin/open", append([]string{"-W", "-n", "-a", background.Path}, files...)...)
		if err != nil {
			t.Fatalf("native file-open: %v\n%s", err, out)
		}
		verifyLast(files)
	}
	status("current") // Native execution must not mutate an installed applet.
	// A real Terminal consumer supplies a TTY. Accessible input keeps this
	// functional journey independent of terminal size or visual presentation.
	foreground := findEntry("interactive")
	macString := func(s string) string { return "\"" + strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(s) + "\"" }
	terminal := func(script string) string {
		out, err := run("/usr/bin/osascript", "-e", script)
		if err != nil {
			t.Fatalf("Terminal consumer: %v\n%s", err, out)
		}
		return string(out)
	}
	for _, entry := range []Entry{foreground, background} {
		if entry.ID == background.ID {
			cli("config", "set", "drop.mode", "interactive", "--json")
		}
		// Set the native terminal's type through its ordinary profile environment,
		// by prepending a persistent CLI flag in this owned qualification bundle.
		// Reinstall records the exact resulting launcher; no installed bytes are patched.
		entry.Arguments = append([]string{"--accessible"}, entry.Arguments...)
		if entry.ID == background.ID {
			entry.Arguments = append([]string{"drop", "--accessible"}, entry.Arguments[2:]...)
		}
		loc, _ := NativeLocations()
		manager := Manager{StatePath: RegistryName(filepath.Join(home, "state")), Version: "native-test", Allowed: loc, Renderer: Native{}}
		if err := manager.Install(context.Background(), []Entry{entry}); err != nil {
			t.Fatal(err)
		}
		_ = os.Remove(lastPath)
		out, err := run("/usr/bin/open", append([]string{"-W", "-n", "-a", entry.Path}, inputs...)...)
		if err != nil {
			t.Fatalf("foreground open: %v\n%s", err, out)
		}
		// Find only the tab displaying this fixture's exact source filename.
		query := "tell application \"Terminal\"\nrepeat with w in windows\nrepeat with tb in tabs of w\nif (contents of tb) contains " + macString(filepath.Base(inputs[0])) + " then return tty of tb\nend repeat\nend repeat\nreturn \"\"\nend tell"
		tty := ""
		deadline := time.Now().Add(20 * time.Second)
		for time.Now().Before(deadline) {
			tty = strings.TrimSpace(terminal(query))
			if tty != "" {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		if tty == "" {
			t.Fatal("actual Terminal source/TTY not observed")
		}
		contents := func() string {
			return terminal("tell application \"Terminal\"\nrepeat with w in windows\nrepeat with tb in tabs of w\nif tty of tb is " + macString(tty) + " then return contents of tb\nend repeat\nend repeat\nend tell")
		}
		if !strings.Contains(contents(), "Choose a recipe") {
			t.Fatal("foreground recipe screen not observed")
		}
		send := "tell application \"Terminal\"\nrepeat with w in windows\nrepeat with tb in tabs of w\nif tty of tb is " + macString(tty) + " then\ndo script \"\" in tb\nreturn\nend if\nend repeat\nend repeat\nend tell"
		terminal(send)
		deadline = time.Now().Add(60 * time.Second)
		for time.Now().Before(deadline) {
			if _, err := os.Stat(lastPath); err == nil {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		verifyLast(inputs)
		deadline = time.Now().Add(10 * time.Second)
		for !strings.Contains(contents(), "Enter: done") && time.Now().Before(deadline) {
			time.Sleep(250 * time.Millisecond)
		}
		if !strings.Contains(contents(), "Enter: done") {
			t.Fatal("result screen not observed")
		}
		terminal(send)
		t.Logf("actual Terminal TTY %s completed %s with original file arguments", tty, entry.ID)
		// Close exactly the owned fixture tab after the CLI has returned.
		terminal("tell application \"Terminal\"\nrepeat with w in windows\nrepeat with tb in tabs of w\nif tty of tb is " + macString(tty) + " then\nif busy of tb then error \"fixture terminal still busy\"\nclose tb\nreturn\nend if\nend repeat\nend repeat\nend tell")
	}
	status("current")
	// Preserve a user's extra file, refusing whole-plan replacement and removal.
	extra := filepath.Join(background.Path, "keep-user.txt")
	if err := os.WriteFile(extra, []byte("user content"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"integrate", "install", "--desktop", "--apply", "--json"}, {"integrate", "remove", "--apply", "--json"}} {
		out, err := run(app, append([]string{"--home", home, "--no-project"}, args...)...)
		if err == nil || !bytes.Contains(out, []byte("modified")) {
			t.Fatalf("modified bundle refusal: %v %s", err, out)
		}
	}
	data, err := os.ReadFile(extra)
	if err != nil || string(data) != "user content" {
		t.Fatal("extra user file lost")
	}
	if err := os.Remove(extra); err != nil {
		t.Fatal(err)
	}
	status("current")
	cli("integrate", "remove", "--apply", "--json")
	for _, entry := range installed.Entries {
		if _, err := os.Stat(entry.Path); !os.IsNotExist(err) {
			t.Fatalf("launcher remains: %s %v", entry.Path, err)
		}
	}
	var removed struct {
		Entries []Status `json:"entries"`
	}
	if err := json.Unmarshal(cli("integrate", "status", "--json"), &removed); err != nil || len(removed.Entries) != 0 {
		t.Fatalf("removal status: %v %+v", err, removed)
	}
	if data, err := os.ReadFile(configPath); err != nil || len(data) == 0 {
		t.Fatal("configuration lost")
	}
}
