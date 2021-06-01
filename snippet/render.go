package snippet

import (
	"fmt"
	"strings"

	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/charmbracelet/glamour"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rs/zerolog/log"
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
	idColLen = 60 + padLen
)

func RenderSnippetNodeView(entry *config.Entry, nv *graph.NodeView) error {
	nr := &nodeRender{
		view:       nv,
		serviceURL: getServiceURL(entry),
	}
	nr.Render()
	return nil
}

func RenderSnippetNodes(entry *config.Entry, nodes []graph.Node, kind graph.Kind) error {
	if len(nodes) == 0 {
		tools.LogStdout("No results found!")
		return nil
	}
	nr := &nodesRender{Nodes: nodes, kind: kind, serviceURL: getServiceURL(entry)}
	nr.Render()
	return nil
}

func getServiceURL(entry *config.Entry) string {
	serviceURL := entry.ServiceUrl
	if !strings.HasSuffix(serviceURL, "/") {
		serviceURL += "/"
	}
	return serviceURL
}

type nodeRender struct {
	view       *graph.NodeView
	serviceURL string
}

func (nr *nodeRender) Render() {
	w, err := tools.OpenOutputWriter()
	if err != nil {
		tools.LogStdout(fmt.Sprintf("Render: tools.OpenOutputWriter: err = %v", err))
		return
	}
	tools.Log.Info().Msgf("Render: OpenOutputWriter opened SupportRawControlChars: %b",
		w.SupportRawControlChars())

	defer func() {
		if err := w.Close(); err != nil {
			tools.Log.Err(err).Msgf("Render: defer: error closing writer")
			return
		}
	}()

	nr.renderNodeView(nr.view, w)
	if nr.view.Children != nil && len(nr.view.Children) > 0 {
		for _, child := range nr.view.Children {
			nr.renderNodeView(child, w)
		}
	}

	tools.Log.Info().Msgf("Render complete for node %v", nr.view.Node.ID)
}

func (nr *nodeRender) renderNodeView(nv *graph.NodeView, w tools.OutputWriteCloser) {
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(w)
	t.AppendRow(table.Row{"Id", fmtID(nr.serviceURL, nv.Node.ID)})
	t.AppendRow(table.Row{"Kind", nv.Node.Kind})
	t.AppendSeparator()
	t.AppendRows([]table.Row{
		{"Date", tools.FmtTime(nv.Node.OccurredAt)},
		{"Agent", nv.Node.Metadata.Agent},
	})
	t.AppendSeparator()
	t.Render()

	var renderOpts []glamour.TermRendererOption
	if w.SupportRawControlChars() {
		colLen := nr.getColumnLen()
		renderOpts = []glamour.TermRendererOption{
			glamour.WithAutoStyle(), glamour.WithWordWrap(colLen),
		}
	} else {
		renderOpts = []glamour.TermRendererOption{
			glamour.WithStylePath("notty"),
		}
	}

	var out string
	r, err := glamour.NewTermRenderer(renderOpts...)
	if err != nil {
		tools.Log.Err(err).Msgf("newTermRender %v failed %v", renderOpts, err)
		if _, err := fmt.Fprint(w, out); err != nil {
			tools.Log.Err(err).Msgf("renderNodeView: id = %v fmt.Fprintf()", nv.Node.ID)
		}
		return
	}

	out, err = r.Render(nv.View)
	if err != nil {
		tools.Log.Err(err).Msgf("renderNodeView: id = %v fmt.Fprintf()", nv.Node.ID)
		return
	}
	if _, err := fmt.Fprint(w, out); err != nil {
		tools.Log.Err(err).Msgf("renderNodeView: id = %v fmt.Fprintf()", nv.Node.ID)
		return
	}
	tools.Log.Info().Msgf("renderNodeView: id=%v complete", nv.Node.ID)
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
	w, err := tools.OpenOutputWriter()
	if err != nil {
		tools.LogStdout(fmt.Sprintf("Render: tools.OpenOutputWriter: err = %v", err))
		return
	}
	defer func() {
		if err := w.Close(); err != nil {
			tools.Log.Err(err).Msgf("Render defer: error closing writer")
			return
		}
	}()

	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.SetOutputMirror(w)
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
	if len(n.Markdown) > 0 {
		return tools.TrimLength(n.Markdown, colLen)
	}
	if len(n.Snippets) > 0 {
		for _, snippet := range n.Snippets {
			return tools.TrimLength(snippet.Content, colLen)
		}
	}
	return "No-Content"
}

func fmtID(serviceURL, id string) string {
	return fmt.Sprintf("%snodes/%s", serviceURL, id)
}
