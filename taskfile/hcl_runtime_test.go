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

func TestTaskFunctionReturnsTaskInfo(t *testing.T) {
	tasks := ast.NewTasks()
	tasks.Set("build", &ast.Task{})
	eval := hclext.NewHCLEvaluator(nil, env.GetEnviron(), tasks)
	expr, diags := hclsyntax.ParseExpression([]byte(`exec(task.build)`), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	
	// Use EvalCommand instead of EvalString since exec() now returns task info
	cmd, err := eval.EvalCommand(expr)
	require.NoError(t, err)
	require.True(t, cmd.IsTaskCall)
	require.Equal(t, "build", cmd.TaskName)
	require.Nil(t, cmd.TaskVars)
}

func TestInvalidReference(t *testing.T) {
	eval := hclext.NewHCLEvaluator(ast.NewVars(), env.GetEnviron(), nil)
	expr, diags := hclsyntax.ParseExpression([]byte(`FOO`), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	_, err := eval.EvalString(expr)
	require.Error(t, err)
}

func TestInvalidTaskReference(t *testing.T) {
	tasks := ast.NewTasks()
	tasks.Set("known", &ast.Task{})
	eval := hclext.NewHCLEvaluator(ast.NewVars(), env.GetEnviron(), tasks)
	expr, diags := hclsyntax.ParseExpression([]byte(`task.unknown`), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	_, err := eval.EvalString(expr)
	require.Error(t, err)
	require.Contains(t, err.Error(), "test.hcl")
}
