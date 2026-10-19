// SPDX-License-Identifier: MIT
// qualify exercises an already-built smolpng against real local optimizers.
// It does not install tools, publish data, use the clipboard or rebuild smolpng.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/domain"
	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"github.com/Wolfsblvt/smolpng/internal/jpegimage"
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
	SchemaVersion    int               `json:"schemaVersion"`
	QualifierCommit  string            `json:"qualifierCommit"`
	ExpectedCommit   string            `json:"expectedCommit"`
	Platform         string            `json:"platform"`
	Host             hostIdentity      `json:"host"`
	Started          time.Time         `json:"started"`
	Executable       string            `json:"executable"`
	ExecutableSHA256 string            `json:"executableSHA256"`
	Tools            json.RawMessage   `json:"tools,omitempty"`
	OptimizerSHA256  map[string]string `json:"optimizerSHA256,omitempty"`
	Checks           []check           `json:"checks"`
	Formats          []string          `json:"formats"`
	Complete         bool              `json:"complete"`
	Passed           bool              `json:"passed"`
	Qualification    string            `json:"qualification"`
}
type options struct {
	Executable, Directory, PNGQuant, OxiPNG, JPEGOptim, Expected, Formats string
	Timeout                                                               time.Duration
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
	flag.StringVar(&o.JPEGOptim, "jpegoptim", "", "Optional explicit jpegoptim executable")
	flag.StringVar(&o.Formats, "formats", "png,jpeg", "Formats to exercise: png, jpeg, or png,jpeg")
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
func run(ctx context.Context, o options) (err error) {
	formats, e := parseFormats(o.Formats)
	if e != nil {
		return e
	}
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
	s.data.Formats = formats
	s.data.QualifierCommit = Commit
	s.data.Host = currentHostIdentity()
	defer func() {
		if !s.data.Complete {
			err = errors.Join(err, s.save())
		}
	}()
	for _, p := range []struct{ name, path string }{{"--pngquant-path", o.PNGQuant}, {"--oxipng-path", o.OxiPNG}, {"--jpegoptim-path", o.JPEGOptim}} {
		if p.path != "" {
			a, e := filepath.Abs(p.path)
			if e != nil {
				return e
			}
			s.common = append(s.common, p.name, a)
		}
	}
	s.data.ExecutableSHA256, e = executableHash(exe)
	if e != nil {
		return e
	}
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
	d.Errors = append(d.Errors, dependencyErrors(d.Report, formats)...)
	s.data.Tools = d.Report
	s.add(d)
	if s.data.Passed {
		s.data.OptimizerSHA256, e = optimizerHashes(d.Report, formats)
		identity := check{Name: "optimizer-provenance"}
		if e != nil {
			identity.Errors = append(identity.Errors, e.Error())
		}
		s.add(identity)
	}
	// Persist identities before corpus work; an interrupted run must not lose them.
	if e = s.save(); e != nil {
		return e
	}
	if !s.data.Passed {
		return fmt.Errorf("qualification preflight failed; inspect %s", filepath.Join(dir, "report.json"))
	}
	fixtures := map[string][]byte{}
	samples := []string{}
	for _, format := range formats {
		var corpus map[string][]byte
		var sample string
		if format == "png" {
			corpus = testutil.Corpus()
			sample = "flat-icon.png"
		} else {
			corpus = testutil.JPEGCorpus()
			sample = "jpeg-baseline.jpg"
		}
		for name, data := range corpus {
			fixtures[name] = data
		}
		samples = append(samples, sample)
	}
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
	for _, sample := range samples {
		s.exerciseOutputs(ctx, sourceDir, sample, fixtures)
		if e = s.save(); e != nil {
			return e
		}
	}
	if len(samples) > 1 {
		s.exerciseMixedBatch(ctx, sourceDir, samples, fixtures)
	}
	for name, original := range fixtures {
		current, e := os.ReadFile(filepath.Join(sourceDir, name))
		if e != nil || !bytes.Equal(current, original) {
			s.add(check{Name: "final-original/" + name, Errors: []string{"source bytes changed"}})
		}
	}
	if e = ctx.Err(); e != nil {
		return e
	}
	s.verifyOptimizerIdentity()
	s.data.Complete = true
	if e = s.save(); e != nil {
		return e
	}
	if !s.data.Passed {
		return fmt.Errorf("qualification found failures; see %s", filepath.Join(dir, "report.json"))
	}
	fmt.Printf("Passed %d local checks; evidence: %s\n", len(s.data.Checks), filepath.Join(dir, "report.json"))
	return nil
}

func (s *suite) exerciseOutputs(ctx context.Context, sourceDir, name string, fixtures map[string][]byte) {
	dir := s.root
	source := filepath.Join(sourceDir, name)
	ext := filepath.Ext(name)
	prefix := ""
	if ext == ".jpg" {
		prefix = "jpeg/"
	}
	for _, target := range []struct {
		name, size string
		bytes      int64
		code       int
		nearest    bool
	}{{"target-met", "10KiB", 10 << 10, 0, false}, {"target-missed", "1B", 1, 6, false}, {"target-nearest", "1B", 1, 6, true}} {
		dest := filepath.Join(dir, "outputs", prefix+target.name+ext)
		args := []string{"compress", source, "--target-size", target.size, "--output", dest, "--json"}
		if target.nearest {
			args = append(args, "--accept-nearest")
		}
		c := s.invoke(ctx, prefix+target.name, target.code, nil, args...)
		var report domain.Report
		if e := json.Unmarshal(c.Report, &report); e != nil || len(report.Results) != 1 {
			c.Errors = append(c.Errors, "missing target result")
		} else {
			result := report.Results[0]
			if result.ExitCode != target.code || report.ExitCode != target.code || result.Target == nil || result.Target.RequestedBytes != target.bytes || result.Target.Met != (target.code == 0) {
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
					c.Errors = append(c.Errors, validateImage(fixtures[name], output, result, false)...)
				}
			}
		}
		s.add(c)
	}
	binary := s.invoke(ctx, prefix+"binary-stdout", 0, nil, "compress", source, "--output", "-", "--quiet")
	// invoke leaves binary bytes in Report only until this assertion; they are never
	// embedded in report.json, which is evidence metadata, not another image copy.
	_, original, sourceErr := decodeImage(fixtures[name])
	_, output, outputErr := decodeImage(binary.Report)
	if sourceErr != nil || outputErr != nil || output.Format != original.Format || output.Dimensions != original.Dimensions {
		binary.Errors = append(binary.Errors, "stdout is not a clean image with the expected format and dimensions")
	}
	binary.Report = nil
	s.add(binary)
	partial := s.invoke(ctx, prefix+"partial-batch", 8, nil, "compress", source, filepath.Join(sourceDir, "missing"+ext), "--output-dir", filepath.Join(dir, "partial", prefix), "--json")
	var report domain.Report
	if e := json.Unmarshal(partial.Report, &report); e != nil || len(report.Results) != 2 || report.ExitCode != 8 || report.Results[0].Output == "" || report.Results[1].ExitCode != 3 {
		partial.Errors = append(partial.Errors, "partial batch did not preserve successful output and failed item")
	} else if data, err := os.ReadFile(report.Results[0].Output); err != nil {
		partial.Errors = append(partial.Errors, "successful partial item lacks its artifact")
	} else {
		partial.Errors = append(partial.Errors, validateImage(fixtures[name], data, report.Results[0], false)...)
	}
	s.add(partial)
}

func (s *suite) exerciseMixedBatch(ctx context.Context, sourceDir string, names []string, fixtures map[string][]byte) {
	dir := filepath.Join(s.root, "mixed")
	args := []string{"compress"}
	pending := map[string][]byte{}
	for _, name := range names {
		source := filepath.Join(sourceDir, name)
		args = append(args, source)
		pending[source] = fixtures[name]
	}
	args = append(args, "--output-dir", dir, "--json")
	c := s.invoke(ctx, "mixed-batch", 0, nil, args...)
	var report domain.Report
	if err := json.Unmarshal(c.Report, &report); err != nil || len(report.Results) != len(names) || report.ExitCode != 0 {
		c.Errors = append(c.Errors, "mixed batch lacks all successful results")
	} else {
		for _, result := range report.Results {
			original, expected := pending[result.Input]
			if !expected {
				c.Errors = append(c.Errors, "mixed batch returned an unexpected or duplicate input")
				continue
			}
			delete(pending, result.Input)
			data, err := os.ReadFile(result.Output)
			if err != nil || filepath.Dir(result.Output) != dir || result.ExitCode != 0 {
				c.Errors = append(c.Errors, "mixed batch lacks a successful artifact in its output directory")
				continue
			}
			c.Errors = append(c.Errors, validateImage(original, data, result, false)...)
		}
		if len(pending) != 0 {
			c.Errors = append(c.Errors, "mixed batch omitted a requested input")
		}
	}
	s.add(c)
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
	snapshot := s.data
	snapshot.Passed = snapshot.Complete && snapshot.Passed
	b, e := json.MarshalIndent(snapshot, "", "  ")
	if e != nil {
		return e
	}
	return fsutil.AtomicWrite(context.Background(), filepath.Join(s.root, "report.json"), append(b, '\n'), 0600, true)
}
func validateImage(original, output []byte, r domain.Result, strict bool) []string {
	failures := []string{}
	a, before, e := decodeImage(original)
	if e != nil {
		return []string{"invalid original specimen"}
	}
	b, after, e := decodeImage(output)
	if e != nil {
		return []string{"optimizer output is not a valid supported image"}
	}
	if after.Format != before.Format || r.Format != before.Format {
		failures = append(failures, "artifact/report format differs from source")
	}
	if after.Animated || after.Dimensions != before.Dimensions || r.Dimensions.Before != before.Dimensions || r.Dimensions.After != after.Dimensions {
		failures = append(failures, "dimensions/animation differ from the promised source")
	}
	if before.Orientation != after.Orientation {
		failures = append(failures, "orientation changed without a requested geometry change")
	}
	if before.Transparency && !after.Transparency {
		failures = append(failures, "transparency was discarded")
	}
	if strict && !pngimage.PixelEqual(a, b) {
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

func decodeImage(data []byte) (image.Image, domain.ImageInfo, error) {
	if jpegimage.Is(data) {
		a, err := jpegimage.Decode(data, 1_000_000)
		if err != nil {
			return nil, domain.ImageInfo{}, err
		}
		return a.Image, a.Info, nil
	}
	a, err := pngimage.Decode(data, 1_000_000)
	if err != nil {
		return nil, domain.ImageInfo{}, err
	}
	return a.Image, a.Info, nil
}
