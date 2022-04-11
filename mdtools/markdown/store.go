package markdown

import (
	"fmt"
	"github.com/1xyz/pryrite/graph"
	"net/url"
	"strings"
)
import "errors"

// fileStore implements the graph.Store interface. In essence it encapsulates
// a single node represented by the single markdown file. The markdown file
// itself can have multiple blocks (the markdown elements as such)
type fileStore struct {
	mdFile string
	Node   *graph.Node
}

func NewMDFileStore(mdFile string) (graph.Store, error) {
	node, err := CreateNodeFromMarkdownFile(mdFile)
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
	idOrURL := strings.TrimSpace(input)
	u, err := url.Parse(idOrURL)
	if err != nil {
		return "", err
	}
	tokens := strings.Split(u.Path, "/")
	if len(tokens) == 0 || len(idOrURL) == 0 {
		return "", fmt.Errorf("empty id")
	}
	return tokens[len(tokens)-1], nil
}
