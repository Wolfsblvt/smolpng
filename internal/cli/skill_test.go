package cli

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/skill"
)

func TestSkillCLIInstallPreviewTrackingAndModifiedProtection(t *testing.T) {
	h := setup(t)
	// Canonical temporary base avoids the platform-owned /var -> /private/var alias on macOS.
	root, e := filepath.EvalSymlinks(h.dir)
	if e != nil {
		t.Fatal(e)
	}
	h.Paths.State = filepath.Join(root, "state")
	parent := filepath.Join(root, "custom-skills")
	dest := filepath.Join(parent, "smolpng")
	if c := h.run("skill", "install", "--path", parent, "--json"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	if _, e = os.Stat(parent); !os.IsNotExist(e) {
		t.Fatal("preview wrote files")
	}
	if c := h.run("skill", "install", "--path", parent, "--apply", "--json"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	if c := h.run("skill", "status", "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	var result struct {
		Status skill.Snapshot `json:"status"`
	}
	if e = json.Unmarshal(h.out.Bytes(), &result); e != nil {
		t.Fatal(e)
	}
	if len(result.Status.Installations) != 1 || result.Status.Installations[0].Destination != dest || result.Status.Installations[0].Managed {
		t.Fatal(h.out)
	}
	if e = os.WriteFile(filepath.Join(dest, "SKILL.md"), []byte("local instructions"), 0600); e != nil {
		t.Fatal(e)
	}
	if c := h.run("skill", "sync", "--destination", dest, "--apply", "--json"); c != 5 || !json.Valid(h.out.Bytes()) {
		t.Fatal(c, h.out, h.err)
	}
	if c := h.run("skill", "uninstall", "--destination", dest, "--apply", "--json"); c != 5 {
		t.Fatal(c, h.out, h.err)
	}
	if c := h.run("doctor", "--json"); c != 0 || !strings.Contains(h.out.String(), `"modified"`) {
		t.Fatal("doctor missed modified installation", c, h.out)
	}
	if c := h.run("skill", "forget", "--destination", dest, "--apply", "--json"); c != 0 {
		t.Fatal(c, h.out)
	}
	b, _ := os.ReadFile(filepath.Join(dest, "SKILL.md"))
	if string(b) != "local instructions" {
		t.Fatal("forget deleted user data")
	}
}
func TestSkillCLIInstructionPackageAndNoClobber(t *testing.T) {
	h := setup(t)
	output := filepath.Join(h.dir, "skill.zip")
	if c := h.run("skill", "package", "--target", "chatgpt", "--output", output, "--json"); c != 0 {
		t.Fatal(c, h.out, h.err)
	}
	z, e := zip.OpenReader(output)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	found := false
	for _, f := range z.File {
		if f.Name == "smolpng/SKILL.md" {
			found = true
		}
		if strings.HasSuffix(f.Name, ".exe") {
			t.Fatal("executable bundled in Skill")
		}
	}
	if !found {
		t.Fatal("missing entry point")
	}
	if c := h.run("skill", "package", "--output", output, "--json"); c != 5 {
		t.Fatal("existing archive overwritten", c, h.out)
	}
	if c := h.run("skill", "print"); c != 0 || !strings.HasPrefix(h.out.String(), "---\nname: smolpng") {
		t.Fatal(c, h.out)
	}
}
