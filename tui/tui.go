package tui

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/tui/common"

	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/run"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v2"
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

	gCtx        *snippet.Context
	info        *info
	PbTree      *PlayBookTree
	snippetView *snippetView
	consoleView *consoleView
	execView    *executionsView
	//statusView   *detailView
	navHelp      *navView
	activityView *activityView

	pages *tview.Pages // different pages in this UI
	grid  *tview.Grid  //  layout for the run page
	run   *run.Run

	Nav *common.Navigator

	// curNodeID refers the current node id selected. "" if unselected
	curNodeID string
}

func NewTui(gCtx *snippet.Context, name string) (*Tui, error) {
	run, err := run.NewRun(gCtx, name)
	if err != nil {
		return nil, err
	}

	ui := &Tui{
		gCtx: gCtx,
		App:  tview.NewApplication(),
		run:  run,
	}

	ui.info = newInfo(ui, gCtx)
	ui.snippetView = newSnippetView(ui)
	ui.consoleView = newExecutionOutputView(ui)
	ui.execView = newExecutionsView(ui)
	ui.activityView = newActivityView(ui)
	//ui.statusView = newDetailView("", false, ui)
	ui.navHelp = newNavView(ui)
	pbTree, err := NewPlaybookTree(ui, run.Root)
	if err != nil {
		return nil, fmt.Errorf("newPlaybookTree: err = %v", err)
	}
	ui.PbTree = pbTree

	ui.setupNavigator()
	ui.grid = tview.NewGrid().
		SetRows(4, 0, 6, 0, 5, 8).
		SetColumns(-1, -4).
		AddItem(ui.info, 0, 0, 1, 2, 0, 0, false).
		AddItem(ui.PbTree, 1, 0, 3, 1, 0, 0, true).
		AddItem(ui.snippetView, 1, 1, 1, 1, 0, 0, false).
		AddItem(ui.execView, 2, 1, 1, 1, 0, 0, false).
		AddItem(ui.consoleView, 3, 1, 2, 1, 0, 0, false).
		AddItem(ui.navHelp, 4, 0, 1, 1, 0, 0, false).
		AddItem(ui.activityView, 5, 0, 1, 2, 0, 0, false)
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
	t.Nav = &common.Navigator{
		RootUI:  t.App,
		Entries: []common.Navigable{t.PbTree, t.snippetView, t.execView, t.consoleView, t.activityView},
	}
}

func (t *Tui) UpdateCurrentNodeID(nodeID string) error {
	tools.Log.Info().Msgf("UpdateCurrentNodeID nodeID: %s", nodeID)
	t.curNodeID = nodeID
	return t.Refresh()
}

// NOTE: this must never be run directly on a child goroutine (it will clash with Draw)
func (t *Tui) Refresh() error {
	if t.curNodeID == "" {
		// clear out the views
		t.snippetView.Refresh(nil)
		t.consoleView.Refresh(nil)
		//t.execResView.Refresh(nil)
		t.execView.Refresh(nil)
		return nil
	}

	view, err := t.run.ViewIndex.Get(t.curNodeID)
	if err != nil {
		return err
	}
	t.snippetView.Refresh(view)

	execResults, _ := t.run.ExecIndex.Get(t.curNodeID)
	t.execView.Refresh(execResults)
	//t.execResView.Refresh(execResults)
	t.consoleView.Refresh(execResults)
	return nil
}

// This is safe to run outside of the main goroutine
func (t *Tui) QueueRefresh(location string) {
	t.App.QueueUpdateDraw(func() {
		if err := t.Refresh(); err != nil {
			t.StatusErrorf("%s: Refresh: err = %v", location, err)
		}
	})
}

func (t *Tui) SetCurrentNodeView(nodeView *graph.NodeView) {
	t.curNodeID = nodeView.Node.ID
	t.snippetView.Refresh(nodeView)
}

func (t *Tui) SetExecutionInProgress() {
	t.consoleView.Refresh(nil)
	//t.execResView.InProgress()
}

func (t *Tui) setFocusedItem() {
	n, found := t.Nav.CurrentFocusedItem()
	if !found {
		return
	}
	t.navHelp.Refresh(n)
}

func (t *Tui) StatusErrorf(format string, v ...interface{}) { t.activityView.Log(Error, format, v...) }
func (t *Tui) StatusInfof(format string, v ...interface{})  { t.activityView.Log(Info, format, v...) }
func (t *Tui) GetContext() *snippet.Context                 { return t.gCtx }

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

	if _, err := t.run.ExecuteBlock(view.Node, b, t.consoleView, t.consoleView); err != nil {
		t.StatusErrorf("ExecuteSelectedBlock  id:[%s]: err = %v", t.curNodeID, err)
		return err
	}

	t.QueueRefresh("ExecuteSelectedBlock")

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

	if err := t.run.ExecuteNode(view.Node, t.consoleView, t.consoleView); err != nil {
		return err
	}

	t.QueueRefresh("ExecuteCurrentNode")

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

func (t *Tui) displayInspect(data, title, page string) {
	text := tview.NewTextView()
	text.SetTitle(title).
		SetTitleAlign(tview.AlignLeft)
	text.SetBorder(true)
	text.SetRegions(true)
	text.SetDynamicColors(true)
	text.SetText(data)
	text.Write([]byte("\n[yellow]Press ESC to go back.[white]"))

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

func (t *Tui) InspectBlockExecution(requestID string) {
	res, err := t.run.GetBlockExecutionResult(t.curNodeID, requestID)
	if err != nil {
		t.StatusErrorf("InspectBlockExecution: req-id: %s err = %v", requestID, err)
		return
	}

	d := struct {
		*graph.BlockExecutionResult
		Stdout string `yaml:"stdout"`
		Stderr string `yaml:"stderr"`
	}{res, string(res.Stdout), string(res.Stderr)}
	b, err := yaml.Marshal(&d)
	if err != nil {
		t.StatusErrorf("InspectBlockExecution: req-id: %s err = %v", requestID, err)
		return
	}

	t.displayInspect(string(b), "Execution detail", "blocks")
}

func (t *Tui) InspectActivity(a *activity) {
	b, err := yaml.Marshal(a)
	if err != nil {
		t.StatusErrorf("InspectActivity: yaml.Marshall err = %v", err)
		return
	}

	t.displayInspect(string(b), "Activity", "activity")
}
