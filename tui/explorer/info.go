package explorer

import (
	"fmt"

	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type infoView struct {
	*tview.Table
	rootUI *UI
}

func (i *infoView) SetBlock(b *graph.Block) {
	i.Clear()
	entries := [][]string{
		{"Block ID", b.ID},
		{"Content Type", b.ContentType.String()},
		{"MD5", b.MD5},
		{"Is Code", fmt.Sprintln(b.IsCode())},
		{"Created At", tools.FormatTime(b.CreatedAt)},
	}
	i.set(entries)
}

func (i *infoView) SetNode(n *graph.Node) {
	i.Clear()
	entries := [][]string{
		{"Node ID", n.ID},
		{"Created At", tools.FormatTime(n.CreatedAt)},
		{"Last Executed At", tools.FormatTime(n.LastExecutedAt)},
		{"Last Executed By", n.LastExecutedBy},
	}
	i.set(entries)
}

func (i *infoView) set(entries [][]string) {
	for index, e := range entries {
		i.SetCell(index, 0, &tview.TableCell{
			Text:            e[0],
			NotSelectable:   true,
			Align:           tview.AlignRight,
			Color:           tcell.ColorDarkCyan,
			BackgroundColor: tcell.ColorDefault,
			Attributes:      tcell.AttrBold,
		})
		i.SetCell(index, 1, &tview.TableCell{
			Text:            e[1],
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           tcell.ColorLightCyan,
			BackgroundColor: tcell.ColorDefault,
			Attributes:      tcell.AttrNone,
		})
	}
}

func newInfoView(rootUI *UI) *infoView {
	n := &infoView{
		Table: tview.NewTable().
			SetSelectable(false, false).
			Select(0, 0).
			SetFixed(1, 1),
		rootUI: rootUI,
	}
	n.SetBorderPadding(2, 2, 2, 2)
	return n
}
