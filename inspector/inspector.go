package inspector

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
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

		switch nextAction.Action {
		case BlockActionQuit:
			break
		case BlockActionNext:
			i++
		case BlockActionPrev:
			i--
			if i <= 0 {
				i = 0
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

	ni := &NodeInspector{
		run:        r,
		codeBlocks: []*codeBlock{},
	}
	ni.populateCodeBlocks(r.Root, "")
	return ni, err
}

type NodeInspector struct {
	run        *run.Run
	codeBlocks []*codeBlock
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
			curIndex := len(n.codeBlocks) + 1
			n.codeBlocks = append(n.codeBlocks,
				newCodeBlock(curIndex, pfx, node.Blocks[i], node))
		}
	}

	if p.Children != nil && len(p.Children) > 0 {
		for i := range p.Children {
			n.populateCodeBlocks(p.Children[i], pfx)
		}
	}
}

func md(content, style string) string {
	mr, err := tools.NewMarkdownRenderer(style)
	if err != nil {
		tools.LogStderr(err, "md: NewTermRender:")
		return content
	}
	out, err := mr.Render(content)
	if err != nil {
		tools.LogStderr(err, "md: tr.Render(node.Markdown):")
		return content
	}
	return out
}
