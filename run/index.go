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
