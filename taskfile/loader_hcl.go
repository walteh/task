package taskfile

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"

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

// LoadTaskfile loads and parses an HCL Taskfile
func (l *HCLLoader) LoadTaskfile(ctx context.Context, node Node, content []byte) (*ast.Taskfile, error) {
	// Parse the HCL file
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(content, node.Location())
	if diags.HasErrors() {
		return nil, l.formatHCLError(node.Location(), diags)
	}

	// Decode the HCL into our struct
	var hclTaskfile HCLTaskfile
	diags = gohcl.DecodeBody(file.Body, nil, &hclTaskfile)
	if diags.HasErrors() {
		return nil, l.formatHCLError(node.Location(), diags)
	}

	// Convert to internal Taskfile format
	taskfile, err := l.convertToTaskfile(&hclTaskfile, node.Location())
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

// convertToTaskfile converts an HCLTaskfile to the internal ast.Taskfile format
func (l *HCLLoader) convertToTaskfile(hclTf *HCLTaskfile, location string) (*ast.Taskfile, error) {
	// Parse version string
	var version *semver.Version
	if hclTf.Version != nil && *hclTf.Version != "" {
		var err error
		version, err = semver.NewVersion(*hclTf.Version)
		if err != nil {
			return nil, &errors.TaskfileInvalidError{
				URI: filepathext.TryAbsToRel(location),
				Err: fmt.Errorf("invalid version format: %s", *hclTf.Version),
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

	// Check that version is set
	if tf.Version == nil {
		return nil, &errors.TaskfileVersionCheckError{URI: location}
	}

	// Convert tasks
	for _, hclTask := range hclTf.Tasks {
		task := &ast.Task{
			Task: hclTask.Name,
			Location: &ast.Location{
				Taskfile: location,
			},
		}

		// Set description
		if hclTask.Desc != nil {
			task.Desc = *hclTask.Desc
		}

		// Convert commands to simple shell commands
		for _, cmdStr := range hclTask.Cmds {
			cmd := &ast.Cmd{
				Cmd:    cmdStr,
				Silent: false, // TODO: Handle silent attribute in future
			}
			task.Cmds = append(task.Cmds, cmd)
		}

		// Add task to the Tasks map
		tf.Tasks.Set(hclTask.Name, task)
	}

	return tf, nil
}