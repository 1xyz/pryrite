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
	if e.execResult == nil {
		return
	}
	if e.execResult.Stdout != nil && len(e.execResult.Stdout) > 0 {
		if _, err := e.Write(e.execResult.Stdout); err != nil {
			e.rootUI.Statusf("setExecutionResult: e.Write(stdout): err = %v\n", err)
		}
	}
	if e.execResult.Stderr != nil && len(e.execResult.Stderr) > 0 {
		if _, err := e.Write(e.execResult.Stderr); err != nil {
			e.rootUI.Statusf("setExecutionResult: e.Write(stderr): err = %v\n", err)
		}
	}
}

func newExecutionView(rootUI *Tui) *executionView {
	e := &executionView{
		DetailPane: NewDetailPane("execution", rootUI),
	}
	return e
}
