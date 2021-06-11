package tui

import (
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

func (e *detailView) NavHelp() [][]string {
	return [][]string{
		{"Tab", "Navigate to the next pane"},
		{"Shift + Tab", "Navigate to the previous pane"},
	}
}

func (e *detailView) Focus(delegate func(p tview.Primitive)) {
	e.SetTitleColor(e.rootUI.focusColor)
	e.SetBorderColor(e.rootUI.focusColor)
	e.TextView.Focus(delegate)
}

func (e *detailView) Blur() {
	e.SetTitleColor(tcell.ColorDefault)
	e.SetBorderColor(tcell.ColorDefault)
	e.TextView.Blur()
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
