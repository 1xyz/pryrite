package tui

import (
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type info struct {
	*tview.Table
	rootUI *Tui
	gCtx   *snippet.Context
}

func newInfo(rootUI *Tui, gCtx *snippet.Context) *info {
	i := &info{
		Table: tview.NewTable().
			SetSelectable(false, false).
			Select(0, 0).
			SetFixed(1, 1),
		rootUI: rootUI,
		gCtx:   gCtx,
	}

	i.display()
	return i
}

func (i *info) display() {
	headers := [][]string{
		{"Version", i.gCtx.Metadata.Agent},
		{"Server", i.gCtx.ConfigEntry.ServiceUrl},
		{"User", i.gCtx.ConfigEntry.Email},
	}

	for index, entries := range headers {
		i.SetCell(index, 0, &tview.TableCell{
			Text:            entries[0] + ":",
			NotSelectable:   true,
			Align:           tview.AlignRight,
			Color:           tcell.ColorYellow,
			BackgroundColor: tcell.ColorDefault,
			Attributes:      tcell.AttrBold,
		})
		i.SetCell(index, 1, &tview.TableCell{
			Text:            entries[1],
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           tcell.ColorWhite,
			BackgroundColor: tcell.ColorDefault,
			Attributes:      tcell.AttrNone,
		})
	}
}
