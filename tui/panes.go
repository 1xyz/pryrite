package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type DetailPane struct {
	// Underlying view
	*tview.TextView

	// Reference ot the root UI component
	rootUI *Tui
}

func NewDetailPane(title string, rootUI *Tui) *DetailPane {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true)
	textView.SetBorder(true).
		SetTitle(title).
		SetTitleAlign(tview.AlignLeft)
	textView.SetDoneFunc(func(key tcell.Key) { rootUI.Navigate(key) })

	return &DetailPane{
		rootUI:   rootUI,
		TextView: textView,
	}
}
