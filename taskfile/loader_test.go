package taskfile

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/go-task/task/v3/taskfile/ast"
)

func TestYAMLLoader(t *testing.T) {
	t.Parallel()

	data := []byte("version: '3'\n\ntasks:\n  default:\n    cmds:\n      - echo hello\n")
	loader := YAMLLoader{}
	tf, err := loader.Load(data, "Taskfile.yml")
	require.NoError(t, err)

	var expected ast.Taskfile
	require.NoError(t, yaml.Unmarshal(data, &expected))
	require.Equal(t, &expected, tf)
}
