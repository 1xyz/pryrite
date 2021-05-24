package tui

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/snippet"
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
	nodeDetails, err := NewNodeDetails()
	if err != nil {
		return nil, fmt.Errorf("newNodeDetails: err = %v", err)
	}

	pbTree, err := NewPlaybookTree(rc.Root, nodeDetails)
	if err != nil {
		return nil, fmt.Errorf("newPlaybookTree: err = %v", err)
	}

	flex := tview.NewFlex().
		AddItem(pbTree.RootView, 0, 1, true).
		AddItem(nodeDetails.TextArea, 0, 4, false)
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

func createTextView() (*tview.TextView, error) {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true)
	textView.SetBorder(true)

	return textView, nil
}
