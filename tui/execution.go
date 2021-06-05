package tui

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/run"
)

// executionOutputView is a  rendered textview of a  NodeExecutionResult's stdout and stderr
type executionOutputView struct {
	*detailView
}

func (e *executionOutputView) Refresh(results run.BlockExecutionResults) {
	e.Clear()
	if results == nil || len(results) == 0 {
		return
	}
	for i, res := range results {
		if err := e.writeBytes(res.Stdout); err != nil {
			e.rootUI.StatusErrorf("[%d] Refresh: e.writeBytes(stdout): err = %v\n", i, err)
		}
		if err := e.writeBytes(res.Stderr); err != nil {
			e.rootUI.StatusErrorf("[%d] Refresh: e.writeBytes(stderr): err = %v\n", i, err)
		}
	}
}

func (e *executionOutputView) writeBytes(p []byte) error {
	if p == nil || len(p) == 0 {
		return nil
	}
	if _, err := e.Write(p); err != nil {
		return err
	}
	return nil
}

func (e *executionOutputView) NavHelp() string {
	help := " ctrl+r: run selected node"
	navigate := " tab: next pane, shift+tab: previous pane"
	navHelp := fmt.Sprintf(" commands \t| %s\n navigate \t| %s\n", help, navigate)
	return navHelp
}

func newExecutionOutputView(rootUI *Tui) *executionOutputView {
	view := &executionOutputView{
		detailView: newDetailView("execution output", true, rootUI),
	}
	// this is necessary to have the view update with the latest contents written
	view.SetChangedFunc(func() { rootUI.App.Draw() })
	view.ScrollToEnd()
	view.SetInputCapture(rootUI.CommonKeyBindings)
	return view
}
