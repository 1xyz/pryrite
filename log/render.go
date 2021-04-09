package log

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
)

type eventsRender struct {
	E []graph.Node
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
			tools.FmtTime(e.OccurredAt),
			getSummary(&e),
			e.Kind})
	}
	t.AppendHeader(table.Row{"Id", "Date", "Summary", "Type"})
	t.AppendSeparator()
	t.Render()
}

func getSummary(n *graph.Node) string {
	summary := n.Description
	if len(summary) == 0 {
		if d, err := n.DecodeDetails(); err != nil {
			tools.Log.Err(err).Msgf("DecodeDetails")
			summary = "N/A"
		} else {
			summary = d.Summary()
		}
	}
	return tools.TrimLength(summary, maxColumnLen)
}

type eventRender struct {
	E            *graph.Node
	renderDetail bool
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
		{"Date", tools.FmtTime(er.E.OccurredAt)},
		{"CreatedOn", tools.FmtTime(er.E.CreatedAt)},
		{"Agent", er.E.Metadata.Agent},
		{"Summary", tools.TrimLength(er.E.Description, maxColumnLen)},
	})
	t.AppendSeparator()
	t.Render()

	if er.renderDetail {
		body, err := er.E.DecodeDetails()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(body.Summary())
		}
	}
}
