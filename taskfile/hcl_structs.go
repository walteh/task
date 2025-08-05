package taskfile

import (
	"github.com/hashicorp/hcl/v2"
)

// HCLTaskfile represents the root structure of an HCL Taskfile with low-level parsing
type HCLTaskfile struct {
	// We'll use low-level parsing instead of struct tags for expression capture
}

// HCLTaskData holds parsed HCL task data with expressions
type HCLTaskData struct {
	Name     string
	Desc     *string
	Label    *string
	Summary  *string
	Dir      *string
	Method   *string
	Prefix   *string
	Run      *string
	
	// Boolean attributes
	Silent      *bool
	Interactive *bool
	Internal    *bool
	IgnoreError *bool
	Watch       *bool
	
	// String slice attributes
	Aliases   []string
	Sources   []string
	Generates []string
	Status    []string
	Set       []string
	Shopt     []string
	Dotenv    []string
	Platforms []string
	
	// Expression-based attributes (for runtime evaluation)
	Cmds hcl.Expression // Raw HCL expression for commands
	Deps hcl.Expression // Raw HCL expression for dependencies
	Vars hcl.Body       // Raw HCL body for variables
	Env  hcl.Body       // Raw HCL body for environment variables
}

// HCLTaskfileData holds parsed HCL Taskfile data
type HCLTaskfileData struct {
	Version *string
	Output  *string
	Method  *string
	Run     *string
	
	// Boolean attributes
	Silent *bool
	
	// String slice attributes
	Set    []string
	Shopt  []string
	Dotenv []string
	
	// Expression-based global attributes
	Vars     hcl.Body // Raw HCL body for global variables
	Env      hcl.Body // Raw HCL body for global environment
	Includes hcl.Body // Raw HCL body for includes
	
	// Tasks are parsed separately using block extraction
	Tasks []*HCLTaskData
}