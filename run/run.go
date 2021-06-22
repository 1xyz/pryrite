package run

import (
	"container/list"
	"fmt"
	"github.com/aardlabs/terminal-poc/graph/log"
	"github.com/aardlabs/terminal-poc/tools/queue"
	"go.uber.org/atomic"
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
type ExecutionUpdateFn func(entry *log.ResultLogEntry)

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
	ExecIndex log.ResultLogIndex

	// Store refers to the graph store
	Store graph.Store

	// Register is the execution library
	Register *executor.Register

	// isRunning indicates if the Run can accept requests to execute
	isRunning *atomic.Bool

	// requestQ is where incoming blockExecutionRequest(s) are queued
	requestQ queue.Queue

	blockReqCh      chan *graph.BlockExecutionRequest
	blockCancelCh   chan *graph.BlockCancelRequest
	logRecvCh       chan *log.ResultLogEntry
	executionDoneCh chan *log.ResultLogEntry
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

	execIndex, err := log.NewResultLogIndex(log.IndexFileSystem)
	if err != nil {
		return nil, err
	}

	run := &Run{
		gCtx:       gCtx,
		ID:         uuid.New().String(),
		PlaybookID: id,
		Root:       nil,
		ViewIndex:  NewNodeViewIndex(),
		ExecIndex:  execIndex,
		Store:      store,
		Register:   register,
		isRunning:  atomic.NewBool(false),
		requestQ:   queue.NewConcurrentQueue(),

		blockReqCh:      make(chan *graph.BlockExecutionRequest),
		blockCancelCh:   make(chan *graph.BlockCancelRequest),
		logRecvCh:       make(chan *log.ResultLogEntry),
		executionDoneCh: make(chan *log.ResultLogEntry),
		stopCh:          make(chan bool),
		statusCh:        make(chan *Status),
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

		children, err := r.Store.GetChildren(parentView.Node.ID)
		if err != nil {
			return err
		}
		for i := range children {
			child := &children[i]
			childView := graph.NodeView{Node: child, View: child.View, Children: []*graph.NodeView{}}
			if err := r.ViewIndex.Add(&childView); err != nil {
				tools.Log.Err(err).Msgf("buildGraph: ViewIndex.Add(%s) failed", childView.Node.ID)
				continue
			}
			parentView.Children = append(parentView.Children, &childView)
			q.PushBack(&childView)
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
func (r *Run) GetBlockExecutionResult(nodeID, logID string) (*log.ResultLogEntry, error) {
	execResults, err := r.ExecIndex.Get(nodeID)
	if err != nil {
		return nil, err
	}
	result, err := execResults.Find(logID)
	if err != nil {
		return nil, fmt.Errorf("LogID=%s not found in node=%s err = %v", logID, nodeID, err)
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
		r.ExecuteBlock(n, b, stdout, stderr)
	}
	return nil
}

func (r *Run) reqDispatchLoop() {
	for {
		item := r.requestQ.WaitForItem()
		req, ok := item.(*graph.BlockExecutionRequest)
		if !ok {
			tools.Log.Fatal().Msgf("reqDispatch: cast-error. item %T is not (*graph.BlockExecutionRequest)", item)
		}

		if !r.isRunning.Load() {
			tools.Log.Info().Msgf("reqDispatch: system is shutdown")
			return
		}

		r.sendLog(req, log.ExecStateStarted)
		result := r.executeBlock(req)
		if r.isRunning.Load() {
			r.executionDoneCh <- result
		}
	}
}

// Start a loop to receive messages
func (r *Run) Start() {
	requests := map[string]*graph.BlockExecutionRequest{}
	if !r.isRunning.CAS(false, true) {
		tools.Log.Info().Msgf("System is already running")
		return
	}

	go r.reqDispatchLoop()

	for {
		select {
		case req := <-r.blockReqCh:
			requests[req.ID] = req
			r.statusInfof("NewRequest: Recd. new request:%v", req.ID)
			// ensure that the enqueue is done here to FIFO ordering
			// not in a separate go-routine
			r.requestQ.Enqueue(req)
			go func(blockReq *graph.BlockExecutionRequest) {
				r.sendLog(blockReq, log.ExecStateQueued)
			}(req)

		case entry := <-r.executionDoneCh:
			_, ok := requests[entry.RequestID]
			r.statusInfof("ExecutionDone: request:%s found:%v completed processing",
				entry.RequestID, ok)
			delete(requests, entry.RequestID)
			go func() {
				if r.isRunning.Load() {
					r.logRecvCh <- entry
				}
			}()

		case logEntry := <-r.logRecvCh:
			go func() {
				if err := r.ExecIndex.Append(logEntry); err != nil {
					tools.Log.Err(err).Msgf("logEntryRecv: ExecIndex.Append:")
					return
				}
				if r.executionDoneFn != nil {
					r.executionDoneFn(logEntry)
				}
			}()

		case cancelReq := <-r.blockCancelCh:
			req, found := requests[cancelReq.RequestID]
			if !found {
				r.statusErrf("CancelRequest: cannot cancel req:%v", cancelReq.RequestID)
			} else {
				go func() {
					r.statusInfof("CancelRequest: req:%s received", cancelReq.RequestID)
					r.sendLog(req, log.ExecStateCanceled)
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
			tools.Log.Info().Msgf("Stop: Request to shutdown received")
			r.isRunning.Store(false)
			close(r.statusCh)
			close(r.logRecvCh)
			close(r.blockCancelCh)
			close(r.blockReqCh)
			close(r.executionDoneCh)

			done := make(chan bool)
			go func() {
				for _, r := range requests {
					r.CancelFn()
				}
				done <- true
			}()

			<-done
			close(done)
			tools.Log.Info().Msgf("Stop: Request to shutdown complete")
			return
		}
	}
}

// Shutdown stops this system. You cannot use Start to start running
func (r *Run) Shutdown() {
	tools.Log.Info().Msgf("Shutdown: Request to shutdown received")
	if !r.isRunning.Load() {
		tools.Log.Info().Msgf("System is already stopped")
		return
	}
	r.Register.Cleanup()
	r.stopCh <- true
	close(r.stopCh)
}

func (r *Run) SetStatusUpdateFn(fn StatusUpdateFn)       { r.statusUpdateFn = fn }
func (r *Run) SetExecutionUpdateFn(fn ExecutionUpdateFn) { r.executionDoneFn = fn }

func (r *Run) statusErrf(format string, v ...interface{})  { r.sendStatus(StatusError, format, v...) }
func (r *Run) statusInfof(format string, v ...interface{}) { r.sendStatus(StatusInfo, format, v...) }
func (r *Run) sendStatus(level StatusLevel, format string, v ...interface{}) {
	go func() {
		msg := fmt.Sprintf(format, v...)
		if r.isRunning.Load() {
			r.statusCh <- &Status{Level: level, Message: msg}
		}
	}()
}

func (r *Run) sendLog(req *graph.BlockExecutionRequest, state log.ExecState) {
	go func() {
		entry := graph.NewResultLogEntryFromRequest(req)
		entry.State = state
		r.logRecvCh <- entry
	}()
}

func (r *Run) CancelBlock(nodeID, requestID string) {
	if !r.isRunning.Load() {
		tools.Log.Warn().Msgf("CancelBlock: Run system is not started")
		return
	}
	r.blockCancelCh <- graph.NewBlockCancelRequest(nodeID, requestID)
}

// ExecuteBlock executes the specified block in the context of this node
func (r *Run) ExecuteBlock(n *graph.Node, b *graph.Block, stdout, stderr io.Writer) {
	if !r.isRunning.Load() {
		tools.Log.Warn().Msgf("CancelBlock: Run system is not started")
		return
	}

	timeout := r.gCtx.ConfigEntry.ExecutionTimeout.Duration()
	if timeout == 0 {
		timeout = time.Hour
	}

	req := graph.NewBlockExecutionRequest(n, b, stdout, stderr, r.ID, r.gCtx.ConfigEntry.Email, timeout)
	tools.Log.Info().Msgf("ExecuteBlock: node:%s block:%s req:%s content:%s",
		n.ID, b.ID, req.ID, tools.TrimLength(b.Content, 6))
	r.blockReqCh <- req
}

func (r *Run) executeBlock(req *graph.BlockExecutionRequest) *log.ResultLogEntry {
	contentType := executor.ContentType(req.Block.ContentType)
	tools.Log.Info().Msgf("ExecuteBlock: req %v", req)
	execResult := graph.NewResultLogEntryFromRequest(req)

	if contentType == executor.Empty {
		execResult.State = log.ExecStateFailed
		execResult.SetError(fmt.Errorf("cannot execute. No contentType specified"))
		return execResult
	}

	exec, err := r.Register.Get(contentType)
	if err != nil {
		execResult.State = log.ExecStateFailed
		execResult.SetError(fmt.Errorf("cannot execute. No contentType specified"))
		return execResult
	}

	stdoutWriter := tools.NewBytesWriter()
	stderrWriter := tools.NewBytesWriter()
	defer func() {
		execResult.Stdout = stdoutWriter.GetString()
		execResult.Stderr = stderrWriter.GetString()
	}()

	// Write to both the real TUI passed in (with buffering to avoid delays) and
	// a capture writer in the result.
	outWriter := tools.NewBufferedWriteCloser(io.MultiWriter(stdoutWriter, req.Stdout))
	errWriter := tools.NewBufferedWriteCloser(io.MultiWriter(stderrWriter, req.Stderr))

	startMarker := fmt.Sprintf("\n[yellow]>> executing node:%s req-id:%s [white]\n", req.Node.ID, req.ID)
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

	req.Node.LastExecutedAt = execResult.ExecutedAt
	req.Node.LastExecutedBy = execResult.ExecutedBy
	// ToDo: update this to reflect lastexecuted at BlockLevel
	err = r.Store.UpdateNode(&graph.Node{ID: req.Node.ID, LastExecutedAt: req.Node.LastExecutedAt})
	if err != nil {
		tools.Log.Err(err).Msgf("ExecuteBlock: r.Store.UpdateNode: failed to record run with service")
	}

	execResult.ExitStatus = strconv.Itoa(res.ExitStatus)
	execResult.SetError(res.Err)
	if res.ExitStatus != 0 || res.Err != nil {
		execResult.State = log.ExecStateFailed
	} else {
		execResult.State = log.ExecStateCompleted
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
