package run

import (
	"fmt"
	"sync"

	"github.com/1xyz/pryrite/graph"
)

type NodeIndex struct {
	sync.Map
}

func NewNodeViewIndex() *NodeIndex {
	return &NodeIndex{}
}

func (ni *NodeIndex) Add(n *graph.Node) error {
	id := n.ID
	_, loaded := ni.LoadOrStore(id, n)
	if loaded {
		return fmt.Errorf("an entry with id=%s exists", id)
	}
	return nil
}

func (ni *NodeIndex) ContainsID(id string) bool {
	_, found := ni.Load(id)
	return found
}

func (ni *NodeIndex) Get(id string) (*graph.Node, error) {
	e, found := ni.Load(id)
	if !found {
		return nil, fmt.Errorf("nodeIndex.Get(id=%s) not found", id)
	}
	return e.(*graph.Node), nil
}
