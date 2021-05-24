package tui

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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

	PbTree     *PlayBookTree
	Detail     *DetailPane
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
	}

	ui.Detail = NewDetailPane("Node detail", ui)
	ui.ExecDetail = NewDetailPane("Execution result", ui)
	ui.Status = NewDetailPane("Status", ui)
	pbTree, err := NewPlaybookTree(ui, rc.Root, ui.Detail)
	if err != nil {
		return nil, fmt.Errorf("newPlaybookTree: err = %v", err)
	}
	ui.PbTree = pbTree

	ui.setupNavigator()
	// 100 x 100 grid
	ui.grid = tview.NewGrid().
		AddItem(pbTree.View, 0, 0, 4, 1, 0, 0, true).
		AddItem(ui.Detail.View, 0, 1, 2, 4, 0, 0, false).
		AddItem(ui.ExecDetail.View, 2, 1, 2, 4, 0, 0, false).
		AddItem(ui.Status.View, 4, 0, 1, 5, 0, 0, false)
	ui.pages = tview.NewPages().AddPage("main", ui.grid, true, true)
	ui.App.SetRoot(ui.pages, true)
	return ui, nil
}

func (t *Tui) Run() error             { return t.App.Run() }
func (t *Tui) Navigate(key tcell.Key) { t.Nav.Navigate(key) }
func (t *Tui) setupNavigator() {
	t.Nav = &Navigator{
		rootUI:  t.App,
		Entries: []tview.Primitive{t.PbTree.View, t.Detail.View, t.ExecDetail.View, t.Status.View},
	}
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
