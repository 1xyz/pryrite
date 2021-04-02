package events

import (
	"encoding/json"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
)

type eventsRender struct {
	E []Event
}

const (
	maxColumnLen = 40
)

func (er *eventsRender) Render() {
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	for _, e := range er.E {
		t.AppendRow(table.Row{
			e.ID,
			tools.FmtTime(e.CreatedAt),
			tools.TrimLength(e.Metadata.Title, maxColumnLen),
			e.Kind})
	}
	t.AppendHeader(table.Row{"Id", "Date", "Summary", "Type"})
	t.AppendSeparator()
	t.Render()
}

type eventRender struct {
	E *Event
}

func (er *eventRender) Render() {
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	t.AppendRow(table.Row{"Id", er.E.ID})
	t.AppendRow(table.Row{"Kind", er.E.Kind})
	t.AppendSeparator()

	t.AppendRows([]table.Row{
		{"Title", er.E.Metadata.Title},
		{"SessionID", er.E.Metadata.SessionID},
		{"CreatedOn", tools.FmtTime(er.E.CreatedAt)},
	})
	t.AppendSeparator()

	switch er.E.Kind {
	case "Console":
		cm := ConsoleMetadata{}
		if err := json.Unmarshal(er.E.Details, &cm); err != nil {
			t.AppendRow(table.Row{"Error", err.Error()})
		} else {
			t.AppendRows([]table.Row{
				{"Content", cm.Raw},
			})
		}
	case "PageClose", "PageOpen":
		t.AppendRow(table.Row{"Url", er.E.Metadata.URL})
	}
	t.Render()
}
