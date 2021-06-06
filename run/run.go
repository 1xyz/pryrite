package run

import (
	"bytes"
	"container/list"
	"context"
	"fmt"
	executor "github.com/aardlabs/terminal-poc/executors"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/google/uuid"
	"io"
	"strings"
)

// Run encapsulates a playbook's execution
type Run struct {
	// snippet context
	gCtx *snippet.Context

	// Unique ID for this execution
	ID string

	// PlaybookID of the playbook snippet to run
	PlaybookID string

	// NodeView of the Root node of execution
	Root *graph.NodeView

	// NodeViews indexed by NodeID
	ViewIndex NodeViewIndex

	// NodeExecutionResult indexed by NodeID
	ExecIndex NodeExecResultIndex

	// Store refers to the graph store
	Store graph.Store

	// Register is the execution library
	Register executor.Register
}

// NewRun constructs a new run for the provided playbook for the
func NewRun(gCtx *snippet.Context, playbookIDOrURL string) (*Run, error) {
	id, err := snippet.GetID(playbookIDOrURL)
	if err != nil {
		return nil, err
	}

	store, err := snippet.NewStoreFromContext(gCtx)
	if err != nil {
		return nil, err
	}

	register, err := executor.NewRegister()
	if err != nil {
		return nil, err
	}

	run := &Run{
		gCtx:       gCtx,
		ID:         uuid.New().String(),
		PlaybookID: id,
		Root:       nil,
		ViewIndex:  make(NodeViewIndex),
		ExecIndex:  make(NodeExecResultIndex),
		Store:      store,
		Register:   register,
	}

	if err := run.buildGraph(); err != nil {
		return nil, err
	}

	return run, nil
}

func (r *Run) buildGraph() error {
	// fetch the node
	view, err := r.getNodeView(r.PlaybookID)
	if err != nil {
		return err
	}
	r.Root = view

	q := list.New()
	q.PushBack(view)

	for q.Len() > 0 {
		e := q.Front()
		parentView := e.Value.(*graph.NodeView)
		q.Remove(e)

		childIDs := parentView.Node.GetChildIDs()
		for _, childID := range childIDs {
			childView, err := r.getNodeView(childID)
			if err != nil {
				return err
			}
			parentView.Children = append(parentView.Children, childView)
			q.PushBack(childView)
		}
	}

	return nil
}

func (r *Run) getNodeView(id string) (*graph.NodeView, error) {
	// fetch the node
	view, err := r.Store.GetNodeView(id)
	if err != nil {
		return nil, fmt.Errorf("getNodeView: id = %s err = %v", id, err)
	}
	// check to see if this node is already found (not a tree, but can be a DAG)
	if err := r.ViewIndex.Add(view); err != nil {
		return nil, err
	}
	if view.Children == nil {
		view.Children = []*graph.NodeView{}
	}
	return view, nil
}

// GetBlock returns the specified block from the local index
func (r *Run) GetBlock(nodeID, blockID string) (*graph.Block, error) {
	view, err := r.ViewIndex.Get(nodeID)
	if err != nil {
		return nil, err
	}
	block, found := view.Node.GetBlock(blockID)
	if !found {
		return nil, fmt.Errorf("block [%s] not found in node[%s]", blockID, nodeID)
	}
	return block, nil
}

// GetBlockExecutionResult retrieves the execution result for the specified requestID within that node
func (r *Run) GetBlockExecutionResult(nodeID, requestID string) (*graph.BlockExecutionResult, error) {
	execResults, found := r.ExecIndex.Get(nodeID)
	if !found {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}
	result, found := execResults.Find(requestID)
	if !found {
		return nil, fmt.Errorf("request=%s not found in node=%s", requestID, nodeID)
	}
	return result, nil
}

// ExecuteNode executes all code blocks in the context of this node
// If any block fails executions, ExecuteNode will return an error and not continue executing
func (r *Run) ExecuteNode(n *graph.Node, stdout, stderr io.Writer) error {
	for _, b := range n.Blocks {
		if !b.IsCode() {
			continue
		}
		if _, err := r.ExecuteBlock(n, b, stdout, stderr); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteBlock executes the specified block in the context of this node
// Return value:
// On an execution
//		a non nil BlockExecutionResult which encapsulates the full execution result is returned
//		a nil error is returned
// If an execution did not happen for some reason them
//		a nil BlockExecutionResult is returned
//  	a non nil  error value is returned
func (r *Run) ExecuteBlock(n *graph.Node, b *graph.Block, stdout, stderr io.Writer) (*graph.BlockExecutionResult, error) {
	contentType := executor.ContentType(b.ContentType)
	if contentType == executor.Empty {
		return nil, fmt.Errorf("cannot execute. No contentType specified")
	}

	exec, err := r.Register.Get(contentType)
	if err != nil {
		return nil, err
	}

	executedBy := r.gCtx.ConfigEntry.Email
	executionID := r.ID
	nodeID := n.ID
	blockID := b.ID
	reqID := tools.RandAlphaNumericStr(8)
	tools.Log.Info().Msgf("ExecuteBlock: execn-id; %s node:%s block:%s content-type:%v req-id:%s",
		executionID, nodeID, blockID, contentType, reqID)

	// Define a new execution result
	execResult := graph.NewBlockExecutionResult(executionID, nodeID, blockID, reqID, executedBy, b.Content)
	defer func() {
		// This can take a while for rendering large captures…
		if _, err := io.Copy(stdout, bytes.NewBuffer(execResult.Stdout)); err != nil {
			tools.Log.Err(err).Msg("Execute: io.Copy(stdout..))")
		}
		if _, err := io.Copy(stderr, bytes.NewBuffer(execResult.Stderr)); err != nil {
			tools.Log.Err(err).Msg("Execute: io.Copy(stderr..))")
		}
		if err := execResult.Close(); err != nil {
			tools.Log.Err(err).Msg("Execute: execResult.close()")
		}
	}()

	// Write to both the real TUI passed in (with buffering to avoid delays) and
	// a capture writer in the result.
	outWriter := tools.NewBufferedWriteCloser(io.MultiWriter(execResult.StdoutWriter, stdout))
	errWriter := tools.NewBufferedWriteCloser(io.MultiWriter(execResult.StderrWriter, stderr))

	startMarker := fmt.Sprintf("\n[yellow]>> executing node:%s[white]\n", nodeID)
	outWriter.Write([]byte(startMarker))
	cmdInfo := fmt.Sprintf("[yellow]>> %s[white]\n", b.Content)
	outWriter.Write([]byte(cmdInfo))
	req := &executor.ExecRequest{
		Hdr:     &executor.RequestHdr{ID: reqID, ExecutionID: executionID, NodeID: nodeID},
		Content: []byte(b.Content),
		Stdout:  outWriter,
		Stderr:  errWriter,
	}

	// Call the underlying executor
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	res := exec.Execute(ctx, req)

	n.LastExecutedAt = execResult.ExecutedAt
	n.LastExecutedBy = execResult.ExecutedBy

	// ToDo: update this to reflect lastexecuted at BlockLevel
	err = r.Store.UpdateNode(&graph.Node{ID: n.ID, LastExecutedAt: n.LastExecutedAt})
	if err != nil {
		tools.Log.Err(err).Msgf("ExecuteBlock: r.Store.UpdateNode: failed to record run with service")
	}

	// Update the execution result and add it to the index
	execResult.ExitStatus = res.ExitStatus
	execResult.Err = res.Err
	r.ExecIndex.Append(execResult)
	return execResult, nil
}

func (r *Run) EditSnippet(nodeID string) (*graph.NodeView, error) {
	view, err := r.ViewIndex.Get(nodeID)
	if err != nil {
		tools.Log.Err(err).Msgf("EditSnippet: err = %v", err)
		return nil, err
	}

	// ToDo: support edit without saving
	if _, err := snippet.EditSnippetNode(r.gCtx, nodeID, true); err != nil {
		tools.Log.Err(err).Msgf("EditSnippet: snippet.EditSnippetNode: err = %v", err)
		return nil, err
	}

	updatedView, err := r.Store.GetNodeView(nodeID)
	if err != nil {
		tools.Log.Err(err).Msgf("EditSnippet: store.GetNodeView (%s) err = %v", nodeID, err)
	}

	view.Node = updatedView.Node
	view.View = updatedView.View
	return view, nil
}

func (r *Run) EditBlock(nodeID, blockID string, save bool) (*graph.NodeView, *graph.Block, error) {
	view, err := r.ViewIndex.Get(nodeID)
	if err != nil {
		tools.Log.Err(err).Msgf("EditBlock: err = %v", err)
		return nil, nil, err
	}

	block, found := view.Node.GetBlock(blockID)
	if !found {
		return nil, nil, fmt.Errorf("block [%s] not found in node [%s]", blockID, nodeID)
	}

	newBlock, err := snippet.EditNodeBlock(r.gCtx, view.Node, block, save)
	if err != nil {
		return nil, nil, err
	}

	return view, newBlock, nil
}

func (r Run) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("RootID = %s\n", r.PlaybookID))
	sb.WriteString(fmt.Sprintf("RootView = %v\n", r.Root))
	return sb.String()
}
