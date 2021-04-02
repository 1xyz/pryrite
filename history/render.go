package history

import (
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
)

type historyRender struct {
	H     []Item
	Limit int
}

func (hr *historyRender) Render() {
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Index", "Date", "Command"})
	rows := make([]table.Row, 0)
	n := len(hr.H)
	for i, h := range hr.H {
		if hr.Limit == 0 || n-i-1 < hr.Limit {
			rows = append(rows, table.Row{i, tools.FmtTime(h.CreatedAt), h.Command})
		}
	}
	t.AppendRows(rows)
	t.Render()
}
