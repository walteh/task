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

func TestHCLComprehensiveFeatures(t *testing.T) {
	t.Parallel()

	t.Run("Complex HCL with all task attributes", func(t *testing.T) {
		hclContent := `version = "3"
output = "prefixed"
method = "timestamp"
silent = true
set = ["pipefail"]
shopt = ["expand_aliases"]
dotenv = [".env"]

task "complex" {
  desc = "Complex task with all attributes"
  label = "Complex Task"
  summary = "A comprehensive test task"
  dir = "./subdir"
  method = "checksum"
  prefix = "[TASK]"
  run = "once"
  silent = true
  interactive = false
  internal = true
  ignore_error = false
  watch = true
  
  aliases = ["comp", "c"]
  sources = ["src/**/*.go"]
  generates = ["bin/app"]
  status = ["test -f bin/app"]
  set = ["errexit"]
  shopt = ["nullglob"]
  dotenv = [".env.local"]
  platforms = ["linux", "darwin"]
  
  cmds = [
    "echo 'Building application'",
    "go build -o bin/app",
    "echo 'Build complete'"
  ]
  
  deps = ["prepare", "test"]
}`

		node := &mockNode{
			location: "/path/to/complex.hcl",
			content:  []byte(hclContent),
		}

		loader := NewHCLLoader()
		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(hclContent))

		require.NoError(t, err)
		require.NotNil(t, taskfile)

		// Verify global attributes
		assert.Equal(t, "3.0.0", taskfile.Version.String())
		assert.Equal(t, "prefixed", taskfile.Output.Name)
		assert.Equal(t, "timestamp", taskfile.Method)
		assert.True(t, taskfile.Silent)
		assert.Equal(t, []string{"pipefail"}, taskfile.Set)
		assert.Equal(t, []string{"expand_aliases"}, taskfile.Shopt)
		assert.Equal(t, []string{".env"}, taskfile.Dotenv)

		// Verify task attributes
		complexTask, exists := taskfile.Tasks.Get("complex")
		require.True(t, exists)
		assert.Equal(t, "complex", complexTask.Task)
		assert.Equal(t, "Complex task with all attributes", complexTask.Desc)
		assert.Equal(t, "Complex Task", complexTask.Label)
		assert.Equal(t, "A comprehensive test task", complexTask.Summary)
		assert.Equal(t, "./subdir", complexTask.Dir)
		assert.Equal(t, "checksum", complexTask.Method)
		assert.Equal(t, "[TASK]", complexTask.Prefix)
		assert.Equal(t, "once", complexTask.Run)
		assert.True(t, complexTask.Silent)
		assert.False(t, complexTask.Interactive)
		assert.True(t, complexTask.Internal)
		assert.False(t, complexTask.IgnoreError)
		assert.True(t, complexTask.Watch)

		// Verify string slice attributes
		assert.Equal(t, []string{"comp", "c"}, complexTask.Aliases)
		assert.Equal(t, []string{"errexit"}, complexTask.Set)
		assert.Equal(t, []string{"nullglob"}, complexTask.Shopt)
		assert.Equal(t, []string{".env.local"}, complexTask.Dotenv)
		assert.Equal(t, []string{"test -f bin/app"}, complexTask.Status)

		// Verify glob attributes
		require.Len(t, complexTask.Sources, 1)
		assert.Equal(t, "src/**/*.go", complexTask.Sources[0].Glob)
		require.Len(t, complexTask.Generates, 1)
		assert.Equal(t, "bin/app", complexTask.Generates[0].Glob)

		// Verify platform attributes
		require.Len(t, complexTask.Platforms, 2)
		assert.Equal(t, "linux", complexTask.Platforms[0].OS)
		assert.Equal(t, "darwin", complexTask.Platforms[1].OS)

		// Verify commands
		require.Len(t, complexTask.Cmds, 3)
		assert.Equal(t, "echo 'Building application'", complexTask.Cmds[0].Cmd)
		assert.Equal(t, "go build -o bin/app", complexTask.Cmds[1].Cmd)
		assert.Equal(t, "echo 'Build complete'", complexTask.Cmds[2].Cmd)

		// Verify dependencies
		require.Len(t, complexTask.Deps, 2)
		assert.Equal(t, "prepare", complexTask.Deps[0].Task)
		assert.Equal(t, "test", complexTask.Deps[1].Task)
	})

	t.Run("HCL with single command and dependency", func(t *testing.T) {
		hclContent := `version = "3"

task "single" {
  desc = "Task with single command and dependency"
  cmds = "echo 'Single command'"
  deps = "prepare"
}`

		node := &mockNode{
			location: "/path/to/single.hcl",
			content:  []byte(hclContent),
		}

		loader := NewHCLLoader()
		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(hclContent))

		require.NoError(t, err)
		require.NotNil(t, taskfile)

		singleTask, exists := taskfile.Tasks.Get("single")
		require.True(t, exists)
		assert.Equal(t, "Task with single command and dependency", singleTask.Desc)

		// Verify single command
		require.Len(t, singleTask.Cmds, 1)
		assert.Equal(t, "echo 'Single command'", singleTask.Cmds[0].Cmd)

		// Verify single dependency
		require.Len(t, singleTask.Deps, 1)
		assert.Equal(t, "prepare", singleTask.Deps[0].Task)
	})

	t.Run("HCL with interpolated commands (expression capture)", func(t *testing.T) {
		hclContent := `version = "3"

task "interpolated" {
  desc = "Task with interpolated commands"
  cmds = [
    "echo 'Hello ${USER}'",
    "echo 'Building for ${ARCH}'",
    "go build -o bin/app-${VERSION}"
  ]
}`

		node := &mockNode{
			location: "/path/to/interpolated.hcl",
			content:  []byte(hclContent),
		}

		loader := NewHCLLoader()
		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(hclContent))

		require.NoError(t, err)
		require.NotNil(t, taskfile)

		interpolatedTask, exists := taskfile.Tasks.Get("interpolated")
		require.True(t, exists)
		assert.Equal(t, "Task with interpolated commands", interpolatedTask.Desc)

		// Verify that commands with ${} interpolation are captured
		require.Len(t, interpolatedTask.Cmds, 3)
		assert.Equal(t, "echo 'Hello ${USER}'", interpolatedTask.Cmds[0].Cmd)
		assert.Equal(t, "echo 'Building for ${ARCH}'", interpolatedTask.Cmds[1].Cmd)
		assert.Equal(t, "go build -o bin/app-${VERSION}", interpolatedTask.Cmds[2].Cmd)
	})

	t.Run("HCL with multiple tasks", func(t *testing.T) {
		hclContent := `version = "3"

task "prepare" {
  desc = "Prepare for build"
  cmds = ["mkdir -p bin", "go mod download"]
}

task "build" {
  desc = "Build the application"
  deps = ["prepare"]
  cmds = ["go build -o bin/app"]
}

task "test" {
  desc = "Run tests"
  cmds = ["go test ./..."]
}

task "all" {
  desc = "Run all tasks"
  deps = ["build", "test"]
}`

		node := &mockNode{
			location: "/path/to/multi.hcl",
			content:  []byte(hclContent),
		}

		loader := NewHCLLoader()
		taskfile, err := loader.LoadTaskfile(context.Background(), node, []byte(hclContent))

		require.NoError(t, err)
		require.NotNil(t, taskfile)

		// Verify all tasks exist
		prepareTask, exists := taskfile.Tasks.Get("prepare")
		require.True(t, exists)
		assert.Equal(t, "Prepare for build", prepareTask.Desc)
		assert.Len(t, prepareTask.Cmds, 2)
		assert.Len(t, prepareTask.Deps, 0)

		buildTask, exists := taskfile.Tasks.Get("build")
		require.True(t, exists)
		assert.Equal(t, "Build the application", buildTask.Desc)
		assert.Len(t, buildTask.Cmds, 1)
		assert.Len(t, buildTask.Deps, 1)
		assert.Equal(t, "prepare", buildTask.Deps[0].Task)

		testTask, exists := taskfile.Tasks.Get("test")
		require.True(t, exists)
		assert.Equal(t, "Run tests", testTask.Desc)
		assert.Len(t, testTask.Cmds, 1)
		assert.Len(t, testTask.Deps, 0)

		allTask, exists := taskfile.Tasks.Get("all")
		require.True(t, exists)
		assert.Equal(t, "Run all tasks", allTask.Desc)
		assert.Len(t, allTask.Cmds, 0)
		assert.Len(t, allTask.Deps, 2)
		assert.Equal(t, "build", allTask.Deps[0].Task)
		assert.Equal(t, "test", allTask.Deps[1].Task)
	})
}