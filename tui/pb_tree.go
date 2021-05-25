package tui

import (
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// PlayBookTree encapsulates the Tree view of a playbook (that is shown on the left pane)
type PlayBookTree struct {
	// Reference to the TreeView UI component
	View *tview.TreeView

	// Reference ot the root UI component
	rootUI *Tui

	// Refers to the underlying playbook
	playbook *graph.NodeView
}

func NewPlaybookTree(root *Tui, playbook *graph.NodeView) (*PlayBookTree, error) {
	tn := tview.NewTreeNode(playbook.Node.Title).
		SetColor(tcell.ColorRed)
	tree := tview.NewTreeView().
		SetRoot(tn).
		SetCurrentNode(tn)
	tree.SetBorder(true).
		SetTitle("Playbook").
		SetTitleAlign(tview.AlignLeft)
	// A helper function which adds the child nodes to the given target node.
	add := func(target *tview.TreeNode, view *graph.NodeView) error {
		for _, child := range view.Children {
			hasChildren := len(child.Children) > 0
			tNode := tview.NewTreeNode(child.Node.Title).
				SetReference(child).
				SetSelectable(true)
			if hasChildren {
				tNode.SetColor(tcell.ColorGreen)
			}
			target.AddChild(tNode)
		}
		return nil
	}

	if err := add(tn, playbook); err != nil {
		return nil, err
	}

	// If a node was selected, open it and show it in the details window
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		ref := node.GetReference()
		if ref == nil {
			return
		}
		view := ref.(*graph.NodeView)
		//fmt.Printf("view = %s", view.Node.Title)
		children := node.GetChildren()
		if len(children) == 0 {
			if err := add(node, view); err != nil {
				tools.Log.Err(err).Msgf("SetSelectedFunc: add: err = %v", err)
			}
		} else {
			// Collapse if visible, expand if collapsed.
			node.SetExpanded(!node.IsExpanded())
		}
		root.SetCurrentNodeView(view)
		res, _ := root.run.ExecIndex.Get(view.Node.ID)
		root.SetCurrentNodeExecutionResult(res)
	})

	tree.SetDoneFunc(func(key tcell.Key) { root.Navigate(key) })

	return &PlayBookTree{
		rootUI:   root,
		View:     tree,
		playbook: playbook,
	}, nil
}
