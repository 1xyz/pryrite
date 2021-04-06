package events

import (
	"fmt"
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
		{"SessionID", er.E.Metadata.SessionID},
		{"CreatedOn", tools.FmtTime(er.E.CreatedAt)},
		{"Title", tools.TrimLength(er.E.Metadata.Title, maxColumnLen)},
	})
	t.AppendSeparator()
	t.Render()

	body, err := er.E.DecodeDetails()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(body.Body())
	}
}
