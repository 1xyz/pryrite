package run

import (
	"bufio"
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

	// NodeExecutionResult indexed by Id
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

// Execute runs the specified node with the executor
func (r *Run) Execute(n *graph.Node, stdout, stderr io.Writer) (*graph.NodeExecutionResult, error) {
	// find a matching executor.
	contentType := executor.ContentType(n.ContentLanguage)
	if contentType == executor.Empty {
		contentType = executor.Shell
	}
	exec, err := r.Register.Get(contentType)
	if err != nil {
		return nil, err
	}

	tools.Log.Info().Msgf("execute node %v content-type %v", n.ID, contentType)
	nodeID := n.ID
	executionID := r.ID
	reqID := uuid.NewString()

	// Define a new execution result
	execResult := graph.NewNodeExecutionResult(executionID, nodeID, reqID)
	defer func() {
		if err := execResult.Close(); err != nil {
			tools.Log.Err(err).Msgf("Execute: execResult.close() err = %v", err)
		}
	}()

	// The execution stdout and stderr will be mirrored to the executionResult
	outWriter := bufio.NewWriter(io.MultiWriter(stdout, execResult.StdoutWriter))
	errWriter := bufio.NewWriter(io.MultiWriter(stderr, execResult.StderrWriter))
	req := &executor.ExecRequest{
		Hdr:     &executor.RequestHdr{ID: reqID, ExecutionID: executionID, NodeID: nodeID},
		Content: []byte(n.Content),
		Stdout:  outWriter,
		Stderr:  errWriter,
	}

	// Call the underlying executor
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	res := exec.Execute(ctx, req)

	// Update the execution result and add it to the index
	execResult.ExitStatus = res.ExitStatus
	execResult.Err = res.Err
	r.ExecIndex.Set(execResult)
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
	view.ContentMarkdown = updatedView.ContentMarkdown
	return view, nil
}

func (r Run) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("RootID = %s\n", r.PlaybookID))
	sb.WriteString(fmt.Sprintf("RootView = %v\n", r.Root))
	return sb.String()
}
