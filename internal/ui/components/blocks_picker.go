package components

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/manifoldco/promptui"
	"strings"
)

// BlockPickEntry is a single entry in the BlockPickList
type BlockPickEntry interface {
	QualifiedID() string
	Block() *graph.Block
	// Index of this entry in the BlockPickList
	Index() int
	// Len of the BlockPickList
	Len() int
}

type BlockPickList []BlockPickEntry

func RenderBlockPicker(entries BlockPickList, header string, pageSize, startIndex int) (BlockPickEntry, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("an empty block entroes provided")
	}

	rows := createDisplayBlockRows(entries)
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0001F449  [{{.Index | red | bold }}/{{ .RowLen | red | bold }}] {{ .DisplayTitle | yellow  | bold }} - {{ .Date | white }}",
		Inactive: "\U00002026   [{{.Index | red | bold }}/{{ .RowLen | red | bold }}] {{ .DisplayTitle | yellow  }} - {{ .Date | white }}",
		Selected: "\U0001F449  [{{.Index | red | bold }}/{{ .RowLen | red | bold }}] {{ .DisplayTitle | cyan | bold }} - {{ .Date | white }}",
		Details: `
- {{.DisplayID | white }}
- {{.ContentType | yellow | bold }}
- {{.Summary | yellow | bold }}
`,
	}

	searcher := func(input string, index int) bool {
		row := rows[index]
		name := strings.Replace(strings.ToLower(row.DisplayTitle), " ", "", -1)
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

	tools.Log.Info().Msgf("RenderBlockPicker: You choose number %d: %s\n", i+1, entries[i].QualifiedID())
	return entries[i], nil
}

type displayBlockRow struct {
	Index        int
	RowLen       int
	DisplayTitle string
	DisplayID    string
	ContentType  string
	Summary      string
	Date         string
	entry        BlockPickEntry
}

func createDisplayBlockRows(entries BlockPickList) []*displayBlockRow {
	rows := make([]*displayBlockRow, len(entries))
	for i, e := range entries {
		index := e.Index() + 1
		displayTitle := strings.Trim(e.Block().Content, "\n")
		displayTitle = tools.TrimLength(displayTitle, 30)
		rows[i] = &displayBlockRow{
			Index:        index,
			RowLen:       e.Len(),
			DisplayTitle: displayTitle,
			DisplayID:    e.QualifiedID(),
			ContentType:  e.Block().ContentType.String(),
			Summary:      e.Block().Content,
			Date:         tools.FmtTime(e.Block().CreatedAt),
			entry:        entries[i],
		}
	}
	return rows
}
