package explorer

import (
	"fmt"
	"os"

	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/aardlabs/terminal-poc/tui/common"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	app   *tview.Application
	grid  *tview.Grid
	pages *tview.Pages

	focusColor tcell.Color

	explorer *NodeExplorer

	nodeTreeView *nodeTreeView
	navView      *navView
	statusView   *statusView
	contentView  *contentView
	navigator    *common.Navigator
	infoView     *infoView
	helpview     *tview.TextView
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
func (u *UI) SetNavHelp(entries [][]string)                { u.navView.SetHelp(entries) }
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

func (u *UI) ShowHelpScreen() {
	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		g := tview.NewGrid().
			SetColumns(0, width, 0).
			SetRows(0, height, 0).
			AddItem(p, 1, 1, 1, 1, 0, 0, true)
		return g
	}

	help := [][]string{
		{"Enter", "Execute code block"},
		{"Ctrl + R", "Open node in Runner UI"},
		{"Ctrl + Space", "Print code block and exit"},
	}

	dlg := newNavView(u)
	dlg.SetHelp(help)
	dlg.SetBorder(true).
		SetTitle("Help...")
	dlg.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEsc {
			u.pages.RemovePage("help")
		}
	})
	dlg.SetBorderPadding(1, 1, 2, 1)
	m := modal(dlg, 43, 10)
	u.pages.AddPage("help", m, true, true)
}

// ExecuteCmdDialog shows a modal dialog navigating a user to execute the provided command
func (u *UI) ExecuteCmdDialog(cmd, title string) {
	// Returns a new primitive which puts the provided primitive in the center and
	// sets its size to the given width and height.
	// Returns a new primitive which puts the provided primitive in the center and
	// sets its size to the given width and height.
	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewGrid().
			SetColumns(0, width, 0).
			SetRows(0, height, 0).
			AddItem(p, 1, 1, 1, 1, 0, 0, true)
	}

	dlg := tview.NewModal().
		SetText(title).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				u.app.Stop()
				fmt.Printf(">> %s\n", cmd)
				if err := tools.BashExec(cmd); err != nil {
					fmt.Printf("error = %v", err)
					os.Exit(1)
				} else {
					os.Exit(0)
				}
			} else {
				u.pages.RemovePage("execute")
			}
		})
	m := modal(dlg, 40, 10)
	u.pages.AddPage("execute", m, true, true)
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
		app:        app,
		explorer:   explorer,
		focusColor: tcell.ColorYellow,
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
	ui.navigator = common.NewNavigator(
		ui.app,
		[]common.Navigable{ui.nodeTreeView, ui.contentView},
		ui.Stop,
	)
	ui.helpview = tview.NewTextView()
	ui.helpview.SetTextColor(tcell.ColorYellow)
	ui.helpview.SetText("Ctrl+H for help")
	ui.helpview.SetTextAlign(tview.AlignRight)

	ui.grid = tview.NewGrid().
		SetRows(0, 2).
		SetColumns(0, 0).
		AddItem(ui.nodeTreeView, 0, 0, 1, 1, 0, 0, true).
		AddItem(ui.contentView, 0, 1, 1, 1, 0, 0, false).
		//AddItem(ui.infoView, 1, 1, 1, 1, 0, 0, false).
		AddItem(ui.statusView, 1, 0, 1, 1, 0, 0, false).
		AddItem(ui.helpview, 1, 1, 1, 1, 0, 0, false)

	ui.pages = tview.NewPages().
		AddPage("main", ui.grid, true, true)
	ui.app.SetRoot(ui.pages, true)
	return ui, nil
}
