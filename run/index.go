package run

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
)

type NodeViewIndex map[string]*graph.NodeView

func (ni NodeViewIndex) Add(view *graph.NodeView) error {
	id := view.Node.ID
	if ni.ContainsID(id) {
		return fmt.Errorf("an entry with id=%s exists", id)
	}
	ni[id] = view
	return nil
}

func (ni NodeViewIndex) ContainsID(id string) bool {
	_, found := ni[id]
	return found
}

func (ni NodeViewIndex) Get(id string) (*graph.NodeView, error) {
	e, found := ni[id]
	if !found {
		return nil, fmt.Errorf("nodeIndex.Get(id=%s) not found", id)
	}
	return e, nil
}

type NodeExecResultIndex map[string]*graph.NodeExecutionResult

func (n NodeExecResultIndex) Set(res *graph.NodeExecutionResult) { n[res.NodeID] = res }
func (n NodeExecResultIndex) Get(nodeID string) (*graph.NodeExecutionResult, bool) {
	entry, found := n[nodeID]
	return entry, found
}
