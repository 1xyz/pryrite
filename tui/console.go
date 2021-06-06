package tui

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/run"
)

// consoleView is a  rendered textview of a  NodeExecutionResult's stdout and stderr
type consoleView struct {
	*detailView
}

func (c *consoleView) Refresh(results run.BlockExecutionResults) {
	c.Clear()
	if results == nil || len(results) == 0 {
		return
	}
	for i, res := range results {
		if err := c.writeBytes(res.Stdout); err != nil {
			c.rootUI.StatusErrorf("[%d] Refresh: writeBytes(stdout): err = %v\n", i, err)
		}
		if err := c.writeBytes(res.Stderr); err != nil {
			c.rootUI.StatusErrorf("[%d] Refresh: writeBytes(stderr): err = %v\n", i, err)
		}
	}
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

func (c *consoleView) NavHelp() string {
	help := " ctrl+r: run selected node"
	navigate := " tab: next pane, shift+tab: previous pane"
	navHelp := fmt.Sprintf(" commands \t| %s\n navigate \t| %s\n", help, navigate)
	return navHelp
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
