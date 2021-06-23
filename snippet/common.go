package snippet

import (
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
)

func kindGlyph(n *graph.Node) string {
	if n.HasChildren() {
		return "\U0001F4C2"
	}

	return "\U0001F4DC"
}

// createNodeSummary tries to pick the first available attribute as
// a representation of the node summary
func createNodeSummary(n *graph.Node) string {
	if len(n.Title) > 0 {
		return n.Title
	}
	if len(n.Markdown) > 0 {
		return n.Markdown
	}
	if len(n.Blocks) > 0 {
		for _, block := range n.Blocks {
			return block.Content
		}
	}
	return "No-Content"
}

func renderNodeMarkdown(n *graph.Node, style string) string {
	mr, err := tools.NewMarkdownRenderer(style)
	if err != nil {
		tools.LogStderr(err, "renderNodeView: NewTermRender:")
		return n.Markdown
	}
	out, err := mr.Render(n.Markdown)
	if err != nil {
		tools.LogStderr(err, "renderNodeView: tr.Render(node.Markdown):")
		return n.Markdown
	}
	return out
}
