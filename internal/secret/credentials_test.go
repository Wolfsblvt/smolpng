package secret

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

type store struct {
	value string
	reads int
}

func (s *store) Get() (string, error) {
	s.reads++
	if s.value == "" {
		return "", ErrMissing
	}
	return s.value, nil
}
func (s *store) Set(v string) error { s.value = v; return nil }
func (s *store) Delete() error      { s.value = ""; return nil }
func TestEnvironmentPrecedenceAndNoSecretInStatus(t *testing.T) {
	s := &store{value: "keyring-secret"}
	r := Resolver{Store: s, Lookup: func(string) (string, bool) { return "environment-secret", true }}
	v, src, e := r.Resolve(context.Background())
	if e != nil || v != "environment-secret" || src != "environment" || s.reads != 0 {
		t.Fatal(src, e)
	}
	b, _ := json.Marshal(r.Status(context.Background()))
	if strings.Contains(string(b), "secret") {
		t.Fatal("secret was serialized")
	}
	r.Lookup = func(string) (string, bool) { return "", true }
	if _, _, e = r.Resolve(context.Background()); e == nil || s.reads != 0 {
		t.Fatal("invalid environment silently fell back")
	}
	r.Lookup = func(string) (string, bool) { return "", false }
	v, src, e = r.Resolve(context.Background())
	if e != nil || v != s.value || src != "keyring" {
		t.Fatal(src, e)
	}
}
