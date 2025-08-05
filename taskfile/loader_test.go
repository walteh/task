package taskfile

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNode implements the Node interface for testing
type mockNode struct {
	location string
	content  []byte
}

func (m *mockNode) Read() ([]byte, error)                             { return m.content, nil }
func (m *mockNode) Parent() Node                                      { return nil }
func (m *mockNode) Location() string                                  { return m.location }
func (m *mockNode) Dir() string                                       { return "" }
func (m *mockNode) Checksum() string                                  { return "" }
func (m *mockNode) Verify(checksum string) bool                       { return true }
func (m *mockNode) ResolveEntrypoint(entrypoint string) (string, error) { return "", nil }
func (m *mockNode) ResolveDir(dir string) (string, error)             { return "", nil }

func TestYAMLLoader(t *testing.T) {
	t.Parallel()

	loader := NewYAMLLoader()

	t.Run("SupportsExtension", func(t *testing.T) {
		assert.True(t, loader.SupportsExtension(".yml"))
		assert.True(t, loader.SupportsExtension(".yaml"))
		assert.True(t, loader.SupportsExtension(".YML"))
		assert.True(t, loader.SupportsExtension(".YAML"))
		assert.False(t, loader.SupportsExtension(".hcl"))
		assert.False(t, loader.SupportsExtension(".json"))
		assert.False(t, loader.SupportsExtension(""))
	})

	t.Run("FormatName", func(t *testing.T) {
		assert.Equal(t, "YAML", loader.FormatName())
	})

	t.Run("LoadTaskfile", func(t *testing.T) {
		yamlContent := `version: "3"
tasks:
  test:
    cmds:
      - echo "hello world"`

		node := &mockNode{
			location: "/path/to/Taskfile.yml",
			content:  []byte(yamlContent),
		}

		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(yamlContent))
		require.NoError(t, err)
		require.NotNil(t, taskfile)

		assert.Equal(t, "/path/to/Taskfile.yml", taskfile.Location)
		assert.NotNil(t, taskfile.Version)
		assert.Equal(t, "3.0.0", taskfile.Version.String())
		assert.NotNil(t, taskfile.Tasks)

		task, exists := taskfile.Tasks.Get("test")
		assert.True(t, exists)
		assert.NotNil(t, task)
		assert.Equal(t, "/path/to/Taskfile.yml", task.Location.Taskfile)
	})

	t.Run("LoadTaskfile_InvalidYAML", func(t *testing.T) {
		invalidYAML := `invalid: yaml: content: [`

		node := &mockNode{
			location: "/path/to/Taskfile.yml",
			content:  []byte(invalidYAML),
		}

		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(invalidYAML))
		assert.Error(t, err)
		assert.Nil(t, taskfile)
	})

	t.Run("LoadTaskfile_NoVersion", func(t *testing.T) {
		yamlContent := `tasks:
  test:
    cmds:
      - echo "hello world"`

		node := &mockNode{
			location: "/path/to/Taskfile.yml",
			content:  []byte(yamlContent),
		}

		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(yamlContent))
		assert.Error(t, err)
		assert.Nil(t, taskfile)
	})
}

func TestLoaderRegistry(t *testing.T) {
	t.Parallel()

	t.Run("NewLoaderRegistry", func(t *testing.T) {
		registry := NewLoaderRegistry()
		assert.NotNil(t, registry)
		assert.Len(t, registry.loaders, 2) // Should have YAML and HCL loaders by default
	})

	t.Run("RegisterLoader", func(t *testing.T) {
		registry := NewLoaderRegistry()
		initialCount := len(registry.loaders)

		// Create a mock loader
		mockLoader := NewYAMLLoader()
		registry.RegisterLoader(mockLoader)

		assert.Len(t, registry.loaders, initialCount+1)
	})

	t.Run("GetLoader_YAMLExtensions", func(t *testing.T) {
		registry := NewLoaderRegistry()

		loader := registry.GetLoader("Taskfile.yml")
		assert.NotNil(t, loader)
		assert.Equal(t, "YAML", loader.FormatName())

		loader = registry.GetLoader("Taskfile.yaml")
		assert.NotNil(t, loader)
		assert.Equal(t, "YAML", loader.FormatName())

		loader = registry.GetLoader("/path/to/Taskfile.YML")
		assert.NotNil(t, loader)
		assert.Equal(t, "YAML", loader.FormatName())
	})

	t.Run("GetLoader_UnsupportedExtension", func(t *testing.T) {
		registry := NewLoaderRegistry()

		// Should use appropriate loader for supported extensions
		loader := registry.GetLoader("Taskfile.hcl")
		assert.NotNil(t, loader)
		assert.Equal(t, "HCL", loader.FormatName())

		// Should default to YAML loader for extensionless files (backward compatibility)
		loader = registry.GetLoader("Taskfile")
		assert.NotNil(t, loader)
		assert.Equal(t, "YAML", loader.FormatName())

		// Should default to YAML loader for truly unsupported extensions
		loader = registry.GetLoader("Taskfile.json")
		assert.NotNil(t, loader)
		assert.Equal(t, "YAML", loader.FormatName())
	})

	t.Run("LoadTaskfile", func(t *testing.T) {
		registry := NewLoaderRegistry()

		yamlContent := `version: "3"
tasks:
  test:
    cmds:
      - echo "hello world"`

		node := &mockNode{
			location: "/path/to/Taskfile.yml",
			content:  []byte(yamlContent),
		}

		taskfile, err := registry.LoadTaskfile(context.Background(), node, []byte(yamlContent))
		require.NoError(t, err)
		require.NotNil(t, taskfile)

		assert.Equal(t, "/path/to/Taskfile.yml", taskfile.Location)
		assert.NotNil(t, taskfile.Version)
		assert.Equal(t, "3.0.0", taskfile.Version.String())
	})
}

func TestLoaderIntegration(t *testing.T) {
	t.Parallel()

	// Test that the refactored reader works with the same YAML content as before
	yamlContent := `version: "3"
output: prefixed
tasks:
  build:
    desc: Build the application
    cmds:
      - go build -o app ./cmd/app
    sources:
      - "**/*.go"
    generates:
      - app
  test:
    desc: Run tests
    cmds:
      - go test ./...
    deps:
      - build`

	node := &mockNode{
		location: "/path/to/Taskfile.yml",
		content:  []byte(yamlContent),
	}

	reader := NewReader()
	
	// Test that the reader can load the taskfile using the new loader system
	taskfile, err := reader.readNode(context.Background(), node)
	require.NoError(t, err)
	require.NotNil(t, taskfile)

	// Verify the taskfile was parsed correctly
	assert.Equal(t, "/path/to/Taskfile.yml", taskfile.Location)
	assert.NotNil(t, taskfile.Version)
	assert.Equal(t, "3.0.0", taskfile.Version.String())
	assert.Equal(t, "prefixed", taskfile.Output.Name)

	// Verify tasks
	buildTask, exists := taskfile.Tasks.Get("build")
	assert.True(t, exists)
	assert.Equal(t, "Build the application", buildTask.Desc)
	assert.Len(t, buildTask.Cmds, 1)
	assert.Equal(t, "go build -o app ./cmd/app", buildTask.Cmds[0].Cmd)

	testTask, exists := taskfile.Tasks.Get("test")
	assert.True(t, exists)
	assert.Equal(t, "Run tests", testTask.Desc)
	assert.Len(t, testTask.Deps, 1)
	assert.Equal(t, "build", testTask.Deps[0].Task)
}