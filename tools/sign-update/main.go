// SPDX-License-Identifier: MIT
// Command sign-update signs an already reviewed local release payload. It does
// not generate a key, publish a signature, or upload a private key anywhere.
package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"github.com/Wolfsblvt/smolpng/internal/fsutil"
	"github.com/Wolfsblvt/smolpng/internal/updates"
	"os"
	"runtime"
	"time"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, "sign-update:", e)
		os.Exit(1)
	}
}
func run() error {
	payloadPath := flag.String("payload", "", "Reviewed release.UNSIGNED.json")
	keyFile := flag.String("key-file", "", "Protected PKCS#8 Ed25519 PEM file; never a command-line secret value")
	id := flag.String("key-id", "", "Public key identifier already established for this release authority")
	dest := flag.String("out", "", "New signed envelope destination")
	flag.Parse()
	if *keyFile == "" || *payloadPath == "" || *id == "" || *dest == "" || flag.NArg() != 0 {
		return fmt.Errorf("payload, key-file, key-id and out are required")
	}
	st, e := os.Stat(*keyFile)
	if e != nil {
		return fmt.Errorf("signing key file unavailable")
	}
	if runtime.GOOS != "windows" && st.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("signing key file must not be readable by group/other")
	}
	b, e := updates.ReadFile(*keyFile, 16384)
	if e != nil {
		return fmt.Errorf("cannot read protected signing key")
	}
	defer clear(b)
	block, rest := pem.Decode(b)
	if block == nil || len(bytes.TrimSpace(rest)) != 0 || block.Type != "PRIVATE KEY" {
		return fmt.Errorf("a PKCS#8 PEM private key is required")
	}
	defer clear(block.Bytes)
	parsed, e := x509.ParsePKCS8PrivateKey(block.Bytes)
	if e != nil {
		return fmt.Errorf("invalid signing key")
	}
	key, ok := parsed.(ed25519.PrivateKey)
	if !ok {
		return fmt.Errorf("Ed25519 signing key required")
	}
	defer clear(key)
	payload, e := updates.ReadFile(*payloadPath, updates.MaxManifest)
	if e != nil {
		return e
	}
	pub := key.Public().(ed25519.PublicKey)
	signed, e := json.MarshalIndent(updates.Envelope{Payload: base64.StdEncoding.EncodeToString(payload), Signatures: []updates.Signature{{KeyID: *id, Value: base64.StdEncoding.EncodeToString(ed25519.Sign(key, payload))}}}, "", "  ")
	if e != nil {
		return e
	}
	trust := updates.Trust{Keys: map[string]string{*id: base64.StdEncoding.EncodeToString(pub)}}
	// An arbitrary JSON document is not accepted merely because it was signed.
	if _, e = updates.Verify(signed, trust, time.Now().UTC()); e != nil {
		return e
	}
	return fsutil.AtomicWrite(context.Background(), *dest, append(signed, '\n'), 0644, false)
}
