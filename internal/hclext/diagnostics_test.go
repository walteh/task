package hclext

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/require"

	"github.com/go-task/task/v3/taskfile/ast"
)

func TestUnsupportedTupleErrorIncludesRange(t *testing.T) {
	expr, diags := hclsyntax.ParseExpression([]byte(`tuple("a", "b")`), "Taskfile.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	vars := ast.NewVars()
	vars.Set("INVALID", ast.Var{Expr: expr})
	r := NewResolver(vars, nil, nil)
	_, _, err := r.Resolve()
	require.Error(t, err)
	rng := expr.Range()
	expected := fmt.Sprintf("Taskfile.hcl:%d,%d-%d,%d", rng.Start.Line, rng.Start.Column, rng.End.Line, rng.End.Column)
	require.Contains(t, err.Error(), "unsupported value type tuple")
	require.Contains(t, err.Error(), expected)
}
