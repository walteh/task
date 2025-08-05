package taskfile

import (
	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"

	"github.com/go-task/task/v3/errors"
	"github.com/go-task/task/v3/internal/filepathext"
	"github.com/go-task/task/v3/taskfile/ast"
)

// hclTaskfile mirrors the HCL structure of a Taskfile.
type hclTaskfile struct {
	Version string         `hcl:"version,attr"`
	Tasks   []hclTaskBlock `hcl:"task,block"`
}

// hclTaskBlock represents an individual task block.
type hclTaskBlock struct {
	Name string   `hcl:"name,label"`
	Desc *string  `hcl:"desc,attr"`
	Cmds []string `hcl:"cmds,attr"`
}

// Load parses the given data as HCL into a Taskfile structure.
func (HCLLoader) Load(data []byte, location string) (*ast.Taskfile, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(data, location)
	if diags.HasErrors() {
		return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: diags}
	}

	var htf hclTaskfile
	diags = gohcl.DecodeBody(file.Body, nil, &htf)
	if diags.HasErrors() {
		return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: diags}
	}

	version, err := semver.NewVersion(htf.Version)
	if err != nil {
		return nil, &errors.TaskfileInvalidError{URI: filepathext.TryAbsToRel(location), Err: err}
	}

	tf := &ast.Taskfile{
		Version: version,
		Tasks:   ast.NewTasks(),
	}

	for _, t := range htf.Tasks {
		task := &ast.Task{Task: t.Name}
		if t.Desc != nil {
			task.Desc = *t.Desc
		}
		for _, cmd := range t.Cmds {
			task.Cmds = append(task.Cmds, &ast.Cmd{Cmd: cmd})
		}
		tf.Tasks.Set(t.Name, task)
	}

	return tf, nil
}
