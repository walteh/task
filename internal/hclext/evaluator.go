package hclext

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	"github.com/go-task/task/v3/taskfile/ast"
)

type HCLEvaluator struct {
	EvalCtx *hcl.EvalContext
}

func NewHCLEvaluator(vars *ast.Vars) *HCLEvaluator {
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: builtinFunctions(),
	}
	if vars != nil {
		for k, v := range vars.All() {
			if v.Value != nil {
				ctx.Variables[k] = cty.StringVal(fmt.Sprint(v.Value))
			}
		}
	}
	return &HCLEvaluator{EvalCtx: ctx}
}

func builtinFunctions() map[string]function.Function {
	return map[string]function.Function{
		"upper": stringFunc(strings.ToUpper),
		"lower": stringFunc(strings.ToLower),
		"join":  joinFunc(),
		"split": splitFunc(),
		"env":   envFunc(),
	}
}

func stringFunc(fn func(string) string) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{{Name: "s", Type: cty.String}},
		Type:   function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return cty.StringVal(fn(args[0].AsString())), nil
		},
	})
}

func joinFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{Name: "list", Type: cty.List(cty.String)},
			{Name: "delim", Type: cty.String},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			vals := args[0].AsValueSlice()
			parts := make([]string, len(vals))
			for i, v := range vals {
				parts[i] = v.AsString()
			}
			return cty.StringVal(strings.Join(parts, args[1].AsString())), nil
		},
	})
}

func splitFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{Name: "s", Type: cty.String},
			{Name: "delim", Type: cty.String},
		},
		Type: func(args []cty.Value) (cty.Type, error) {
			return cty.List(cty.String), nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			parts := strings.Split(args[0].AsString(), args[1].AsString())
			vals := make([]cty.Value, len(parts))
			for i, p := range parts {
				vals[i] = cty.StringVal(p)
			}
			return cty.ListVal(vals), nil
		},
	})
}

func envFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{{Name: "name", Type: cty.String}},
		Type:   function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return cty.StringVal(os.Getenv(args[0].AsString())), nil
		},
	})
}

func (e *HCLEvaluator) SetVar(name, value string) {
	if e.EvalCtx.Variables == nil {
		e.EvalCtx.Variables = map[string]cty.Value{}
	}
	e.EvalCtx.Variables[name] = cty.StringVal(value)
}

func (e *HCLEvaluator) EvalString(expr hcl.Expression) (string, error) {
	val, diags := expr.Value(e.EvalCtx)
	if diags.HasErrors() {
		return "", diags
	}
	switch {
	case val.Type() == cty.String:
		return val.AsString(), nil
	case val.Type() == cty.Number:
		bf := val.AsBigFloat()
		return bf.Text('f', -1), nil
	case val.Type() == cty.Bool:
		if val.True() {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("unsupported value type %s", val.Type().FriendlyName())
	}
}
