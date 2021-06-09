package explorer

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"os"
	"strings"
)

type nodeTreeView struct {
	*tview.TreeView
	rootUI *UI
	title  string
}

func (view *nodeTreeView) addCodeBlock(root *tview.TreeNode, b *graph.Block) {
	title := strings.TrimSpace(b.Content)
	title = tools.TrimLength(title, 30)
	tn := tview.NewTreeNode(title).
		SetReference(b).
		SetSelectable(true)
	tn.SetColor(tcell.ColorGreen)
	root.AddChild(tn)
}

func (view *nodeTreeView) addNode(root *tview.TreeNode, n *graph.Node) *tview.TreeNode {
	title := strings.TrimSpace(n.Title)
	title = tools.TrimLength(title, 30)
	tn := tview.NewTreeNode(title).
		SetReference(n).
		SetSelectable(true)
	if n.HasChildren() {
		tn.SetColor(tcell.ColorYellow)
	}
	root.AddChild(tn)
	return tn
}

func (view *nodeTreeView) buildTree(nodes []*graph.Node) error {
	root := tview.NewTreeNode(view.title).
		SetColor(tcell.ColorDarkCyan).
		SetSelectable(false)
	view.TreeView.SetRoot(root).SetCurrentNode(root)
	view.TreeView.SetTitleAlign(tview.AlignLeft)
	for _, n := range nodes {
		view.addNode(root, n)
	}
	view.setupSelection()
	view.setupChangeNavigation()
	view.setupInputCapture()
	view.SetDoneFunc(view.rootUI.Navigate)
	return nil
}

func (view *nodeTreeView) setupChangeNavigation() {
	view.SetChangedFunc(func(tn *tview.TreeNode) {
		ref := tn.GetReference()
		if ref == nil {
			view.rootUI.SetNavHelp([][]string{
				{"up/down", "navigate through this tree"},
			})
			return
		}

		navhelp := view.getNavHelp(tn)
		view.rootUI.SetNavHelp(navhelp)

		b, isBlock := ref.(*graph.Block)
		n, isNode := ref.(*graph.Node)

		if isBlock {
			view.rootUI.SetContentBlock(b)
			view.rootUI.SetInfoBlock(b)
		} else if isNode {
			view.rootUI.SetContentNode(n)
			view.rootUI.SetInfoNode(n)
		}
	})
}

func (view *nodeTreeView) getNavHelp(tn *tview.TreeNode) [][]string {
	ref := tn.GetReference()
	if ref == nil {
		view.rootUI.SetNavHelp([][]string{
			{"up/down", "navigate through this tree"},
		})
		return [][]string{}
	}

	_, isBlock := ref.(*graph.Block)
	if len(tn.GetChildren()) > 0 {
		return [][]string{
			{"ctrl + R", "run this node"},
			{"up/down", "navigate through this tree"},
		}
	} else if isBlock {
		return [][]string{
			{"Enter", "Execute code snippet in bash"},
			{"Ctrl + Space", "print code snippet to stdout"},
			{"up/down", "navigate through this tree"},
		}
	} else {
		return [][]string{
			{"up/down", "navigate through this tree"},
		}
	}
}

func (view *nodeTreeView) setupSelection() {
	// If a node was selected, open it and show it in the details window
	view.SetSelectedFunc(func(tn *tview.TreeNode) {
		ref := tn.GetReference()
		if ref == nil {
			return
		}

		if len(tn.GetChildren()) == 0 {
			if n, ok := ref.(*graph.Node); ok {
				children := view.rootUI.GetChildren(n.ID)
				if children != nil && len(children) > 0 {
					for _, c := range children {
						view.addNode(tn, c)
					}
				}
				if n.HasBlocks() {
					for _, b := range n.Blocks {
						if !b.IsCode() {
							continue
						}
						view.addCodeBlock(tn, b)
					}
				}
			}
		} else {
			tn.SetExpanded(!tn.IsExpanded())
		}
	})
}

func (view *nodeTreeView) setupInputCapture() {
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		tn := view.GetCurrentNode()
		if tn == nil {
			return event
		}
		ref := tn.GetReference()
		if ref == nil {
			return event
		}

		switch event.Key() {
		case tcell.KeyEnter:
			tools.Log.Info().Msgf("nodeTreeView: Ctrl+Enter request to exec block")
			if b, ok := ref.(*graph.Block); ok {
				view.rootUI.Stop()
				fmt.Printf(">> %s\n", b.Content)
				if err := tools.BashExec(b.Content); err != nil {
					fmt.Printf("error = %v", err)
					os.Exit(1)
				} else {
					os.Exit(0)
				}
			}

		case tcell.KeyCtrlSpace:
			tools.Log.Info().Msgf("nodeTreeView: Ctrl+Space request to print block to stdout")
			if b, ok := ref.(*graph.Block); ok {
				view.rootUI.Stop()
				fmt.Printf("%s", b.Content)
				os.Exit(0)
			}

		case tcell.KeyCtrlR:
			tools.Log.Info().Msgf("nodeTreeView: Ctrl+R request to run a node")
			if n, ok := ref.(*graph.Node); ok {
				view.rootUI.Stop()
				cmd := fmt.Sprintf("%s run %s", os.Args[0], n.ID)
				fmt.Printf(">> %s", cmd)
				if err := tools.BashExec(cmd); err != nil {
					fmt.Printf("error = %v\n", err)
					os.Exit(1)
				} else {
					os.Exit(0)
				}
			}
		}
		return event
	})
}

func (view *nodeTreeView) NavHelp() [][]string {
	return view.getNavHelp(view.GetCurrentNode())
}

func newNodeTreeView(rootUI *UI, nodes []*graph.Node, title, mainTitle string) (*nodeTreeView, error) {
	view := &nodeTreeView{
		TreeView: tview.NewTreeView(),
		rootUI:   rootUI,
		title:    title,
	}
	if err := view.buildTree(nodes); err != nil {
		return nil, err
	}
	view.SetBorder(true)
	view.SetBorderColor(tcell.ColorDarkCyan)
	if len(mainTitle) > 0 {
		view.SetTitle(mainTitle)
	}
	return view, nil
}
