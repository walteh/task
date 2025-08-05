package task

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHCLE2E(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/task", "-t", filepath.Join("testdata", "HCLE2ETest", "Taskfile.hcl"), "all")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("task run failed: %v\n%s", err, out)
	}
	output := string(out)
	if !strings.Contains(output, "BUILD:1.2.3") {
		t.Fatalf("missing build output: %s", output)
	}
	one := strings.Index(output, "ONE-DONE")
	two := strings.Index(output, "TWO-DONE")
	if one == -1 || two == -1 || one > two {
		t.Fatalf("dependency order wrong: %s", output)
	}
	if !strings.Contains(output, "LINT MODE fast") {
		t.Fatalf("missing lint output: %s", output)
	}
	if !strings.Contains(output, "FINAL foo") {
		t.Fatalf("missing final output: %s", output)
	}
	idx := strings.Index(output, "PATH=")
	if idx == -1 || idx+5 >= len(output) || output[idx+5] == '\n' {
		t.Fatalf("missing path output: %s", output)
	}
}
