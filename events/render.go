package events

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
)

type eventsRender struct {
	E []Event
}

const maxColumnLen = 50

func trimLength(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func (er *eventsRender) Render() {
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	for _, e := range er.E {
		t.AppendRow(table.Row{
			e.ID,
			e.CreatedAt.Format("Jan _2 3:04PM"),
			trimLength(e.Metadata.Title, maxColumnLen)})
	}
	t.AppendHeader(table.Row{"Id", "Date", "Summary"})
	t.AppendSeparator()
	t.Render()
}

type eventRender struct {
	E *Event
}

func (er *eventRender) Render() {
	data := make([]table.Row, 0)

	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	t.AppendRow(table.Row{"Id", er.E.ID})
	t.AppendRow(table.Row{"Kind", er.E.Kind})
	t.AppendSeparator()
	t.AppendRows([]table.Row{
		{"Title", er.E.Metadata.Title},
		{"Url", er.E.Metadata.URL},
		{"SessionID", er.E.Metadata.SessionID},
		{"CreatedOn", er.E.CreatedAt},
	})
	t.AppendRows(data)
	t.Render()
}
