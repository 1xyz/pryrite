package tui

import (
	"strconv"

	"github.com/aardlabs/terminal-poc/graph/log"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v2"
)

type executionsView struct {
	*tview.Table
	rootUI *Tui
}

func newExecutionsView(root *Tui) *executionsView {
	view := &executionsView{
		Table: tview.NewTable().
			SetSelectable(true, false).
			Select(0, 0).
			SetFixed(1, 1),
		rootUI: root,
	}

	view.SetTitle("Execution Log").
		SetTitleAlign(tview.AlignLeft)
	view.SetBorder(true)
	view.SetDoneFunc(root.Navigate)
	view.setKeybinding()
	return view
}

func (b *executionsView) Refresh(results log.ResultLog) {
	table := b.Clear()
	if results == nil {
		return
	}

	headers := []string{
		"#",
		"Log Id",
		"Block Id",
		"Request Id",
		"Code Snippet",
		"State",
		"Exit Status",
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
	results.Each(func(_ int, r *log.ResultLogEntry) bool {
		showColor := tcell.ColorWhite
		switch r.State {
		case log.ExecStateQueued:
			showColor = tcell.ColorDarkCyan
		case log.ExecStateStarted:
			showColor = tcell.ColorYellow
		case log.ExecStateCompleted:
			showColor = tcell.ColorGreen
		case log.ExecStateFailed:
			showColor = tcell.ColorRed
		case log.ExecStateCanceled:
			showColor = tcell.ColorPurple
		}

		table.SetCell(j, 0, tview.NewTableCell(strconv.Itoa(j)).
			SetTextColor(showColor).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(j, 1, tview.NewTableCell(r.ID).
			SetTextColor(showColor).
			SetMaxWidth(4).
			SetExpansion(4))

		table.SetCell(j, 2, tview.NewTableCell(r.BlockID).
			SetTextColor(showColor).
			SetMaxWidth(4).
			SetExpansion(4))

		table.SetCell(j, 3, tview.NewTableCell(r.RequestID).
			SetTextColor(showColor).
			SetMaxWidth(4).
			SetExpansion(4))

		table.SetCell(j, 4, tview.NewTableCell(r.Content).
			SetTextColor(showColor).
			SetMaxWidth(4).
			SetExpansion(4))

		table.SetCell(j, 5, tview.NewTableCell(string(r.State)).
			SetTextColor(showColor).
			SetMaxWidth(3).
			SetExpansion(3))

		table.SetCell(j, 6, tview.NewTableCell(r.ExitStatus).
			SetTextColor(showColor).
			SetMaxWidth(4).
			SetExpansion(4))

		errStr := "N/A"
		expansion := 5
		if len(r.Err) > 0 {
			errStr = r.Err
			expansion = 15
		} else if len(r.Stderr) > 0 {
			errStr = string(r.Stderr)
			expansion = 15
		}
		table.SetCell(j, 7, tview.NewTableCell(errStr).
			SetTextColor(showColor).
			SetMaxWidth(5).
			SetExpansion(expansion))

		executedAt := "N/A"
		if r.ExecutedAt != nil {
			executedAt = r.ExecutedAt.Format("2006/01/02 15:04:05")
		}
		table.SetCell(j, 8, tview.NewTableCell(executedAt).
			SetTextColor(showColor).
			SetMaxWidth(4).
			SetExpansion(4))

		table.SetCell(j, 9, tview.NewTableCell(r.ExecutedBy).
			SetTextColor(showColor).
			SetMaxWidth(4).
			SetExpansion(4))
		table.Select(j, 0)
		j++
		return true
	})
	b.ScrollToEnd()
}

func (b *executionsView) setKeybinding() {
	b.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		event = b.rootUI.Nav.GlobalKeyBindings(event)
		switch event.Key() {
		case tcell.KeyEnter:
			r, _ := b.GetSelection()
			logID := b.GetCell(r, 1).Text
			b.rootUI.InspectBlockExecution(logID)

		case tcell.KeyCtrlK:
			r, _ := b.GetSelection()
			requestID := b.GetCell(r, 3).Text
			if err := b.rootUI.CancelBlockExecution(requestID); err != nil {
				b.rootUI.StatusErrorf("Cancel request failed (requesID: %s error: %v)",
					requestID, err)
			}
		}

		return event
	})
}

func (b *executionsView) NavHelp() [][]string {
	return [][]string{
		{"Enter", "Inspect Execution Detail"},
		{"Ctrl + K", "Kill current task"},
		{"Ctrl + R", "Run selected node"},
		{"⇵ Down/Up", "Navigate through executions"},
		{"Tab", "Navigate to the next pane"},
		{"Shift + Tab", "Navigate to the previous pane"},
	}
}

func (b *executionsView) renderYaml(result *log.ResultLogEntry) ([]byte, error) {
	return yaml.Marshal(&result)
}
