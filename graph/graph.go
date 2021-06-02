package graph

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/tools"
	"io"
	"strings"
	"time"
)

type Metadata struct {
	SourceTitle string `json:"source_title"`
	SourceURI   string `json:"source_uri"`
	Agent       string `json:"Agent"`
}

func NewMetadata(agent, version string) *Metadata {
	return &Metadata{
		Agent: fmt.Sprintf("%s:%s", agent, version),
	}
}

type Node struct {
	ID             string     `json:"id,omitempty"`
	CreatedAt      *time.Time `json:"created_at"`
	OccurredAt     *time.Time `json:"occurred_at,omitempty"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
	Kind           Kind       `json:"kind"`
	Metadata       Metadata   `json:"metadata"`
	Title          string     `json:"title,omitempty"`
	Markdown       string     `json:"markdown,omitempty"`
	View           string     `json:"view"`
	Blocks         []*Block   `json:"blocks,omitempty"`
	Snippets       []*Snippet `json:"snippets,omitempty"`
	Children       string     `json:"children,omitempty"`
	IsShared       bool       `json:"is_shared"`
	LastExecutedAt *time.Time `json:"last_executed_at"`
	LastExecutedBy string     `json:"last_executed_by"`
}

type Block struct {
	ID          string     `json:"id,omitempty"`
	CreatedAt   *time.Time `json:"created_at"`
	Content     string     `json:"content"`
	ContentType string     `json:"content_type"`
	MD5         string     `json:"md5"`
}

func (block *Block) IsCode() bool {
	return strings.HasPrefix(block.ContentType, "text/") &&
		!strings.HasSuffix(block.ContentType, "markdown")
}

type Snippet Block

func (n *Node) GetChildIDs() []string {
	ids := strings.Split(n.Children, ",")
	result := []string{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if len(id) > 0 {
			result = append(result, id)
		}
	}
	return result
}

func (n *Node) HasBlocks() bool {
	return n.Blocks != nil && len(n.Blocks) > 0
}

func (n Node) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("ID=%s kind=%v Title=%s MD=%s Children=%s",
		n.ID, n.Kind, n.Title, n.Markdown, n.Children))
	return sb.String()
}

func NewNode(kind Kind, title, description, content, contentType string, metadata Metadata) (*Node, error) {
	now := time.Now().UTC()
	snippet := &Snippet{
		CreatedAt:   &now,
		Content:     content,
		ContentType: contentType,
	}
	return &Node{
		CreatedAt:  &now,
		OccurredAt: &now,
		Kind:       kind,
		Title:      title,
		Markdown:   description,
		Snippets:   []*Snippet{snippet},
		Metadata:   metadata,
	}, nil
}

type NodeView struct {
	Node     *Node  `json:"node"`
	View     string `json:"view"`
	Children []*NodeView
}

func (n NodeView) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Node: %v\n", n.Node))
	if n.Children != nil && len(n.Children) > 0 {
		for i, c := range n.Children {
			sb.WriteString(fmt.Sprintf("Child[%d] = %v\n", i, c))
		}
	}
	return sb.String()
}

// NodeExecutionResult encapsulates a node's execution result
type NodeExecutionResult struct {
	ExecutionID string `json:"execution_id"`
	NodeID      string `json:"node_id"`
	RequestID   string `json:"request_id"`
	Stdout      []byte `json:"stdout"`
	Stderr      []byte `json:"stderr"`
	Err         error  `json:"err"`
	ExitStatus  int    `json:"exit_status"`

	StdoutWriter io.WriteCloser
	StderrWriter io.WriteCloser
}

func (n *NodeExecutionResult) Close() error {
	w := []io.Closer{n.StderrWriter, n.StdoutWriter}
	for _, c := range w {
		if c != nil {
			if err := c.Close(); err != nil {
				tools.Log.Err(err).Msgf("close writer %v", err)
			}
		}
	}
	return nil
}

func NewNodeExecutionResult(executionID, nodeID, requestID string) *NodeExecutionResult {
	res := &NodeExecutionResult{
		ExecutionID: executionID,
		NodeID:      nodeID,
		RequestID:   requestID,
		Err:         nil,
		ExitStatus:  0,
	}
	res.StdoutWriter = newByteWriter(func(bytes []byte) {
		res.Stdout = bytes
	})
	res.StderrWriter = newByteWriter(func(bytes []byte) {
		res.Stderr = bytes
	})
	return res
}

type bytesSetFn = func(bytes []byte)

type byteWriter struct {
	bytes []byte
	fn    bytesSetFn
}

func newByteWriter(fn bytesSetFn) io.WriteCloser {
	return &byteWriter{
		bytes: []byte{},
		fn:    fn,
	}
}

func (b *byteWriter) Write(p []byte) (int, error) {
	b.bytes = append(b.bytes, p...)
	return len(p), nil
}

func (b *byteWriter) Close() error {
	if b.fn != nil {
		b.fn(b.bytes)
	}
	return nil
}

// BlockExecutionResult encapsulates a node's execution result
type BlockExecutionResult struct {
	ExecutionID string `json:"execution_id"`
	NodeID      string `json:"node_id"`
	BlockID     string `json:"block_id"`
	RequestID   string `json:"request_id"`
	Err         error  `json:"err"`
	ExitStatus  int    `json:"exit_status"`
	Stdout      []byte `json:"stdout"`
	Stderr      []byte `json:"stderr"`

	StdoutWriter io.WriteCloser
	StderrWriter io.WriteCloser
}

func (b *BlockExecutionResult) Close() error {
	w := []io.Closer{b.StderrWriter, b.StdoutWriter}
	for _, c := range w {
		if c != nil {
			if err := c.Close(); err != nil {
				tools.Log.Err(err).Msgf("close writer %v", err)
			}
		}
	}
	return nil
}

func NewBlockExecutionResult(executionID, nodeID, blockID, requestID string) *BlockExecutionResult {
	res := &BlockExecutionResult{
		ExecutionID: executionID,
		NodeID:      nodeID,
		BlockID:     blockID,
		RequestID:   requestID,
		Err:         nil,
		ExitStatus:  0,
	}
	res.StdoutWriter = newByteWriter(func(bytes []byte) {
		res.Stdout = bytes
	})
	res.StderrWriter = newByteWriter(func(bytes []byte) {
		res.Stderr = bytes
	})
	return res
}
