package taskfile

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"

	"github.com/go-task/task/v3/errors"
	"github.com/go-task/task/v3/internal/filepathext"
	"github.com/go-task/task/v3/taskfile/ast"
)

// HCLLoader implements TaskfileLoader for HCL format
type HCLLoader struct{}

// NewHCLLoader creates a new HCLLoader instance
func NewHCLLoader() *HCLLoader {
	return &HCLLoader{}
}

// LoadTaskfile loads and parses an HCL Taskfile using low-level parsing
func (l *HCLLoader) LoadTaskfile(ctx context.Context, node Node, content []byte) (*ast.Taskfile, error) {
	// Parse the HCL file
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(content, node.Location())
	if diags.HasErrors() {
		return nil, l.formatHCLError(node.Location(), diags)
	}

	// Parse HCL using low-level block/attribute extraction
	hclData, err := l.parseHCLBody(file.Body, node.Location())
	if err != nil {
		return nil, err
	}

	// Convert to internal Taskfile format
	taskfile, err := l.convertToTaskfile(hclData, node.Location())
	if err != nil {
		return nil, err
	}

	return taskfile, nil
}

// SupportsExtension returns true if the loader supports the given file extension
func (l *HCLLoader) SupportsExtension(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".hcl" || ext == ""  // Support both .hcl and extensionless files
}

// FormatName returns the name of the format this loader supports
func (l *HCLLoader) FormatName() string {
	return "HCL"
}

// formatHCLError converts HCL diagnostics to a user-friendly error
func (l *HCLLoader) formatHCLError(filename string, diags hcl.Diagnostics) error {
	if len(diags) == 0 {
		return fmt.Errorf("unknown HCL error in file: %s", filename)
	}
	
	// Use the first diagnostic for the error message
	diag := diags[0]
	var location string
	if diag.Subject != nil {
		location = fmt.Sprintf("%s:%d:%d", filename, diag.Subject.Start.Line, diag.Subject.Start.Column)
	} else {
		location = filename
	}
	
	return &errors.TaskfileInvalidError{
		URI: filepathext.TryAbsToRel(filename),
		Err: fmt.Errorf("HCL parse error at %s: %s", location, diag.Summary),
	}
}

// parseHCLBody parses the HCL body using low-level extraction
func (l *HCLLoader) parseHCLBody(body hcl.Body, location string) (*HCLTaskfileData, error) {
	hclData := &HCLTaskfileData{
		Tasks: []*HCLTaskData{},
	}

	// Get the content to access both attributes and blocks with proper schema
	content, _, diags := body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "version"},
			{Name: "output"},
			{Name: "method"},
			{Name: "run"},
			{Name: "silent"},
			{Name: "set"},
			{Name: "shopt"},
			{Name: "dotenv"},
			{Name: "vars"},
			{Name: "env"},
			{Name: "includes"},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "task", LabelNames: []string{"name"}},
		},
	})
	if diags.HasErrors() {
		return nil, l.formatHCLError(location, diags)
	}

	// Parse top-level attributes
	for name, attr := range content.Attributes {
		switch name {
		case "version":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.String {
				versionStr := val.AsString()
				hclData.Version = &versionStr
			}
		case "output":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.String {
				str := val.AsString()
				hclData.Output = &str
			}
		case "method":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.String {
				str := val.AsString()
				hclData.Method = &str
			}
		case "run":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.String {
				str := val.AsString()
				hclData.Run = &str
			}
		case "silent":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.Bool {
				bool := val.True()
				hclData.Silent = &bool
			}
		case "set":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type().IsTupleType() {
				hclData.Set = l.extractStringSlice(val)
			}
		case "shopt":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type().IsTupleType() {
				hclData.Shopt = l.extractStringSlice(val)
			}
		case "dotenv":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type().IsTupleType() {
				hclData.Dotenv = l.extractStringSlice(val)
			}
		case "vars":
			// For now, skip vars parsing at global level
			// TODO: Implement proper vars parsing with expression capture
		case "env":
			// For now, skip env parsing at global level
			// TODO: Implement proper env parsing with expression capture
		case "includes":
			// For now, skip includes parsing at global level
			// TODO: Implement proper includes parsing with expression capture
		}
	}

	// Parse task blocks
	for _, block := range content.Blocks {
		if block.Type == "task" && len(block.Labels) > 0 {
			taskData, err := l.parseTaskBlock(block, location)
			if err != nil {
				return nil, err
			}
			hclData.Tasks = append(hclData.Tasks, taskData)
		}
	}

	return hclData, nil
}

// parseTaskBlock parses a task block
func (l *HCLLoader) parseTaskBlock(block *hcl.Block, location string) (*HCLTaskData, error) {
	taskData := &HCLTaskData{
		Name: block.Labels[0],
	}

	// Parse the task block content with proper schema
	content, _, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "desc"}, {Name: "label"}, {Name: "summary"}, {Name: "dir"},
			{Name: "method"}, {Name: "prefix"}, {Name: "run"},
			{Name: "silent"}, {Name: "interactive"}, {Name: "internal"},
			{Name: "ignore_error"}, {Name: "watch"},
			{Name: "aliases"}, {Name: "sources"}, {Name: "generates"},
			{Name: "status"}, {Name: "set"}, {Name: "shopt"},
			{Name: "dotenv"}, {Name: "platforms"},
			{Name: "cmds"}, {Name: "deps"}, {Name: "vars"}, {Name: "env"},
		},
	})
	if diags.HasErrors() {
		return nil, l.formatHCLError(location, diags)
	}

	attrs := content.Attributes

	// Parse task attributes
	for name, attr := range attrs {
		switch name {
		case "desc":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.String {
				str := val.AsString()
				taskData.Desc = &str
			}
		case "label":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.String {
				str := val.AsString()
				taskData.Label = &str
			}
		case "summary":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.String {
				str := val.AsString()
				taskData.Summary = &str
			}
		case "dir":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.String {
				str := val.AsString()
				taskData.Dir = &str
			}
		case "method":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.String {
				str := val.AsString()
				taskData.Method = &str
			}
		case "prefix":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.String {
				str := val.AsString()
				taskData.Prefix = &str
			}
		case "run":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.String {
				str := val.AsString()
				taskData.Run = &str
			}
		case "silent":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.Bool {
				bool := val.True()
				taskData.Silent = &bool
			}
		case "interactive":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.Bool {
				bool := val.True()
				taskData.Interactive = &bool
			}
		case "internal":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.Bool {
				bool := val.True()
				taskData.Internal = &bool
			}
		case "ignore_error":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.Bool {
				bool := val.True()
				taskData.IgnoreError = &bool
			}
		case "watch":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type() == cty.Bool {
				bool := val.True()
				taskData.Watch = &bool
			}
		case "aliases":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type().IsTupleType() {
				taskData.Aliases = l.extractStringSlice(val)
			}
		case "sources":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type().IsTupleType() {
				taskData.Sources = l.extractStringSlice(val)
			}
		case "generates":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type().IsTupleType() {
				taskData.Generates = l.extractStringSlice(val)
			}
		case "status":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type().IsTupleType() {
				taskData.Status = l.extractStringSlice(val)
			}
		case "set":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type().IsTupleType() {
				taskData.Set = l.extractStringSlice(val)
			}
		case "shopt":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type().IsTupleType() {
				taskData.Shopt = l.extractStringSlice(val)
			}
		case "dotenv":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type().IsTupleType() {
				taskData.Dotenv = l.extractStringSlice(val)
			}
		case "platforms":
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return nil, l.formatHCLError(location, diags)
			}
			if val.Type().IsTupleType() {
				taskData.Platforms = l.extractStringSlice(val)
			}
		case "cmds":
			// Store the raw expression for commands for later evaluation
			taskData.Cmds = attr.Expr
		case "deps":
			// Store the raw expression for dependencies for later evaluation
			taskData.Deps = attr.Expr
		case "vars":
			// For now, skip vars parsing at task level
			// TODO: Implement proper vars parsing with expression capture
		case "env":
			// For now, skip env parsing at task level
			// TODO: Implement proper env parsing with expression capture
		}
	}

	return taskData, nil
}

// extractStringSlice extracts a string slice from a cty.Value
func (l *HCLLoader) extractStringSlice(val cty.Value) []string {
	if !val.Type().IsTupleType() {
		return nil
	}
	
	var result []string
	for it := val.ElementIterator(); it.Next(); {
		_, elem := it.Element()
		if elem.Type() == cty.String {
			result = append(result, elem.AsString())
		}
	}
	return result
}

// convertToTaskfile converts HCLTaskfileData to the internal ast.Taskfile format
func (l *HCLLoader) convertToTaskfile(hclData *HCLTaskfileData, location string) (*ast.Taskfile, error) {
	// Parse version string
	var version *semver.Version
	if hclData.Version != nil && *hclData.Version != "" {
		var err error
		version, err = semver.NewVersion(*hclData.Version)
		if err != nil {
			return nil, &errors.TaskfileInvalidError{
				URI: filepathext.TryAbsToRel(location),
				Err: fmt.Errorf("invalid version format: %s", *hclData.Version),
			}
		}
	}

	// Create the internal Taskfile
	tf := &ast.Taskfile{
		Location: location,
		Version:  version,
		Tasks:    ast.NewTasks(),
		Vars:     ast.NewVars(),
		Env:      ast.NewVars(),
		Includes: ast.NewIncludes(),
	}

	// Set global attributes
	if hclData.Output != nil {
		tf.Output = ast.Output{Name: *hclData.Output}
	}
	if hclData.Method != nil {
		tf.Method = *hclData.Method
	}
	if hclData.Run != nil {
		tf.Run = *hclData.Run
	}
	if hclData.Silent != nil {
		tf.Silent = *hclData.Silent
	}
	if hclData.Set != nil {
		tf.Set = hclData.Set
	}
	if hclData.Shopt != nil {
		tf.Shopt = hclData.Shopt
	}
	if hclData.Dotenv != nil {
		tf.Dotenv = hclData.Dotenv
	}

	// Check that version is set
	if tf.Version == nil {
		return nil, &errors.TaskfileVersionCheckError{URI: location}
	}

	// TODO: Parse global vars, env, and includes
	// For now, these are skipped as they require more complex expression parsing

	// Convert tasks
	for _, hclTask := range hclData.Tasks {
		task := &ast.Task{
			Task: hclTask.Name,
			Location: &ast.Location{
				Taskfile: location,
			},
			Vars: ast.NewVars(),
			Env:  ast.NewVars(),
		}

		// Set string attributes
		if hclTask.Desc != nil {
			task.Desc = *hclTask.Desc
		}
		if hclTask.Label != nil {
			task.Label = *hclTask.Label
		}
		if hclTask.Summary != nil {
			task.Summary = *hclTask.Summary
		}
		if hclTask.Dir != nil {
			task.Dir = *hclTask.Dir
		}
		if hclTask.Method != nil {
			task.Method = *hclTask.Method
		}
		if hclTask.Prefix != nil {
			task.Prefix = *hclTask.Prefix
		}
		if hclTask.Run != nil {
			task.Run = *hclTask.Run
		}

		// Set boolean attributes
		if hclTask.Silent != nil {
			task.Silent = *hclTask.Silent
		}
		if hclTask.Interactive != nil {
			task.Interactive = *hclTask.Interactive
		}
		if hclTask.Internal != nil {
			task.Internal = *hclTask.Internal
		}
		if hclTask.IgnoreError != nil {
			task.IgnoreError = *hclTask.IgnoreError
		}
		if hclTask.Watch != nil {
			task.Watch = *hclTask.Watch
		}

		// Set string slice attributes
		task.Aliases = hclTask.Aliases
		task.Set = hclTask.Set
		task.Shopt = hclTask.Shopt
		task.Dotenv = hclTask.Dotenv
		task.Status = hclTask.Status

		// Convert sources and generates to Glob objects
		for _, source := range hclTask.Sources {
			task.Sources = append(task.Sources, &ast.Glob{Glob: source})
		}
		for _, generate := range hclTask.Generates {
			task.Generates = append(task.Generates, &ast.Glob{Glob: generate})
		}

		// Convert platforms
		for _, platform := range hclTask.Platforms {
			task.Platforms = append(task.Platforms, &ast.Platform{OS: platform})
		}

		// Parse commands (for now, basic implementation - TODO: enhance for expressions)
		if hclTask.Cmds != nil {
			cmds, err := l.parseCmdsFromExpression(hclTask.Cmds)
			if err != nil {
				return nil, fmt.Errorf("error parsing commands for task %s: %w", hclTask.Name, err)
			}
			task.Cmds = cmds
		}

		// Parse dependencies (for now, basic implementation - TODO: enhance for expressions)
		if hclTask.Deps != nil {
			deps, err := l.parseDepsFromExpression(hclTask.Deps)
			if err != nil {
				return nil, fmt.Errorf("error parsing dependencies for task %s: %w", hclTask.Name, err)
			}
			task.Deps = deps
		}

		// TODO: Parse task vars and env
		// For now, these are skipped as they require more complex expression parsing

		// Add task to the Tasks map
		tf.Tasks.Set(hclTask.Name, task)
	}

	return tf, nil
}

// parseCmdsFromExpression parses commands from an HCL expression
func (l *HCLLoader) parseCmdsFromExpression(expr hcl.Expression) ([]*ast.Cmd, error) {
	if expr == nil {
		return nil, nil
	}

	// For now, evaluate the expression to get a list of strings
	// TODO: Enhance this to capture expressions for runtime evaluation
	val, diags := expr.Value(nil)
	if diags.HasErrors() {
		return nil, fmt.Errorf("error evaluating cmds expression: %s", diags.Error())
	}

	var cmds []*ast.Cmd
	if val.Type().IsTupleType() {
		for it := val.ElementIterator(); it.Next(); {
			_, elem := it.Element()
			if elem.Type() == cty.String {
				cmd := &ast.Cmd{
					Cmd: elem.AsString(),
				}
				cmds = append(cmds, cmd)
			} else if elem.Type().IsObjectType() {
				// Handle task calls or complex commands
				// For now, just skip - TODO: implement proper task call parsing
			}
		}
	} else if val.Type() == cty.String {
		// Single command
		cmd := &ast.Cmd{
			Cmd: val.AsString(),
		}
		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

// parseDepsFromExpression parses dependencies from an HCL expression
func (l *HCLLoader) parseDepsFromExpression(expr hcl.Expression) ([]*ast.Dep, error) {
	if expr == nil {
		return nil, nil
	}

	// For now, evaluate the expression to get a list of strings
	// TODO: Enhance this to capture expressions for runtime evaluation
	val, diags := expr.Value(nil)
	if diags.HasErrors() {
		return nil, fmt.Errorf("error evaluating deps expression: %s", diags.Error())
	}

	var deps []*ast.Dep
	if val.Type().IsTupleType() {
		for it := val.ElementIterator(); it.Next(); {
			_, elem := it.Element()
			if elem.Type() == cty.String {
				dep := &ast.Dep{
					Task: elem.AsString(),
				}
				deps = append(deps, dep)
			} else if elem.Type().IsObjectType() {
				// Handle dependency with vars
				// For now, just skip - TODO: implement proper dependency with vars parsing
			}
		}
	} else if val.Type() == cty.String {
		// Single dependency
		dep := &ast.Dep{
			Task: val.AsString(),
		}
		deps = append(deps, dep)
	}

	return deps, nil
}