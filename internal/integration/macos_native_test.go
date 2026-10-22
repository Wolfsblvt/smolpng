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
	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		// Capture the actual failed native surface before the disposable VM ends.
		// No privacy settings or permission grants are changed by the journey.
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, "/bin/ps", "-axo", "pid=,ppid=,tty=,command=").CombinedOutput()
		owned := []string{fmt.Sprint(err)}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, root) || strings.Contains(line, "/Terminal.app/Contents/MacOS/Terminal") {
				owned = append(owned, line)
			}
			if strings.Contains(line, root) && strings.Contains(line, ".app/Contents/MacOS/droplet") {
				pid := strings.Fields(line)[0]
				sample, err := exec.CommandContext(ctx, "/usr/bin/sample", pid, "1", "1").CombinedOutput()
				_ = os.WriteFile(filepath.Join(evidence, "diagnostic-sample-"+pid+".txt"), []byte(fmt.Sprintf("%v\n%s", err, sample)), 0600)
			}
		}
		_ = os.WriteFile(filepath.Join(evidence, "diagnostic-processes.txt"), []byte(strings.Join(owned, "\n")), 0600)
		out, err = exec.CommandContext(ctx, "/usr/sbin/screencapture", "-x", filepath.Join(evidence, "diagnostic-screen.png")).CombinedOutput()
		_ = os.WriteFile(filepath.Join(evidence, "diagnostic-screen.txt"), []byte(fmt.Sprintf("%v\n%s", err, out)), 0600)
	})
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
			// Native aliases may return decomposed Unicode on macOS. Compare
			// actual file identity, not an assumed normalization of its name.
			matched := false
			actual, statErr := os.Stat(result.Input)
			for _, path := range paths {
				wanted, err := os.Stat(path)
				matched = matched || statErr == nil && err == nil && os.SameFile(actual, wanted)
			}
			if !matched || result.Output == "" {
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

const nativeTerminalDriver = `on run argv
	set sourcePrefix to item 1 of argv
	tell application "Terminal"
		set fixtureTab to missing value
		repeat 80 times
			repeat with w in windows
				repeat with tb in tabs of w
					if (contents of tb) contains sourcePrefix then set fixtureTab to tb
				end repeat
			end repeat
			if fixtureTab is not missing value then exit repeat
			delay 0.25
		end repeat
		if fixtureTab is missing value then error "fixture source tab missing"
		set fixtureTTY to tty of fixtureTab
		set beforeText to contents of fixtureTab
		if beforeText does not contain "Choose a recipe" then error "recipe screen missing"
		do script "" in fixtureTab
		repeat 240 times
			set resultText to contents of fixtureTab
			if resultText contains "Enter: done" then exit repeat
			delay 0.25
		end repeat
		if resultText does not contain "Enter: done" then error "result screen missing"
		do script "" in fixtureTab
		set finished to false
		repeat 40 times
			try
				if not (busy of fixtureTab) then
					close fixtureTab
					set finished to true
					exit repeat
				end if
			on error
				-- Terminal may close a successfully exited .command automatically.
				set finished to true
				exit repeat
			end try
			delay 0.25
		end repeat
		if not finished then error "fixture remains busy"
		return fixtureTTY & linefeed & beforeText & linefeed & resultText
	end tell
end run
`
