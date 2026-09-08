// SPDX-License-Identifier: MIT
// Package secret retrieves provider credentials without exposing them to callers' reports.
package secret

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/zalando/go-keyring"
)

const Service = "smolpng"
const Account = "tinypng"

var ErrMissing = errors.New("no TinyPNG credential; use providers login tinypng or TINIFY_API_KEY")
var ErrUnavailable = errors.New("native keyring unavailable or locked; unlock it or supply TINIFY_API_KEY for this process")

type Store interface {
	Get() (string, error)
	Set(string) error
	Delete() error
}
type Native struct{}

func (Native) Get() (string, error) {
	v, e := keyring.Get(Service, Account)
	if errors.Is(e, keyring.ErrNotFound) {
		return "", ErrMissing
	}
	if e != nil {
		return "", ErrUnavailable
	}
	return v, nil
}
func (Native) Set(v string) error {
	if e := Validate(v); e != nil {
		return e
	}
	if e := keyring.Set(Service, Account, v); e != nil {
		return ErrUnavailable
	}
	return nil
}
func (Native) Delete() error {
	e := keyring.Delete(Service, Account)
	if e == nil || errors.Is(e, keyring.ErrNotFound) {
		return nil
	}
	return ErrUnavailable
}
func Validate(v string) error {
	if len(v) < 1 || len(v) > 512 || strings.TrimSpace(v) != v {
		return fmt.Errorf("API key must be 1..512 visible ASCII characters without surrounding whitespace")
	}
	for _, c := range v {
		if c < 33 || c > 126 {
			return fmt.Errorf("API key contains unsupported whitespace or control characters")
		}
	}
	return nil
}

// Credentials must not be serialized. Status returns only configuration/source.
type Resolver struct {
	Store  Store
	Lookup func(string) (string, bool)
}
type Status struct {
	Provider     string `json:"provider"`
	Configured   bool   `json:"configured"`
	Source       string `json:"credentialSource"`
	Verification string `json:"verification"`
	Reason       string `json:"reason,omitempty"`
}

func (r Resolver) Resolve(ctx context.Context) (string, string, error) {
	if e := ctx.Err(); e != nil {
		return "", "", e
	}
	lookup := r.Lookup
	if lookup == nil {
		lookup = os.LookupEnv
	}
	if v, present := lookup("TINIFY_API_KEY"); present {
		if e := Validate(v); e != nil {
			return "", "environment", fmt.Errorf("TINIFY_API_KEY is present but invalid; it takes precedence over the keyring")
		}
		return v, "environment", nil
	}
	store := r.Store
	if store == nil {
		store = Native{}
	}
	v, e := store.Get()
	if e != nil {
		return "", "keyring", e
	}
	if e = Validate(v); e != nil {
		return "", "keyring", fmt.Errorf("stored TinyPNG key is invalid; replace it using providers login")
	}
	return v, "keyring", nil
}
func (r Resolver) Status(ctx context.Context) Status {
	_, source, e := r.Resolve(ctx)
	s := Status{Provider: "tinypng", Source: source, Configured: e == nil, Verification: "not-probed"}
	if e != nil {
		s.Reason = e.Error()
	}
	return s
}
