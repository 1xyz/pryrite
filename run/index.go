package run

import (
	"fmt"
	"sync"

	"github.com/aardlabs/terminal-poc/graph"
)

type NodeViewIndex struct {
	sync.Map
}

func NewNodeViewIndex() *NodeViewIndex {
	return &NodeViewIndex{}
}

func (ni *NodeViewIndex) Add(view *graph.NodeView) error {
	id := view.Node.ID
	_, loaded := ni.LoadOrStore(id, view)
	if loaded {
		return fmt.Errorf("an entry with id=%s exists", id)
	}
	return nil
}

func (ni *NodeViewIndex) ContainsID(id string) bool {
	_, found := ni.Load(id)
	return found
}

func (ni *NodeViewIndex) Get(id string) (*graph.NodeView, error) {
	e, found := ni.Load(id)
	if !found {
		return nil, fmt.Errorf("nodeIndex.Get(id=%s) not found", id)
	}
	return e.(*graph.NodeView), nil
}

// BlockExecutionResults is a log of execution results
type BlockExecutionResults struct {
	list []*graph.BlockExecutionResult
	lock sync.RWMutex
}

// NodeExecResultIndex is map of BlockExecutionResults indexed by nodeID
type NodeExecResultIndex struct {
	sync.Map
}

func NewNodeExecResultIndex() *NodeExecResultIndex {
	return &NodeExecResultIndex{}
}

func (n *NodeExecResultIndex) Append(res *graph.BlockExecutionResult) {
	blockResults, _ := n.LoadOrStore(res.NodeID, &BlockExecutionResults{})
	blockResults.(*BlockExecutionResults).Append(res)
}

func (n *NodeExecResultIndex) Get(nodeID string) (*BlockExecutionResults, bool) {
	e, ok := n.Load(nodeID)
	if ok {
		return e.(*BlockExecutionResults), true
	}
	return nil, false
}

func (b *BlockExecutionResults) Len() int {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return len(b.list)
}

func (b *BlockExecutionResults) Append(blockResult *graph.BlockExecutionResult) {
	b.lock.Lock()
	b.list = append(b.list, blockResult)
	b.lock.Unlock()
}

func (b *BlockExecutionResults) Each(cb func(int, *graph.BlockExecutionResult) bool) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	for i, entry := range b.list {
		if !cb(i, entry) {
			return
		}
	}
}

func (b *BlockExecutionResults) EachFromEnd(cb func(int, *graph.BlockExecutionResult) bool) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	for i := len(b.list) - 1; i >= 0; i-- {
		entry := b.list[i]
		if !cb(i, entry) {
			return
		}
	}
}

// Find scans the slice of BlockExecutionResults by ID
func (b *BlockExecutionResults) Find(id string) (*graph.BlockExecutionResult, bool) {
	var found *graph.BlockExecutionResult
	b.Each(func(_ int, entry *graph.BlockExecutionResult) bool {
		if entry.ID == id {
			found = entry
			return false
		}
		return true
	})
	return found, found != nil
}
