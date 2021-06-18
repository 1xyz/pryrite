package explorer

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/snippet"
)

type nodeW struct {
	*graph.Node
	childNodes []*graph.Node
}

type NodeExplorer struct {
	gCtx  *snippet.Context
	nodes map[string]*nodeW
	store graph.Store
}

func (exp *NodeExplorer) GetChildren(nodeID string) ([]*graph.Node, error) {
	n, found := exp.nodes[nodeID]
	if !found {
		return nil, fmt.Errorf("entry with node %s not found", nodeID)
	}

	if n.childNodes == nil {
		children, err := exp.store.GetChildren(n.ID)
		if err != nil {
			return nil, err
		}

		n.childNodes = make([]*graph.Node, 0)
		for i := range children {
			n.childNodes = append(n.childNodes, &children[i])
			if _, found := exp.nodes[children[i].ID]; !found {
				exp.nodes[children[i].ID] = &nodeW{Node: &children[i], childNodes: nil}
			}
		}
	}

	return n.childNodes, nil
}

func NewNodeExplorer(gCtx *snippet.Context, nodes []*graph.Node) (*NodeExplorer, error) {
	nodeMap := make(map[string]*nodeW)
	for _, n := range nodes {
		nodeMap[n.ID] = &nodeW{
			Node:       n,
			childNodes: nil,
		}
	}
	store, err := snippet.NewStoreFromContext(gCtx)
	if err != nil {
		return nil, err
	}
	return &NodeExplorer{
		gCtx:  gCtx,
		nodes: nodeMap,
		store: store,
	}, nil
}
