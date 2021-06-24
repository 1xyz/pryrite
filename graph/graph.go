package graph

import (
	"context"
	"fmt"
	"github.com/aardlabs/terminal-poc/graph/log"
	"io"
	"strings"
	"time"

	executor "github.com/aardlabs/terminal-poc/executors"
	"github.com/aardlabs/terminal-poc/tools"
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
	Children       string     `json:"children,omitempty"`
	IsShared       bool       `json:"is_shared"`
	LastExecutedAt *time.Time `json:"last_executed_at"`
	LastExecutedBy string     `json:"last_executed_by"`
}

func (n *Node) GetBlock(blockID string) (*Block, bool) {
	if n.Blocks == nil {
		return nil, false
	}
	for _, b := range n.Blocks {
		if b.ID == blockID {
			return b, true
		}
	}
	return nil, false
}

type Block struct {
	ID          string                `json:"id,omitempty"`
	CreatedAt   *time.Time            `json:"created_at"`
	Content     string                `json:"content"`
	ContentType *executor.ContentType `json:"content_type"`
	MD5         string                `json:"md5"`
}

func (block *Block) IsCode() bool {
	return block.ContentType != nil &&
		block.ContentType.Type == "text" && block.ContentType.Subtype != "markdown"
}

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

func (n *Node) HasChildren() bool {
	return len(n.Children) > 0
}

func (n Node) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("ID=%s kind=%v Title=%s MD=%s Children=%s",
		n.ID, n.Kind, n.Title, n.Markdown, n.Children))
	return sb.String()
}

func NewNode(kind Kind, content, contentType string, metadata Metadata) (*Node, error) {
	now := time.Now().UTC()
	var language string
	vals := strings.Split(contentType, "/")
	if len(vals) > 1 {
		language = vals[1]
	}
	markdown := fmt.Sprintf("```%s\n%s\n```\n", language, content)
	return &Node{
		CreatedAt:  &now,
		OccurredAt: &now,
		Kind:       kind,
		Markdown:   markdown,
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

type BlockExecutionRequest struct {
	// CancelFn - can be called by multiple go routines, and is idempotent
	CancelFn    context.CancelFunc
	Ctx         context.Context
	ID          string
	Node        *Node
	Block       *Block
	Stdout      io.Writer
	Stderr      io.Writer
	ExecutionID string
	ExecutedBy  string
}

func (b *BlockExecutionRequest) String() string {
	return fmt.Sprintf("BlockExecutionRequest ID:%s NodeID:%s BlockID:%s ExecutionID:%s ExecutedBy:%s",
		b.ID, b.Node.ID, b.Block.ID, b.ExecutionID, b.ExecutedBy)
}

func NewBlockExecutionRequest(n *Node, b *Block, stdout, stderr io.Writer,
	executionID, executedBy string, executionTimeout time.Duration) *BlockExecutionRequest {
	tools.Log.Info().Msgf("NewBlockExecutionRequest executionTimeout = %v", executionTimeout)
	ctx, cancelFn := context.WithTimeout(context.Background(), executionTimeout)

	req := &BlockExecutionRequest{
		Ctx:         ctx,
		CancelFn:    cancelFn,
		ID:          tools.RandAlphaNumericStr(8),
		Node:        n,
		Block:       b,
		Stdout:      stdout,
		Stderr:      stderr,
		ExecutionID: executionID,
		ExecutedBy:  executedBy,
	}

	return req
}

func NewResultLogEntryFromRequest(req *BlockExecutionRequest) *log.ResultLogEntry {
	res := log.NewResultLogEntry(
		req.ExecutionID,
		req.Node.ID,
		req.Block.ID,
		req.ID,
		req.ExecutedBy,
		req.Block.Content)
	return res
}

type BlockCancelRequest struct {
	NodeID    string
	RequestID string
}

func NewBlockCancelRequest(nodeID, requestID string) *BlockCancelRequest {
	return &BlockCancelRequest{
		NodeID:    nodeID,
		RequestID: requestID,
	}
}
