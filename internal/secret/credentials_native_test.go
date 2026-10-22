//go:build nativekeyring

// SPDX-License-Identifier: MIT
package secret_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/process"
	"github.com/Wolfsblvt/smolpng/internal/secret"
)

// This journey owns only generated values in a disposable hosted credential store.
// The tag excludes it from ordinary verification; an existing credential refuses it.
func TestNativeCredentialsLifecycle(t *testing.T) {
	c := newCredentialCLI(t)
	native := secret.Native{}
	if _, err := native.Get(); !errors.Is(err, secret.ErrMissing) {
		if errors.Is(err, secret.ErrUnavailable) {
			t.Fatal("native store is unavailable; no credential mutation attempted")
		}
		t.Fatal("native store already contains a credential; refusing mutation")
	}
	t.Cleanup(func() {
		value, err := native.Get()
		if errors.Is(err, secret.ErrMissing) {
			return
		}
		if err != nil || !c.owns(value) {
			t.Error("fixture cleanup could not establish ownership; credential preserved")
			return
		}
		if native.Delete() != nil {
			t.Error("generated credential cleanup failed")
		}
	})
	step := func(name string, run func(*testing.T)) {
		if !t.Run(name, run) {
			t.FailNow()
		}
	}
	step("missing-status", func(t *testing.T) {
		assertCredentialStatus(t, c.run(t, nil, "", 0, "providers", "status"), false, "keyring")
	})
	step("stdin-login-and-native-read", func(t *testing.T) {
		out := c.run(t, nil, c.values[0]+"\n", 0, "providers", "login", "tinypng", "--stdin")
		var receipt struct {
			Stored              string `json:"stored"`
			Verification        string `json:"verification"`
			RemotePolicyChanged bool   `json:"remotePolicyChanged"`
		}
		if json.Unmarshal(out, &receipt) != nil || receipt.Stored != "keyring" || receipt.Verification != "not-probed" || receipt.RemotePolicyChanged {
			t.Fatal("login did not return the expected local-only receipt")
		}
		if value, err := native.Get(); err != nil || value != c.values[0] {
			t.Fatal("native read differs from the stdin fixture")
		}
	})
	step("cross-process-status-and-resolution", func(t *testing.T) {
		assertCredentialStatus(t, c.run(t, nil, "", 0, "providers", "status", "tinypng"), true, "keyring")
		resolver := secret.Resolver{Lookup: func(string) (string, bool) { return "", false }}
		value, source, err := resolver.Resolve(context.Background())
		if err != nil || source != "keyring" || value != c.values[0] {
			t.Fatal("native credential did not resolve across processes")
		}
	})
	step("native-replacement-visible-to-cli", func(t *testing.T) {
		if native.Set(c.values[1]) != nil {
			t.Fatal("native replacement failed")
		}
		assertCredentialStatus(t, c.run(t, nil, "", 0, "providers", "status"), true, "keyring")
		if value, err := native.Get(); err != nil || value != c.values[1] {
			t.Fatal("native replacement did not persist")
		}
	})
	step("environment-precedence-without-store-mutation", func(t *testing.T) {
		assertCredentialStatus(t, c.run(t, &c.values[0], "", 0, "providers", "status"), true, "environment")
		invalid := " "
		assertCredentialStatus(t, c.run(t, &invalid, "", 0, "providers", "status"), false, "environment")
		if value, err := native.Get(); err != nil || value != c.values[1] {
			t.Fatal("environment override changed the stored fixture")
		}
	})
	step("invalid-stdin-preserves-store", func(t *testing.T) {
		c.run(t, nil, "two words\n", 2, "providers", "login", "tinypng", "--stdin")
		if value, err := native.Get(); err != nil || value != c.values[1] {
			t.Fatal("invalid stdin changed the stored fixture")
		}
	})
	step("logout-removes-only-stored-credential", func(t *testing.T) {
		c.run(t, &c.values[0], "", 0, "providers", "logout", "tinypng")
		if _, err := native.Get(); !errors.Is(err, secret.ErrMissing) {
			t.Fatal("logout did not remove the native fixture")
		}
		assertCredentialStatus(t, c.run(t, &c.values[0], "", 0, "providers", "status"), true, "environment")
		assertCredentialStatus(t, c.run(t, nil, "", 0, "providers", "status"), false, "keyring")
		c.run(t, nil, "", 0, "providers", "logout", "tinypng")
	})
	step("no-plaintext-application-files", c.assertNoPlaintext)
}

func TestNativeUnavailableCredentials(t *testing.T) {
	if runtime.GOOS != "linux" || os.Getenv("SMOLPNG_KEYRING_MODE") != "unavailable" {
		t.Fatal("this journey requires the explicitly unavailable disposable Linux bus")
	}
	c := newCredentialCLI(t)
	if !strings.HasPrefix(os.Getenv("DBUS_SESSION_BUS_ADDRESS"), "unix:path="+c.home) {
		// The workflow chooses a nonexistent socket below the isolated application home.
		t.Fatal("unavailable-store journey must name its isolated nonexistent bus socket")
	}
	if _, err := (secret.Native{}).Get(); !errors.Is(err, secret.ErrUnavailable) {
		t.Fatal("native store was not unavailable")
	}
	status := assertCredentialStatus(t, c.run(t, nil, "", 0, "providers", "status"), false, "keyring")
	if status.Reason != secret.ErrUnavailable.Error() {
		t.Fatal("unavailable status lost its actionable reason")
	}
	c.run(t, nil, c.values[0]+"\n", 5, "providers", "login", "tinypng", "--stdin")
	assertCredentialStatus(t, c.run(t, &c.values[1], "", 0, "providers", "status"), true, "environment")
	c.assertNoPlaintext(t)
}

type credentialCLI struct {
	executable string
	home       string
	values     [2]string
}

func newCredentialCLI(t *testing.T) *credentialCLI {
	t.Helper()
	if os.Getenv("SMOLPNG_NATIVE_KEYRING") != "disposable-host" || os.Getenv("GITHUB_ACTIONS") != "true" {
		t.Fatal("native keyring journeys run only on explicitly selected disposable hosted runners")
	}
	c := &credentialCLI{executable: os.Getenv("SMOLPNG_KEYRING_CLI"), home: os.Getenv("SMOLPNG_KEYRING_HOME")}
	if !filepath.IsAbs(c.executable) || !filepath.IsAbs(c.home) {
		t.Fatal("native journey needs absolute executable and isolated home paths")
	}
	for i := range c.values {
		var random [32]byte
		if _, err := rand.Read(random[:]); err != nil {
			t.Fatal("could not generate credential fixture")
		}
		c.values[i] = "smolpng-fixture-" + hex.EncodeToString(random[:])
	}
	return c
}

func (c *credentialCLI) owns(value string) bool {
	return value == c.values[0] || value == c.values[1]
}

func (c *credentialCLI) run(t *testing.T, environment *string, input string, wantExit int, arguments ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	args := append([]string{"--home", c.home, "--no-project", "--json"}, arguments...)
	cmd := process.CommandContext(ctx, c.executable, args...)
	cmd.WaitDelay = 2 * time.Second
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(entry), "TINIFY_API_KEY=") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	if environment != nil {
		cmd.Env = append(cmd.Env, "TINIFY_API_KEY="+*environment)
	}
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if c.containsCredential(stdout.Bytes()) || c.containsCredential(stderr.Bytes()) {
		t.Fatal("CLI output disclosed a generated credential; output withheld")
	}
	exit := 0
	if err != nil {
		var failure *exec.ExitError
		if !errors.As(err, &failure) {
			t.Fatal("CLI process failed to finish; raw process output withheld")
		}
		exit = failure.ExitCode()
	}
	if exit != wantExit {
		t.Fatalf("CLI exit %d, expected %d; output withheld", exit, wantExit)
	}
	return stdout.Bytes()
}

func assertCredentialStatus(t *testing.T, output []byte, configured bool, source string) secret.Status {
	t.Helper()
	var status secret.Status
	if json.Unmarshal(output, &status) != nil || status.Provider != "tinypng" || status.Configured != configured || status.Source != source || status.Verification != "not-probed" {
		t.Fatal("credential status differs from the expected configuration/source; output withheld")
	}
	return status
}

func (c *credentialCLI) assertNoPlaintext(t *testing.T) {
	t.Helper()
	err := filepath.WalkDir(c.home, func(path string, entry fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) && path == c.home {
			return nil
		}
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if c.containsCredential(data) {
			return errors.New("generated credential found in application files")
		}
		return nil
	})
	if err != nil {
		t.Fatal("application-home credential containment check failed; contents withheld")
	}
}

func (c *credentialCLI) containsCredential(data []byte) bool {
	for _, value := range c.values {
		// macOS stores the pinned adapter's base64 representation of the value.
		for _, form := range []string{value, base64.StdEncoding.EncodeToString([]byte(value))} {
			if bytes.Contains(data, []byte(form)) {
				return true
			}
		}
	}
	return false
}
