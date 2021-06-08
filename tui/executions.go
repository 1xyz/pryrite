package tui

import (
	"strconv"

	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/run"
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

	view.SetTitle("Executions").
		SetTitleAlign(tview.AlignLeft)
	view.SetBorder(true)
	view.SetDoneFunc(root.Navigate)
	view.setKeybinding()
	return view
}

func (b *executionsView) Refresh(results *run.BlockExecutionResults) {
	table := b.Clear()
	if results == nil {
		return
	}

	headers := []string{
		"Block Id",
		"Request Id",
		"Code Snippet",
		"Exit Code",
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
	results.EachFromEnd(func(_ int, r *graph.BlockExecutionResult) bool {
		showColor := tcell.ColorGreen
		if r.Err != nil || r.ExitStatus > 0 {
			showColor = tcell.ColorRed
		}

		table.SetCell(j, 0, tview.NewTableCell(r.BlockID).
			SetTextColor(showColor).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(j, 1, tview.NewTableCell(r.RequestID).
			SetTextColor(showColor).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(j, 2, tview.NewTableCell(r.Content).
			SetTextColor(showColor).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(j, 3, tview.NewTableCell(strconv.Itoa(r.ExitStatus)).
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
		table.SetCell(j, 4, tview.NewTableCell(errStr).
			SetTextColor(showColor).
			SetMaxWidth(1).
			SetExpansion(expansion))

		executedAt := "N/A"
		if r.ExecutedAt != nil {
			executedAt = r.ExecutedAt.Format("2006/01/02 15:04:05")
		}
		table.SetCell(j, 5, tview.NewTableCell(executedAt).
			SetTextColor(showColor).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(j, 6, tview.NewTableCell(r.ExecutedBy).
			SetTextColor(showColor).
			SetMaxWidth(1).
			SetExpansion(1))
		j++

		return true
	})
}

func (b *executionsView) setKeybinding() {
	b.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			r, _ := b.GetSelection()
			requestID := b.GetCell(r, 1).Text
			b.rootUI.InspectBlockExecution(requestID)
		}
		return event
	})
}

func (b *executionsView) NavHelp() [][]string {
	return [][]string{
		{"Enter", "Inspect Execution Detail"},
		{"Ctrl + R", "Run selected node"},
		{"⇵ Down/Up", "Navigate through executions"},
		{"Tab", "Navigate to the next pane"},
		{"Shift + Tab", "Navigate to the previous pane"},
	}
}

func (b *executionsView) renderYaml(result *graph.BlockExecutionResult) ([]byte, error) {
	resultView := struct {
		*graph.BlockExecutionResult
		Stdout string `yaml:"stdout"`
		Stderr string `yaml:"stderr"`
	}{result,
		string(result.Stdout),
		string(result.Stderr)}
	return yaml.Marshal(&resultView)
}
