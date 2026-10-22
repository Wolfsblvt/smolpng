// SPDX-License-Identifier: MIT
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wolfsblvt/smolpng/internal/process"
)

func TestExplorerLaunchHelper(t *testing.T) {
	switch os.Getenv("SMOLPNG_TEST_EXPLORER_LAUNCH") {
	case "parent":
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		child := process.CommandContext(ctx, os.Getenv("SMOLPNG_TEST_CLI_BINARY"), "-test.run=^TestExplorerLaunchHelper$")
		child.Env = append(os.Environ(), "SMOLPNG_TEST_EXPLORER_LAUNCH=child")
		child.Stdout, child.Stderr = os.Stdout, os.Stderr
		if err := child.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	case "child":
		os.Exit(Run(context.Background(), []string{"version"}, Environment{Out: os.Stdout, Err: os.Stderr}))
	}
}

func TestExplorerLaunchDispatchesCommand(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	image, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	// Cobra's Windows guard identifies Explorer by the immediate parent's image
	// name. Reproduce that boundary without requiring a desktop or moving focus.
	parent := filepath.Join(t.TempDir(), "explorer.exe")
	if err := os.WriteFile(parent, image, 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := process.CommandContext(ctx, parent, "-test.run=^TestExplorerLaunchHelper$")
	command.Env = append(os.Environ(), "SMOLPNG_TEST_EXPLORER_LAUNCH=parent", "SMOLPNG_TEST_CLI_BINARY="+executable)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Explorer-parented command did not complete: %v\n%s", err, output)
	}
	var identity struct {
		SchemaVersion int    `json:"schemaVersion"`
		Version       string `json:"version"`
	}
	if err := json.Unmarshal(output, &identity); err != nil || identity.SchemaVersion != 1 || identity.Version == "" {
		t.Fatalf("Explorer-parented command did not emit its build identity: %s; decode: %v", output, err)
	}
}
