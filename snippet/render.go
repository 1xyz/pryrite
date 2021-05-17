package snippet

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rs/zerolog/log"
	"io"
	"strings"
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

type RenderNodeViewOpts struct {
	RenderMarkdown bool
}

func RenderSnippetNodeView(cfg *config.Config, nv *graph.NodeView, opts *RenderNodeViewOpts) error {
	nr := &nodeRender{
		view:       nv,
		serviceURL: getServiceURL(cfg),
		opts:       opts,
	}
	nr.Render()
	return nil
}

func RenderSnippetNodes(cfg *config.Config, nodes []graph.Node, kind graph.Kind) error {
	if len(nodes) == 0 {
		tools.LogStdout("No results found!")
		return nil
	}
	nr := &nodesRender{Nodes: nodes, kind: kind, serviceURL: getServiceURL(cfg)}
	nr.Render()
	return nil
}

func getServiceURL(cfg *config.Config) string {
	entry, found := cfg.GetDefaultEntry()
	if !found {
		tools.Log.Warn().Msgf("RenderSnippetNodeView(s): cannot find default entry!")
	}
	serviceURL := entry.ServiceUrl
	if !strings.HasSuffix(serviceURL, "/") {
		serviceURL += "/"
	}
	return serviceURL
}

type nodeRender struct {
	view       *graph.NodeView
	serviceURL string
	opts       *RenderNodeViewOpts
}

func (nr *nodeRender) Render() {
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
	nr.renderNodeView(nr.view, w, err)
	if nr.view.Children != nil && len(nr.view.Children) > 0 {
		for _, child := range nr.view.Children {
			nr.renderNodeView(child, w, err)
		}
	}
	tools.Log.Info().Msgf("Render complete for node %v", nr.view.Node.ID)
}

func (nr *nodeRender) renderNodeView(nv *graph.NodeView, w io.WriteCloser, err error) {
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

	out := nv.ContentMarkdown
	//colLen := nr.getColumnLen()
	//r, _ := glamour.NewTermRenderer(
	//	// detect background color and pick either the default dark or light theme
	//	glamour.WithAutoStyle(),
	//	// wrap output at specific width
	//	glamour.WithWordWrap(colLen),
	//)
	//
	//// Glamour rendering preserves carriage return characters in code blocks, but
	//// we need to ensure that no such characters are present in the output.
	//// text := strings.ReplaceAll(nv.ContentMarkdown, "\r\n", "\n")
	//text := nv.ContentMarkdown
	//out, err := r.Render(text)
	//if err != nil {
	//	tools.Log.Err(err).Msgf("renderNodeView: id = %v fmt.Fprintf()", nv.Node.ID)
	//	return
	//}
	if _, err := fmt.Fprint(w, out); err != nil {
		tools.Log.Err(err).Msgf("renderNodeView: id = %v fmt.Fprintf()", nv.Node.ID)
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
