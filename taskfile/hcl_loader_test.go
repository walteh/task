package taskfile

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-task/task/v3/taskfile/ast"
)

func TestHCLLoader(t *testing.T) {
	t.Parallel()

	data := []byte(`version = "3"
        task "build" {
            desc = "Build the project"
            cmds = ["echo hello ${USER}"]
            vars = { USER = "world" }
            env = { GREETING = "hi" }
        }
    `)

	loader := HCLLoader{}
	tf, err := loader.Load(data, "Taskfile.hcl")
	require.NoError(t, err)

	require.NotNil(t, tf.Version)
	require.True(t, tf.Version.Equal(ast.V3))

	build, ok := tf.Tasks.Get("build")
	require.True(t, ok)
	require.Equal(t, "Build the project", build.Desc)
	require.Len(t, build.Cmds, 1)
	require.NotNil(t, build.Cmds[0].Expr)

	v, ok := build.Vars.Get("USER")
	require.True(t, ok)
	require.NotNil(t, v.Expr)

	e, ok := build.Env.Get("GREETING")
	require.True(t, ok)
	require.NotNil(t, e.Expr)
}

func TestHCLLoaderInvalid(t *testing.T) {
	t.Parallel()

	data := []byte(`version = "3"
        task "build" {
            desc = "Missing brace"
    `)

	loader := HCLLoader{}
	_, err := loader.Load(data, "Taskfile.hcl")
	require.Error(t, err)
	require.Contains(t, err.Error(), "Taskfile.hcl")
	require.Contains(t, err.Error(), ":2")
}

func TestHCLLoaderRejectsGoTemplates(t *testing.T) {
	t.Parallel()

	data := []byte(`version = "3"
        task "demo" {
            cmds = ["echo {{.FOO}}"]
        }
    `)

	loader := HCLLoader{}
	_, err := loader.Load(data, "Taskfile.hcl")
	require.Error(t, err)
}
