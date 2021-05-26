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
	statusView  *DetailPane

	pages *tview.Pages // different pages in this UI
	grid  *tview.Grid  //  layout for the run page
	run   *run.Run

	Nav *Navigator
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
	ui.execResView = newExecutionResultView()
	ui.statusView = NewDetailPane("Status", ui)
	pbTree, err := NewPlaybookTree(ui, run.Root)
	if err != nil {
		return nil, fmt.Errorf("newPlaybookTree: err = %v", err)
	}
	ui.PbTree = pbTree

	ui.setupNavigator()
	// 100 x 100 grid
	ui.grid = tview.NewGrid().
		SetRows(3, 0, 0, 3, 3).
		AddItem(ui.info, 0, 0, 1, 5, 0, 0, false).
		AddItem(pbTree.View, 1, 0, 3, 1, 0, 0, true).
		AddItem(ui.snippetView, 1, 1, 1, 4, 0, 0, false).
		AddItem(ui.execOutView, 2, 1, 1, 4, 0, 0, false).
		AddItem(ui.execResView, 3, 1, 1, 4, 0, 0, false).
		AddItem(ui.statusView, 4, 0, 1, 5, 0, 0, false)
	ui.pages = tview.NewPages().AddPage("main", ui.grid, true, true)
	ui.App.SetRoot(ui.pages, true)
	return ui, nil
}

func (t *Tui) Run() error             { return t.App.Run() }
func (t *Tui) Navigate(key tcell.Key) { t.Nav.Navigate(key) }
func (t *Tui) setupNavigator() {
	t.Nav = &Navigator{
		rootUI:  t.App,
		Entries: []tview.Primitive{t.PbTree.View, t.snippetView, t.execOutView, t.statusView},
	}
}

func (t *Tui) SetCurrentNodeView(nodeView *graph.NodeView) {
	t.snippetView.SetCurrentNodeView(nodeView)
}
func (t *Tui) ResetNodeExecutionResult() { t.SetCurrentNodeExecutionResult(nil) }
func (t *Tui) SetCurrentNodeExecutionResult(res *graph.NodeExecutionResult) {
	t.execOutView.setExecutionResult(res)
}

func (t *Tui) Statusf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	tools.Log.Info().Msg(s)
	if _, err := t.statusView.Write([]byte(s)); err != nil {
		tools.Log.Err(err).Msgf("WriteStatus: err = %v", err)
	}
}

func (t *Tui) Execute(n *graph.Node, stdout, stderr io.Writer) (*graph.NodeExecutionResult, error) {
	return t.run.Execute(n, stdout, stderr)
}

type Navigator struct {
	rootUI  *tview.Application
	Entries []tview.Primitive
}

func (n *Navigator) Navigate(key tcell.Key) {
	switch key {
	case tcell.KeyBacktab:
		n.Prev()
	case tcell.KeyTab:
		n.Next()
	}
}

func (n *Navigator) Next() {
	index := n.getCurrentFocus()
	next := 0
	if index == -1 || index == len(n.Entries)-1 {
		next = 0
	} else {
		next = index + 1
	}
	n.rootUI.SetFocus(n.Entries[next])
}

func (n *Navigator) Prev() {
	index := n.getCurrentFocus()
	next := 0
	if index == 0 {
		next = len(n.Entries) - 1
	} else if index == -1 {
		next = 0
	} else {
		next = index - 1
	}
	n.rootUI.SetFocus(n.Entries[next])
}

func (n *Navigator) getCurrentFocus() int {
	for i, e := range n.Entries {
		if e.HasFocus() {
			return i
		}
	}
	return -1
}
