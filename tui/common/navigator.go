package common

import (
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Navigable interface {
	tview.Primitive

	// Required for automatic focus handling
	SetTitleColor(color tcell.Color) *tview.Box
	SetBorderColor(color tcell.Color) *tview.Box

	// NavHelp returns Navigable help info
	// The return type is a 2D string slice where column zero represents the
	// key combination and column 1 represents the help context.
	NavHelp() [][]string
}

type Navigator struct {
	RootUI     *tview.Application
	Entries    []Navigable
	BlurColor  tcell.Color
	FocusColor tcell.Color
	Quit       func()
}

func NewNavigator(app *tview.Application, entries []Navigable, quit func()) *Navigator {
	n := &Navigator{
		RootUI:     app,
		Entries:    entries,
		BlurColor:  tcell.ColorDarkCyan,
		FocusColor: tcell.ColorYellow,
		Quit:       quit,
	}
	n.SetCurrentFocusedIndex(0)
	return n
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
	n.SetCurrentFocusedIndex(0)
}

func (n *Navigator) Next() {
	index := n.GetCurrentFocusedIndex()
	next := 0
	if index == -1 || index == len(n.Entries)-1 {
		next = 0
	} else {
		next = index + 1
	}
	n.SetCurrentFocusedIndex(next)
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
	n.SetCurrentFocusedIndex(next)
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
	if focused, found := n.CurrentFocusedItem(); found {
		focused.SetTitleColor(n.BlurColor)
		focused.SetBorderColor(n.BlurColor)
	}
	newEntry := n.Entries[index]
	n.RootUI.SetFocus(newEntry)
	newEntry.SetTitleColor(n.FocusColor)
	newEntry.SetBorderColor(n.FocusColor)
}

func (n *Navigator) GlobalKeyBindings(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		tools.Log.Info().Msg("Global: ESC request to go home")
		n.Home()
	case tcell.KeyCtrlQ:
		tools.Log.Info().Msg("Global: Ctrl+Q request to quit")
		n.Quit()
	}
	return event
}
