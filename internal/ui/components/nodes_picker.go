package components

import (
	"fmt"
	"strings"

	"github.com/1xyz/pryrite/config"
	"github.com/1xyz/pryrite/graph"
	"github.com/1xyz/pryrite/internal/common"
	"github.com/1xyz/pryrite/tools"
	"github.com/manifoldco/promptui"
)

type NodePickResult struct {
	Node *graph.Node
}

func RenderNodesPicker(entry *config.Entry, nodes []graph.Node, header string, pageSize, startIndex int) (*NodePickResult, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("an empty node list provided")
	}

	rows := newDisplayRows(nodes, entry)

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0001F449 {{.KindGlyph | white }} {{ .Summary | yellow  | bold }} - {{ .Date | white }}",
		Inactive: "\U00002026  {{.KindGlyph | white }} {{ .Summary | yellow  }} - {{ .Date | white }}",
		Selected: "\U0001F449 {{.KindGlyph | white }} {{ .Name | cyan | bold }} - {{ .Date | white }}",
		Details: `
 • URL      {{.NodeID | white }}
 • Created  {{.DateLong | white }}
 • Summary  {{.SummaryLong | yellow | bold }}
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
		Stdout:    newBellSkipper(),
	}

	if startIndex >= 0 {
		prompt.CursorPos = startIndex
	}

	i, _, err := prompt.Run()

	if err != nil {
		if err == promptui.ErrExit {
			return nil, ErrNoEntryPicked
		}
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
	DateLong    string
	KindGlyph   string
	node        *graph.Node
}

func newDisplayRows(nodes []graph.Node, entry *config.Entry) []displayRow {
	cols := colLen()
	summaryLen := 30
	if cols < summaryLen {
		summaryLen = cols
	}

	summaryLongLen := cols - 20 //20 padding
	if summaryLongLen < 0 {
		summaryLongLen = 30
	}

	rows := make([]displayRow, len(nodes))
	for i, n := range nodes {
		nodeID := common.GetNodeURL(entry, n.ID).String()
		summary := common.CreateNodeSummary(&nodes[i])
		summary = tools.RemoveNewLines(summary, " ")
		date := tools.FmtTime(n.OccurredAt)
		dateLong := n.OccurredAt.Format("Mon, 02 Jan 2006 15:04:05 MST")

		rows[i] = displayRow{
			Index:       i + 1,
			NodeID:      nodeID,
			Summary:     tools.TrimLength(summary, summaryLen),
			SummaryLong: tools.TrimLength(summary, summaryLongLen),
			Markdown:    common.GenerateNodeMarkdown(&nodes[i], entry.Style),
			Date:        date,
			DateLong:    dateLong,
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
