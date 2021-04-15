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
	Nodes []graph.Node
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
func getSummaryColLen() int {
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
	colLen := getSummaryColLen()
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	for _, e := range nr.Nodes {
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

type nodeRender struct {
	Node         *graph.Node
	renderDetail bool
}

func (nr *nodeRender) Render() {
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	t.AppendRow(table.Row{"Id", nr.Node.ID})
	t.AppendRow(table.Row{"Kind", nr.Node.Kind})
	t.AppendSeparator()
	summary, err := getSummary(nr.Node, getSummaryColLen())
	if err != nil {
		log.Err(err).Msgf("getSummary")
	}
	t.AppendRows([]table.Row{
		{"SessionID", nr.Node.Metadata.SessionID},
		{"Date", tools.FmtTime(nr.Node.OccurredAt)},
		{"Agent", nr.Node.Metadata.Agent},
		{"Summary", summary},
		{"CreatedBy", nr.Node.User.Username},
	})
	t.AppendSeparator()
	t.Render()

	if nr.renderDetail {
		if len(nr.Node.Description) > 0 {
			fmt.Println(nr.Node.Description)
		}
		d, err := nr.Node.DecodeDetails()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if len(d.GetTitle()) > 0 {
			fmt.Printf("Title: %v\n", d.GetTitle())
		}
		if len(d.GetUrl()) > 0 {
			fmt.Printf("URL: %s\n", d.GetUrl())
		}
		if len(d.GetBody()) > 0 {
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
