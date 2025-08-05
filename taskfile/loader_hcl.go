package taskfile

import (
	"context"
	"fmt"
	"strings"

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
	return nil, fmt.Errorf("HCL parsing is not yet implemented for file: %s", node.Location())
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