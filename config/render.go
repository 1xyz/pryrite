package config

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

type TableRender struct {
	Config *Config
}

func (tr *TableRender) Render() {
	data := make([]table.Row, 0)
	for _, entry := range tr.Config.Entries {
		defaultStr := "[ ]"
		if entry.Name == tr.Config.DefaultEntry {
			defaultStr = "[*]"
		}
		row := table.Row{entry.Name, entry.ServiceUrl, entry.Email, defaultStr}
		data = append(data, row)
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Service URL", "Email", "Active?"})
	t.AppendRows(data)
	t.AppendSeparator()
	t.Render()
}
