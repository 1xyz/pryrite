package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type navigable interface {
	tview.Primitive

	// NavHelp returns navigable help info
	// The return type is a 2D string slice where column zero represents the
	// key combination and column 1 represents the help context.
	NavHelp() [][]string
}

type navigator struct {
	rootUI  *tview.Application
	Entries []navigable
}

func (n *navigator) Navigate(key tcell.Key) {
	switch key {
	case tcell.KeyBacktab:
		n.Prev()
	case tcell.KeyTab:
		n.Next()
	}
}

func (n *navigator) Next() {
	index := n.GetCurrentFocusedIndex()
	next := 0
	if index == -1 || index == len(n.Entries)-1 {
		next = 0
	} else {
		next = index + 1
	}
	n.rootUI.SetFocus(n.Entries[next])
}

func (n *navigator) Prev() {
	index := n.GetCurrentFocusedIndex()
	next := 0
	if index == 0 {
		next = len(n.Entries) - 1
	} else if index == -1 {
		next = 0
	} else {
		next = index - 1
	}
	n.rootUI.SetFocus(n.Entries[next])
}

func (n *navigator) GetCurrentFocusedIndex() int {
	for i, e := range n.Entries {
		if e.HasFocus() {
			return i
		}
	}
	return -1
}

func (n *navigator) CurrentFocusedItem() (navigable, bool) {
	index := n.GetCurrentFocusedIndex()
	if index == -1 {
		return nil, false
	}
	return n.Entries[index], true
}

func (n *navigator) SetCurrentFocusedIndex(index int) {
	n.rootUI.SetFocus(n.Entries[index])
}

// ui component
type navView struct {
	*tview.Table
	rootUI *Tui
}

func newNavView(rootUI *Tui) *navView {
	n := &navView{
		Table: tview.NewTable().
			SetSelectable(false, false).
			Select(0, 0).
			SetFixed(1, 1),
		rootUI: rootUI,
	}

	return n
}

func (n *navView) Refresh(nav navigable) {
	n.Clear()
	entries := nav.NavHelp()
	if entries == nil || len(entries) == 0 {
		return
	}

	for index, e := range entries {
		n.SetCell(index, 0, &tview.TableCell{
			Text:            e[0],
			NotSelectable:   true,
			Align:           tview.AlignRight,
			Color:           tcell.ColorDarkCyan,
			BackgroundColor: tcell.ColorDefault,
			Attributes:      tcell.AttrBold,
		})
		n.SetCell(index, 1, &tview.TableCell{
			Text:            e[1],
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           tcell.ColorLightCyan,
			BackgroundColor: tcell.ColorDefault,
			Attributes:      tcell.AttrNone,
		})
	}
}
