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
		// populate the child nodes
		n.childNodes = make([]*graph.Node, 0)
		childIDs := n.GetChildIDs()
		for _, childID := range childIDs {
			// check if already found and store
			if cachedNode, found := exp.nodes[childID]; found {
				n.childNodes = append(n.childNodes, cachedNode.Node)
				continue
			}

			// request the download the child from remote store
			childNode, err := exp.store.GetNode(childID)
			if err != nil {
				return nil, err
			}
			exp.nodes[childID] = &nodeW{Node: childNode, childNodes: nil}
			n.childNodes = append(n.childNodes, childNode)
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
