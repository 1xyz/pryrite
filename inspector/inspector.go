package inspector

import (
	_ "embed"
	"fmt"
	"github.com/1xyz/pryrite/graph/log"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"

	"github.com/1xyz/pryrite/config"
	"github.com/1xyz/pryrite/graph"
	"github.com/1xyz/pryrite/internal/completer"
	"github.com/1xyz/pryrite/internal/history"
	"github.com/1xyz/pryrite/internal/ui/components"
	"github.com/1xyz/pryrite/markdown"
	"github.com/1xyz/pryrite/run"
	"github.com/1xyz/pryrite/snippet"
	"github.com/1xyz/pryrite/tools"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

func InspectNode(gCtx *snippet.Context, nodeID string) error {
	ni, err := NewNodeInspector(gCtx, nodeID)
	if err != nil {
		return err
	}
	if len(ni.codeBlocks) == 0 {
		return fmt.Errorf("no code blocks found for %s", nodeID)
	}

	go ni.runner.Start()
	defer func() {
		ni.runner.Shutdown()
	}()

	if err := introMsg(gCtx.ConfigEntry); err != nil {
		tools.Log.Err(err).Msgf("introMsg")
	}

	ni.openRepl()
	return nil
}

func NewNodeInspector(graphCtx *snippet.Context, nodeID string) (*NodeInspector, error) {
	r, err := run.NewRun(graphCtx, nodeID)
	if err != nil {
		return nil, err
	}

	hist, err := history.New(fmt.Sprintf("%s/%s", history.HistoryDir, nodeID))
	if err != nil {
		return nil, err
	}

	ni := &NodeInspector{
		runner:       r,
		codeBlocks:   []*codeBlock{},
		codeBlockPos: 0,
		hist:         hist,
	}
	ni.populateCodeBlocks(r.Root, "")
	for _, b := range ni.codeBlocks {
		b.SetNBlocks(len(ni.codeBlocks))
	}
	return ni, err
}

type NodeInspector struct {
	runner       *run.Run
	codeBlocks   []*codeBlock
	codeBlockPos int
	hist         history.History
}

func (n *NodeInspector) NewRootCmd() *cobra.Command {
	return newRootCmd(n)
}

func (n *NodeInspector) HistoryAppend(cmd string) error {
	return n.hist.Append(cmd)
}

func (n *NodeInspector) openRepl() {
	runner := tools.NewRunner(n)
	cc := completer.NewCobraCommandCompleter(newRootCmd(n))
	pt := prompt.New(
		runner.Execute,
		cc.Complete,
		prompt.OptionTitle("interactive inspector"),
		prompt.OptionLivePrefix(n.updatePromptPrefix),
		prompt.OptionPrefixTextColor(prompt.Green),
		prompt.OptionInputTextColor(prompt.Yellow),
	)

	n.currentBlock().WhereAmI()

	pt.Run()
}

func (n *NodeInspector) updatePromptPrefix() (string, bool) {
	prefix := fmt.Sprintf("[Step %d of %d] >>> ", n.codeBlockPos+1, len(n.codeBlocks))
	return prefix, true
}

func (n *NodeInspector) currentBlock() *codeBlock {
	return n.codeBlocks[n.codeBlockPos]
}

func (n *NodeInspector) processAction(nextAction *BlockAction) {
	if nextAction.Action == BlockActionQuit {
		// FIXME quit!
		return
	}

	switch nextAction.Action {
	case BlockActionExecutionDone:
		// This indicates the step is done, so lets move ahead
		// but unlike next; don't show an error message
		if n.codeBlockPos < len(n.codeBlocks)-1 {
			n.codeBlockPos++
		}
	case BlockActionNext:
		if n.codeBlockPos < len(n.codeBlocks)-1 {
			n.codeBlockPos++
		} else {
			tools.LogStderr(nil, "You are at the last step\n")
			return
		}
	case BlockActionPrev:
		if n.codeBlockPos > 0 {
			n.codeBlockPos--
		} else {
			tools.LogStderr(nil, "You are at the first step\n")
			return
		}
	case BlockActionJump:
		entries := components.BlockPickList{}
		for j := range n.codeBlocks {
			entries = append(entries, n.codeBlocks[j])
		}

		selEntry, err := components.RenderBlockPicker(entries, "Switch to code-block", 10, n.codeBlockPos)
		if err != nil {
			if err != components.ErrNoEntryPicked {
				tools.LogStdError("RenderBlockPicker err = %v", err)
				return
			}
		} else {
			n.codeBlockPos = selEntry.Index()
		}
	}

	n.currentBlock().WhereAmI()
}

// populateCodeBlocks Flatten the tree into a list in a pre-order
// depth traversal first a node's code blocks are added;
// followed by a traversal for the first child and so on...
func (n *NodeInspector) populateCodeBlocks(node *graph.Node, prefix string) {
	pfx := fmt.Sprintf("%s/%s", prefix, node.ID)
	if node.HasBlocks() {
		for i := range node.Blocks {
			if !node.Blocks[i].IsCode() {
				continue
			}
			curIndex := len(n.codeBlocks)
			n.codeBlocks = append(n.codeBlocks,
				newCodeBlock(curIndex, pfx, node.Blocks[i], node, n.runner, n.hist))
		}
	}

	if node.ChildNodes != nil {
		for i := range node.ChildNodes {
			n.populateCodeBlocks(node.ChildNodes[i], pfx)
		}
	}
}

func (n *NodeInspector) ResultLog() (log.ResultLog, error) {
	return n.runner.ExecIndex.Get(n.runner.PlaybookID)
}

func (n *NodeInspector) IterateLogEntries(limit int) error {
	rl, err := n.ResultLog()
	if err != nil {
		return err
	}

	count := 0
	if err := rl.Each(func(i int, entry *log.ResultLogEntry) bool {
		if entry.State == log.ExecStateStarted || entry.State == log.ExecStateQueued {
			// skip the started and queued states. There is not much to show
			return true
		}

		count++
		if count > limit {
			return false
		}

		executedAt := "unknown"
		if entry.ExecutedAt != nil {
			executedAt = entry.ExecutedAt.Format(timeLayout)
		}

		renderRows(table.Row{"log entry:", fmt.Sprintf("log entry %d", count)})
		renderRows([]table.Row{
			{"Block", entry.BlockID},
			{"Executed On", executedAt},
			{"Exit Status", entry.ExitStatus},
			{"Error", entry.Err},
		}...)

		renderRows(table.Row{"Command"})
		fmt.Fprintf(os.Stdout, Strip(entry.Content))
		fmt.Fprintf(os.Stdout, "\n")

		if len(entry.Stdout) > 0 {
			renderRows(table.Row{"stdout"})
			fmt.Fprintf(os.Stdout, Strip(entry.Stdout))
			fmt.Fprintf(os.Stdout, "\n")
		}

		if len(entry.Stderr) > 0 {
			renderRows(table.Row{"stderr"})
			fmt.Fprintf(os.Stdout, Strip(entry.Stdout))
			fmt.Fprintf(os.Stdout, "\n")
		}

		return true
	}); err != nil {
		return fmt.Errorf("rleach Err = %w", err)
	}

	return nil
}

var (
	//go:embed intro.md
	intro string
)

func introMsg(entry *config.Entry) error {
	if entry.HideInspectIntro {
		return nil
	}
	tr, err := markdown.NewTermRenderer("")
	if err != nil {
		return err
	}
	content, err := tr.Render(intro)
	if err != nil {
		return err
	}
	fmt.Println(content)
	entry.HideInspectIntro = !components.ShowYNQuestionPrompt("Show this message again?")
	return config.SetEntry(entry)
}

func md(content, style string, cursor *markdown.Cursor) string {
	mr, err := markdown.NewLinenumRenderer(style)
	if err != nil {
		tools.LogStderr(err, "md: NewTermRender:")
		return content
	}

	result, err := mr.Render(content, cursor)
	if err != nil {
		tools.LogStderr(err, "md: tr.Render(node.Markdown):")
		return content
	}

	return result
}
