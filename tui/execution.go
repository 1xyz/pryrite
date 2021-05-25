package tui

import (
	"github.com/aardlabs/terminal-poc/graph"
)

// executionView is a  rendered textview of a  NodeExecutionResult
type executionView struct {
	*DetailPane

	// execResult refers to the result being rendered
	execResult *graph.NodeExecutionResult
}

func (e *executionView) setExecutionResult(res *graph.NodeExecutionResult) {
	e.execResult = res
	e.Clear()
	if e.execResult != nil {

	}
}

func newExecutionView(rootUI *Tui) *executionView {
	e := &executionView{
		DetailPane: NewDetailPane("execution", rootUI),
	}
	return e
}
