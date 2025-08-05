package taskfile

// HCLTaskfile represents the root structure of an HCL Taskfile
type HCLTaskfile struct {
	Version *string          `hcl:"version,attr"`
	Tasks   []*HCLTaskBlock  `hcl:"task,block"`
	
	// TODO: Add in future tasks
	// Vars     []*HCLVarBlock     `hcl:"var,block"`
	// Env      []*HCLEnvBlock     `hcl:"env,block"`
	// Includes []*HCLIncludeBlock `hcl:"include,block"`
	
	// Global attributes (to be implemented later)
	// Output   *string          `hcl:"output,attr"`
	// Method   *string          `hcl:"method,attr"`
	// Set      []string         `hcl:"set,attr"`
	// Shopt    []string         `hcl:"shopt,attr"`
	// Silent   *bool            `hcl:"silent,attr"`
	// Dotenv   []string         `hcl:"dotenv,attr"`
	// Run      *string          `hcl:"run,attr"`
}

// HCLTaskBlock represents a task block in HCL format
// Example: task "build" { desc = "Build the project"; cmds = ["go build"] }
type HCLTaskBlock struct {
	Name string   `hcl:"name,label"`
	Desc *string  `hcl:"desc,attr"`
	Cmds []string `hcl:"cmds,optional"`
	
	// TODO: Add in future tasks
	// Deps          []string                `hcl:"deps,attr"`
	// Label         *string                 `hcl:"label,attr"`
	// Summary       *string                 `hcl:"summary,attr"`
	// Aliases       []string                `hcl:"aliases,attr"`
	// Sources       []string                `hcl:"sources,attr"`
	// Generates     []string                `hcl:"generates,attr"`
	// Status        []string                `hcl:"status,attr"`
	// Dir           *string                 `hcl:"dir,attr"`
	// Set           []string                `hcl:"set,attr"`
	// Shopt         []string                `hcl:"shopt,attr"`
	// Dotenv        []string                `hcl:"dotenv,attr"`
	// Silent        *bool                   `hcl:"silent,attr"`
	// Interactive   *bool                   `hcl:"interactive,attr"`
	// Internal      *bool                   `hcl:"internal,attr"`
	// Method        *string                 `hcl:"method,attr"`
	// Prefix        *string                 `hcl:"prefix,attr"`
	// IgnoreError   *bool                   `hcl:"ignore_error,attr"`
	// Run           *string                 `hcl:"run,attr"`
	// Platforms     []string                `hcl:"platforms,attr"`
	// Watch         *bool                   `hcl:"watch,attr"`
	
	// Complex nested structures (to be implemented later)
	// Preconditions []*HCLPreconditionBlock `hcl:"precondition,block"`
	// Vars          []*HCLVarBlock          `hcl:"var,block"`
	// Env           []*HCLEnvBlock          `hcl:"env,block"`
}