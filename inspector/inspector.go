package inspector

import (
	"fmt"

	"github.com/aardlabs/terminal-poc/internal/history"
	"github.com/aardlabs/terminal-poc/markdown"

	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/internal/ui/components"
	"github.com/aardlabs/terminal-poc/run"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
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
	i := 0
	for {
		if i >= len(ni.codeBlocks) {
			break
		}

		currentBlock := ni.codeBlocks[i]
		currentBlock.OpenRepl()

		nextAction, found := currentBlock.GetExitAction()
		if !found {
			break
		}

		if nextAction.Action == BlockActionQuit {
			break
		}

		switch nextAction.Action {
		case BlockActionNext:
			i++
		case BlockActionPrev:
			i--
			if i <= 0 {
				i = 0
			}
		case BlockActionJump:
			entries := components.BlockPickList{}
			for j := range ni.codeBlocks {
				entries = append(entries, ni.codeBlocks[j])
			}

			selEntry, err := components.RenderBlockPicker(entries, "Switch to code-block", 10, i)
			if err != nil {
				if err != components.ErrNoEntryPicked {
					tools.LogStdError("RenderBlockPicker err = %v", err)
					return err
				}
			} else {
				i = selEntry.Index()
			}
		}
	}
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
		runner:     r,
		codeBlocks: []*codeBlock{},
		hist:       hist,
	}
	ni.populateCodeBlocks(r.Root, "")
	for _, b := range ni.codeBlocks {
		b.SetNBlocks(len(ni.codeBlocks))
	}
	return ni, err
}

type NodeInspector struct {
	runner     *run.Run
	codeBlocks []*codeBlock
	hist       history.History
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
