package markdown

import (
	"github.com/1xyz/pryrite/graph"
)
import "errors"

// fileStore implements the graph.Store interface. In essence it encapsulates
// a single node represented by the single markdown file. The markdown file
// itself can have multiple blocks (the markdown elements as such)
type fileStore struct {
	mdFile string
	Node   *graph.Node
}

func NewMDFileStore(id, mdFile string) (graph.Store, error) {
	node, err := CreateNodeFromMarkdownFile(id, mdFile)
	if err != nil {
		return nil, err
	}

	return &fileStore{
		mdFile: mdFile,
		Node:   node,
	}, nil
}

var UnsupportedErr = errors.New("not supported")
var NodeNotFoundErr = errors.New("not found")

func (f *fileStore) GetNodes(int, graph.Kind) ([]graph.Node, error) {
	return []graph.Node{*f.Node}, nil
}
func (f *fileStore) AddNode(*graph.Node) (*graph.Node, error)                 { return nil, UnsupportedErr }
func (f *fileStore) GetChildren(string) ([]graph.Node, error)                 { return []graph.Node{}, nil }
func (f *fileStore) UpdateNodeBlockExecution(*graph.Node, *graph.Block) error { return UnsupportedErr }
func (f *fileStore) UpdateNodeBlock(*graph.Node, *graph.Block) error          { return UnsupportedErr }
func (f *fileStore) UpdateNode(*graph.Node) error                             { return nil }

func (f *fileStore) SearchNodes(string, int, graph.Kind) ([]graph.Node, error) {
	return nil, UnsupportedErr
}

func (f *fileStore) GetNode(id string) (*graph.Node, error) {
	if id != f.Node.ID {
		return nil, NodeNotFoundErr
	}
	return f.Node, nil
}

func (f *fileStore) ExtractID(input string) (string, error) {
	return ExtractIDFromFilePath(input)
}
