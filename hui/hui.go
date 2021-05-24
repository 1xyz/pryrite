package hui

import (
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/charmbracelet/glamour"
	tcell2 "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func LaunchUI(gCtx *snippet.Context, name string) error {
	ui, err := setupRunUI(gCtx, name)
	if err != nil {
		return err
	}
	if err := ui.Run(); err != nil {
		return err
	}
	return nil
}

type RunUI struct {
	app   *tview.Application // primary UI application
	pages *tview.Pages       // different pages in this UI
	flex  *tview.Flex        // Flex layout for the run page
	rc    *RunContext
}

func setupRunUI(gCtx *snippet.Context, name string) (*RunUI, error) {
	rc, err := BuildRunContext(gCtx, name)
	if err != nil {
		return nil, err
	}

	app := tview.NewApplication()
	textArea, err := createTextView()
	if err != nil {
		return nil, err
	}
	tree, err := createTreeView(rc.Root, textArea)
	if err != nil {
		return nil, err
	}

	flex := tview.NewFlex().
		AddItem(tree, 0, 1, true).
		AddItem(textArea, 0, 4, false)
	pages := tview.NewPages().AddPage("main", flex, true, true)
	app.SetRoot(pages, true)

	rui := &RunUI{
		app:   app,
		pages: pages,
		flex:  flex,
		rc:    rc,
	}
	return rui, nil
}

func (ui *RunUI) Run() error {
	return ui.app.Run()
}

func createTreeView(root *graph.NodeView, textArea *tview.TextView) (*tview.TreeView, error) {
	troot := tview.NewTreeNode(root.Node.Title).
		SetColor(tcell2.ColorRed)
	tree := tview.NewTreeView().
		SetRoot(troot).
		SetCurrentNode(troot)
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
				tNode.SetColor(tcell2.ColorGreen)
			}
			target.AddChild(tNode)
		}
		return nil
	}

	// Register the current directory to the troot node.
	if err := add(troot, root); err != nil {
		return nil, err
	}

	// If a directory was selected, open it.
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			//fmt.Printf("reference is nil")
			return // Selecting the root node does nothing.
		}
		view := reference.(*graph.NodeView)
		//fmt.Printf("view = %s", view.Node.Title)
		children := node.GetChildren()
		if len(children) == 0 {
			add(node, view)
		} else {
			// Collapse if visible, expand if collapsed.
			node.SetExpanded(!node.IsExpanded())
		}
		textArea.Clear()
		var out string
		r, err := glamour.NewTermRenderer(glamour.WithStylePath("notty"))
		if err != nil {
			tools.Log.Err(err).Msgf("newTermRenderfailed %v", err)
			return
		}

		out, err = r.Render(view.ContentMarkdown)
		textArea.Write([]byte(out))
	})

	return tree, nil
}

func createTextView() (*tview.TextView, error) {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true)
	textView.SetBorder(true)

	return textView, nil
}
