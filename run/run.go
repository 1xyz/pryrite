package run

import (
	"container/list"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/uuid"
	"github.com/muesli/termenv"
	"github.com/pkg/errors"
	"go.uber.org/atomic"

	executor "github.com/1xyz/pryrite/executors"
	"github.com/1xyz/pryrite/graph"
	"github.com/1xyz/pryrite/graph/log"
	"github.com/1xyz/pryrite/snippet"
	"github.com/1xyz/pryrite/tools"
	"github.com/1xyz/pryrite/tools/queue"
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
	Root *graph.Node

	// NodeViews indexed by NodeID
	ViewIndex *NodeIndex

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

	blockReqCh      chan *BlockExecutionRequest
	blockCancelCh   chan *BlockCancelRequest
	logRecvCh       chan *log.ResultLogEntry
	executionDoneCh chan *log.ResultLogEntry
	stopCh          chan bool
	statusCh        chan *Status
	executionDoneFn ExecutionUpdateFn
	statusUpdateFn  StatusUpdateFn
}

// NewRun constructs a new run for the provided playbook for the
func NewRun(gCtx *snippet.Context, playbookIDOrURL string) (*Run, error) {
	store, err := gCtx.GetStore()
	if err != nil {
		return nil, err
	}

	id, err := store.ExtractID(playbookIDOrURL)
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

		blockReqCh:      make(chan *BlockExecutionRequest),
		blockCancelCh:   make(chan *BlockCancelRequest),
		logRecvCh:       make(chan *log.ResultLogEntry),
		executionDoneCh: make(chan *log.ResultLogEntry),
		stopCh:          make(chan bool),
		statusCh:        make(chan *Status),
	}

	spin := spinner.New(spinner.CharSets[70], 100*time.Millisecond)
	spin.Color("white")
	spin.Start()
	spin.Suffix = "Fetching content from service..."
	defer func() {
		spin.FinalMSG = "[Completed]"
		spin.Stop()
		time.Sleep(300 * time.Millisecond) // teeny wait to let them see the message
		termenv.ClearLine()
		fmt.Print("\r") // not sure why this is needed, but clearline leaves the cursor position untouched
	}()

	start := time.Now()
	if err := run.buildGraph(); err != nil {
		return nil, err
	}
	tools.TimeTrack(start, "run.buildGraph")

	return run, nil
}

func (r *Run) buildGraph() error {
	n, err := r.getNode(r.PlaybookID)
	if err != nil {
		return err
	}

	r.Root = n
	q := list.New()
	q.PushBack(n)

	var fetchDuration time.Duration
	s1 := time.Now()
	for q.Len() > 0 {

		e := q.Front()
		parent := e.Value.(*graph.Node)
		q.Remove(e)

		s3 := time.Now()
		if err := parent.LoadChildNodes(r.Store, false); err != nil {
			return err
		}
		tools.TimeTrack(s3, "r.Store.GetChildren"+parent.ID)
		fetchDuration += time.Since(s3)

		for i := range parent.ChildNodes {
			child := parent.ChildNodes[i]
			if err := r.ViewIndex.Add(child); err != nil {
				tools.Log.Err(err).Msgf("buildGraph: ViewIndex.Add(%s) failed", child.ID)
				continue
			}
			q.PushBack(parent.ChildNodes[i])
		}
	}
	tools.TimeTrack(s1, "queue stuff")
	tools.Log.Info().Msgf("FetchDuration for get children %v", fetchDuration)
	return nil
}

func (r *Run) getNode(id string) (*graph.Node, error) {
	// fetch the node
	view, err := r.Store.GetNode(id)
	if err != nil {
		return nil, fmt.Errorf("getNode: id = %s err = %v", id, err)
	}
	// check to see if this node is already found (not a tree, but can be a DAG)
	if err := r.ViewIndex.Add(view); err != nil {
		return nil, err
	}
	return view, nil
}

// GetBlock returns the specified block from the local index
func (r *Run) GetBlock(nodeID, blockID string) (*graph.Block, error) {
	n, err := r.ViewIndex.Get(nodeID)
	if err != nil {
		return nil, err
	}
	block, found := n.GetBlock(blockID)
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
		req, ok := item.(*BlockExecutionRequest)
		if !ok {
			tools.Log.Fatal().Msgf("reqDispatch: cast-error. item %T is not (*BlockExecutionRequest)", item)
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
	requests := map[string]*BlockExecutionRequest{}
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
			go func(blockReq *BlockExecutionRequest) {
				r.sendLog(blockReq, log.ExecStateQueued)
			}(req)

		case entry := <-r.executionDoneCh:
			_, ok := requests[entry.RequestID]
			r.statusInfof("ExecutionDone: request:%s found:%v completed processing",
				entry.RequestID, ok)
			delete(requests, entry.RequestID)
			go func() {
				if err := r.ReportExecutionInfo(entry); err != nil {
					tools.Log.Err(err).Msgf("ReportExecutionInfo:")
				}
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

func (r *Run) sendLog(req *BlockExecutionRequest, state log.ExecState) {
	go func() {
		entry := NewResultLogEntryFromRequest(req)
		entry.State = state
		r.logRecvCh <- entry
	}()
}

func (r *Run) CancelBlock(nodeID, requestID string) {
	if !r.isRunning.Load() {
		tools.Log.Warn().Msgf("CancelBlock: Run system is not started")
		return
	}
	r.blockCancelCh <- NewBlockCancelRequest(nodeID, requestID)
}

// ExecuteBlock executes the specified block in the context of this node
func (r *Run) ExecuteBlock(n *graph.Node, b *graph.Block, stdout, stderr io.Writer) (string, error) {
	if !r.isRunning.Load() {
		tools.Log.Warn().Msgf("ExecuteBlock: Run system is not started")
		return "", fmt.Errorf("run system is not started")
	}

	timeout := r.gCtx.ConfigEntry.ExecutionTimeout.GetDuration()
	if timeout == 0 {
		timeout = time.Hour * 48
	}

	req := NewBlockExecutionRequest(n, b, stdout, stderr, r.ID, r.gCtx.ConfigEntry.Email, timeout)
	tools.Log.Info().
		Str("nodeID", n.ID).
		Str("requestID", req.ID).
		Str("blockID", b.ID).
		Str("ContentTrimmed", tools.TrimLength(b.Content, 6)).
		Msg("ExecuteBlock")
	r.blockReqCh <- req
	return req.ID, nil
}

func (r *Run) executeBlock(req *BlockExecutionRequest) *log.ResultLogEntry {
	tools.Log.Info().Msgf("ExecuteBlock: req %v", req)
	execResult := NewResultLogEntryFromRequest(req)

	exec, err := r.Register.Get([]byte(req.Block.Content), req.Block.ContentType)
	if err != nil {
		execResult.State = log.ExecStateFailed
		execResult.SetError(errors.Wrap(err, "cannot execute"))
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

	tools.Log.Info().Msgf("executeBlock node:%s req-id:%s content:%s",
		req.Node.ID, req.ID, req.Block.Content)
	execReq := &executor.ExecRequest{
		Hdr:         &executor.RequestHdr{ID: req.ID, ExecutionID: req.ExecutionID, NodeID: req.Node.ID},
		Content:     []byte(req.Block.Content),
		ContentType: req.Block.ContentType,
		In:          os.Stdin,
		Stdout:      outWriter,
		Stderr:      errWriter,
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

func (r *Run) EditSnippet(nodeID string) (*graph.Node, error) {
	n, err := r.ViewIndex.Get(nodeID)
	if err != nil {
		tools.Log.Err(err).Msgf("EditSnippet: err = %v", err)
		return nil, err
	}

	// ToDo: support edit without saving
	if _, err := snippet.EditSnippetNode(r.gCtx, nodeID, true); err != nil {
		tools.Log.Err(err).Msgf("EditSnippet: snippet.EditSnippetNode: err = %v", err)
		return nil, err
	}

	updatedNode, err := r.Store.GetNode(nodeID)
	if err != nil {
		tools.Log.Err(err).Msgf("EditSnippet: store.GetNodeView (%s) err = %v", nodeID, err)
	}

	n = updatedNode
	return n, nil
}

func (r *Run) EditBlock(nodeID, blockID string, save bool) (*graph.Node, *graph.Block, error) {
	n, err := r.ViewIndex.Get(nodeID)
	if err != nil {
		tools.Log.Err(err).Msgf("EditBlock: err = %v", err)
		return nil, nil, err
	}

	block, found := n.GetBlock(blockID)
	if !found {
		return nil, nil, fmt.Errorf("block [%s] not found in node [%s]", blockID, nodeID)
	}

	newBlock, err := snippet.EditNodeBlock(r.gCtx, n, block, save)
	if err != nil {
		return nil, nil, err
	}

	return n, newBlock, nil
}

func (r *Run) ReportExecutionInfo(entry *log.ResultLogEntry) error {
	if entry.State != log.ExecStateCompleted && entry.State != log.ExecStateFailed {
		return fmt.Errorf("report only completed/failed executions")
	}

	n, err := r.ViewIndex.Get(entry.NodeID)
	if err != nil {
		return err
	}
	b, found := n.GetBlock(entry.BlockID)
	if !found {
		return fmt.Errorf("block %s not found in node %s", entry.BlockID, entry.NodeID)
	}

	execInfo := &snippet.ExecutionInfo{
		ExecutedAt:    entry.ExecutedAt,
		ExecutedBy:    entry.ExecutedBy,
		ExitStatus:    entry.ExitStatus,
		ExecutionInfo: entry.Err + entry.Stderr,
	}
	return snippet.UpdateNodeBlockExecution(r.gCtx, n, b, execInfo)
}

func (r Run) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("RootID = %s\n", r.PlaybookID))
	sb.WriteString(fmt.Sprintf("RootView = %v\n", r.Root))
	return sb.String()
}
