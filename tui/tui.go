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

func (t *Tui) SetCurrentNodeView(nodeView *graph.NodeView) {
	t.snippetView.SetCurrentNodeView(nodeView)
}

func (t *Tui) SetCurrentNodeExecutionResult(res *graph.NodeExecutionResult) {
	t.execOutView.setExecutionResult(res)
	t.execResView.setExecutionResult(res)
}

func (t *Tui) SetExecutionInProgress() {
	t.execOutView.setExecutionResult(nil)
	t.execResView.setExecutionInProgress()
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
