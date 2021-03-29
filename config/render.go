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
	for name, entry := range tr.config.Entries {
		defaultStr := "[ ]"
		if name == tr.config.DefaultEntry {
			defaultStr = "[*]"
		}
		row := table.Row{name, entry.ServiceUrl, defaultStr}
		data = append(data, row)
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Service URL", "Default?"})
	t.AppendRows(data)
	t.AppendSeparator()
	t.Render()
}
