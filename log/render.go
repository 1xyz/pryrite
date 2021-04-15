package log

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

type nodesRender struct {
	G []graph.Node
}

const (
	maxColumnLen = 40
)

func (nr *nodesRender) Render() {
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	for _, e := range nr.G {
		summary, err := getSummary(&e, maxColumnLen)
		if err != nil {
			log.Err(err).Msgf("getSummary = %v", err)
		}
		t.AppendRow(table.Row{
			e.ID,
			tools.FmtTime(e.OccurredAt),
			summary,
			e.Kind})
	}
	t.AppendHeader(table.Row{"Id", "Date", "Summary", "Type"})
	t.AppendSeparator()
	t.Render()
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
	summary, err := getSummary(er.E, maxColumnLen)
	if err != nil {
		log.Err(err).Msgf("getSummary")
	}
	t.AppendRows([]table.Row{
		{"SessionID", er.E.Metadata.SessionID},
		{"Date", tools.FmtTime(er.E.OccurredAt)},
		{"CreatedOn", tools.FmtTime(er.E.CreatedAt)},
		{"Agent", er.E.Metadata.Agent},
		{"Summary", summary},
	})
	t.AppendSeparator()
	t.Render()

	if er.renderDetail {
		d, err := er.E.DecodeDetails()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(d.GetBody())
		}
	}
}

func getSummary(n *graph.Node, columnLen int) (string, error) {
	ns := &nodeSummary{
		b:         strings.Builder{},
		n:         n,
		columnLen: columnLen,
	}
	if err := ns.createSummary(); err != nil {
		return "", err
	}
	return ns.b.String(), nil
}

type nodeSummary struct {
	b         strings.Builder
	n         *graph.Node
	columnLen int
}

func (ns *nodeSummary) createSummary() error {
	if err := ns.append(ns.n.Description); err != nil {
		return err
	}
	d, err := ns.n.DecodeDetails()
	if err != nil {
		return err
	}
	return ns.appendAll(d.GetTitle(), d.GetUrl(), d.GetBody())
}

func (ns *nodeSummary) appendAll(vals ...string) error {
	for _, s := range vals {
		if err := ns.append(s); err != nil {
			return err
		}
	}
	return nil
}

func (ns *nodeSummary) append(s string) error {
	if len(strings.TrimSpace(s)) == 0 {
		return nil
	}
	if ns.b.Len() > 0 {
		if _, err := fmt.Fprintln(&ns.b); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(&ns.b, "%s", tools.TrimLength(s, ns.columnLen)); err != nil {
		return err
	}
	return nil
}
