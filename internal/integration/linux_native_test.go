//go:build linux && nativedesktop

// SPDX-License-Identifier: MIT
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/state"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
	"github.com/godbus/dbus/v5"
)

// This opt-in consumer journey belongs only on the disposable hosted display.
// It deliberately does not join the ordinary workstation test command.
type linuxJourney struct {
	t                         *testing.T
	root, evidence, app, home string
	sequence                  int
	a11y                      *dbus.Conn
	completed                 time.Time
}

func (j *linuxJourney) write(name string, value any) {
	j.t.Helper()
	b, e := json.MarshalIndent(value, "", "  ")
	if e != nil {
		j.t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(j.evidence, name), append(b, '\n'), 0600); e != nil {
		j.t.Fatal(e)
	}
}

func (j *linuxJourney) command(program string, args ...string) []byte {
	j.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, program, args...)
	out, err := cmd.CombinedOutput()
	j.sequence++
	j.write(fmt.Sprintf("command-%02d.json", j.sequence), map[string]any{"program": program, "args": args, "output": string(out), "error": fmt.Sprint(err)})
	if err != nil {
		j.t.Fatalf("%s %q: %v\n%s", program, args, err, out)
	}
	return out
}

func (j *linuxJourney) cli(args ...string) []byte {
	j.t.Helper()
	return j.command(j.app, append([]string{"--home", j.home, "--no-project"}, args...)...)
}

func (j *linuxJourney) wait(description string, test func() bool) {
	j.t.Helper()
	until := time.Now().Add(25 * time.Second)
	for time.Now().Before(until) {
		if test() {
			return
		}
		time.Sleep(150 * time.Millisecond)
	}
	j.t.Fatalf("timed out: %s", description)
}

func (j *linuxJourney) screen(name string) {
	j.t.Helper()
	j.command("scrot", "--overwrite", filepath.Join(j.evidence, name+".png"))
}

func (j *linuxJourney) start(program string, args ...string) {
	j.t.Helper()
	f, e := os.Create(filepath.Join(j.evidence, fmt.Sprintf("process-%s-%d.log", filepath.Base(program), j.sequence)))
	if e != nil {
		j.t.Fatal(e)
	}
	cmd := exec.Command(program, args...)
	cmd.Stdout, cmd.Stderr = f, f
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if e = cmd.Start(); e != nil {
		f.Close()
		j.t.Fatal(e)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait(); f.Close() }()
	j.t.Cleanup(func() {
		select {
		case <-done:
			return
		default:
		}
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			<-done
		}
	})
}

type accessibleRef struct {
	Bus  string
	Path dbus.ObjectPath
}
type accessibleNode struct {
	Ref        accessibleRef
	Name, Role string
}

func (j *linuxJourney) connectAccessibility() {
	j.t.Helper()
	session, e := dbus.ConnectSessionBus()
	if e != nil {
		j.t.Fatal(e)
	}
	defer session.Close()
	var address string
	if e = session.Object("org.a11y.Bus", "/org/a11y/bus").Call("org.a11y.Bus.GetAddress", 0).Store(&address); e != nil {
		j.t.Fatal(e)
	}
	j.a11y, e = dbus.Dial(address)
	if e != nil {
		j.t.Fatal(e)
	}
	j.t.Cleanup(func() { j.a11y.Close() })
	if e = j.a11y.Auth(nil); e != nil {
		j.t.Fatal(e)
	}
	if e = j.a11y.Hello(); e != nil {
		j.t.Fatal(e)
	}
}

// AT-SPI is the desktop's existing accessibility protocol. The fixture uses the
// already-owned D-Bus library, rather than adding another GUI automation runtime.
func (j *linuxJourney) tree() []accessibleNode {
	queue := []accessibleRef{{"org.a11y.atspi.Registry", "/org/a11y/atspi/accessible/root"}}
	seen := map[accessibleRef]bool{}
	nodes := []accessibleNode{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for len(queue) > 0 && len(seen) < 1500 && ctx.Err() == nil {
		r := queue[0]
		queue = queue[1:]
		if seen[r] || r.Bus == "" || r.Path == "/org/a11y/atspi/null" {
			continue
		}
		seen[r] = true
		o := j.a11y.Object(r.Bus, r.Path)
		var name dbus.Variant
		var role string
		if e := o.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, "org.a11y.atspi.Accessible", "Name").Store(&name); e != nil {
			continue
		}
		_ = o.CallWithContext(ctx, "org.a11y.atspi.Accessible.GetRoleName", 0).Store(&role)
		nodes = append(nodes, accessibleNode{r, fmt.Sprint(name.Value()), role})
		var children []accessibleRef
		if o.CallWithContext(ctx, "org.a11y.atspi.Accessible.GetChildren", 0).Store(&children) == nil {
			queue = append(queue, children...)
		}
	}
	return nodes
}

func (j *linuxJourney) action(label string) {
	j.t.Helper()
	var found accessibleNode
	j.wait("accessible action "+label, func() bool {
		nodes := j.tree()
		j.write("accessibility-current.json", nodes)
		for _, n := range nodes {
			if strings.ReplaceAll(n.Name, "&", "") == label && strings.Contains(n.Role, "menu") {
				found = n
				return true
			}
		}
		return false
	})
	var accepted bool
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if e := j.a11y.Object(found.Ref.Bus, found.Ref.Path).CallWithContext(ctx, "org.a11y.atspi.Action.DoAction", 0, int32(0)).Store(&accepted); e != nil || !accepted {
		j.t.Fatalf("activate %q: accepted %v, error %v", label, accepted, e)
	}
}

func (j *linuxJourney) last(inputs []string) {
	j.t.Helper()
	var last state.Last
	j.wait("last result for selected original files", func() bool {
		b, e := os.ReadFile(filepath.Join(j.home, "state", "last-result.json"))
		if e != nil || json.Unmarshal(b, &last) != nil || !last.Completed.After(j.completed) || len(last.Report.Results) != len(inputs) {
			return false
		}
		for i, r := range last.Report.Results {
			if r.Input != inputs[i] || r.Output == "" {
				return false
			}
		}
		return true
	})
	j.completed = last.Completed
	for _, r := range last.Report.Results {
		b, e := os.ReadFile(r.Output)
		if e != nil {
			j.t.Fatal(e)
		}
		if _, _, e = image.Decode(bytes.NewReader(b)); e != nil {
			j.t.Fatalf("decode result: %v", e)
		}
	}
	j.write(fmt.Sprintf("last-%02d.json", j.sequence), last)
	j.cli("last", "--json")
}

func TestNativeLinuxDesktop(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "true" || os.Getenv("SMOLPNG_DISPOSABLE_LINUX") != "yes" || os.Getenv("DISPLAY") == "" {
		t.Fatal("requires explicitly selected disposable hosted X11 desktop")
	}
	evidence, source := os.Getenv("SMOLPNG_NATIVE_EVIDENCE"), os.Getenv("SMOLPNG_NATIVE_APP")
	if !filepath.IsAbs(evidence) || !filepath.IsAbs(source) {
		t.Fatal("absolute evidence and source-bound app paths required")
	}
	root, e := os.MkdirTemp("", "smolpng-linux-desktop-")
	if e != nil {
		t.Fatal(e)
	}
	j := &linuxJourney{t: t, root: root, evidence: evidence, home: filepath.Join(root, "config ' ü $()")}
	j.write("report.json", map[string]any{"complete": false, "step": "preflight"})
	t.Cleanup(func() {
		if t.Failed() {
			j.screen("failure")
			if j.a11y != nil {
				j.write("accessibility-failure.json", j.tree())
			}
		}
	})
	appDir := filepath.Join(root, "application ' ü $(touch injected)")
	if e = os.MkdirAll(appDir, 0700); e != nil {
		t.Fatal(e)
	}
	j.app = filepath.Join(appDir, "smolpng")
	b, e := os.ReadFile(source)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(j.app, b, 0700); e != nil {
		t.Fatal(e)
	}
	j.cli("version", "--json")
	for _, setting := range [][2]string{{"clipboard.single", "none"}, {"clipboard.batch", "none"}, {"drop.notify", "false"}} {
		j.cli("config", "set", setting[0], setting[1])
	}
	if e = os.MkdirAll(filepath.Join(j.home, "state"), 0700); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(j.home, "state", "setup.json"), []byte("{\"schemaVersion\":1,\"acknowledged\":true}\n"), 0600); e != nil {
		t.Fatal(e)
	}
	j.cli("integrate", "install", "--desktop", "--apply", "--json")
	j.cli("integrate", "install", "--desktop", "--apply", "--json")
	var installed struct {
		Locations Locations `json:"locations"`
	}
	if e = json.Unmarshal(j.cli("integrate", "status", "--json"), &installed); e != nil {
		t.Fatal(e)
	}
	j.cli("integrate", "shell", "install", "--mode", "single", "--apply", "--json")
	j.cli("integrate", "shell", "status", "--json")
	j.connectAccessibility()
	t.Cleanup(func() {
		if t.Failed() {
			j.write("accessibility-before-close.json", j.tree())
		}
	})

	corpus := testutil.Corpus()
	inputDir := filepath.Join(root, "images ' ü $(touch injected)")
	if e = os.Mkdir(inputDir, 0700); e != nil {
		t.Fatal(e)
	}
	inputs := []string{filepath.Join(inputDir, "first ' ü $(touch injected).png"), filepath.Join(inputDir, "second \" & ü.jpg")}
	originals := [][]byte{corpus["flat-icon.png"], testutil.JPEGCorpus()["jpeg-baseline.jpg"]}
	for i, p := range inputs {
		if _, _, e = image.Decode(bytes.NewReader(originals[i])); e != nil {
			t.Fatalf("invalid generated fixture %s: %v", p, e)
		}
		if e = os.WriteFile(p, originals[i], 0600); e != nil {
			t.Fatal(e)
		}
	}

	// GLib is an independent native .desktop consumer, including its %F parser.
	j.command("gio", "launch", filepath.Join(installed.Locations.Menu, "smolpng-auto.desktop"), inputs[0])
	j.last(inputs[:1])
	j.command("gio", "launch", filepath.Join(installed.Locations.Desktop, "smolpng-desktop-auto.desktop"), inputs[0], inputs[1])
	j.last(inputs)
	j.start("dolphin", "--new-window", "--select", inputs[0])
	var window string
	j.wait("visible Dolphin", func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		out, e := exec.CommandContext(ctx, "xdotool", "search", "--onlyvisible", "--class", "dolphin").Output()
		if e != nil {
			return false
		}
		window = strings.Fields(string(out))[0]
		return true
	})
	j.command("xdotool", "windowactivate", "--sync", window)
	j.command("xdotool", "key", "--clearmodifiers", "Shift+F10")
	j.screen("dolphin-single-menu")
	j.action("Make smol")
	j.last(inputs[:1])
	j.screen("dolphin-single-result")
	for i, p := range inputs {
		b, e := os.ReadFile(p)
		if e != nil || !bytes.Equal(b, originals[i]) {
			t.Fatalf("original changed: %s, %v", p, e)
		}
	}
	if _, e = os.Stat(filepath.Join(os.Getenv("HOME"), "injected")); !os.IsNotExist(e) {
		t.Fatal("argument text executed as a command")
	}
	j.cli("integrate", "shell", "remove", "--apply", "--json")
	j.cli("integrate", "remove", "--apply", "--json")
	j.write("report.json", map[string]any{"complete": true, "passed": true, "meaning": "Native X11 GLib desktop-entry and Dolphin single-action consumer; remaining interactive/submenu/multiple-selection journeys are not yet included."})
}
