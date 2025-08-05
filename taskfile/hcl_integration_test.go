package taskfile

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHCLTaskfileRun(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	data := []byte(`version = "3"
    task "hello" {
        cmds = ["echo hi"]
    }
`)
	path := filepath.Join(dir, "Taskfile.hcl")
	require.NoError(t, os.WriteFile(path, data, 0o644))

	cmd := exec.Command("go", "run", "./cmd/task", "-t", path, "hello")
	err := cmd.Run()
	require.Error(t, err)
}
