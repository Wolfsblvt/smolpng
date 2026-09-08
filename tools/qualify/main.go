// SPDX-License-Identifier: MIT
// qualify exercises an already-built smolpng against real local optimizers.
// It does not install tools, publish data, use the clipboard or rebuild smolpng.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
	"github.com/Wolfsblvt/smolpng/internal/process"
	"github.com/Wolfsblvt/smolpng/internal/testutil"
)

// Bound by the candidate builder, independently of the program being tested.
var Commit = "unknown"

type check struct {
	Name      string          `json:"name"`
	Arguments []string        `json:"arguments"`
	ExitCode  int             `json:"exitCode"`
	Passed    bool            `json:"passed"`
	Errors    []string        `json:"errors"`
	Report    json.RawMessage `json:"report,omitempty"`
	Stderr    string          `json:"stderr,omitempty"`
}
type evidence struct {
	SchemaVersion    int             `json:"schemaVersion"`
	ExpectedCommit   string          `json:"expectedCommit"`
	Platform         string          `json:"platform"`
	Started          time.Time       `json:"started"`
	Executable       string          `json:"executable"`
	ExecutableSHA256 string          `json:"executableSHA256"`
	Tools            json.RawMessage `json:"tools,omitempty"`
	Checks           []check         `json:"checks"`
	Passed           bool            `json:"passed"`
	Qualification    string          `json:"qualification"`
}
type options struct {
	Executable, Directory, PNGQuant, OxiPNG, Expected string
	Timeout                                           time.Duration
}
type suite struct {
	opt    options
	data   evidence
	common []string
	root   string
	runner process.Runner
}

func main() {
	var o options
	flag.StringVar(&o.Executable, "executable", "", "Path to the existing smolpng executable")
	flag.StringVar(&o.Directory, "output", "", "New directory for qualification fixtures, outputs and report")
	flag.StringVar(&o.PNGQuant, "pngquant", "", "Optional explicit pngquant executable")
	flag.StringVar(&o.OxiPNG, "oxipng", "", "Optional explicit OxiPNG executable")
	flag.StringVar(&o.Expected, "expected-commit", Commit, "Require this exact smolpng version-report commit")
	flag.DurationVar(&o.Timeout, "timeout", 2*time.Minute, "Deadline per invocation")
	flag.Parse()
	if o.Executable == "" || o.Directory == "" || o.Timeout <= 0 {
		fmt.Fprintln(os.Stderr, "Provide -executable PATH and a new -output DIRECTORY; timeout must be positive.")
		os.Exit(2)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if e := run(ctx, o); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run(ctx context.Context, o options) error {
	exe, e := filepath.Abs(o.Executable)
	if e != nil {
		return e
	}
	dir, e := filepath.Abs(o.Directory)
	if e != nil {
		return e
	}
	if st, e := os.Stat(exe); e != nil || !st.Mode().IsRegular() {
		return fmt.Errorf("application executable is not a regular file")
	}
	if e = os.MkdirAll(filepath.Dir(dir), 0755); e != nil {
		return e
	}
	if e = os.Mkdir(dir, 0700); e != nil {
		return fmt.Errorf("qualification output must be a new directory: %w", e)
	}
	o.Executable = exe
	o.Directory = dir
	s := suite{opt: o, root: dir, runner: process.OSRunner{}, common: []string{"--home", filepath.Join(dir, "home"), "--no-project"}, data: evidence{SchemaVersion: 1, ExpectedCommit: o.Expected, Platform: runtime.GOOS + "/" + runtime.GOARCH, Started: time.Now().UTC(), Executable: exe, Checks: []check{}, Qualification: "real local executable/optimizer and filesystem checks; does not qualify graphical clipboard, drag/drop, notifications, signing or remote providers"}}
	for _, p := range []struct{ name, path string }{{"--pngquant-path", o.PNGQuant}, {"--oxipng-path", o.OxiPNG}} {
		if p.path != "" {
			a, e := filepath.Abs(p.path)
			if e != nil {
				return e
			}
			s.common = append(s.common, p.name, a)
		}
	}
	binaryBytes, e := os.ReadFile(exe)
	if e != nil {
		return e
	}
	s.data.ExecutableSHA256 = fmt.Sprintf("%x", sha256.Sum256(binaryBytes))
	s.data.Passed = true
	v := s.invoke(ctx, "build-provenance", 0, nil, "version", "--json")
	var info struct {
		Commit string `json:"commit"`
	}
	if e = json.Unmarshal(v.Report, &info); e != nil {
		v.Errors = append(v.Errors, "version report is not JSON")
	} else if o.Expected != "" && o.Expected != "unknown" && info.Commit != o.Expected {
		v.Errors = append(v.Errors, "executable does not match expected source commit")
	}
	s.add(v)
	d := s.invoke(ctx, "dependency-preflight", 0, nil, "deps", "status", "--json")
	var health struct {
		Tools struct {
			PNGQuant struct {
				Available bool `json:"available"`
			} `json:"pngquant"`
			OxiPNG struct {
				Available bool `json:"available"`
			} `json:"oxipng"`
		} `json:"tools"`
	}
	if e = json.Unmarshal(d.Report, &health); e != nil || !health.Tools.PNGQuant.Available || !health.Tools.OxiPNG.Available {
		d.Errors = append(d.Errors, "both real compatible optimizers are required; install separately or supply explicit paths")
	}
	s.data.Tools = d.Report
	s.add(d)
	if !s.data.Passed {
		_ = s.save()
		return fmt.Errorf("qualification preflight failed; inspect %s", filepath.Join(dir, "report.json"))
	}
	fixtures := testutil.Corpus()
	names := []string{}
	for n := range fixtures {
		names = append(names, n)
	}
	sort.Strings(names)
	sourceDir := filepath.Join(dir, "sources")
	if e = os.Mkdir(sourceDir, 0700); e != nil {
		return e
	}
	for _, name := range names {
		if e = os.WriteFile(filepath.Join(sourceDir, name), fixtures[name], 0600); e != nil {
			return e
		}
	}
	for _, preset := range []string{"auto", "lossless"} {
		for _, name := range names {
			if ctx.Err() != nil {
				_ = s.save()
				return ctx.Err()
			}
			source := filepath.Join(sourceDir, name)
			dest := filepath.Join(dir, "outputs", preset, name)
			c := s.invoke(ctx, preset+"/"+name, 0, nil, "compress", source, "--preset", preset, "--output", dest, "--json")
			var report domain.Report
			if e = json.Unmarshal(c.Report, &report); e != nil || len(report.Results) != 1 {
				c.Errors = append(c.Errors, "missing per-image result")
			} else {
				result := report.Results[0]
				if result.Output != dest {
					c.Errors = append(c.Errors, "returned path differs from requested artifact")
				}
				output, e := os.ReadFile(dest)
				if e != nil {
					c.Errors = append(c.Errors, "output file is unavailable")
				} else {
					c.Errors = append(c.Errors, validateImage(fixtures[name], output, result, preset == "lossless")...)
				}
				now, e := os.ReadFile(source)
				if e != nil || !bytes.Equal(now, fixtures[name]) {
					c.Errors = append(c.Errors, "original file changed")
				}
			}
			s.add(c)
			if e = s.save(); e != nil {
				return e
			}
		}
	}
	source := filepath.Join(sourceDir, "flat-icon.png")
	for _, target := range []struct {
		name, size string
		code       int
		nearest    bool
	}{{"target-met", "10KiB", 0, false}, {"target-missed", "1B", 6, false}, {"target-nearest", "1B", 6, true}} {
		dest := filepath.Join(dir, "outputs", target.name+".png")
		args := []string{"compress", source, "--target-size", target.size, "--output", dest, "--json"}
		if target.nearest {
			args = append(args, "--accept-nearest")
		}
		c := s.invoke(ctx, target.name, target.code, nil, args...)
		var report domain.Report
		if e = json.Unmarshal(c.Report, &report); e != nil || len(report.Results) != 1 {
			c.Errors = append(c.Errors, "missing target result")
		} else {
			result := report.Results[0]
			if result.ExitCode != target.code || report.ExitCode != target.code || result.Target == nil || result.Target.Met != (target.code == 0) {
				c.Errors = append(c.Errors, "target status/exit mismatch")
			}
			if target.code == 6 && !target.nearest {
				if _, e := os.Stat(dest); !os.IsNotExist(e) || result.Output != "" {
					c.Errors = append(c.Errors, "unaccepted target miss produced an output")
				}
			} else {
				output, e := os.ReadFile(dest)
				if e != nil {
					c.Errors = append(c.Errors, "accepted target result lacks an artifact")
				} else {
					c.Errors = append(c.Errors, validateImage(fixtures["flat-icon.png"], output, result, false)...)
				}
			}
		}
		s.add(c)
	}
	binary := s.invoke(ctx, "binary-stdout", 0, nil, "compress", source, "--output", "-", "--quiet")
	// invoke leaves binary bytes in Report only until this assertion; they are never
	// embedded in report.json, which is evidence metadata, not another image copy.
	if output, e := pngimage.Decode(binary.Report, 1_000_000); e != nil || output.Info.Dimensions != (domain.Dimensions{Width: 64, Height: 64}) {
		binary.Errors = append(binary.Errors, "stdout is not a clean PNG with expected dimensions")
	}
	binary.Report = nil
	s.add(binary)
	partial := s.invoke(ctx, "partial-batch", 8, nil, "compress", source, filepath.Join(sourceDir, "missing.png"), "--output-dir", filepath.Join(dir, "partial"), "--json")
	var report domain.Report
	if e = json.Unmarshal(partial.Report, &report); e != nil || len(report.Results) != 2 || report.ExitCode != 8 || report.Results[0].Output == "" || report.Results[1].ExitCode != 3 {
		partial.Errors = append(partial.Errors, "partial batch did not preserve successful output and failed item")
	}
	s.add(partial)
	for name, original := range fixtures {
		current, e := os.ReadFile(filepath.Join(sourceDir, name))
		if e != nil || !bytes.Equal(current, original) {
			s.add(check{Name: "final-original/" + name, Errors: []string{"source bytes changed"}})
		}
	}
	if e = s.save(); e != nil {
		return e
	}
	if !s.data.Passed {
		return fmt.Errorf("qualification found failures; see %s", filepath.Join(dir, "report.json"))
	}
	fmt.Printf("Passed %d local checks; evidence: %s\n", len(s.data.Checks), filepath.Join(dir, "report.json"))
	return nil
}
func (s *suite) invoke(ctx context.Context, name string, want int, stdin []byte, args ...string) check {
	if len(args) > 0 && args[0] == "compress" {
		args = append(args, "--offline", "--remote-policy", "never")
	}
	c := check{Name: name, Arguments: append(append([]string{}, s.common...), args...), Errors: []string{}}
	bounded, cancel := context.WithTimeout(ctx, s.opt.Timeout)
	defer cancel()
	result, e := s.runner.Run(bounded, process.Invocation{Executable: s.opt.Executable, Args: c.Arguments, Stdin: bytes.NewReader(stdin)})
	c.ExitCode = result.ExitCode
	c.Stderr = result.Stderr
	c.Report = json.RawMessage(result.Stdout)
	if e != nil {
		c.Errors = append(c.Errors, "command could not complete: "+e.Error())
	}
	if result.ExitCode != want {
		c.Errors = append(c.Errors, fmt.Sprintf("expected exit %d, got %d", want, result.ExitCode))
	}
	return c
}
func (s *suite) add(c check) {
	c.Passed = len(c.Errors) == 0
	s.data.Checks = append(s.data.Checks, c)
	s.data.Passed = s.data.Passed && c.Passed
	fmt.Printf("%s: %t\n", c.Name, c.Passed)
}
func (s *suite) save() error {
	// Unexpected non-JSON stdout is evidence of a failed contract, not a reason
	// to lose every earlier passing result by making the manifest unmarshalable.
	for i := range s.data.Checks {
		c := &s.data.Checks[i]
		if len(c.Report) > 0 && !json.Valid(c.Report) {
			c.Errors = append(c.Errors, "non-JSON output omitted from manifest")
			c.Report = nil
			c.Passed = false
			s.data.Passed = false
		}
	}
	if len(s.data.Tools) > 0 && !json.Valid(s.data.Tools) {
		s.data.Tools = nil
	}
	b, e := json.MarshalIndent(s.data, "", "  ")
	if e != nil {
		return e
	}
	return fsutil.AtomicWrite(context.Background(), filepath.Join(s.root, "report.json"), append(b, '\n'), 0600, true)
}
func validateImage(original, output []byte, r domain.Result, strict bool) []string {
	failures := []string{}
	a, e := pngimage.Decode(original, 1_000_000)
	if e != nil {
		return []string{"invalid original specimen"}
	}
	b, e := pngimage.Decode(output, 1_000_000)
	if e != nil {
		return []string{"optimizer output is not a valid static PNG"}
	}
	if b.Info.Animated || b.Info.Dimensions != a.Info.Dimensions || r.Dimensions.Before != a.Info.Dimensions || r.Dimensions.After != b.Info.Dimensions {
		failures = append(failures, "dimensions/animation differ from the promised source")
	}
	if a.Info.Transparency && !b.Info.Transparency {
		failures = append(failures, "transparency was discarded")
	}
	if strict && !pngimage.PixelEqual(a.Image, b.Image) {
		failures = append(failures, "strict lossless changed decoded pixels, including hidden RGB")
	}
	if len(output) > len(original) {
		failures = append(failures, "normal recipe emitted a larger file")
	}
	if r.OriginalBytes != int64(len(original)) || r.OutputBytes != int64(len(output)) {
		failures = append(failures, "reported byte counts differ from artifact")
	}
	if r.Target != nil && r.Target.Met != (int64(len(output)) <= r.Target.RequestedBytes) {
		failures = append(failures, "reported strict target does not match artifact bytes")
	}
	if r.RemoteUsed {
		failures = append(failures, "local recipe reports remote use")
	}
	return failures
}
