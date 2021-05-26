package tui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// detailView is a resuable abstraction of TextView
type detailView struct {
	// Underlying view
	*tview.TextView

	// Reference ot the root UI component
	rootUI *Tui
}

func (e *detailView) NavHelp() string {
	navigate := " tab: next pane, shift+tab: previous pane"
	navHelp := fmt.Sprintf("navigate \t| %s\n", navigate)
	return navHelp
}

func newDetailView(title string, showBorder bool, rootUI *Tui) *detailView {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true)
	textView.SetBorder(showBorder).
		SetTitle(title).
		SetTitleAlign(tview.AlignLeft)
	textView.SetDoneFunc(func(key tcell.Key) { rootUI.Navigate(key) })

	return &detailView{
		rootUI:   rootUI,
		TextView: textView,
	}
}
