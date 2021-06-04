package tui

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/run"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strconv"
)

type blockExecutionsView struct {
	*tview.Table
	rootUI *Tui
}

func newBlockExecutionsView(root *Tui) *blockExecutionsView {
	view := &blockExecutionsView{
		Table: tview.NewTable().
			SetSelectable(true, false).
			Select(0, 0).
			SetFixed(1, 1),
		rootUI: root,
	}

	view.SetTitle("completed executions").
		SetTitleAlign(tview.AlignLeft)
	view.SetBorder(true)
	view.SetDoneFunc(func(key tcell.Key) {
		root.Navigate(key)
	})
	return view
}

func (b *blockExecutionsView) Refresh(results run.BlockExecutionResults) {
	table := b.Clear()
	if results == nil {
		return
	}

	headers := []string{
		"Node",
		"Code-Snippet",
		"Exit-Code",
		"Error",
		"Time",
		"User",
	}

	for i, header := range headers {
		table.SetCell(0, i, &tview.TableCell{
			Text:            header,
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           tcell.ColorWhite,
			BackgroundColor: tcell.ColorDefault,
			Attributes:      tcell.AttrBold,
		})
	}

	j := 1
	for i := len(results) - 1; i >= 0; i-- {
		r := results[i]
		showColor := tcell.ColorGreen
		if r.Err != nil || r.ExitStatus > 0 {
			showColor = tcell.ColorRed
		}

		table.SetCell(j, 0, tview.NewTableCell(r.NodeID).
			SetTextColor(showColor).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(j, 1, tview.NewTableCell(r.Content).
			SetTextColor(showColor).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(j, 2, tview.NewTableCell(strconv.Itoa(r.ExitStatus)).
			SetTextColor(showColor).
			SetMaxWidth(1).
			SetExpansion(1))

		errStr := "N/A"
		expansion := 1
		if r.Err != nil {
			errStr = r.Err.Error()
			expansion = 2
		} else if r.Stderr != nil && len(r.Stderr) > 0 {
			errStr = string(r.Stderr)
			expansion = 2
		}
		table.SetCell(j, 3, tview.NewTableCell(errStr).
			SetTextColor(showColor).
			SetMaxWidth(1).
			SetExpansion(expansion))

		executedAt := "N/A"
		if r.ExecutedAt != nil {
			executedAt = r.ExecutedAt.Format("2006/01/02 15:04:05")
		}
		table.SetCell(j, 4, tview.NewTableCell(executedAt).
			SetTextColor(showColor).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(j, 5, tview.NewTableCell(r.ExecutedBy).
			SetTextColor(showColor).
			SetMaxWidth(1).
			SetExpansion(1))
		j++
	}
}

func (b *blockExecutionsView) NavHelp() string {
	navigate := " tab: next pane, shift+tab: previous pane"
	navHelp := fmt.Sprintf("navigate \t| %s\n", navigate)
	return navHelp
}
