package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type navigable interface {
	tview.Primitive

	// NavHelp returns navigable help info
	NavHelp() string
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
	index := n.getCurrentFocus()
	next := 0
	if index == -1 || index == len(n.Entries)-1 {
		next = 0
	} else {
		next = index + 1
	}
	n.rootUI.SetFocus(n.Entries[next])
}

func (n *navigator) Prev() {
	index := n.getCurrentFocus()
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

func (n *navigator) getCurrentFocus() int {
	for i, e := range n.Entries {
		if e.HasFocus() {
			return i
		}
	}
	return -1
}

func (n *navigator) CurrentFocusedItem() (navigable, bool) {
	index := n.getCurrentFocus()
	if index == -1 {
		return nil, false
	}
	return n.Entries[index], true
}
