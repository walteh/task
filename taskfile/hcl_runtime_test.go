package taskfile

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/require"

	"github.com/go-task/task/v3/internal/env"
	"github.com/go-task/task/v3/internal/hclext"
	"github.com/go-task/task/v3/taskfile/ast"
)

func TestBlockSyntaxEnforced(t *testing.T) {
	data := []byte(`version = "3"
        vars = { FOO = "bar" }
        `)
	loader := HCLLoader{}
	_, err := loader.Load(data, "Taskfile.hcl")
	require.Error(t, err)
}

func TestShFunctionSuccess(t *testing.T) {
	eval := hclext.NewHCLEvaluator(nil, env.GetEnviron(), nil)
	expr, diags := hclsyntax.ParseExpression([]byte(`sh("echo hi")`), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	v, err := eval.EvalString(expr)
	require.NoError(t, err)
	require.Equal(t, "hi", v)
}

func TestShFunctionFail(t *testing.T) {
	eval := hclext.NewHCLEvaluator(nil, env.GetEnviron(), nil)
	expr, diags := hclsyntax.ParseExpression([]byte(`sh("exit 1")`), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	_, err := eval.EvalString(expr)
	require.Error(t, err)
}

func TestTaskFunctionStdoutCapture(t *testing.T) {
	runner := func(name string, vars *ast.Vars) (string, error) {
		return "output", nil
	}
	eval := hclext.NewHCLEvaluator(nil, env.GetEnviron(), runner)
	expr, diags := hclsyntax.ParseExpression([]byte(`task("build")`), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	v, err := eval.EvalString(expr)
	require.NoError(t, err)
	require.Equal(t, "output", v)
}

func TestInvalidReference(t *testing.T) {
	eval := hclext.NewHCLEvaluator(ast.NewVars(), env.GetEnviron(), nil)
	expr, diags := hclsyntax.ParseExpression([]byte(`FOO`), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	_, err := eval.EvalString(expr)
	require.Error(t, err)
}
