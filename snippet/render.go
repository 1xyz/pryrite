package snippet

import (
	"fmt"
	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rs/zerolog/log"
	"strings"

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
	// id URL column length - a rough estimate
	idColLen = 32 + padLen
)

func RenderSnippetNode(cfg *config.Config, n *graph.Node, showContent bool) error {
	nr := &nodeRender{
		Node:        n,
		showContent: showContent,
		serviceURL:  getServiceURL(cfg),
	}
	nr.Render()
	return nil
}

func RenderSnippetNodes(cfg *config.Config, nodes []graph.Node, kind graph.Kind) error {
	nr := &nodesRender{Nodes: nodes, kind: kind, serviceURL: getServiceURL(cfg)}
	nr.Render()
	return nil
}

func getServiceURL(cfg *config.Config) string {
	entry, found := cfg.GetDefaultEntry()
	if !found {
		tools.Log.Warn().Msgf("RenderSnippetNode(s): cannot find default entry!")
	}
	serviceURL := entry.ServiceUrl
	if !strings.HasSuffix(serviceURL, "/") {
		serviceURL += "/"
	}
	return serviceURL
}

type nodeRender struct {
	Node        *graph.Node
	showContent bool
	serviceURL  string
}

func (nr *nodeRender) Render() {
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	t.AppendRow(table.Row{"Id", fmtID(nr.serviceURL, nr.Node.ID)})
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
			result := markdown.Render("* Title:-", colLen, padLen)
			fmt.Println(string(result))
			result = markdown.Render(nr.Node.Title, colLen, padLen)
			fmt.Println(string(result))
		}
		if len(nr.Node.Description) > 0 {
			result := markdown.Render("* Description:-", colLen, padLen)
			fmt.Println(string(result))
			result = markdown.Render(nr.Node.Description, colLen, padLen)
			fmt.Println(string(result))
			fmt.Println()
		}

		result := markdown.Render("* Content:-", colLen, padLen)
		fmt.Println(string(result))
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
	Nodes      []graph.Node
	kind       graph.Kind
	serviceURL string
}

func (nr *nodesRender) Render() {
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(os.Stdout)
	for _, e := range nr.Nodes {
		summary := nr.getSummary(&e)
		idURL := fmtID(nr.serviceURL, e.ID)
		if nr.kind == graph.Unknown {
			t.AppendRow(table.Row{
				idURL,
				tools.FmtTime(e.OccurredAt),
				summary,
				e.Kind})
		} else {
			t.AppendRow(table.Row{
				idURL,
				tools.FmtTime(e.OccurredAt),
				summary})
		}
		t.AppendRow(table.Row{})
	}
	if nr.kind == graph.Unknown {
		t.AppendHeader(table.Row{"Id", "Date", "Summary", "Kind"})
	} else {
		t.AppendHeader(table.Row{"Id", "Date", "Summary"})
	}
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

func fmtID(serviceURL, id string) string {
	return fmt.Sprintf("%snodes/%s", serviceURL, id)
}
