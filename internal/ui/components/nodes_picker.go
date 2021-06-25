package components

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/internal/common"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/manifoldco/promptui"
	"strings"
)

type NodePickResult struct {
	Node *graph.Node
}

func RenderNodesPicker(entry *config.Entry, nodes []graph.Node, header string, pageSize, startIndex int) (*NodePickResult, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("an empty node list provided")
	}

	serviceUrl := common.GetServiceURL(entry)
	rows := newDisplayRows(nodes, serviceUrl, entry.Style)

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0001F449 {{.KindGlyph | white }} {{ .Summary | yellow  | bold }} - {{ .Date | white }}",
		Inactive: "\U00002026  {{.KindGlyph | white }} {{ .Summary | yellow  }} - {{ .Date | white }}",
		Selected: "\U0001F449 {{.KindGlyph | white }} {{ .Name | cyan | bold }} - {{ .Date | white }}",
		Details: `
- {{.NodeID | white }}
- {{.SummaryLong | yellow | bold }}
`,
	}

	searcher := func(input string, index int) bool {
		row := rows[index]
		name := strings.Replace(strings.ToLower(row.Summary), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     header,
		Items:     rows,
		Templates: templates,
		Size:      pageSize,
		Searcher:  searcher,
	}

	if startIndex >= 0 {
		prompt.CursorPos = startIndex
	}

	i, _, err := prompt.Run()

	if err != nil {
		return nil, err
	}

	tools.Log.Info().Msgf("RenderNodesPicker: You choose number %d: %s\n", i+1, rows[i].Summary)
	return &NodePickResult{Node: rows[i].node}, nil
}

type displayRow struct {
	Index       int
	NodeID      string
	Summary     string
	SummaryLong string
	Markdown    string
	Date        string
	KindGlyph   string
	node        *graph.Node
}

func newDisplayRows(nodes []graph.Node, serviceURL, style string) []displayRow {
	rows := make([]displayRow, len(nodes))
	const summaryLen = 40
	for i, n := range nodes {
		nodeID := common.GetNodeURL(serviceURL, n.ID)
		summary := common.CreateNodeSummary(&nodes[i])
		date := tools.FmtTime(n.OccurredAt)

		rows[i] = displayRow{
			Index:       i + 1,
			NodeID:      nodeID,
			Summary:     tools.TrimLength(summary, summaryLen),
			SummaryLong: summary,
			Markdown:    common.GenerateNodeMarkdown(&nodes[i], style),
			Date:        date,
			KindGlyph:   kindGlyph(&nodes[i]),
			node:        &nodes[i],
		}
	}
	return rows
}

func kindGlyph(n *graph.Node) string {
	if n.HasChildren() {
		return "\U0001F4C2"
	}

	return "\U0001F4DC"
}
