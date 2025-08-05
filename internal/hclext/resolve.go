package hclext

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/go-task/task/v3/taskfile/ast"
)

// Resolver evaluates HCL expressions for vars and env allowing recursive references.
// It resolves variables on demand, detecting cycles and caching results.
type Resolver struct {
	vars     *ast.Vars
	env      *ast.Vars
	runner   TaskRunner
	varCache map[string]string
	envCache map[string]string
	varStack map[string]bool
	envStack map[string]bool
}

// NewResolver creates a new Resolver.
func NewResolver(vars, env *ast.Vars, runner TaskRunner) *Resolver {
	r := &Resolver{
		vars:     vars,
		env:      env,
		runner:   runner,
		varCache: map[string]string{},
		envCache: map[string]string{},
		varStack: map[string]bool{},
		envStack: map[string]bool{},
	}
	if vars != nil {
		for k, v := range vars.All() {
			if v.Expr == nil {
				if v.Value != nil {
					r.varCache[k] = fmt.Sprint(v.Value)
				}
			}
		}
	}
	if env != nil {
		for k, v := range env.All() {
			if v.Expr == nil {
				if v.Value != nil {
					r.envCache[k] = fmt.Sprint(v.Value)
				}
			}
		}
	}
	return r
}

// Resolve evaluates all expressions and returns new vars and env with values set.
func (r *Resolver) Resolve() (*ast.Vars, *ast.Vars, error) {
	if r.vars != nil {
		for k := range r.vars.All() {
			if _, ok := r.varCache[k]; !ok {
				if _, err := r.resolveVar(k); err != nil {
					return nil, nil, err
				}
			}
		}
	}
	if r.env != nil {
		for k := range r.env.All() {
			if _, ok := r.envCache[k]; !ok {
				if _, err := r.resolveEnv(k); err != nil {
					return nil, nil, err
				}
			}
		}
	}
	vars := ast.NewVars()
	for k, v := range r.varCache {
		vars.Set(k, ast.Var{Value: v})
	}
	env := ast.NewVars()
	for k, v := range r.envCache {
		env.Set(k, ast.Var{Value: v})
	}
	return vars, env, nil
}

func (r *Resolver) resolveVar(name string) (string, error) {
	if v, ok := r.varCache[name]; ok {
		return v, nil
	}
	if r.varStack[name] {
		return "", fmt.Errorf("cyclic variable reference for %s", name)
	}
	if r.vars == nil {
		return "", fmt.Errorf("undefined variable %s", name)
	}
	v, ok := r.vars.Get(name)
	if !ok {
		return "", fmt.Errorf("undefined variable %s", name)
	}
	if v.Expr == nil {
		val := fmt.Sprint(v.Value)
		r.varCache[name] = val
		return val, nil
	}
	r.varStack[name] = true
	defer delete(r.varStack, name)
	depsVars, depsEnv := findDeps(v.Expr)
	for dv := range depsVars {
		if _, err := r.resolveVar(dv); err != nil {
			return "", err
		}
	}
	for de := range depsEnv {
		if _, err := r.resolveEnv(de); err != nil {
			return "", err
		}
	}
	eval := NewHCLEvaluator(varsFromCache(r.varCache), envFromCache(r.envCache), r.runner)
	val, err := eval.EvalString(v.Expr)
	if err != nil {
		return "", err
	}
	r.varCache[name] = val
	return val, nil
}

func (r *Resolver) resolveEnv(name string) (string, error) {
	if v, ok := r.envCache[name]; ok {
		return v, nil
	}
	if r.envStack[name] {
		return "", fmt.Errorf("cyclic env reference for %s", name)
	}
	if r.env != nil {
		if v, ok := r.env.Get(name); ok {
			if v.Expr == nil {
				val := fmt.Sprint(v.Value)
				r.envCache[name] = val
				return val, nil
			}
			r.envStack[name] = true
			defer delete(r.envStack, name)
			depsVars, depsEnv := findDeps(v.Expr)
			for dv := range depsVars {
				if _, err := r.resolveVar(dv); err != nil {
					return "", err
				}
			}
			for de := range depsEnv {
				if _, err := r.resolveEnv(de); err != nil {
					return "", err
				}
			}
			eval := NewHCLEvaluator(varsFromCache(r.varCache), envFromCache(r.envCache), r.runner)
			val, err := eval.EvalString(v.Expr)
			if err != nil {
				return "", err
			}
			r.envCache[name] = val
			return val, nil
		}
	}
	// Not defined; return empty string
	r.envCache[name] = ""
	return "", nil
}

func varsFromCache(m map[string]string) *ast.Vars {
	vs := ast.NewVars()
	for k, v := range m {
		vs.Set(k, ast.Var{Value: v})
	}
	return vs
}

func envFromCache(m map[string]string) *ast.Vars {
	vs := ast.NewVars()
	for k, v := range m {
		vs.Set(k, ast.Var{Value: v})
	}
	return vs
}

func findDeps(expr hcl.Expression) (vars map[string]struct{}, env map[string]struct{}) {
	vars = map[string]struct{}{}
	env = map[string]struct{}{}
	if expr == nil {
		return
	}
	for _, tr := range expr.Variables() {
		if len(tr) != 2 {
			continue
		}
		root := tr.RootName()
		attr, ok := tr[1].(hcl.TraverseAttr)
		if !ok {
			continue
		}
		switch root {
		case "vars":
			vars[attr.Name] = struct{}{}
		case "env":
			env[attr.Name] = struct{}{}
		}
	}
	return
}

// Helper to satisfy linter for unused imports
var _ = cty.String
