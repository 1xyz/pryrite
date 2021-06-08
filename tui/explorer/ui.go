package explorer

import (
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/rivo/tview"
)

type UI struct {
	app  *tview.Application
	grid *tview.Grid

	explorer *NodeExplorer
	query    string

	nodeTreeView *nodeTreeView
	navView      *navView
	statusView   *statusView
}

func (u *UI) Run() error {
	return u.app.Run()
}

func (u *UI) Stop() {
	u.app.Stop()
}

func (u *UI) GetChildren(nodeID string) []*graph.Node {
	children, err := u.explorer.GetChildren(nodeID)
	if err != nil {
		tools.Log.Err(err).Msg("GetChildren: explorer.GetChildren")
		u.StatusErrorf("GetChildren failed for node %s err = %v", nodeID, err)
		return nil
	}
	u.StatusInfof("GetChildren completed for node %s", nodeID)
	return children
}

func (u *UI) SetNavHelp(entries [][]string) {
	u.navView.setHelp(entries)
}

func (u *UI) StatusInfof(format string, v ...interface{}) {
	u.statusView.Infof(format, v)
}

func (u *UI) StatusErrorf(format string, v ...interface{}) {
	u.statusView.Errorf(format, v)
}

func NewUI(gCtx *snippet.Context, query string, nodes []*graph.Node) (*UI, error) {
	explorer, err := NewNodeExplorer(gCtx, nodes)
	if err != nil {
		return nil, err
	}

	app := tview.NewApplication()
	ui := &UI{
		app:      app,
		query:    query,
		explorer: explorer,
	}

	entriesView, err := newNodeTreeView(ui, nodes, "Search result")
	if err != nil {
		return nil, err
	}
	ui.nodeTreeView = entriesView
	ui.navView = newNavView(ui)
	ui.statusView = newStatusView(ui)

	ui.grid = tview.NewGrid().
		SetRows(-10, 0, 2).
		SetColumns(-4, 0).
		AddItem(ui.nodeTreeView, 0, 0, 1, 1, 0, 0, true).
		AddItem(ui.navView, 0, 1, 3, 1, 0, 0, false).
		AddItem(ui.statusView, 2, 0, 1, 2, 0, 0, false)

	ui.app.SetRoot(ui.grid, true)
	return ui, nil
}
