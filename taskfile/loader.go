package taskfile

import (
	stdErrors "errors"

	"gopkg.in/yaml.v3"

	"github.com/go-task/task/v3/errors"
	"github.com/go-task/task/v3/internal/filepathext"
	"github.com/go-task/task/v3/taskfile/ast"
)

// Loader defines the behavior required to load a Taskfile from raw data.
//
// Note: the returned [ast.Taskfile] is still backed by YAML-specific
// unmarshalling logic within the ast package. Loaders for alternative
// formats must populate the AST structures directly without relying on the
// YAML-only helpers. Future work will extract those bindings out of the ast
// package so it becomes truly format agnostic.
type Loader interface {
	Load(data []byte, location string) (*ast.Taskfile, error)
}

// YAMLLoader implements [Loader] using YAML as the configuration format.
type YAMLLoader struct{}

// Load parses the given data as YAML into a Taskfile structure.
func (YAMLLoader) Load(data []byte, location string) (*ast.Taskfile, error) {
	var tf ast.Taskfile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		taskfileDecodeErr := &errors.TaskfileDecodeError{}
		if stdErrors.As(err, &taskfileDecodeErr) {
			snippet := NewSnippet(data,
				WithLine(taskfileDecodeErr.Line),
				WithColumn(taskfileDecodeErr.Column),
				WithPadding(2),
			)
			return nil, taskfileDecodeErr.WithFileInfo(location, snippet.String())
		}
		return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: err}
	}
	return &tf, nil
}

// HCLLoader implements [Loader] using HCL as the configuration format.
//
// Note: HCL parsing is not yet implemented. This loader currently returns a
// descriptive error so that discovery logic can recognize HCL Taskfiles without
// breaking program flow.
type HCLLoader struct{}

// Load returns a not implemented error until HCL support is completed.
func (HCLLoader) Load(data []byte, location string) (*ast.Taskfile, error) {
	return nil, errors.New("HCL parsing not implemented")
}
