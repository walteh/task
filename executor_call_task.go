package task

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/go-task/task/v3/internal/logger"
	"github.com/go-task/task/v3/taskfile/ast"
)

func (e *Executor) callTask(name string, vars *ast.Vars) (string, error) {
	buf := &bytes.Buffer{}
	origStdout, origStderr, origLogger := e.Stdout, e.Stderr, e.Logger
	e.Stdout, e.Stderr = buf, buf
	e.Logger = &logger.Logger{Stdout: io.Discard, Stderr: io.Discard}
	err := e.RunTask(context.Background(), &Call{Task: name, Vars: vars, Silent: true, Indirect: true})
	e.Stdout, e.Stderr, e.Logger = origStdout, origStderr, origLogger
	return strings.TrimSpace(buf.String()), err
}
