package snippet

import (
	"fmt"
	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rs/zerolog/log"

	"os"
)

const (
	// min. column length
	minDisplayColLen = 10
	// max column length for the node column
	maxDisplayColLen = 120
	// total pad length. one column on either side
	padLen = 2
	// sample date column length w/ padding length
	dateColLen = len("Apr 15 10:59AM") + padLen
	// sample Kind col. len w/ padding length
	kindColLen = len(graph.ClipboardCopy) + padLen
	// id column length - a rough estimate
	idColLen = 4 + padLen
)

func RenderSnippetNode(n *graph.Node, showContent bool) error {
	nr := &nodeRender{
		Node:        n,
		showContent: showContent,
	}
	nr.Render()
	return nil
}

func RenderSnippetNodes(nodes []graph.Node) error {
	nr := &nodesRender{nodes}
	nr.Render()
	return nil
}

type nodeRender struct {
	Node        *graph.Node
	showContent bool
}

func (nr *nodeRender) Render() {
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	t.AppendRow(table.Row{"Id", nr.Node.ID})
	t.AppendRow(table.Row{"Kind", nr.Node.Kind})
	t.AppendSeparator()
	colLen := nr.getColumnLen()
	t.AppendRows([]table.Row{
		{"Date", tools.FmtTime(nr.Node.OccurredAt)},
		{"Agent", nr.Node.Metadata.Agent},
	})
	t.AppendSeparator()
	t.Render()

	if nr.showContent {
		if len(nr.Node.Title) > 0 {
			result := markdown.Render("# "+nr.Node.Title, colLen, padLen)
			fmt.Println(string(result))
		}

		fmt.Println()
		fmt.Println("  ────────────")
		fmt.Println()

		if len(nr.Node.Description) > 0 {
			result := markdown.Render(nr.Node.Description, colLen, padLen)
			fmt.Println(string(result))
		}

		fmt.Println()
		fmt.Println("  ────────────")
		fmt.Println()

		if len(nr.Node.Content) > 0 {
			result := markdown.Render(nr.Node.Content, colLen, padLen)
			fmt.Println(string(result))
		} else {
			fmt.Println("  No content")
		}
	}
}

func (nr *nodeRender) getColumnLen() int {
	_, maxCols, err := tools.GetTermWindowSize()
	if err != nil {
		log.Err(err).Msg("tools.GetTermWindowSize()")
		return minDisplayColLen
	}
	// allowedLen is the maximum columns allowed
	allowedLen := maxCols - (len("description") + 6) // add 5 for column bars
	if allowedLen < minDisplayColLen {
		// this would make things unreadable, but we can revisit this
		return minDisplayColLen
	}
	if allowedLen > maxDisplayColLen {
		return maxDisplayColLen
	}
	return allowedLen
}

type nodesRender struct {
	Nodes []graph.Node
}

func (nr *nodesRender) Render() {
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	for _, e := range nr.Nodes {
		summary := nr.getSummary(&e)
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

func (nr *nodesRender) getColumnLen() int {
	_, maxCols, err := tools.GetTermWindowSize()
	if err != nil {
		log.Err(err).Msg("tools.GetTermWindowSize()")
		return minDisplayColLen
	}
	// allowedLen is the maximum columns allowed
	allowedLen := maxCols - (dateColLen + kindColLen + padLen + idColLen + 6) // add 5 for column bars
	if allowedLen < minDisplayColLen {
		// this would make things unreadable, but we can revisit this
		return minDisplayColLen
	}
	if allowedLen > maxDisplayColLen {
		return maxDisplayColLen
	}
	return allowedLen
}

func (nr *nodesRender) getSummary(n *graph.Node) string {
	colLen := nr.getColumnLen()
	if len(n.Title) > 0 {
		return tools.TrimLength(n.Title, colLen)
	}
	if len(n.Description) > 0 {
		return tools.TrimLength(n.Description, colLen)
	}
	if len(n.Content) > 0 {
		return tools.TrimLength(n.Content, colLen)
	}
	return "No-Content"
}
