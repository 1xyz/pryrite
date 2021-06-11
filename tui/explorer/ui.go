package explorer

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/aardlabs/terminal-poc/tui/common"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	app  *tview.Application
	grid *tview.Grid

	explorer *NodeExplorer

	nodeTreeView *nodeTreeView
	navView      *navView
	statusView   *statusView
	contentView  *contentView
	navigator    *common.Navigator
	infoView     *infoView
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

func (u *UI) Run() error                                   { return u.app.Run() }
func (u *UI) Stop()                                        { u.app.Stop() }
func (u *UI) SetNavHelp(entries [][]string)                { u.navView.setHelp(entries) }
func (u *UI) SetContentNode(n *graph.Node)                 { u.contentView.SetNode(n) }
func (u *UI) SetContentBlock(b *graph.Block)               { u.contentView.SetBlock(b) }
func (u *UI) SetInfoBlock(b *graph.Block)                  { u.infoView.SetBlock(b) }
func (u *UI) SetInfoNode(n *graph.Node)                    { u.infoView.SetNode(n) }
func (u *UI) StatusInfof(format string, v ...interface{})  { u.statusView.Infof(format, v...) }
func (u *UI) StatusErrorf(format string, v ...interface{}) { u.statusView.Errorf(format, v...) }
func (u *UI) GetContext() *snippet.Context                 { return u.explorer.gCtx }

func (u *UI) Navigate(key tcell.Key) {
	u.navigator.Navigate(key)
	n, ok := u.navigator.CurrentFocusedItem()
	if ok {
		u.SetNavHelp(n.NavHelp())
	}
}

func NewUI(gCtx *snippet.Context, title, borderTitle string, nodes []*graph.Node) (*UI, error) {
	if nodes == nil || len(nodes) == 0 {
		return nil, fmt.Errorf("no entries found")
	}

	explorer, err := NewNodeExplorer(gCtx, nodes)
	if err != nil {
		return nil, err
	}

	app := tview.NewApplication()
	ui := &UI{
		app:      app,
		explorer: explorer,
	}

	entriesView, err := newNodeTreeView(ui, nodes, title, borderTitle)
	if err != nil {
		return nil, err
	}
	ui.nodeTreeView = entriesView
	ui.navView = newNavView(ui)
	ui.statusView = newStatusView(ui)
	ui.contentView = newContentView(ui)
	ui.infoView = newInfoView(ui)
	ui.navigator = &common.Navigator{
		RootUI:  ui.app,
		Entries: []common.Navigable{ui.nodeTreeView, ui.contentView},
	}
	ui.grid = tview.NewGrid().
		SetRows(-2, 0, 1).
		SetColumns(-1, -4, 0).
		AddItem(ui.nodeTreeView, 0, 0, 1, 2, 0, 0, true).
		AddItem(ui.navView, 0, 2, 1, 1, 0, 0, false).
		AddItem(ui.contentView, 1, 1, 1, 1, 0, 0, false).
		AddItem(ui.infoView, 1, 0, 1, 1, 0, 0, false).
		AddItem(ui.statusView, 2, 0, 1, 2, 0, 0, false)

	ui.app.SetRoot(ui.grid, true)
	return ui, nil
}
