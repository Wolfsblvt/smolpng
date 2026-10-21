package packaging

import (
	"strings"
	"testing"
)

func TestChocolateyControlsGeneratedShims(t *testing.T) {
	files, err := Metadata("0.3.0", []Artifact{{
		Platform: "windows/amd64",
		URL:      "https://example.test/smolpng.zip",
		SHA256:   strings.Repeat("a", 64),
	}})
	if err != nil {
		t.Fatal(err)
	}
	script := string(files["chocolatey/tools/chocolateyinstall.ps1"])
	for _, marker := range []string{"smolpng-qualify.exe.ignore", "smolpngw.exe.gui"} {
		if !strings.Contains(script, marker) {
			t.Fatalf("Chocolatey install script must create %s", marker)
		}
	}
}
