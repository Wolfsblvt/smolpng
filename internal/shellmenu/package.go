// SPDX-License-Identifier: MIT
package shellmenu

import (
	"archive/zip"
	"crypto/sha256"
	"debug/pe"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Bridge struct {
	Schema       int               `json:"schema"`
	Commit       string            `json:"commit"`
	Architecture string            `json:"architecture"`
	Files        map[string]string `json:"files"`
}
type Package struct{ Name, Publisher, Version, Architecture, SHA256 string }

func FileHash(path string) (string, error) {
	// Shell state and package files are trusted-user inputs, never privileged code.
	// Still refuse links to avoid unintentional writes or inconsistent ownership.
	for p := path; ; p = filepath.Dir(p) {
		st, e := os.Lstat(p)
		if e != nil {
			return "", e
		}
		if st.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symlink in native package path")
		}
		if filepath.Dir(p) == p {
			break
		}
	}
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	s, e := f.Stat()
	if e != nil || !s.Mode().IsRegular() || s.Size() > 256<<20 {
		return "", fmt.Errorf("native file is not a bounded regular file")
	}
	h := sha256.New()
	if _, e = io.Copy(h, io.LimitReader(f, 256<<20+1)); e != nil {
		return "", e
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// CheckPackage qualifies the sparse package's purpose and its separate payload.
// Trust-chain verification is deliberately left to Add-AppxPackage/Windows, not
// replaced by trusting an arbitrary local checksum or installing a certificate.
func CheckPackage(path, external, commit, arch string) (Package, error) {
	var out Package
	if !filepath.IsAbs(path) || !strings.EqualFold(filepath.Ext(path), ".msix") {
		return out, fmt.Errorf("select an absolute .msix path")
	}
	hash, e := FileHash(path)
	if e != nil {
		return out, e
	}
	out.SHA256 = hash
	z, e := zip.OpenReader(path)
	if e != nil {
		return out, e
	}
	defer z.Close()
	if len(z.File) > 32 {
		return out, fmt.Errorf("unexpected sparse package file count")
	}
	signed := false
	var manifest []byte
	for _, f := range z.File {
		if f.UncompressedSize64 > 1<<20 {
			return out, fmt.Errorf("unexpectedly large sparse package member")
		}
		switch f.Name {
		case "AppxSignature.p7x":
			signed = f.UncompressedSize64 > 0
		case "AppxManifest.xml":
			r, e := f.Open()
			if e != nil {
				return out, e
			}
			manifest, e = io.ReadAll(io.LimitReader(r, (1<<20)+1))
			r.Close()
			if e != nil {
				return out, e
			}
		case "AppxBlockMap.xml", "[Content_Types].xml", "Assets/StoreLogo.png", "Assets/Square44x44Logo.png", "Assets/Square150x150Logo.png":
		default:
			return out, fmt.Errorf("unexpected sparse package member %s", f.Name)
		}
	}
	if !signed || len(manifest) == 0 {
		return out, fmt.Errorf("package must contain a signed sparse manifest; smolpng never installs certificate trust")
	}
	d := xml.NewDecoder(strings.NewReader(string(manifest)))
	apps, classes, verbs := 0, 0, 0
	externalContent := false
	itemTypes := map[string]bool{}
	for {
		token, e := d.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			return out, e
		}
		s, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		attrs := map[string]string{}
		for _, a := range s.Attr {
			attrs[a.Name.Local] = a.Value
		}
		switch s.Name.Local {
		case "Identity":
			out.Name = attrs["Name"]
			out.Publisher = attrs["Publisher"]
			out.Version = attrs["Version"]
			out.Architecture = attrs["ProcessorArchitecture"]
		case "AllowExternalContent":
			var text string
			if e = d.DecodeElement(&text, &s); e != nil {
				return out, e
			}
			externalContent = text == "true"
		case "Application":
			apps++
			if attrs["Id"] != "SmolpngShell" || attrs["Executable"] != "smolpng.exe" {
				return out, fmt.Errorf("unexpected sparse application")
			}
		case "Class":
			classes++
			if !strings.EqualFold(attrs["Id"], ClassID) || attrs["Path"] != "SmolpngShell.dll" || attrs["ThreadingModel"] != "STA" {
				return out, fmt.Errorf("unexpected native shell class")
			}
		case "Extension":
			if attrs["Category"] != "windows.comServer" && attrs["Category"] != "windows.fileExplorerContextMenus" {
				return out, fmt.Errorf("unexpected package extension")
			}
		case "ItemType":
			if attrs["Type"] != ".png" && attrs["Type"] != ".jpg" && attrs["Type"] != ".jpeg" {
				return out, fmt.Errorf("context menu may target only supported image formats")
			}
			itemTypes[attrs["Type"]] = true
		case "Verb":
			verbs++
			if attrs["Id"] != "MakeSmol" || !strings.EqualFold(attrs["Clsid"], ClassID) {
				return out, fmt.Errorf("unexpected shell verb")
			}
		}
	}
	wantArch := map[string]string{"amd64": "x64", "arm64": "arm64"}[arch]
	if wantArch == "" || out.Architecture != wantArch || out.Name != PackageName || out.Publisher == "" || out.Version == "" || !externalContent || apps != 1 || classes != 1 || verbs != 3 || len(itemTypes) != 3 {
		return out, fmt.Errorf("package does not match smolpng's sparse shell contract/architecture")
	}
	if e = CheckBridge(external, commit, arch); e != nil {
		return out, e
	}
	return out, nil
}
func CheckBridge(dir, commit, arch string) error {
	if commit == "" || commit == "unknown" {
		return fmt.Errorf("native menu needs a source-bound candidate, not an unidentified development executable")
	}
	path := filepath.Join(dir, "shell-bridge.json")
	if _, e := FileHash(path); e != nil {
		return e
	}
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	if len(b) > 8192 {
		return fmt.Errorf("bridge receipt too large")
	}
	var receipt Bridge
	decoder := json.NewDecoder(strings.NewReader(string(b)))
	decoder.DisallowUnknownFields()
	if e = decoder.Decode(&receipt); e != nil {
		return e
	}
	if receipt.Schema != 1 || receipt.Commit != commit || receipt.Architecture != arch || len(receipt.Files) != 3 {
		return fmt.Errorf("native bridge receipt does not match this application's source/architecture")
	}
	machine := uint16(0x8664)
	if arch == "arm64" {
		machine = 0xaa64
	} else if arch != "amd64" {
		return fmt.Errorf("unsupported Explorer architecture")
	}
	for _, name := range []string{"smolpng.exe", "smolpngw.exe", "SmolpngShell.dll"} {
		h, e := FileHash(filepath.Join(dir, name))
		if e != nil || h != receipt.Files[name] {
			return fmt.Errorf("native payload hash mismatch: %s", name)
		}
		f, e := pe.Open(filepath.Join(dir, name))
		if e != nil {
			return e
		}
		ok := f.Machine == machine && (name != "SmolpngShell.dll" || f.Characteristics&0x2000 != 0)
		f.Close()
		if !ok {
			return fmt.Errorf("wrong native payload architecture/type: %s", name)
		}
	}
	return nil
}
