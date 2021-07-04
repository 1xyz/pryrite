package inspector

import (
	"fmt"

	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/internal/completer"
	"github.com/aardlabs/terminal-poc/internal/history"
	"github.com/aardlabs/terminal-poc/internal/ui/components"
	"github.com/aardlabs/terminal-poc/markdown"
	"github.com/aardlabs/terminal-poc/run"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"

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

	ni.openRepl()
	return nil
}

func NewNodeInspector(graphCtx *snippet.Context, nodeID string) (*NodeInspector, error) {
	r, err := run.NewRun(graphCtx, nodeID)
	if err != nil {
		return nil, err
	}

	hist, err := history.New(fmt.Sprintf("%s/%s.json", history.HistoryDir, nodeID))
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
	case BlockActionNext:
		if n.codeBlockPos < len(n.codeBlocks)-1 {
			n.codeBlockPos++
		} else {
			tools.LogStderr(nil, "Already at end\n")
			return
		}
	case BlockActionPrev:
		if n.codeBlockPos > 0 {
			n.codeBlockPos--
		} else {
			tools.LogStderr(nil, "Already at begining\n")
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
func (n *NodeInspector) populateCodeBlocks(p *graph.NodeView, prefix string) {
	node := p.Node
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

	if p.Children != nil && len(p.Children) > 0 {
		for i := range p.Children {
			n.populateCodeBlocks(p.Children[i], pfx)
		}
	}
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
