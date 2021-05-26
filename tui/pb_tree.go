package tui

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// PlayBookTree encapsulates the Tree view of a playbook (that is shown on the left pane)
type PlayBookTree struct {
	// Reference to the TreeView UI component
	*tview.TreeView

	// Reference ot the root UI component
	rootUI *Tui

	// Refers to the underlying playbook
	playbook *graph.NodeView

	// treeNodes is a map from nodeID to a treeNode object
	treeNodes map[string]*tview.TreeNode
}

func NewPlaybookTree(root *Tui, playbook *graph.NodeView) (*PlayBookTree, error) {
	treeNodes := make(map[string]*tview.TreeNode)
	tn := tview.NewTreeNode(playbook.Node.Title).
		SetColor(tcell.ColorYellow).
		SetReference(playbook).
		SetSelectable(true)
	treeNodes[playbook.Node.ID] = tn
	tree := tview.NewTreeView().
		SetRoot(tn).
		SetCurrentNode(tn)
	tree.SetBorder(true).
		SetTitle("playbook").
		SetTitleAlign(tview.AlignLeft)
	// A helper function which adds the child nodes to the given target node.
	add := func(target *tview.TreeNode, view *graph.NodeView) error {
		for _, child := range view.Children {
			hasChildren := len(child.Children) > 0
			tNode := tview.NewTreeNode(child.Node.Title).
				SetReference(child).
				SetSelectable(true)
			treeNodes[child.Node.ID] = tNode
			if hasChildren {
				tNode.SetColor(tcell.ColorYellow)
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

		if err := root.UpdateCurrentNodeID(view.Node.ID); err != nil {
			root.StatusErrorf("UpdateCurrentNodeID: id=%s err = %v", view.Node.ID, err)
		}
	})

	tree.SetDoneFunc(func(key tcell.Key) { root.Navigate(key) })

	return &PlayBookTree{
		treeNodes: treeNodes,
		rootUI:    root,
		TreeView:  tree,
		playbook:  playbook,
	}, nil
}

func (p *PlayBookTree) NavHelp() string {
	help := " enter: select snippet"
	navigate := " tab: selected snippet pane, shift+tab: previous pane"
	navHelp := fmt.Sprintf(" commands \t| %s\n navigate \t| %s\n", help, navigate)
	return navHelp
}

func (p *PlayBookTree) RefreshNode(nodeID string) error {
	tNode, found := p.treeNodes[nodeID]
	if !found {
		return fmt.Errorf("refreshNode: node with %s not found", nodeID)
	}

	view, err := p.rootUI.run.ViewIndex.Get(nodeID)
	if err != nil {
		return err
	}

	tNode.SetReference(view).
		SetText(view.Node.Title)
	return nil
}
