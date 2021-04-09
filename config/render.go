package config

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
)

type tableRender struct {
	config *Config
}

func (tr *tableRender) Render() {
	data := make([]table.Row, 0)
	for _, entry := range tr.config.Entries {
		defaultStr := "[ ]"
		if entry.Name == tr.config.DefaultEntry {
			defaultStr = "[*]"
		}
		row := table.Row{entry.Name, entry.ServiceUrl, entry.User, defaultStr}
		data = append(data, row)
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Service URL", "User", "Default?"})
	t.AppendRows(data)
	t.AppendSeparator()
	t.Render()
}
