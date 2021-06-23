package snippet

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/manifoldco/promptui"
	"strings"
)

func RenderNodesPicker(entry *config.Entry, nodes []graph.Node, header string, pageSize int) error {
	if len(nodes) == 0 {
		return nil
	}

	serviceUrl := getServiceURL(entry)
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

	i, _, err := prompt.Run()

	if err != nil {
		return err
	}

	fmt.Printf("You choose number %d: %s\n", i+1, rows[i].Summary)
	return nil
}

type displayRow struct {
	Index       int
	NodeID      string
	Summary     string
	SummaryLong string
	Markdown    string
	Date        string
	KindGlyph   string
}

func newDisplayRows(nodes []graph.Node, serviceURL, style string) []displayRow {
	rows := make([]displayRow, len(nodes))
	const summaryLen = 40
	for i, n := range nodes {
		nodeID := fmtID(serviceURL, n.ID)
		summary := createNodeSummary(&nodes[i])
		date := tools.FmtTime(n.OccurredAt)

		rows[i] = displayRow{
			Index:       i + 1,
			NodeID:      nodeID,
			Summary:     tools.TrimLength(summary, summaryLen),
			SummaryLong: summary,
			Markdown:    renderNodeMarkdown(&nodes[i], style),
			Date:        date,
			KindGlyph:   kindGlyph(&nodes[i]),
		}
	}
	return rows
}
