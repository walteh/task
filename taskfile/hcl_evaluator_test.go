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

func TestHCLEvaluatorExpressions(t *testing.T) {
	t.Setenv("HOME", "/home/test")
	vars := ast.NewVars()
	vars.Set("FOO", ast.Var{Value: "bar"})
	eval := hclext.NewHCLEvaluator(vars, env.GetEnviron(), nil, nil)

	expr1, diags := hclsyntax.ParseTemplate([]byte("${vars.FOO}"), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	v, err := eval.EvalString(expr1)
	require.NoError(t, err)
	require.Equal(t, "bar", v)

	expr2, diags := hclsyntax.ParseTemplate([]byte("${upper(vars.FOO)}"), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	v, err = eval.EvalString(expr2)
	require.NoError(t, err)
	require.Equal(t, "BAR", v)

	expr3, diags := hclsyntax.ParseTemplate([]byte("${env(\"HOME\")}"), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	v, err = eval.EvalString(expr3)
	require.NoError(t, err)
	require.Equal(t, "/home/test", v)
}
