package config

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
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
		row := table.Row{entry.Name, entry.ServiceUrl, entry.User, defaultStr}
		data = append(data, row)
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Service URL", "User", "Active?"})
	t.AppendRows(data)
	t.AppendSeparator()
	t.Render()
}
