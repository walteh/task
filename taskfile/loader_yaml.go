package taskfile

import (
	"context"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/go-task/task/v3/errors"
	"github.com/go-task/task/v3/internal/filepathext"
	"github.com/go-task/task/v3/taskfile/ast"
)

// YAMLLoader implements TaskfileLoader for YAML format
type YAMLLoader struct{}

// NewYAMLLoader creates a new YAMLLoader instance
func NewYAMLLoader() *YAMLLoader {
	return &YAMLLoader{}
}

// LoadTaskfile loads and parses a YAML Taskfile
func (l *YAMLLoader) LoadTaskfile(ctx context.Context, node Node, content []byte) (*ast.Taskfile, error) {
	var tf ast.Taskfile
	if err := yaml.Unmarshal(content, &tf); err != nil {
		// Decode the taskfile and add the file info the any errors
		taskfileDecodeErr := &errors.TaskfileDecodeError{}
		if errors.As(err, &taskfileDecodeErr) {
			snippet := NewSnippet(content,
				WithLine(taskfileDecodeErr.Line),
				WithColumn(taskfileDecodeErr.Column),
				WithPadding(2),
			)
			return nil, taskfileDecodeErr.WithFileInfo(node.Location(), snippet.String())
		}
		return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(node.Location()), Err: err}
	}

	// Check that the Taskfile is set and has a schema version
	if tf.Version == nil {
		return nil, &errors.TaskfileVersionCheckError{URI: node.Location()}
	}

	// Set the taskfile/task's locations
	tf.Location = node.Location()
	for task := range tf.Tasks.Values(nil) {
		// If the task is not defined, create a new one
		if task == nil {
			task = &ast.Task{}
		}
		// Set the location of the taskfile for each task
		if task.Location.Taskfile == "" {
			task.Location.Taskfile = tf.Location
		}
	}

	return &tf, nil
}

// SupportsExtension returns true if the loader supports the given file extension
func (l *YAMLLoader) SupportsExtension(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".yml" || ext == ".yaml"
}

// FormatName returns the name of the format this loader supports
func (l *YAMLLoader) FormatName() string {
	return "YAML"
}