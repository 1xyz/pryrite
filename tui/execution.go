package tui

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// executionOutputView is a  rendered textview of a  NodeExecutionResult's stdout and stderr
type executionOutputView struct {
	*detailView
}

func (e *executionOutputView) Refresh(res *graph.NodeExecutionResult) {
	e.Clear()
	if res == nil {
		return
	}
	if err := e.writeBytes(res.Stdout); err != nil {
		e.rootUI.StatusErrorf("Refresh: e.writeBytes(stdout): err = %v\n", err)
	}
	if err := e.writeBytes(res.Stderr); err != nil {
		e.rootUI.StatusErrorf("Refresh: e.writeBytes(stderr): err = %v\n", err)
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

func newExecutionOutputView(rootUI *Tui) *executionOutputView {
	e := &executionOutputView{
		detailView: newDetailView("execution output", true, rootUI),
	}
	return e
}

type executionResultView struct {
	*detailView
}

func (e *executionResultView) Refresh(res *graph.NodeExecutionResult) {
	e.Clear()
	if res == nil {
		return
	}

	execInfo := fmt.Sprintf("request-id:%s exit-code:%d", res.RequestID, res.ExitStatus)
	status := "status:Ok"
	e.SetTextColor(tcell.ColorGreen)
	if res.ExitStatus != 0 || res.Err != nil {
		e.SetTextColor(tcell.ColorRed)
		status = "status:Failed"
	}
	errInfo := "none"
	if res.Err != nil {
		errInfo = res.Err.Error()
	}

	e.SetText(fmt.Sprintf("\n job   \t| %s \n info  \t| %s\n error \t| %s\n",
		status, execInfo, errInfo))
	e.SetTextAlign(tview.AlignLeft)
}

func (e *executionResultView) InProgress() {
	e.Clear()
	status := "status:Busy"
	e.SetTextColor(tcell.ColorYellow)
	e.SetTextAlign(tview.AlignLeft)
	e.SetText(fmt.Sprintf("\n info  \t| %s", status))
}

func newExecutionResultView(rootUI *Tui) *executionResultView {
	return &executionResultView{
		detailView: newDetailView("execution status", true, rootUI),
	}
}
