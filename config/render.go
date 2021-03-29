package config

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
)

type tableRender struct {
	config *Config
}

func (tr *tableRender) Render() {
	data := make([][]string, 0)
	for name, entry := range tr.config.Entries {
		defaultStr := "[ ]"
		if name == tr.config.DefaultEntry {
			defaultStr = "[*]"
		}
		data = append(data, []string{name, entry.ServiceUrl, defaultStr})
	}
	if len(data) == 0 {
		fmt.Println("No entries found!")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Url", "default ?"})
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()
}
