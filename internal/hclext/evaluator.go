package hclext

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	"github.com/go-task/task/v3/internal/filepathext"
	"github.com/go-task/task/v3/taskfile/ast"
)

type TaskRunner func(name string, vars *ast.Vars) (string, error)

type HCLEvaluator struct {
	EvalCtx *hcl.EvalContext
	vars    map[string]cty.Value
	env     map[string]cty.Value
}

func NewHCLEvaluator(vars, env *ast.Vars, runner TaskRunner) *HCLEvaluator {
	varVals := map[string]cty.Value{}
	if vars != nil {
		for k, v := range vars.All() {
			if v.Value != nil {
				varVals[k] = toCty(v.Value)
			}
		}
	}
	envVals := map[string]cty.Value{}
	if env != nil {
		for k, v := range env.All() {
			if v.Value != nil {
				envVals[k] = toCty(v.Value)
			}
		}
	}
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"vars": cty.ObjectVal(varVals),
			"env":  cty.ObjectVal(envVals),
		},
		Functions: builtinFunctions(runner),
	}
	return &HCLEvaluator{EvalCtx: ctx, vars: varVals, env: envVals}
}

func builtinFunctions(runner TaskRunner) map[string]function.Function {
	funcs := map[string]function.Function{
		"upper": stringFunc(strings.ToUpper),
		"lower": stringFunc(strings.ToLower),
		"join":  joinFunc(),
		"split": splitFunc(),
		"env":   envFunc(),
		"sh":    shellFunc("/bin/sh"),
		"bash":  shellFunc("/bin/bash"),
		"zsh":   shellFunc("/bin/zsh"),
		"tuple": tupleFunc(),
	}
	if runner != nil {
		funcs["task"] = taskFunc(runner)
	}
	return funcs
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

func tupleFunc() function.Function {
	return function.New(&function.Spec{
		VarParam: &function.Parameter{Name: "vals", Type: cty.DynamicPseudoType},
		Type: func(args []cty.Value) (cty.Type, error) {
			types := make([]cty.Type, len(args))
			for i, v := range args {
				types[i] = v.Type()
			}
			return cty.Tuple(types), nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return cty.TupleVal(args), nil
		},
	})
}

func toCty(v any) cty.Value {
	switch val := v.(type) {
	case string:
		return cty.StringVal(val)
	case bool:
		return cty.BoolVal(val)
	case int:
		return cty.NumberIntVal(int64(val))
	case int64:
		return cty.NumberIntVal(val)
	case float32:
		return cty.NumberFloatVal(float64(val))
	case float64:
		return cty.NumberFloatVal(val)
	case []any:
		vals := make([]cty.Value, len(val))
		for i, e := range val {
			vals[i] = toCty(e)
		}
		return cty.TupleVal(vals)
	case map[string]any:
		attrs := make(map[string]cty.Value)
		for k, e := range val {
			attrs[k] = toCty(e)
		}
		return cty.ObjectVal(attrs)
	default:
		return cty.StringVal(fmt.Sprint(v))
	}
}

func fromCty(val cty.Value, expr hcl.Expression) (any, error) {
	switch {
	case val.Type() == cty.String:
		return val.AsString(), nil
	case val.Type() == cty.Number:
		bf := val.AsBigFloat()
		if i, acc := bf.Int64(); acc == 1 {
			return i, nil
		}
		f, _ := bf.Float64()
		return f, nil
	case val.Type() == cty.Bool:
		return val.True(), nil
	case val.Type().IsObjectType():
		attrs := make(map[string]any)
		for k := range val.Type().AttributeTypes() {
			v := val.GetAttr(k)
			res, err := fromCty(v, expr)
			if err != nil {
				return nil, err
			}
			attrs[k] = res
		}
		return attrs, nil
	case val.Type().IsTupleType():
		if _, ok := expr.(*hclsyntax.TupleConsExpr); ok {
			vals := val.AsValueSlice()
			res := make([]any, len(vals))
			for i, v := range vals {
				r, err := fromCty(v, expr)
				if err != nil {
					return nil, err
				}
				res[i] = r
			}
			return res, nil
		}
		rng := expr.Range()
		return nil, fmt.Errorf("unsupported value type tuple for %s:%d,%d-%d,%d", filepathext.TryAbsToRel(rng.Filename), rng.Start.Line, rng.Start.Column, rng.End.Line, rng.End.Column)
	default:
		rng := expr.Range()
		return nil, fmt.Errorf("unsupported value type %s for %s:%d,%d-%d,%d", val.Type().FriendlyName(), filepathext.TryAbsToRel(rng.Filename), rng.Start.Line, rng.Start.Column, rng.End.Line, rng.End.Column)
	}
}

func shellFunc(shell string) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{{Name: "cmd", Type: cty.String}},
		Type:   function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			cmd := exec.Command(shell, "-c", args[0].AsString())
			out, err := cmd.CombinedOutput()
			if err != nil {
				return cty.NilVal, fmt.Errorf("%s: %w", shell, err)
			}
			return cty.StringVal(strings.TrimSpace(string(out))), nil
		},
	})
}

func taskFunc(runner TaskRunner) function.Function {
	return function.New(&function.Spec{
		Params:   []function.Parameter{{Name: "name", Type: cty.String}},
		VarParam: &function.Parameter{Name: "vars", Type: cty.DynamicPseudoType},
		Type:     function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			name := args[0].AsString()
			var depVars *ast.Vars
			if len(args) > 1 && args[1].Type().IsObjectType() {
				depVars = ast.NewVars()
				for k := range args[1].Type().AttributeTypes() {
					val := args[1].GetAttr(k)
					var depVal string
					switch {
					case val.Type() == cty.String:
						depVal = val.AsString()
					case val.Type() == cty.Number:
						bf := val.AsBigFloat()
						depVal = bf.Text('f', -1)
					case val.Type() == cty.Bool:
						if val.True() {
							depVal = "true"
						} else {
							depVal = "false"
						}
					default:
						depVal = val.GoString()
					}
					depVars.Set(k, ast.Var{Value: depVal})
				}
			}
			out, err := runner(name, depVars)
			if err != nil {
				return cty.NilVal, err
			}
			return cty.StringVal(out), nil
		},
	})
}

func (e *HCLEvaluator) SetVar(name, value string) {
	if e.vars == nil {
		e.vars = map[string]cty.Value{}
	}
	e.vars[name] = cty.StringVal(value)
	e.EvalCtx.Variables["vars"] = cty.ObjectVal(e.vars)
}

func (e *HCLEvaluator) EvalValue(expr hcl.Expression) (any, error) {
	val, diags := expr.Value(e.EvalCtx)
	if diags.HasErrors() {
		return nil, diags
	}
	return fromCty(val, expr)
}

func (e *HCLEvaluator) EvalString(expr hcl.Expression) (string, error) {
	v, err := e.EvalValue(expr)
	if err != nil {
		return "", err
	}
	switch val := v.(type) {
	case string:
		return val, nil
	case int:
		return fmt.Sprintf("%d", val), nil
	case int64:
		return fmt.Sprintf("%d", val), nil
	case float64:
		return fmt.Sprintf("%v", val), nil
	case bool:
		if val {
			return "true", nil
		}
		return "false", nil
	default:
		rng := expr.Range()
		return "", fmt.Errorf("unsupported value type %T for %s:%d,%d-%d,%d", v, filepathext.TryAbsToRel(rng.Filename), rng.Start.Line, rng.Start.Column, rng.End.Line, rng.End.Column)
	}
}
