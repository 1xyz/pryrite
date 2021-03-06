package common

import (
	"net/url"
	"strings"

	"github.com/1xyz/pryrite/config"
	"github.com/1xyz/pryrite/graph"
	"github.com/1xyz/pryrite/markdown"
	"github.com/1xyz/pryrite/tools"
)

// CreateNodeSummary tries to pick the first available attribute as
// a representation of the node summary
func CreateNodeSummary(n *graph.Node) string {
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

// GenerateNodeMarkdown uses the prescribed style to return a console ready markdown doc.
func GenerateNodeMarkdown(n *graph.Node, style string) string {
	mr, err := markdown.NewTermRenderer(style)
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

// GetServiceURL returns the service URL assoc. with this entry
func GetServiceURL(entry *config.Entry) string {
	serviceURL := entry.ServiceUrl
	if !strings.HasSuffix(serviceURL, "/") {
		serviceURL += "/"
	}
	return serviceURL
}

// GetNodeURL returns the URL representation of this node ID
func GetNodeURL(entry *config.Entry, nodeID string) *url.URL {
	u, err := url.Parse(entry.DashboardUrl)
	if err != nil {
		return nil
	}

	u.Path = "nodes/" + nodeID
	return u
}
