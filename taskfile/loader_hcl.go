package taskfile

import (
	"bytes"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/go-task/task/v3/errors"
	"github.com/go-task/task/v3/internal/filepathext"
	"github.com/go-task/task/v3/taskfile/ast"
)

// Load parses the given data as HCL into a Taskfile structure.
func (HCLLoader) Load(data []byte, location string) (*ast.Taskfile, error) {
	if bytes.Contains(data, []byte("{{")) {
		return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: fmt.Errorf("go templates are not supported in HCL Taskfiles")}
	}
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(data, location)
	if diags.HasErrors() {
		return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: diags}
	}

	schema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "version", Required: true},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "vars"},
			{Type: "env"},
			{Type: "task", LabelNames: []string{"name"}},
		},
	}

	content, diags := file.Body.Content(schema)
	if diags.HasErrors() {
		return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: diags}
	}

	versionVal, diags := content.Attributes["version"].Expr.Value(nil)
	if diags.HasErrors() {
		return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: diags}
	}
	version, err := semver.NewVersion(versionVal.AsString())
	if err != nil {
		return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: err}
	}

	tf := &ast.Taskfile{Version: version, Tasks: ast.NewTasks()}

	if blocks := content.Blocks.OfType("vars"); len(blocks) > 0 {
		vars, err := parseVarsBlock(blocks[0], location)
		if err != nil {
			return nil, err
		}
		tf.Vars = vars
	}
	if blocks := content.Blocks.OfType("env"); len(blocks) > 0 {
		env, err := parseVarsBlock(blocks[0], location)
		if err != nil {
			return nil, err
		}
		tf.Env = env
	}

	for _, block := range content.Blocks.OfType("task") {
		task, err := parseTask(block, location)
		if err != nil {
			return nil, err
		}
		tf.Tasks.Set(task.Task, task)
	}

	if tf.Vars == nil {
		tf.Vars = ast.NewVars()
	}
	if tf.Env == nil {
		tf.Env = ast.NewVars()
	}

	return tf, nil
}

func parseTask(block *hcl.Block, location string) (*ast.Task, error) {
	t := &ast.Task{
		Task:  block.Labels[0],
		Cmds:  []*ast.Cmd{},
		Vars:  ast.NewVars(),
		Env:   ast.NewVars(),
		IsHCL: true,
	}

	schema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "desc"},
			{Name: "cmds"},
			{Name: "deps"},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "vars"},
			{Type: "env"},
		},
	}
	content, diags := block.Body.Content(schema)
	if diags.HasErrors() {
		return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: diags}
	}

	if attr, ok := content.Attributes["desc"]; ok {
		var desc string
		diags := gohcl.DecodeExpression(attr.Expr, nil, &desc)
		if diags.HasErrors() {
			return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: diags}
		}
		t.Desc = desc
	}

	if attr, ok := content.Attributes["cmds"]; ok {
		if tuple, ok := attr.Expr.(*hclsyntax.TupleConsExpr); ok {
			for _, expr := range tuple.Exprs {
				t.Cmds = append(t.Cmds, &ast.Cmd{Expr: expr})
			}
		} else {
			t.Cmds = append(t.Cmds, &ast.Cmd{Expr: attr.Expr})
		}
	}

	if attr, ok := content.Attributes["deps"]; ok {
		deps, err := parseDeps(attr.Expr, location)
		if err != nil {
			return nil, err
		}
		t.Deps = deps
	}

	if blocks := content.Blocks.OfType("vars"); len(blocks) > 0 {
		vars, err := parseVarsBlock(blocks[0], location)
		if err != nil {
			return nil, err
		}
		t.Vars = vars
	}
	if blocks := content.Blocks.OfType("env"); len(blocks) > 0 {
		env, err := parseVarsBlock(blocks[0], location)
		if err != nil {
			return nil, err
		}
		t.Env = env
	}

	return t, nil
}

func parseVarsExpr(expr hcl.Expression, location string) (*ast.Vars, error) {
	obj, ok := expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: hcl.Diagnostics{}}
	}
	vars := ast.NewVars()
	for _, item := range obj.Items {
		key, diags := objectKey(item.KeyExpr)
		if diags.HasErrors() {
			return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: diags}
		}
		v := ast.Var{}
		if inner, ok := item.ValueExpr.(*hclsyntax.ObjectConsExpr); ok {
			for _, innerItem := range inner.Items {
				attrKey, diags := objectKey(innerItem.KeyExpr)
				if diags.HasErrors() {
					return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: diags}
				}
				if attrKey == "sh" {
					v.ShExpr = innerItem.ValueExpr
				}
			}
		} else {
			v.Expr = item.ValueExpr
		}
		vars.Set(key, v)
	}
	return vars, nil
}

func parseVarsBlock(b *hcl.Block, location string) (*ast.Vars, error) {
	body, ok := b.Body.(*hclsyntax.Body)
	if !ok {
		return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: hcl.Diagnostics{}}
	}
	vars := ast.NewVars()
	for name, attr := range body.Attributes {
		v := ast.Var{}
		if obj, ok := attr.Expr.(*hclsyntax.ObjectConsExpr); ok {
			for _, inner := range obj.Items {
				attrKey, diags := objectKey(inner.KeyExpr)
				if diags.HasErrors() {
					return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: diags}
				}
				if attrKey == "sh" {
					v.ShExpr = inner.ValueExpr
				}
			}
		} else {
			v.Expr = attr.Expr
		}
		vars.Set(name, v)
	}
	return vars, nil
}

func parseDeps(expr hcl.Expression, location string) ([]*ast.Dep, error) {
	tuple, ok := expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: hcl.Diagnostics{}}
	}
	deps := make([]*ast.Dep, 0, len(tuple.Exprs))
	for _, e := range tuple.Exprs {
		call, ok := e.(*hclsyntax.FunctionCallExpr)
		if !ok || call.Name != "task" || len(call.Args) == 0 {
			return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: hcl.Diagnostics{}}
		}
		var name string
		diags := gohcl.DecodeExpression(call.Args[0], nil, &name)
		if diags.HasErrors() {
			return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: diags}
		}
		dep := &ast.Dep{Task: name}
		if len(call.Args) > 1 {
			vars, err := parseVarsExpr(call.Args[1], location)
			if err != nil {
				return nil, err
			}
			dep.Vars = vars
		}
		deps = append(deps, dep)
	}
	return deps, nil
}

func objectKey(expr hcl.Expression) (string, hcl.Diagnostics) {
	var key string
	diags := gohcl.DecodeExpression(expr, nil, &key)
	return key, diags
}
