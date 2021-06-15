package explorer

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type navView struct {
	*tview.Table
	rootUI *UI
}

func newNavView(rootUI *UI) *navView {
	n := &navView{
		Table: tview.NewTable().
			SetSelectable(false, false).
			Select(0, 0).
			SetFixed(1, 1),
		rootUI: rootUI,
	}
	return n
}

func (n *navView) SetHelp(entries [][]string) {
	n.Clear()

	for index, e := range entries {
		n.SetCell(index, 0, &tview.TableCell{
			Text:            e[0],
			NotSelectable:   true,
			Align:           tview.AlignRight,
			Color:           tcell.ColorDarkCyan,
			BackgroundColor: tcell.ColorDefault,
			Attributes:      tcell.AttrBold,
			Transparent:     true,
		})
		n.SetCell(index, 1, &tview.TableCell{
			Text:            e[1],
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           tcell.ColorLightCyan,
			BackgroundColor: tcell.ColorDefault,
			Attributes:      tcell.AttrNone,
			Transparent:     true,
		})
	}
}
