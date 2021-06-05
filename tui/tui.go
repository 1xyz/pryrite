package tui

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/run"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	mainPage = "main"
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

	info          *info
	PbTree        *PlayBookTree
	snippetView   *snippetView
	execOutView   *executionOutputView
	blockExecView *blockExecutionsView
	statusView    *detailView

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
	//ui.execResView = newExecutionResultView(ui)
	ui.blockExecView = newBlockExecutionsView(ui)
	ui.statusView = newDetailView("", false, ui)
	pbTree, err := NewPlaybookTree(ui, run.Root)
	if err != nil {
		return nil, fmt.Errorf("newPlaybookTree: err = %v", err)
	}
	ui.PbTree = pbTree

	ui.setupNavigator()
	ui.grid = tview.NewGrid().
		SetRows(5, 0, 6, 0, 4).
		SetColumns(-1, -4).
		AddItem(ui.info, 0, 0, 1, 2, 0, 0, false).
		AddItem(ui.PbTree, 1, 0, 3, 1, 0, 0, true).
		AddItem(ui.snippetView, 1, 1, 1, 1, 0, 0, false).
		AddItem(ui.blockExecView, 2, 1, 1, 1, 0, 0, false).
		AddItem(ui.execOutView, 3, 1, 1, 1, 0, 0, false).
		AddItem(ui.statusView, 4, 0, 1, 2, 0, 0, false)
	ui.pages = tview.NewPages().AddPage(mainPage, ui.grid, true, true)
	ui.App.SetRoot(ui.pages, true)
	if err := ui.UpdateCurrentNodeID(run.Root.Node.ID); err != nil {
		ui.StatusErrorf("NewTui: UpdateCurrentNodeID: id=%s err = %v", run.Root.Node.ID, err)
	}
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
		Entries: []navigable{t.PbTree, t.snippetView, t.blockExecView, t.execOutView, t.statusView},
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
		//t.execResView.Refresh(nil)
		t.blockExecView.Refresh(nil)
		return nil
	}

	view, err := t.run.ViewIndex.Get(t.curNodeID)
	if err != nil {
		return err
	}
	t.snippetView.Refresh(view)

	execResults, _ := t.run.ExecIndex.Get(t.curNodeID)
	t.blockExecView.Refresh(execResults)
	//t.execResView.Refresh(execResults)
	t.execOutView.Refresh(execResults)
	return nil
}

func (t *Tui) SetCurrentNodeView(nodeView *graph.NodeView) {
	t.curNodeID = nodeView.Node.ID
	t.snippetView.Refresh(nodeView)
}

func (t *Tui) SetExecutionInProgress() {
	t.execOutView.Refresh(nil)
	//t.execResView.InProgress()
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

func (t *Tui) GetBlock(blockID string) (*graph.Block, error) {
	return t.run.GetBlock(t.curNodeID, blockID)
}

func (t *Tui) ExecuteSelectedBlock(blockID string) error {
	t.SetExecutionInProgress()
	if err := t.CheckCurrentNode(); err != nil {
		return err
	}

	tools.Log.Info().Msgf("ExecuteSelectedBlock [%s][%s]", t.curNodeID, blockID)
	view, err := t.run.ViewIndex.Get(t.curNodeID)
	if err != nil {
		t.StatusErrorf("ExecuteSelectedBlock: err = %v", err)
		return err
	}

	b, err := t.GetBlock(blockID)
	if err != nil {
		return err
	}

	if _, err := t.run.ExecuteBlock(view.Node, b, t.execOutView, t.execOutView); err != nil {
		t.StatusErrorf("ExecuteSelectedBlock  id:[%s]: err = %v", t.curNodeID, err)
		return err
	}

	if err := t.Refresh(); err != nil {
		t.StatusErrorf("ExecuteSelectedBlock: Refresh: err = %v", err)
		return err
	}
	return nil
}

func (t *Tui) ExecuteCurrentNode() error {
	if err := t.CheckCurrentNode(); err != nil {
		return err
	}

	tools.Log.Info().Msgf("ExecuteCurrentNode [%s]", t.curNodeID)
	view, err := t.run.ViewIndex.Get(t.curNodeID)
	if err != nil {
		return err
	}

	if err := t.run.ExecuteNode(view.Node, t.execOutView, t.execOutView); err != nil {
		return err
	}

	if err := t.Refresh(); err != nil {
		return err
	}
	return nil
}

func (t *Tui) EditSelectedBlock(blockID string) error {
	if err := t.CheckCurrentNode(); err != nil {
		return err
	}
	tools.Log.Info().Msgf("EditSelectedBlock node-id:[%s]", t.curNodeID)
	t.App.Suspend(func() {
		if _, _, err := t.run.EditBlock(t.curNodeID, blockID, true /*save*/); err != nil {
			t.StatusErrorf("EditSelectedBlock: [%s][%s] failed err = %v", t.curNodeID, blockID, err)
		}
	})
	return t.Refresh()
}

func (t *Tui) EditCurrentNode() {
	if err := t.CheckCurrentNode(); err != nil {
		return
	}

	tools.Log.Info().Msgf("EditSnippetContent node-id:[%s]", t.curNodeID)
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
	if err := t.PbTree.RefreshNode(t.curNodeID); err != nil {
		t.StatusErrorf("EditSnippetContent: RefreshNodeTitle: err = %v", err)
	}
}

func (t *Tui) CheckCurrentNode() error {
	if t.curNodeID == "" {
		t.StatusErrorf("IsCurrentNodeSet: no node selected")
		return fmt.Errorf("no node selected")
	}
	return nil
}

func (t *Tui) CommonKeyBindings(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlR:
		tools.Log.Info().Msgf("PlayBookTree: Ctrl+R request to run node")
		go func() {
			if err := t.ExecuteCurrentNode(); err != nil {
				t.StatusErrorf("ExecuteCurrentNode [%s] failed err = %v", t.curNodeID, err)
			}
		}()
	case tcell.KeyCtrlE:
		tools.Log.Info().Msgf("PlayBookTree: Ctrl+E request to edit node")
	}

	return event
}

func (t *Tui) closeAndSwitchPanel(removePanel, switchPanel string) {
	t.pages.RemovePage(removePanel).ShowPage(mainPage)
}

func (t *Tui) displayInspect(data, page string) {
	text := tview.NewTextView()
	text.SetTitle("Detail").SetTitleAlign(tview.AlignLeft)
	text.SetBorder(true)
	text.SetText(data)

	// get the index of the current focused item
	idx := t.Nav.GetCurrentFocusedIndex()
	text.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Rune() == 'q' {
			t.closeAndSwitchPanel("detail", page)
			// reset the focus back again
			t.Nav.SetCurrentFocusedIndex(idx)
		}
		return event
	})

	t.pages.AddAndSwitchToPage("detail", text, true)
}

func (t *Tui) InspectBlockExecution(blockID string) {
	tools.Log.Info().Msgf("InspectBlockExecution blockID: %s", blockID)
	t.displayInspect("hello world", "blocks")
}
