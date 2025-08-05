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

func parseExpr(t *testing.T, s string) hcl.Expression {
	t.Helper()
	expr, diags := hclsyntax.ParseExpression([]byte(s), "test.hcl", hcl.InitialPos)
	require.False(t, diags.HasErrors())
	return expr
}

func TestRecursiveVars(t *testing.T) {
	vars := ast.NewVars()
	vars.Set("GREETING", ast.Var{Expr: parseExpr(t, `"Hello, ${vars.NAME}!"`)})
	vars.Set("NAME", ast.Var{Expr: parseExpr(t, `"BOB"`)})
	vars.Set("UPPER_GREETING", ast.Var{Expr: parseExpr(t, `upper(vars.GREETING)`)})

	resolver := hclext.NewResolver(vars, env.GetEnviron(), nil, nil)
	resolved, _, err := resolver.Resolve()
	require.NoError(t, err)

	g, _ := resolved.Get("GREETING")
	require.Equal(t, "Hello, BOB!", g.Value)
	u, _ := resolved.Get("UPPER_GREETING")
	require.Equal(t, "HELLO, BOB!", u.Value)
}

func TestOrderIndependence(t *testing.T) {
	vars := ast.NewVars()
	vars.Set("FINAL", ast.Var{Expr: parseExpr(t, `upper(vars.INTERMEDIATE)`)})
	vars.Set("INTERMEDIATE", ast.Var{Expr: parseExpr(t, `"${vars.BASE} + ok"`)})
	vars.Set("BASE", ast.Var{Expr: parseExpr(t, `"yup"`)})

	resolver := hclext.NewResolver(vars, env.GetEnviron(), nil, nil)
	resolved, _, err := resolver.Resolve()
	require.NoError(t, err)

	v, _ := resolved.Get("FINAL")
	require.Equal(t, "YUP + OK", v.Value)
}

func TestCyclicReference(t *testing.T) {
	vars := ast.NewVars()
	vars.Set("LOOP", ast.Var{Expr: parseExpr(t, `"${vars.LOOP}"`)})

	resolver := hclext.NewResolver(vars, env.GetEnviron(), nil, nil)
	_, _, err := resolver.Resolve()
	require.Error(t, err)
}
