package common

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Navigable interface {
	tview.Primitive

	// NavHelp returns Navigable help info
	// The return type is a 2D string slice where column zero represents the
	// key combination and column 1 represents the help context.
	NavHelp() [][]string
}

type Navigator struct {
	RootUI  *tview.Application
	Entries []Navigable
}

func (n *Navigator) Navigate(key tcell.Key) {
	switch key {
	case tcell.KeyBacktab:
		n.Prev()
	case tcell.KeyTab:
		n.Next()
	}
}

func (n *Navigator) Home() {
	n.RootUI.SetFocus((n.Entries[0]))
}

func (n *Navigator) Next() {
	index := n.GetCurrentFocusedIndex()
	next := 0
	if index == -1 || index == len(n.Entries)-1 {
		next = 0
	} else {
		next = index + 1
	}
	n.RootUI.SetFocus(n.Entries[next])
}

func (n *Navigator) Prev() {
	index := n.GetCurrentFocusedIndex()
	next := 0
	if index == 0 {
		next = len(n.Entries) - 1
	} else if index == -1 {
		next = 0
	} else {
		next = index - 1
	}
	n.RootUI.SetFocus(n.Entries[next])
}

func (n *Navigator) GetCurrentFocusedIndex() int {
	for i, e := range n.Entries {
		if e.HasFocus() {
			return i
		}
	}
	return -1
}

func (n *Navigator) CurrentFocusedItem() (Navigable, bool) {
	index := n.GetCurrentFocusedIndex()
	if index == -1 {
		return nil, false
	}
	return n.Entries[index], true
}

func (n *Navigator) SetCurrentFocusedIndex(index int) {
	n.RootUI.SetFocus(n.Entries[index])
}
