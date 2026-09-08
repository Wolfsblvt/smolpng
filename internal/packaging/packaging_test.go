package packaging

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"github.com/Wolfsblvt/smolpng/internal/updates"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestArchivesAreDeterministicAndCarryFileModes(t *testing.T) {
	files := map[string]Content{"smolpng": {Data: []byte("binary"), Executable: true}, "LICENSE": {Data: []byte("MIT")}}
	stamp := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	for _, format := range []string{"zip", "tar.gz"} {
		a, e := Archive(files, format, stamp)
		if e != nil {
			t.Fatal(e)
		}
		b, e := Archive(files, format, stamp)
		if e != nil || !bytes.Equal(a, b) {
			t.Fatal("non-deterministic", format, e)
		}
	}
	b, _ := Archive(files, "zip", stamp)
	z, e := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if e != nil {
		t.Fatal(e)
	}
	for _, f := range z.File {
		if f.Name == "smolpng" && f.Mode().Perm() != 0755 {
			t.Fatal(f.Mode())
		}
	}
}
func TestUpdatePackageIsExactlyTheVerifiedSubset(t *testing.T) {
	files := map[string]Content{"smolpng": {Data: []byte("binary"), Executable: true}, "LICENSE": {Data: []byte("MIT")}, "LICENSES/library.txt": {Data: []byte("BSD notice")}, "docs/README.md": {Data: []byte("documentation")}, "config.toml": {Data: []byte("user preference")}, "testdata/image.png": {Data: []byte("input")}}
	for _, format := range []string{"zip", "tar.gz"} {
		b, a, e := UpdateAsset(files, "linux/amd64", "https://example.test/update", format, time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC))
		if e != nil {
			t.Fatal(e)
		}
		dir := t.TempDir()
		if e = updates.Extract(b, a, "linux/amd64", dir); e != nil {
			t.Fatal(e)
		}
		if len(a.Files) != 3 {
			t.Fatal(a.Files)
		}
		if b, e := os.ReadFile(filepath.Join(dir, "smolpng")); e != nil || string(b) != "binary" {
			t.Fatal(e)
		}
		if _, e := os.Stat(filepath.Join(dir, "config.toml")); !os.IsNotExist(e) {
			t.Fatal("user config in update")
		}
	}
}
func TestPackageMetadataIsOfflineAndUsesExactHashes(t *testing.T) {
	as := []Artifact{}
	for _, p := range []string{"windows/amd64", "windows/arm64", "linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64"} {
		as = append(as, Artifact{Platform: p, URL: "https://example.test/" + strings.ReplaceAll(p, "/", "-") + ".zip", SHA256: strings.Repeat("a", 64)})
	}
	files, e := Metadata("0.2.0", as)
	if e != nil {
		t.Fatal(e)
	}
	if !json.Valid(files["scoop/smolpng.json"]) {
		t.Fatal("invalid scoop JSON")
	}
	for _, n := range []string{"winget/Wolfsblvt.smolpng.installer.yaml", "homebrew/smolpng.rb", "chocolatey/tools/chocolateyinstall.ps1"} {
		if !strings.Contains(strings.ToLower(string(files[n])), strings.Repeat("a", 64)) {
			t.Fatal("missing exact hashes", n)
		}
	}
	for _, n := range []string{"nfpm/amd64.yaml", "nfpm/arm64.yaml", "PUBLICATION-NOT-AUTHORIZED.md"} {
		if len(files[n]) == 0 {
			t.Fatal("missing config", n)
		}
	}
	if _, e = Metadata("0.2.0-beta", as); e == nil {
		t.Fatal("prerelease package submission allowed")
	}
}
func TestArchivePathAndMetadataInjectionRejected(t *testing.T) {
	for _, name := range []string{"../x", "/x", "C:/x", `dir\x`} {
		if _, e := Archive(map[string]Content{name: {Data: []byte("x")}}, "zip", time.Now()); e == nil {
			t.Fatal("path accepted", name)
		}
	}
	if _, e := Metadata("1.0.0\nInjected: true", nil); e == nil {
		t.Fatal("version injection")
	}
}

func TestMetadataRejectsDuplicateUnknownAndMalformedArtifactIdentity(t *testing.T) {
	good := Artifact{Platform: "linux/amd64", URL: "https://example.test/pkg.tar.gz", SHA256: strings.Repeat("a", 64)}
	if _, e := Metadata("1.0.0", []Artifact{good, good}); e == nil {
		t.Fatal("duplicate platform accepted")
	}
	bad := good
	bad.SHA256 = strings.Repeat("g", 64)
	if _, e := Metadata("1.0.0", []Artifact{bad}); e == nil {
		t.Fatal("invalid digest accepted")
	}
	bad = good
	bad.Platform = "linux/mips"
	if _, e := Metadata("1.0.0", []Artifact{bad}); e == nil {
		t.Fatal("unknown platform silently ignored")
	}
}
