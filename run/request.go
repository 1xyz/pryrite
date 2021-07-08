package run

import (
	"context"
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/graph/log"
	"github.com/aardlabs/terminal-poc/tools"
	"io"
	"time"
)

type BlockExecutionRequest struct {
	// CancelFn - can be called by multiple go routines, and is idempotent
	CancelFn    context.CancelFunc
	Ctx         context.Context
	ID          string
	Node        *graph.Node
	Block       *graph.Block
	Stdout      io.Writer
	Stderr      io.Writer
	ExecutionID string
	ExecutedBy  string
}

func (b *BlockExecutionRequest) String() string {
	return fmt.Sprintf("BlockExecutionRequest ID:%s NodeID:%s BlockID:%s ExecutionID:%s ExecutedBy:%s",
		b.ID, b.Node.ID, b.Block.ID, b.ExecutionID, b.ExecutedBy)
}

func NewBlockExecutionRequest(n *graph.Node, b *graph.Block, stdout, stderr io.Writer,
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
