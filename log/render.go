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
	// min. column length
	minSummaryColLen = 10
	// max column length for the summary column
	maxSummaryColLen = 80
	// total pad length. one column on either side
	padLen = 2
	// sample date column length w/ padding length
	dateColLen = len("Apr 15 10:59AM") + padLen
	// sample Kind col. len w/ padding length
	kindColLen = len(graph.ClipboardCopy) + padLen
	// id column length - a rough estimate
	idColLen = 4 + padLen
)

// getSummaryColLen basically returns a value between [minSummaryColLen, maxSummaryColLen]
func (nr *nodesRender) getSummaryColLen() int {
	_, maxCols, err := tools.GetTermWindowSize()
	if err != nil {
		log.Err(err).Msg("tools.GetTermWindowSize()")
		return minSummaryColLen
	}
	// allowedLen is the maximum columns allowed
	allowedLen := maxCols - (dateColLen + kindColLen + padLen + idColLen + 6) // add 5 for column bars
	if allowedLen < minSummaryColLen {
		// this would make things unreadable, but we can revisit this
		return minSummaryColLen
	}
	if allowedLen > maxSummaryColLen {
		return maxSummaryColLen
	}
	return allowedLen
}

func (nr *nodesRender) Render() {
	colLen := nr.getSummaryColLen()
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	for _, e := range nr.G {
		summary, err := getSummary(&e, colLen)
		if err != nil {
			log.Err(err).Msgf("getSummary = %v", err)
		}
		t.AppendRow(table.Row{
			e.ID,
			tools.FmtTime(e.OccurredAt),
			summary,
			e.Kind})
		t.AppendRow(table.Row{})
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
	summary, err := getSummary(er.E, maxSummaryColLen)
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
