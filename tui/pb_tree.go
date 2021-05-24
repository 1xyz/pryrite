package tui

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/charmbracelet/glamour"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// PlayBookTree represents the Tree view of a playbook (that is shown on the left pane)
type PlayBookTree struct {
	// Reference to the TreeView UI component
	RootView *tview.TreeView

	// Refers the details component for a selected node on the PlayBookTree
	detail *NodeDetails

	// Refers to the underlying playbook
	playbook *graph.NodeView
}

func NewPlaybookTree(playbook *graph.NodeView, detail *NodeDetails) (*PlayBookTree, error) {
	tn := tview.NewTreeNode(playbook.Node.Title).
		SetColor(tcell.ColorRed)
	tree := tview.NewTreeView().
		SetRoot(tn).
		SetCurrentNode(tn)
	tree.SetBorder(true)

	// A helper function which adds the files and directories of the given path
	// to the given target node.
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

	// If a directory was selected, open it.
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

		detail.Clear()
		var out string
		r, err := glamour.NewTermRenderer(glamour.WithStylePath("notty"))
		if err != nil {
			tools.Log.Err(err).Msgf("SetSelectedFunc: NewTermRenderer err =  %v", err)
			return
		}

		out, err = r.Render(view.ContentMarkdown)
		if err != nil {
			out = fmt.Sprintf("SetSelectedFunc: render markdown: err = %v", err)
		}
		if err := detail.Write([]byte(out)); err != nil {
			tools.Log.Err(err).Msgf("SetSelectedFunc: detail.Write: err = %v", err)
		}
	})

	return &PlayBookTree{
		RootView: tree,
		detail:   detail,
		playbook: playbook,
	}, nil
}
