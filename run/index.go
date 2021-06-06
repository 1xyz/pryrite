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

// BlockExecutionResults is a log of execution results
type BlockExecutionResults []*graph.BlockExecutionResult

// NodeExecResultIndex is map of BlockExecutionResults indexed by nodeID
type NodeExecResultIndex map[string]BlockExecutionResults

func (n NodeExecResultIndex) Append(res *graph.BlockExecutionResult) {
	_, found := n[res.NodeID]
	if !found {
		n[res.NodeID] = make(BlockExecutionResults, 0)
	}
	n[res.NodeID] = append(n[res.NodeID], res)
}

func (n NodeExecResultIndex) Get(nodeID string) (BlockExecutionResults, bool) {
	e, found := n[nodeID]
	return e, found
}

// Find scans the slice of BlockExecutionResults by resultID
func (b BlockExecutionResults) Find(requestID string) (*graph.BlockExecutionResult, bool) {
	for _, entry := range b {
		if entry.RequestID == requestID {
			return entry, true
		}
	}
	return nil, false
}
