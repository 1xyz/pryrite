package tui

import (
	"github.com/aardlabs/terminal-poc/tui/common"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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

func (n *navView) Refresh(nav common.Navigable) {
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
