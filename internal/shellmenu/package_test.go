package shellmenu

import (
	"archive/zip"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Structural PE fixture, never executed. Only tests our bounded architecture and
// receipt parser. It is NOT a native DLL/EXE qualification or signature proof.
func tinyPE(machine uint16, dll bool) []byte {
	b := make([]byte, 128)
	copy(b, []byte{'M', 'Z'})
	binary.LittleEndian.PutUint32(b[0x3c:], 64)
	copy(b[64:], []byte{'P', 'E', 0, 0})
	binary.LittleEndian.PutUint16(b[68:], machine)
	if dll {
		binary.LittleEndian.PutUint16(b[86:], 0x2000)
	}
	return b
}
func packageFixture(t *testing.T) (string, string, string) {
	t.Helper()
	d := canonicalTempDir(t)
	commit := strings.Repeat("1", 40)
	files := map[string]string{}
	for _, n := range []string{"smolpng.exe", "smolpngw.exe", "SmolpngShell.dll"} {
		p := filepath.Join(d, n)
		if e := os.WriteFile(p, tinyPE(0x8664, n == "SmolpngShell.dll"), 0600); e != nil {
			t.Fatal(e)
		}
		h, e := FileHash(p)
		if e != nil {
			t.Fatal(e)
		}
		files[n] = h
	}
	receipt, _ := json.Marshal(Bridge{1, commit, "amd64", files})
	_ = os.WriteFile(filepath.Join(d, "shell-bridge.json"), receipt, 0600)
	template, e := os.ReadFile("../../native/windows-shell/AppxManifest.xml.in")
	if e != nil {
		t.Fatal(e)
	}
	xml := strings.NewReplacer("@PUBLISHER@", "CN=Explicit Test Only", "@VERSION@", "0.3.0.1", "@ARCH@", "x64").Replace(string(template))
	p := filepath.Join(d, "test.msix")
	f, e := os.Create(p)
	if e != nil {
		t.Fatal(e)
	}
	z := zip.NewWriter(f)
	w, _ := z.Create("AppxManifest.xml")
	_, _ = w.Write([]byte(xml))
	w, _ = z.Create("AppxSignature.p7x")
	_, _ = w.Write([]byte("test marker, not a trusted signature"))
	if e = z.Close(); e != nil {
		t.Fatal(e)
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	return p, d, commit
}
func TestSparsePackageStructuralScopeAndExactPayload(t *testing.T) {
	p, d, commit := packageFixture(t)
	info, e := CheckPackage(p, d, commit, "amd64")
	if e != nil || info.Name != PackageName {
		t.Fatal(info, e)
	}
	if _, e = CheckPackage(p, d, commit, "arm64"); e == nil {
		t.Fatal("wrong architecture accepted")
	}
	if _, e = CheckPackage(p, d, "unknown", "amd64"); e == nil {
		t.Fatal("unbound source accepted")
	}
	_ = os.WriteFile(filepath.Join(d, "SmolpngShell.dll"), []byte("changed"), 0600)
	if _, e = CheckPackage(p, d, commit, "amd64"); e == nil {
		t.Fatal("changed external DLL accepted")
	}
}
func TestUnsignedSparsePackageRefusedBeforeActivation(t *testing.T) {
	_, d, c := packageFixture(t)
	p := filepath.Join(d, "unsigned.msix")
	f, _ := os.Create(p)
	z := zip.NewWriter(f)
	w, _ := z.Create("AppxManifest.xml")
	_, _ = w.Write([]byte("<Package/>"))
	z.Close()
	f.Close()
	if _, e := CheckPackage(p, d, c, "amd64"); e == nil {
		t.Fatal("unsigned package accepted")
	}
}
func TestPackageUnknownMemberRejected(t *testing.T) {
	_, d, c := packageFixture(t)
	p := filepath.Join(d, "unexpected.msix")
	f, _ := os.Create(p)
	z := zip.NewWriter(f)
	w, _ := z.Create("other.exe")
	_, _ = w.Write([]byte("x"))
	z.Close()
	f.Close()
	if _, e := CheckPackage(p, d, c, "amd64"); e == nil {
		t.Fatal("unrelated package accepted")
	}
}
