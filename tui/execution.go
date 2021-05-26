package tui

import (
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/rivo/tview"
)

// executionOutputView is a  rendered textview of a  NodeExecutionResult's stdout and stderr
type executionOutputView struct {
	*DetailPane

	// execResult refers to the result being rendered
	execResult *graph.NodeExecutionResult
}

func (e *executionOutputView) setExecutionResult(res *graph.NodeExecutionResult) {
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

func newExecutionOutputView(rootUI *Tui) *executionOutputView {
	e := &executionOutputView{
		DetailPane: NewDetailPane("execution", rootUI),
	}
	return e
}

type executionResultView struct {
	*tview.TextView
	// execResult refers to the result being rendered
	execResult *graph.NodeExecutionResult
}

func (e *executionResultView) setExecutionResult(res *graph.NodeExecutionResult) {
	e.execResult = res
}

func newExecutionResultView() *executionResultView {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(false)
	textView.SetBorder(true).
		SetTitleAlign(tview.AlignLeft)
	return &executionResultView{
		TextView: textView,
	}
}
