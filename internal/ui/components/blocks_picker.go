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
 • Reference-ID :{{.DisplayID | white }}
 • Content-Type :{{.ContentType | yellow | bold }}
 • Summary      :{{.Summary | yellow | bold }}

 • Last known Execution Info
   • Executed By :{{.LastExecutedBy | yellow  }}
   • Executed At :{{.LastExecutedAt | yellow  }}
   • Exit-Status :{{.LastExitStatus | yellow  }}
   • Addon-Info  :{{.ExecInfo | yellow }}
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
		Stdout:    &bellSkipper{},
	}

	i, _, err := prompt.RunCursorAt(startIndex, startIndex)
	if err != nil {
		if err == promptui.ErrExit {
			return nil, ErrNoEntryPicked
		}
		return nil, err
	}

	tools.Log.Info().Msgf("RenderBlockPicker: You choose number %d: %s\n", i+1, entries[i].QualifiedID())
	return entries[i], nil
}

type displayBlockRow struct {
	Index          int
	RowLen         int
	DisplayTitle   string
	DisplayID      string
	ContentType    string
	Summary        string
	Date           string
	LastExecutedBy string
	LastExitStatus string
	LastExecutedAt string
	ExecInfo       string
	entry          BlockPickEntry
}

func createDisplayBlockRows(entries BlockPickList) []*displayBlockRow {
	rows := make([]*displayBlockRow, len(entries))
	cols := colLen()
	summaryLen := 30
	if cols < summaryLen {
		summaryLen = cols
	}

	summaryLongLen := cols - 20 //20 padding
	if summaryLongLen < 0 {
		summaryLongLen = 30
	}

	for i, e := range entries {
		index := e.Index() + 1
		content := tools.RemoveNewLines(e.Block().Content, " ")
		displayTitle := tools.TrimLength(content, summaryLen)
		summary := tools.TrimLength(content, summaryLongLen)
		contentType := tools.RemoveNewLines(e.Block().ContentType.String(), "")
		execInfo := tools.RemoveNewLines(e.Block().LastExecutionInfo, "")
		execInfo = tools.TrimLength(execInfo, summaryLongLen)

		lastExecutedBy := e.Block().LastExecutedBy
		if lastExecutedBy == "" {
			lastExecutedBy = "N/A"
		}
		lastExecutedAt := "N/A"
		if e.Block().LastExecutedAt != nil {
			lastExecutedAt = tools.FormatTime(e.Block().LastExecutedAt)
		}

		rows[i] = &displayBlockRow{
			Index:          index,
			RowLen:         e.Len(),
			DisplayTitle:   displayTitle,
			DisplayID:      e.QualifiedID(),
			ContentType:    contentType,
			Summary:        summary,
			Date:           tools.FmtTime(e.Block().CreatedAt),
			entry:          entries[i],
			LastExecutedBy: lastExecutedBy,
			LastExecutedAt: lastExecutedAt,
			LastExitStatus: e.Block().LastExitStatus,
			ExecInfo:       execInfo,
		}
	}
	return rows
}
