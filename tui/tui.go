package tui

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
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

	info       *info
	PbTree     *PlayBookTree
	Detail     *SnippetPane
	ExecDetail *DetailPane
	Status     *DetailPane

	pages *tview.Pages // different pages in this UI
	grid  *tview.Grid  //  layout for the run page
	rc    *RunContext

	Nav *Navigator
}

func NewTui(gCtx *snippet.Context, name string) (*Tui, error) {
	rc, err := BuildRunContext(gCtx, name)
	if err != nil {
		return nil, err
	}

	ui := &Tui{
		App: tview.NewApplication(),
		rc:  rc,
	}

	ui.info = newInfo(ui, gCtx)
	ui.Detail = NewSnippetPane(ui)
	ui.ExecDetail = NewDetailPane("Execution result", ui)
	ui.Status = NewDetailPane("Status", ui)
	pbTree, err := NewPlaybookTree(ui, rc.Root)
	if err != nil {
		return nil, fmt.Errorf("newPlaybookTree: err = %v", err)
	}
	ui.PbTree = pbTree

	ui.setupNavigator()
	// 100 x 100 grid
	ui.grid = tview.NewGrid().
		SetRows(3, -7, -1, -1).
		AddItem(ui.info, 0, 0, 1, 5, 0, 0, false).
		AddItem(pbTree.View, 1, 0, 3, 1, 0, 0, true).
		AddItem(ui.Detail, 1, 1, 2, 4, 0, 0, false).
		AddItem(ui.ExecDetail, 2, 1, 2, 4, 0, 0, false).
		AddItem(ui.Status, 4, 0, 1, 5, 0, 0, false)
	ui.pages = tview.NewPages().AddPage("main", ui.grid, true, true)
	ui.App.SetRoot(ui.pages, true)
	return ui, nil
}

func (t *Tui) Run() error             { return t.App.Run() }
func (t *Tui) Navigate(key tcell.Key) { t.Nav.Navigate(key) }
func (t *Tui) setupNavigator() {
	t.Nav = &Navigator{
		rootUI:  t.App,
		Entries: []tview.Primitive{t.PbTree.View, t.Detail, t.ExecDetail, t.Status},
	}
}

func (t *Tui) SetCurrentNodeView(nodeView *graph.NodeView) {
	t.Detail.SetCurrentNodeView(nodeView)
}

func (t *Tui) ClearDetail() {
	t.Detail.Clear()
}

func (t *Tui) WriteDetail(p []byte) {
	if _, err := t.Detail.Write(p); err != nil {
		t.Statusf("WriteDetail: err = %v", err)
	}
}

func (t *Tui) Statusf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	tools.Log.Info().Msg(s)
	if _, err := t.Status.Write([]byte(s)); err != nil {
		tools.Log.Err(err).Msgf("WriteStatus: err = %v", err)
	}
}

func (t *Tui) Execute(n *graph.Node, stdout, stderr io.Writer) (*graph.NodeExecutionResult, error) {
	return t.rc.Execute(n, stdout, stderr)
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
