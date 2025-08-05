package taskfile

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskfileDiscovery(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	t.Run("YAML files have priority over HCL", func(t *testing.T) {
		subDir := filepath.Join(tempDir, "yaml_priority")
		require.NoError(t, os.MkdirAll(subDir, 0755))

		// Create both YAML and HCL files
		yamlContent := `version: "3"
tasks:
  test:
    cmds:
      - echo "yaml"`

		hclContent := `version = "3"
tasks = {
  test = {
    cmds = ["echo hcl"]
  }
}`

		yamlFile := filepath.Join(subDir, "Taskfile.yml")
		hclFile := filepath.Join(subDir, "Taskfile.hcl")

		require.NoError(t, os.WriteFile(yamlFile, []byte(yamlContent), 0644))
		require.NoError(t, os.WriteFile(hclFile, []byte(hclContent), 0644))

		// Create a file node and test discovery
		node, err := NewFileNode("", subDir)
		require.NoError(t, err)

		// Should find the YAML file (priority)
		assert.Equal(t, yamlFile, node.Location())
	})

	t.Run("HCL files are found when no YAML exists", func(t *testing.T) {
		subDir := filepath.Join(tempDir, "hcl_only")
		require.NoError(t, os.MkdirAll(subDir, 0755))

		hclContent := `version = "3"
tasks = {
  test = {
    cmds = ["echo hcl"]
  }
}`

		hclFile := filepath.Join(subDir, "Taskfile.hcl")
		require.NoError(t, os.WriteFile(hclFile, []byte(hclContent), 0644))

		// Create a file node and test discovery
		node, err := NewFileNode("", subDir)
		require.NoError(t, err)

		// Should find the HCL file
		assert.Equal(t, hclFile, node.Location())
	})

	t.Run("Extensionless files are found", func(t *testing.T) {
		subDir := filepath.Join(tempDir, "extensionless")
		require.NoError(t, os.MkdirAll(subDir, 0755))

		content := `version: "3"
tasks:
  test:
    cmds:
      - echo "extensionless"`

		extensionlessFile := filepath.Join(subDir, "Taskfile")
		require.NoError(t, os.WriteFile(extensionlessFile, []byte(content), 0644))

		// Create a file node and test discovery
		node, err := NewFileNode("", subDir)
		require.NoError(t, err)

		// Should find the extensionless file
		assert.Equal(t, extensionlessFile, node.Location())
	})

	t.Run("Dist files follow priority order", func(t *testing.T) {
		subDir := filepath.Join(tempDir, "dist_priority")
		require.NoError(t, os.MkdirAll(subDir, 0755))

		yamlContent := `version: "3"
tasks:
  test:
    cmds:
      - echo "yaml dist"`

		hclContent := `version = "3"`

		// Create dist files
		yamlDistFile := filepath.Join(subDir, "Taskfile.dist.yml")
		hclDistFile := filepath.Join(subDir, "Taskfile.dist.hcl")

		require.NoError(t, os.WriteFile(yamlDistFile, []byte(yamlContent), 0644))
		require.NoError(t, os.WriteFile(hclDistFile, []byte(hclContent), 0644))

		// Create a file node and test discovery
		node, err := NewFileNode("", subDir)
		require.NoError(t, err)

		// Should find the YAML dist file (priority)
		assert.Equal(t, yamlDistFile, node.Location())
	})

	t.Run("HCL priority over extensionless", func(t *testing.T) {
		subDir := filepath.Join(tempDir, "hcl_vs_extensionless")
		require.NoError(t, os.MkdirAll(subDir, 0755))

		content := `version: "3"
tasks:
  test:
    cmds:
      - echo "test"`

		// Create HCL file and extensionless file
		hclFile := filepath.Join(subDir, "Taskfile.hcl")
		extensionlessFile := filepath.Join(subDir, "Taskfile")

		require.NoError(t, os.WriteFile(hclFile, []byte(content), 0644))
		require.NoError(t, os.WriteFile(extensionlessFile, []byte(content), 0644))

		// Create a file node and test discovery
		node, err := NewFileNode("", subDir)
		require.NoError(t, err)

		// Should find HCL file (higher priority than extensionless)
		assert.Equal(t, hclFile, node.Location())
	})
}

func TestHCLLoaderRegistration(t *testing.T) {
	t.Parallel()

	registry := NewLoaderRegistry()

	t.Run("HCL loader is registered", func(t *testing.T) {
		// Should have both YAML and HCL loaders
		assert.Len(t, registry.loaders, 2)

		// Test HCL file extension recognition
		loader := registry.GetLoader("Taskfile.hcl")
		assert.Equal(t, "HCL", loader.FormatName())

		// Test YAML file extension recognition
		loader = registry.GetLoader("Taskfile.yml")
		assert.Equal(t, "YAML", loader.FormatName())
	})

	t.Run("HCL loader returns not implemented error", func(t *testing.T) {
		hclContent := `version = "3"`

		node := &mockNode{
			location: "/path/to/Taskfile.hcl",
			content:  []byte(hclContent),
		}

		loader := registry.GetLoader("Taskfile.hcl")
		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(hclContent))

		assert.Error(t, err)
		assert.Nil(t, taskfile)
		assert.Contains(t, err.Error(), "HCL parsing is not yet implemented")
		assert.Contains(t, err.Error(), "/path/to/Taskfile.hcl")
	})
}

func TestLoaderExtensionSupport(t *testing.T) {
	t.Parallel()

	t.Run("YAML loader extensions", func(t *testing.T) {
		loader := NewYAMLLoader()
		
		assert.True(t, loader.SupportsExtension(".yml"))
		assert.True(t, loader.SupportsExtension(".yaml"))
		assert.True(t, loader.SupportsExtension(".YML"))
		assert.True(t, loader.SupportsExtension(".YAML"))
		assert.False(t, loader.SupportsExtension(".hcl"))
		assert.False(t, loader.SupportsExtension(""))
	})

	t.Run("HCL loader extensions", func(t *testing.T) {
		loader := NewHCLLoader()
		
		assert.True(t, loader.SupportsExtension(".hcl"))
		assert.True(t, loader.SupportsExtension(".HCL"))
		assert.True(t, loader.SupportsExtension(""))  // Supports extensionless
		assert.False(t, loader.SupportsExtension(".yml"))
		assert.False(t, loader.SupportsExtension(".yaml"))
	})
}

func TestExtensionlessFileParsing(t *testing.T) {
	t.Parallel()

	registry := NewLoaderRegistry()

	t.Run("Extensionless YAML file parsed correctly", func(t *testing.T) {
		yamlContent := `version: "3"
tasks:
  test:
    cmds:
      - echo "hello world"`

		node := &mockNode{
			location: "/path/to/Taskfile",
			content:  []byte(yamlContent),
		}

		// Should try YAML first and succeed
		taskfile, err := registry.LoadTaskfile(context.Background(), node, []byte(yamlContent))
		require.NoError(t, err)
		require.NotNil(t, taskfile)

		assert.Equal(t, "/path/to/Taskfile", taskfile.Location)
		assert.Equal(t, "3.0.0", taskfile.Version.String())
	})

	t.Run("Extensionless HCL file falls back to HCL loader", func(t *testing.T) {
		hclContent := `version = "3"`

		node := &mockNode{
			location: "/path/to/Taskfile",
			content:  []byte(hclContent),
		}

		// Should try YAML first (fail), then try HCL (get not implemented error)
		taskfile, err := registry.LoadTaskfile(context.Background(), node, []byte(hclContent))
		assert.Error(t, err)
		assert.Nil(t, taskfile)
		assert.Contains(t, err.Error(), "HCL parsing is not yet implemented")
	})
}