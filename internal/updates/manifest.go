// SPDX-License-Identifier: MIT
// Package updates verifies signed standalone updates and owns their transaction.
// It has no built-in signing key and never overwrites package-manager installs.
package updates

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const MaxManifest = 1 << 20
const MaxArchive = 512 << 20
const MaxFile = 128 << 20

type Signature struct {
	KeyID string `json:"keyId"`
	Value string `json:"signature"`
}
type Envelope struct {
	Payload    string      `json:"payload"`
	Signatures []Signature `json:"signatures"`
}
type File struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}
type Asset struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	Files  []File `json:"files"`
}
type Manifest struct {
	Schema  int              `json:"schema"`
	Product string           `json:"product"`
	Version string           `json:"version"`
	Commit  string           `json:"commit"`
	Channel string           `json:"channel"`
	Issued  time.Time        `json:"issuedAt"`
	Expires time.Time        `json:"expiresAt"`
	Assets  map[string]Asset `json:"assets"`
}
type Trust struct {
	Keys map[string]string `json:"keys"`
}
type Verified struct {
	Manifest Manifest `json:"manifest"`
	KeyID    string   `json:"verifiedKeyId"`
	Envelope []byte   `json:"-"`
}

var hashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
var namePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$`)
var versionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z.-]+))?$`)

func VersionParts(v string) ([3]uint64, string, error) {
	m := versionPattern.FindStringSubmatch(v)
	var n [3]uint64
	if m == nil {
		return n, "", fmt.Errorf("version must be semantic MAJOR.MINOR.PATCH with optional prerelease")
	}
	for i := range n {
		v, e := strconv.ParseUint(m[i+1], 10, 32)
		if e != nil {
			return n, "", e
		}
		n[i] = v
	}
	return n, m[4], nil
}
func Newer(next, current string) bool {
	a, ap, e := VersionParts(next)
	if e != nil {
		return false
	}
	b, bp, e := VersionParts(current)
	if e != nil {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return ap == "" && bp != ""
}
func decode(data []byte, v any) error {
	if len(data) > MaxManifest {
		return fmt.Errorf("manifest/record too large")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return fmt.Errorf("invalid JSON contract: %w", e)
	}
	if e := d.Decode(new(any)); e != io.EOF {
		return fmt.Errorf("trailing JSON data")
	}
	return nil
}
func ParseTrust(data []byte) (Trust, error) {
	var t Trust
	e := decode(data, &t)
	if e != nil {
		return t, e
	}
	if len(t.Keys) == 0 || len(t.Keys) > 32 {
		return t, fmt.Errorf("one to 32 trusted public keys required")
	}
	for id, key := range t.Keys {
		b, e := base64.StdEncoding.DecodeString(key)
		if !namePattern.MatchString(id) || e != nil || len(b) != ed25519.PublicKeySize {
			return t, fmt.Errorf("invalid Ed25519 public key")
		}
	}
	return t, nil
}
func Verify(data []byte, trust Trust, now time.Time) (Verified, error) {
	var out Verified
	var e Envelope
	if err := decode(data, &e); err != nil {
		return out, err
	}
	payload, err := base64.StdEncoding.DecodeString(e.Payload)
	if err != nil || len(payload) > MaxManifest {
		return out, fmt.Errorf("invalid signed payload")
	}
	if len(e.Signatures) < 1 || len(e.Signatures) > 32 {
		return out, fmt.Errorf("invalid signature count")
	}
	for _, sig := range e.Signatures {
		key, ok := trust.Keys[sig.KeyID]
		if !ok {
			continue
		}
		public, er := base64.StdEncoding.DecodeString(key)
		s, er2 := base64.StdEncoding.DecodeString(sig.Value)
		if er == nil && er2 == nil && len(public) == ed25519.PublicKeySize && ed25519.Verify(public, payload, s) {
			out.KeyID = sig.KeyID
			break
		}
	}
	if out.KeyID == "" {
		return out, fmt.Errorf("release signature does not match a selected trusted key")
	}
	if err = decode(payload, &out.Manifest); err != nil {
		return out, err
	}
	m := out.Manifest
	if m.Schema != 1 || m.Product != "smolpng" || m.Channel != "stable" {
		return out, fmt.Errorf("unsupported product, schema or channel")
	}
	_, pre, err := VersionParts(m.Version)
	if err != nil || pre != "" || !commitPattern.MatchString(m.Commit) {
		return out, fmt.Errorf("stable release requires a stable semantic version and exact source SHA")
	}
	if m.Issued.After(now.Add(5*time.Minute)) || !m.Expires.After(now) || !m.Expires.After(m.Issued) || m.Expires.Sub(m.Issued) > 366*24*time.Hour {
		return out, fmt.Errorf("release manifest is expired, future-dated or has invalid validity bounds")
	}
	if len(m.Assets) == 0 || len(m.Assets) > 12 {
		return out, fmt.Errorf("invalid platform count")
	}
	for platform, a := range m.Assets {
		if err = ValidateAsset(platform, a); err != nil {
			return out, err
		}
	}
	out.Envelope = append([]byte{}, data...)
	return out, nil
}
func HTTPS(raw string) error {
	u, e := url.Parse(raw)
	if e != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" || u.Opaque != "" {
		return fmt.Errorf("release resources require an absolute credential-free HTTPS URL")
	}
	return nil
}
func allowedFile(name, platform string) bool {
	if strings.ContainsAny(name, `\:`) || strings.ContainsAny(name, "\x00\r\n") || path.Clean(name) != name || strings.HasPrefix(name, "/") {
		return false
	}
	if strings.HasPrefix(name, "LICENSES/") {
		base := strings.TrimPrefix(name, "LICENSES/")
		return namePattern.MatchString(base) && strings.HasSuffix(base, ".txt")
	}
	if name == "LICENSE" || name == "license-map.md" || name == "third-party-notices.md" {
		return true
	}
	suffix := ""
	if strings.HasPrefix(platform, "windows/") {
		suffix = ".exe"
	}
	if name == "smolpng"+suffix || name == "smolpng-update"+suffix || name == "smolpng-qualify"+suffix {
		return true
	}
	return strings.HasPrefix(platform, "windows/") && name == "smolpngw.exe"
}
func ValidateAsset(platform string, a Asset) error {
	switch platform {
	case "windows/amd64", "windows/arm64", "linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64":
	default:
		return fmt.Errorf("unsupported update platform %q", platform)
	}
	if e := HTTPS(a.URL); e != nil {
		return e
	}
	if !hashPattern.MatchString(a.SHA256) || a.Size < 1 || a.Size > MaxArchive || len(a.Files) < 2 || len(a.Files) > 256 {
		return fmt.Errorf("invalid archive hash, size or file count")
	}
	seen := map[string]bool{}
	total := int64(0)
	for _, f := range a.Files {
		key := strings.ToLower(f.Name)
		if seen[key] || !allowedFile(f.Name, platform) || !hashPattern.MatchString(f.SHA256) || f.Size < 1 || f.Size > MaxFile {
			return fmt.Errorf("invalid/duplicate update member %q", f.Name)
		}
		seen[key] = true
		total += f.Size
	}
	name := "smolpng"
	if strings.HasPrefix(platform, "windows/") {
		name += ".exe"
		if !seen["smolpngw.exe"] {
			return fmt.Errorf("Windows update must include the background executable")
		}
	}
	if !seen[name] || !seen["license"] || total > MaxArchive {
		return fmt.Errorf("update must contain main executable, license and bounded total file bytes")
	}
	return nil
}
