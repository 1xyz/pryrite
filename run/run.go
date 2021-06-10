package run

import (
	"container/list"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	executor "github.com/aardlabs/terminal-poc/executors"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/google/uuid"
)

type StatusUpdateFn func(*Status)
type ExecutionUpdateFn func(*graph.BlockExecutionResult)

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
	ViewIndex *NodeViewIndex

	// NodeExecutionResult indexed by NodeID
	ExecIndex *NodeExecResultIndex

	// Store refers to the graph store
	Store graph.Store

	// Register is the execution library
	Register *executor.Register

	blockReqCh      chan *graph.BlockExecutionRequest
	blockCancelCh   chan *graph.BlockCancelRequest
	logRecvCh       chan *graph.BlockExecutionResult
	stopCh          chan bool
	statusCh        chan *Status
	executionDoneFn ExecutionUpdateFn
	statusUpdateFn  StatusUpdateFn
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
		ViewIndex:  NewNodeViewIndex(),
		ExecIndex:  NewNodeExecResultIndex(),
		Store:      store,
		Register:   register,

		blockReqCh:    make(chan *graph.BlockExecutionRequest),
		blockCancelCh: make(chan *graph.BlockCancelRequest),
		logRecvCh:     make(chan *graph.BlockExecutionResult),
		stopCh:        make(chan bool),
		statusCh:      make(chan *Status),
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

// GetBlockExecutionResult retrieves the execution result for the specified LogID within that node
func (r *Run) GetBlockExecutionResult(nodeID, logID string) (*graph.BlockExecutionResult, error) {
	execResults, found := r.ExecIndex.Get(nodeID)
	if !found {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}
	result, found := execResults.Find(logID)
	if !found {
		return nil, fmt.Errorf("LogID=%s not found in node=%s", logID, nodeID)
	}
	return result, nil
}

// ExecuteNode executes all code blocks in the context of this node
// If any block fails executions, ExecuteNode will return an error and not continue executing
func (r *Run) ExecuteNode(n *graph.Node, stdout, stderr io.Writer) error {
	//for _, b := range n.Blocks {
	//	if !b.IsCode() {
	//		continue
	//	}
	//	r.ExecuteBlock(context.Background(), n, b, stdout, stderr)
	//}
	return nil
}

func (r *Run) Start() {
	requests := map[string]*graph.BlockExecutionRequest{}
	executionDoneCh := make(chan *graph.BlockExecutionResult)
	for {
		select {
		case req := <-r.blockReqCh:
			requests[req.ID] = req
			r.statusInfof("NewRequest: Recd. new request:%v", req.ID)
			go func(blockReq *graph.BlockExecutionRequest) {
				r.sendLog(req, graph.BlockStateStarted)
				result := r.executeBlock(blockReq)
				executionDoneCh <- result
			}(req)

		case entry := <-executionDoneCh:
			_, ok := requests[entry.RequestID]
			r.statusInfof("ExecutionDone: request:%s found:%v completed processing", entry.RequestID, ok)
			delete(requests, entry.RequestID)
			go func() {
				r.logRecvCh <- entry
			}()

		case logEntry := <-r.logRecvCh:
			go func() {
				r.ExecIndex.Append(logEntry)
				if r.executionDoneFn != nil {
					go r.executionDoneFn(logEntry)
				}
			}()

		case cancelReq := <-r.blockCancelCh:
			req, found := requests[cancelReq.RequestID]
			if !found {
				r.statusErrf("CancelRequest: cannot cancel req:%v", cancelReq.RequestID)
			} else {
				go func() {
					r.statusInfof("CancelRequest: req:%s received", cancelReq.RequestID)
					r.sendLog(req, graph.BlockStateCanceled)
					req.CancelFn()
				}()
			}

		case s := <-r.statusCh:
			go func() {
				if r.statusUpdateFn != nil {
					r.statusUpdateFn(s)
				}
			}()

		case <-r.stopCh:
			r.statusInfof("Stop: request received")
			// Cancel existing contexts
			close(executionDoneCh)
			return
		}
	}
}

func (r *Run) SetStatusUpdateFn(fn StatusUpdateFn)       { r.statusUpdateFn = fn }
func (r *Run) SetExecutionUpdateFn(fn ExecutionUpdateFn) { r.executionDoneFn = fn }

func (r *Run) statusErrf(format string, v ...interface{})  { r.sendStatus(StatusError, format, v...) }
func (r *Run) statusInfof(format string, v ...interface{}) { r.sendStatus(StatusInfo, format, v...) }
func (r *Run) sendStatus(level StatusLevel, format string, v ...interface{}) {
	go func() {
		msg := fmt.Sprintf(format, v...)
		r.statusCh <- &Status{Level: level, Message: msg}
	}()
}

func (r *Run) sendLog(req *graph.BlockExecutionRequest, state graph.BlockState) {
	go func() {
		entry := graph.NewBlockExecutionResultFromRequest(req)
		entry.State = state
		r.logRecvCh <- entry
	}()
}

func (r *Run) CancelBlock(nodeID, requestID string) {
	r.blockCancelCh <- graph.NewBlockCancelRequest(nodeID, requestID)
}

// ExecuteBlock executes the specified block in the context of this node
func (r *Run) ExecuteBlock(n *graph.Node, b *graph.Block, stdout, stderr io.Writer) {
	timeout := r.gCtx.ConfigEntry.ExecutionTimeout.Duration()
	if timeout == 0 {
		timeout = time.Hour
	}

	r.blockReqCh <- graph.NewBlockExecutionRequest(n, b, stdout, stderr, r.ID, r.gCtx.ConfigEntry.Email, timeout)
}

func (r *Run) executeBlock(req *graph.BlockExecutionRequest) *graph.BlockExecutionResult {
	contentType := executor.ContentType(req.Block.ContentType)
	tools.Log.Info().Msgf("ExecuteBlock: req %v", req)
	execResult := graph.NewBlockExecutionResult(
		req.ExecutionID,
		req.Node.ID,
		req.Block.ID,
		req.ID,
		req.ExecutedBy,
		req.Block.Content)

	if contentType == executor.Empty {
		execResult.SetErr(fmt.Errorf("cannot execute. No contentType specified"))
		return execResult
	}

	exec, err := r.Register.Get(contentType)
	if err != nil {
		execResult.SetErr(fmt.Errorf("cannot execute. No contentType specified"))
		return execResult
	}

	defer func() {
		// This can take a while for rendering large captures…
		if err := execResult.Close(); err != nil {
			tools.Log.Err(err).Msg("Execute: execResult.close()")
		}
	}()

	// Write to both the real TUI passed in (with buffering to avoid delays) and
	// a capture writer in the result.
	outWriter := tools.NewBufferedWriteCloser(io.MultiWriter(execResult.StdoutWriter, req.Stdout))
	errWriter := tools.NewBufferedWriteCloser(io.MultiWriter(execResult.StderrWriter, req.Stderr))

	startMarker := fmt.Sprintf("\n[yellow]>> executing node:%s[white]\n", req.Node.ID)
	outWriter.Write([]byte(startMarker))
	cmdInfo := fmt.Sprintf("[yellow]>> %s[white]\n", req.Block.Content)
	outWriter.Write([]byte(cmdInfo))
	execReq := &executor.ExecRequest{
		Hdr:     &executor.RequestHdr{ID: req.ID, ExecutionID: req.ExecutionID, NodeID: req.Node.ID},
		Content: []byte(req.Block.Content),
		Stdout:  outWriter,
		Stderr:  errWriter,
	}

	res := exec.Execute(req.Ctx, execReq)

	execResult.ExitStatus = strconv.Itoa(res.ExitStatus)
	execResult.SetErr(res.Err)
	if res.ExitStatus != 0 || res.Err != nil {
		execResult.State = graph.BlockStateFailed
	} else {
		execResult.State = graph.BlockStateCompleted
	}
	return execResult
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
