package taskfile

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/go-task/task/v3/taskfile/ast"
)

// TaskfileLoader defines the interface for loading and parsing Taskfiles
// from different formats (YAML, HCL, etc.)
type TaskfileLoader interface {
	LoadTaskfile(ctx context.Context, node Node, content []byte) (*ast.Taskfile, error)
	SupportsExtension(ext string) bool
	FormatName() string
}

// LoaderRegistry manages different TaskfileLoader implementations
type LoaderRegistry struct {
	loaders []TaskfileLoader
}

// NewLoaderRegistry creates a new LoaderRegistry with default loaders
func NewLoaderRegistry() *LoaderRegistry {
	registry := &LoaderRegistry{}
	registry.RegisterLoader(NewYAMLLoader())
	return registry
}

// RegisterLoader adds a new TaskfileLoader to the registry
func (r *LoaderRegistry) RegisterLoader(loader TaskfileLoader) {
	r.loaders = append(r.loaders, loader)
}

// GetLoader returns the appropriate loader for the given file extension
func (r *LoaderRegistry) GetLoader(filename string) TaskfileLoader {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, loader := range r.loaders {
		if loader.SupportsExtension(ext) {
			return loader
		}
	}
	// Default to YAML loader if no specific loader found
	return NewYAMLLoader()
}

// LoadTaskfile loads a taskfile using the appropriate loader based on file extension
func (r *LoaderRegistry) LoadTaskfile(ctx context.Context, node Node, content []byte) (*ast.Taskfile, error) {
	loader := r.GetLoader(node.Location())
	return loader.LoadTaskfile(ctx, node, content)
}