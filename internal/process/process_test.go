package process

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestChild(t *testing.T) {
	if os.Getenv("SMOLPNG_TEST_CHILD") != "1" {
		return
	}
	if os.Getenv("SMOLPNG_TEST_WAIT") == "1" {
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}
	if os.Getenv("TINIFY_API_KEY") != "" {
		os.Exit(9)
	}
	for i, a := range os.Args {
		if a == "--" {
			_ = json.NewEncoder(os.Stdout).Encode(os.Args[i+1:])
			os.Exit(0)
		}
	}
	os.Exit(2)
}
func TestLiteralArgumentsAndNoRemoteCredentialInheritance(t *testing.T) {
	t.Setenv("SMOLPNG_TEST_CHILD", "1")
	t.Setenv("TINIFY_API_KEY", "not-a-real-secret")
	args := []string{"a b.png", "'; touch nope", "$(echo nope)", "雪.png"}
	r, e := (OSRunner{}).Run(context.Background(), Invocation{Executable: os.Args[0], Args: append([]string{"-test.run=^TestChild$", "--"}, args...)})
	if e != nil || r.ExitCode != 0 {
		t.Fatal(r, e)
	}
	var got []string
	if e = json.Unmarshal([]byte(r.Stdout), &got); e != nil || !reflect.DeepEqual(args, got) {
		t.Fatal(r, e)
	}
}
func TestOwnedProcessCancellation(t *testing.T) {
	t.Setenv("SMOLPNG_TEST_CHILD", "1")
	t.Setenv("SMOLPNG_TEST_WAIT", "1")
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	_, e := (OSRunner{}).Run(ctx, Invocation{Executable: os.Args[0], Args: []string{"-test.run=^TestChild$"}})
	if !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
}
