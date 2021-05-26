package tui

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/run"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"io"
)

func LaunchUI(gCtx *snippet.Context, name string) error {
	ui, err := NewTui(gCtx, name)
	if err != nil {
		return err
	}
	return ui.Run()
}

type Tui struct {
	// Primary UI Application component
	App *tview.Application

	info        *info
	PbTree      *PlayBookTree
	snippetView *snippetView
	execOutView *executionOutputView
	execResView *executionResultView
	statusView  *detailView

	pages *tview.Pages // different pages in this UI
	grid  *tview.Grid  //  layout for the run page
	run   *run.Run

	Nav *navigator

	// curNodeID refers the current node id selected. "" if unselected
	curNodeID string
}

func NewTui(gCtx *snippet.Context, name string) (*Tui, error) {
	run, err := run.NewRun(gCtx, name)
	if err != nil {
		return nil, err
	}

	ui := &Tui{
		App: tview.NewApplication(),
		run: run,
	}

	ui.info = newInfo(ui, gCtx)
	ui.snippetView = newSnippetView(ui)
	ui.execOutView = newExecutionOutputView(ui)
	ui.execResView = newExecutionResultView(ui)
	ui.statusView = newDetailView("", false, ui)
	pbTree, err := NewPlaybookTree(ui, run.Root)
	if err != nil {
		return nil, fmt.Errorf("newPlaybookTree: err = %v", err)
	}
	ui.PbTree = pbTree

	ui.setupNavigator()
	ui.grid = tview.NewGrid().
		SetRows(5, 0, 6, 0, 4).
		AddItem(ui.info, 0, 0, 1, 5, 0, 0, false).
		AddItem(ui.PbTree, 1, 0, 3, 1, 0, 0, true).
		AddItem(ui.snippetView, 1, 1, 1, 4, 0, 0, false).
		AddItem(ui.execResView, 2, 1, 1, 4, 0, 0, false).
		AddItem(ui.execOutView, 3, 1, 1, 4, 0, 0, false).
		AddItem(ui.statusView, 4, 0, 1, 5, 0, 0, false)
	ui.pages = tview.NewPages().AddPage("main", ui.grid, true, true)
	ui.App.SetRoot(ui.pages, true)
	ui.setFocusedItem()
	return ui, nil
}

func (t *Tui) Run() error { return t.App.Run() }

func (t *Tui) Navigate(key tcell.Key) {
	t.Nav.Navigate(key)
	t.setFocusedItem()
}

func (t *Tui) setupNavigator() {
	t.Nav = &navigator{
		rootUI:  t.App,
		Entries: []navigable{t.PbTree, t.snippetView, t.execResView, t.execOutView, t.statusView},
	}
}

func (t *Tui) UpdateCurrentNodeID(nodeID string) error {
	tools.Log.Info().Msgf("UpdateCurrentNodeID nodeID: %s", nodeID)
	t.curNodeID = nodeID
	return t.Refresh()
}

func (t *Tui) Refresh() error {
	if t.curNodeID == "" {
		// clear out the views
		t.snippetView.Refresh(nil)
		t.execOutView.Refresh(nil)
		t.execResView.Refresh(nil)
		return nil
	}

	view, err := t.run.ViewIndex.Get(t.curNodeID)
	if err != nil {
		return err
	}
	t.snippetView.Refresh(view)

	execResult, _ := t.run.ExecIndex.Get(t.curNodeID)
	t.execResView.Refresh(execResult)
	t.execOutView.Refresh(execResult)
	return nil
}

func (t *Tui) SetCurrentNodeView(nodeView *graph.NodeView) {
	t.curNodeID = nodeView.Node.ID
	t.snippetView.Refresh(nodeView)
}

func (t *Tui) SetExecutionInProgress() {
	t.execOutView.Refresh(nil)
	t.execResView.InProgress()
}

func (t *Tui) setFocusedItem() {
	n, found := t.Nav.CurrentFocusedItem()
	if !found {
		return
	}
	t.setStatusHelp(n.NavHelp())
}

func (t *Tui) setStatusHelp(s string) {
	t.statusView.Clear()
	t.statusView.SetTextAlign(tview.AlignLeft)
	t.statusView.SetTextColor(tcell.ColorDarkCyan)
	_, err := t.statusView.Write([]byte(s))
	if err != nil {
		tools.Log.Err(err).Msgf("StatusHelp: err = %v", err)
	}
}

func (t *Tui) StatusErrorf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	tools.Log.Info().Msg(s)
	t.statusView.Clear()
	t.statusView.SetTextAlign(tview.AlignLeft)
	t.statusView.SetTextColor(tcell.ColorRed)
	if _, err := t.statusView.Write([]byte(s)); err != nil {
		tools.Log.Err(err).Msgf("WriteStatus: err = %v", err)
	}
}

func (t *Tui) Execute(n *graph.Node, stdout, stderr io.Writer) (*graph.NodeExecutionResult, error) {
	return t.run.Execute(n, stdout, stderr)
}

func (t *Tui) ExecuteCurrentNode() {
	tools.Log.Info().Msgf("Execute current node id:[%s]", t.curNodeID)
	if t.curNodeID == "" {
		t.StatusErrorf("ExecuteCurrentNode: cannot execute empty node")
		return
	}

	// ToDo: for some reason the in-progress is not shown in the UX
	t.SetExecutionInProgress()
	view, err := t.run.ViewIndex.Get(t.curNodeID)
	if err != nil {
		t.StatusErrorf("ExecuteCurrentNode: err = %v", err)
	}

	if _, err := t.Execute(view.Node, t.execOutView, t.execOutView); err != nil {
		t.StatusErrorf("ExecuteCurrentNode  id:[%s]: err = %v", t.curNodeID, err)
		return
	}

	if err := t.Refresh(); err != nil {
		t.StatusErrorf("ExecuteCurrentNode: Refresh: err = %v", err)
	}
}

func (t *Tui) EditCurrentNode() {
	tools.Log.Info().Msgf("EditSnippetContent node-id:[%s]", t.curNodeID)
	if t.curNodeID == "" {
		t.StatusErrorf("EditSnippetContent: cannot edit empty node")
		return
	}

	t.App.Suspend(func() {
		_, err := t.run.EditSnippet(t.curNodeID)
		if err != nil {
			t.StatusErrorf("edit snippet failed %v", err)
			return
		}
	})

	if err := t.UpdateCurrentNodeID(t.curNodeID); err != nil {
		t.StatusErrorf("EditSnippetContent: UpdateCurrentNodeID:  err = %v", err)
	}
}
