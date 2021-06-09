package tui

import (
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/run"
)

// consoleView is a  rendered textview of a  NodeExecutionResult's stdout and stderr
type consoleView struct {
	*detailView
}

func (c *consoleView) Refresh(results *run.BlockExecutionResults) {
	c.Clear()
	if results == nil || results.Len() == 0 {
		return
	}
	results.Each(func(i int, res *graph.BlockExecutionResult) bool {
		if err := c.writeBytes(res.Stdout); err != nil {
			c.rootUI.StatusErrorf("[%d] Refresh: writeBytes(stdout): err = %v\n", i, err)
		}
		if err := c.writeBytes(res.Stderr); err != nil {
			c.rootUI.StatusErrorf("[%d] Refresh: writeBytes(stderr): err = %v\n", i, err)
		}
		return true
	})
}

func (c *consoleView) writeBytes(p []byte) error {
	if p == nil || len(p) == 0 {
		return nil
	}
	if _, err := c.Write(p); err != nil {
		return err
	}
	return nil
}

func (c *consoleView) NavHelp() [][]string {
	return [][]string{
		{"Ctrl + R", "Run current selected node"},
		{"Tab", "Navigate to the next pane"},
		{"Shift + Tab", "Navigate to the previous pane"},
	}
}

func newExecutionOutputView(rootUI *Tui) *consoleView {
	view := &consoleView{
		detailView: newDetailView("Console", true, rootUI),
	}
	// this is necessary to have the view update with the latest contents written
	view.SetChangedFunc(func() { rootUI.App.Draw() })
	view.ScrollToEnd()
	view.SetInputCapture(rootUI.CommonKeyBindings)
	return view
}