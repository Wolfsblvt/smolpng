// SPDX-License-Identifier: MIT
package packaging

import (
	"bytes"
	"debug/pe"
	"errors"
	"os"
	"testing"

	"github.com/tc-hib/winres"
	"github.com/tc-hib/winres/version"
)

func TestVersionWindowsExecutablePreservesGoImageAndOtherResources(t *testing.T) {
	// Reuse the already-built test executable: no fixture compiler, subprocess
	// or write to a running executable is needed for this real Go PE boundary.
	path, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	resources, err := winres.LoadFromEXE(bytes.NewReader(original))
	if err != nil && !errors.Is(err, winres.ErrNoResources) {
		t.Fatal(err)
	}
	if resources == nil {
		resources = &winres.ResourceSet{}
	}
	retained := []byte("existing non-version resource")
	if err := resources.Set(winres.RT_RCDATA, winres.ID(42), 0, retained); err != nil {
		t.Fatal(err)
	}
	var seed bytes.Buffer
	if err := resources.WriteToEXE(&seed, bytes.NewReader(original)); err != nil {
		t.Fatal(err)
	}
	input := seed.Bytes()
	snapshot := bytes.Clone(input)
	result, err := VersionWindowsExecutable(input, "smolpng.exe", "0.3.0-dev.12", "source")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(input, snapshot) {
		t.Fatal("versioning modified the input executable bytes")
	}
	updated, err := winres.LoadFromEXE(bytes.NewReader(result))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(updated.Get(winres.RT_RCDATA, winres.ID(42), 0), retained) {
		t.Fatal("existing resource was lost")
	}
	info, err := version.FromBytes(updated.Get(winres.RT_VERSION, winres.ID(1), version.LangDefault))
	if err != nil {
		t.Fatal(err)
	}
	if info.Table().GetMainTranslation()[version.ProductVersion] != "0.3.0-dev.12" {
		t.Fatal("version resource did not reach the executable")
	}
	before, err := pe.NewFile(bytes.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	defer before.Close()
	after, err := pe.NewFile(bytes.NewReader(result))
	if err != nil {
		t.Fatal(err)
	}
	defer after.Close()
	oldHeader := before.OptionalHeader.(*pe.OptionalHeader64)
	newHeader := after.OptionalHeader.(*pe.OptionalHeader64)
	if before.Machine != after.Machine || oldHeader.Subsystem != newHeader.Subsystem ||
		oldHeader.AddressOfEntryPoint != newHeader.AddressOfEntryPoint {
		t.Fatal("machine, subsystem or entry point changed")
	}
	oldCode, err := before.Section(".text").Data()
	if err != nil {
		t.Fatal(err)
	}
	newCode, err := after.Section(".text").Data()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(oldCode, newCode) {
		t.Fatal("executable code changed while embedding version metadata")
	}
}
