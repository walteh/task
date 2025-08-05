package taskfile

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHCLParsing(t *testing.T) {
	t.Parallel()

	t.Run("Basic HCL Taskfile parsing", func(t *testing.T) {
		hclContent := `version = "3"

task "build" {
  desc = "Build the project"
  cmds = ["go build ./..."]
}

task "test" {
  desc = "Run tests"
  cmds = ["go test ./...", "echo 'Tests completed'"]
}

task "simple" {
  cmds = ["echo 'No description'"]
}`

		node := &mockNode{
			location: "/path/to/Taskfile.hcl",
			content:  []byte(hclContent),
		}

		loader := NewHCLLoader()
		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(hclContent))

		require.NoError(t, err)
		require.NotNil(t, taskfile)

		// Verify basic structure
		assert.Equal(t, "/path/to/Taskfile.hcl", taskfile.Location)
		assert.NotNil(t, taskfile.Version)
		assert.Equal(t, "3.0.0", taskfile.Version.String())
		assert.NotNil(t, taskfile.Tasks)

		// Verify build task
		buildTask, exists := taskfile.Tasks.Get("build")
		require.True(t, exists)
		require.NotNil(t, buildTask)
		assert.Equal(t, "build", buildTask.Task)
		assert.Equal(t, "Build the project", buildTask.Desc)
		assert.Len(t, buildTask.Cmds, 1)
		assert.Equal(t, "go build ./...", buildTask.Cmds[0].Cmd)
		assert.Equal(t, "/path/to/Taskfile.hcl", buildTask.Location.Taskfile)

		// Verify test task (multiple commands)
		testTask, exists := taskfile.Tasks.Get("test")
		require.True(t, exists)
		require.NotNil(t, testTask)
		assert.Equal(t, "test", testTask.Task)
		assert.Equal(t, "Run tests", testTask.Desc)
		assert.Len(t, testTask.Cmds, 2)
		assert.Equal(t, "go test ./...", testTask.Cmds[0].Cmd)
		assert.Equal(t, "echo 'Tests completed'", testTask.Cmds[1].Cmd)

		// Verify simple task (no description)
		simpleTask, exists := taskfile.Tasks.Get("simple")
		require.True(t, exists)
		require.NotNil(t, simpleTask)
		assert.Equal(t, "simple", simpleTask.Task)
		assert.Equal(t, "", simpleTask.Desc) // No description provided
		assert.Len(t, simpleTask.Cmds, 1)
		assert.Equal(t, "echo 'No description'", simpleTask.Cmds[0].Cmd)
	})

	t.Run("Minimal HCL Taskfile", func(t *testing.T) {
		hclContent := `version = "3"

task "hello" {
  cmds = ["echo hello"]
}`

		node := &mockNode{
			location: "/path/to/minimal.hcl",
			content:  []byte(hclContent),
		}

		loader := NewHCLLoader()
		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(hclContent))

		require.NoError(t, err)
		require.NotNil(t, taskfile)

		// Verify minimal structure
		assert.Equal(t, "3.0.0", taskfile.Version.String())
		
		helloTask, exists := taskfile.Tasks.Get("hello")
		require.True(t, exists)
		assert.Equal(t, "hello", helloTask.Task)
		assert.Equal(t, "", helloTask.Desc) // No description
		assert.Len(t, helloTask.Cmds, 1)
		assert.Equal(t, "echo hello", helloTask.Cmds[0].Cmd)
	})

	t.Run("HCL task with no commands", func(t *testing.T) {
		hclContent := `version = "3"

task "empty" {
  desc = "Task with no commands"
}`

		node := &mockNode{
			location: "/path/to/empty.hcl",
			content:  []byte(hclContent),
		}

		loader := NewHCLLoader()
		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(hclContent))

		require.NoError(t, err)
		require.NotNil(t, taskfile)

		emptyTask, exists := taskfile.Tasks.Get("empty")
		require.True(t, exists)
		assert.Equal(t, "empty", emptyTask.Task)
		assert.Equal(t, "Task with no commands", emptyTask.Desc)
		assert.Len(t, emptyTask.Cmds, 0) // No commands
	})
}

func TestHCLErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("Invalid HCL syntax", func(t *testing.T) {
		invalidHCL := `version = "3"

task "broken" {
  desc = "Missing closing brace"
  cmds = ["echo test"]
# Missing closing brace`

		node := &mockNode{
			location: "/path/to/broken.hcl",
			content:  []byte(invalidHCL),
		}

		loader := NewHCLLoader()
		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(invalidHCL))

		assert.Error(t, err)
		assert.Nil(t, taskfile)
		assert.Contains(t, err.Error(), "HCL parse error")
		assert.Contains(t, err.Error(), "/path/to/broken.hcl")
	})

	t.Run("Missing version", func(t *testing.T) {
		hclContent := `task "test" {
  cmds = ["echo test"]
}`

		node := &mockNode{
			location: "/path/to/noversion.hcl",
			content:  []byte(hclContent),
		}

		loader := NewHCLLoader()
		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(hclContent))

		assert.Error(t, err)
		assert.Nil(t, taskfile)
		assert.Contains(t, err.Error(), "Missing schema version")
	})

	t.Run("Invalid version format", func(t *testing.T) {
		hclContent := `version = "invalid-version"

task "test" {
  cmds = ["echo test"]
}`

		node := &mockNode{
			location: "/path/to/badversion.hcl",
			content:  []byte(hclContent),
		}

		loader := NewHCLLoader()
		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(hclContent))

		assert.Error(t, err)
		assert.Nil(t, taskfile)
		assert.Contains(t, err.Error(), "invalid version format")
	})

	t.Run("Empty file", func(t *testing.T) {
		hclContent := ``

		node := &mockNode{
			location: "/path/to/empty.hcl",
			content:  []byte(hclContent),
		}

		loader := NewHCLLoader()
		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(hclContent))

		assert.Error(t, err)
		assert.Nil(t, taskfile)
		assert.Contains(t, err.Error(), "Missing schema version")
	})
}

func TestHCLLoaderMethods(t *testing.T) {
	t.Parallel()

	loader := NewHCLLoader()

	t.Run("SupportsExtension", func(t *testing.T) {
		assert.True(t, loader.SupportsExtension(".hcl"))
		assert.True(t, loader.SupportsExtension(".HCL"))
		assert.True(t, loader.SupportsExtension("")) // Extensionless
		assert.False(t, loader.SupportsExtension(".yml"))
		assert.False(t, loader.SupportsExtension(".yaml"))
		assert.False(t, loader.SupportsExtension(".json"))
	})

	t.Run("FormatName", func(t *testing.T) {
		assert.Equal(t, "HCL", loader.FormatName())
	})
}

func TestHCLIntegration(t *testing.T) {
	t.Parallel()

	t.Run("HCL with LoaderRegistry", func(t *testing.T) {
		hclContent := `version = "3"

task "integration" {
  desc = "Integration test task"
  cmds = ["echo 'Integration test'"]
}`

		node := &mockNode{
			location: "/path/to/integration.hcl",
			content:  []byte(hclContent),
		}

		registry := NewLoaderRegistry()
		taskfile, err := registry.LoadTaskfile(context.Background(), node, []byte(hclContent))

		require.NoError(t, err)
		require.NotNil(t, taskfile)

		assert.Equal(t, "3.0.0", taskfile.Version.String())
		
		integrationTask, exists := taskfile.Tasks.Get("integration")
		require.True(t, exists)
		assert.Equal(t, "Integration test task", integrationTask.Desc)
		assert.Len(t, integrationTask.Cmds, 1)
		assert.Equal(t, "echo 'Integration test'", integrationTask.Cmds[0].Cmd)
	})
}