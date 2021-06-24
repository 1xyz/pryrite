package explorer

import (
	"context"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	executor "github.com/aardlabs/terminal-poc/executors"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/aardlabs/terminal-poc/tui/common"
)

type UI struct {
	app   *tview.Application
	grid  *tview.Grid
	pages *tview.Pages

	focusColor tcell.Color

	explorer *NodeExplorer
	register *executor.Register

	nodeTreeView *nodeTreeView
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
func (u *UI) SetContentNode(n *graph.Node)                 { u.contentView.SetNode(n) }
func (u *UI) SetContentBlock(b *graph.Block)               { u.contentView.SetBlock(b) }
func (u *UI) SetInfoBlock(b *graph.Block)                  { u.infoView.SetBlock(b) }
func (u *UI) SetInfoNode(n *graph.Node)                    { u.infoView.SetNode(n) }
func (u *UI) StatusInfof(format string, v ...interface{})  { u.statusView.Infof(format, v...) }
func (u *UI) StatusErrorf(format string, v ...interface{}) { u.statusView.Errorf(format, v...) }
func (u *UI) GetContext() *snippet.Context                 { return u.explorer.gCtx }
func (u *UI) Navigate(key tcell.Key)                       { u.navigator.Navigate(key) }

func (u *UI) ShowHelpScreen() {
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
	m := newModal(dlg, 43, 10)
	u.pages.AddPage("help", m, true, true)
}

// ExecuteCmdDialog shows a modal dialog navigating a user to execute the provided command
func (u *UI) ExecuteCmdDialog(cmd, title string) {
	u.askExecuteBlockOrCmd(nil, cmd, title)
}
func (u *UI) ExecuteBlockDialog(block *graph.Block, title string) {
	u.askExecuteBlockOrCmd(block, "", title)
}

func NewUI(gCtx *snippet.Context, title, borderTitle string, nodes []*graph.Node) (*UI, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no entries found")
	}

	explorer, err := NewNodeExplorer(gCtx, nodes)
	if err != nil {
		return nil, err
	}

	register, err := executor.NewRegister()
	if err != nil {
		return nil, err
	}

	app := tview.NewApplication()
	ui := &UI{
		app:        app,
		explorer:   explorer,
		register:   register,
		focusColor: tcell.ColorYellow,
	}

	entriesView, err := newNodeTreeView(ui, nodes, title, borderTitle)
	if err != nil {
		return nil, err
	}
	ui.nodeTreeView = entriesView
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

// Returns a new primitive which puts the provided primitive in the center and
// sets its size to the given width and height.
// Returns a new primitive which puts the provided primitive in the center and
// sets its size to the given width and height.
func newModal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
}

func (u *UI) askExecuteBlockOrCmd(block *graph.Block, cmd, title string) {
	dlg := tview.NewModal().
		SetText(title).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				u.app.Stop()

				if block == nil {
					fmt.Printf(">> %s\n", cmd)
					if err := tools.BashExec(cmd); err != nil {
						fmt.Printf("error = %v", err)
						os.Exit(1)
					}

					os.Exit(0)
				}

				req := executor.DefaultRequest()
				req.Content = []byte(block.Content)
				req.ContentType = block.ContentType

				executor, err := u.register.Get(req.Content, req.ContentType)
				if err != nil {
					tools.LogStderrExit(err, "Failed to locate a matching executor\n")
				}

				fmt.Printf(">> %s\n", string(req.Content))

				resp := executor.Execute(context.Background(), req)
				if resp.Err != nil {
					tools.LogStderrExit(resp.Err, "Failed to execute command: %v\n", err)
				}

				os.Exit(resp.ExitStatus)
			} else {
				u.pages.RemovePage("execute")
			}
		})

	m := newModal(dlg, 40, 10)
	u.pages.AddPage("execute", m, true, true)
}
