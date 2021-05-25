package tui

import (
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

type NodeIndex map[string]*graph.NodeView

type RunContext struct {
	ID         string          // Unique ID of the Execution
	PlaybookID string          // PlaybookID of the playbook snippet to run
	Root       *graph.NodeView // NodeView of the Root node of execution
	Index      NodeIndex       // Index of all the nodes in the RunContext
	Store      graph.Store
	Register   executor.Register
}

// BuildRunContext eagerly builds the RunContext
// ToDo: Consider a lazy build
func BuildRunContext(gCtx *snippet.Context, name string) (*RunContext, error) {
	id, err := snippet.GetID(name)
	if err != nil {
		return nil, err
	}

	store, err := snippet.NewStoreFromContext(gCtx)
	if err != nil {
		return nil, err
	}

	runCtx := &RunContext{
		ID:         uuid.New().String(),
		PlaybookID: id,
		Root:       nil,
		Index:      make(NodeIndex),
		Store:      store,
		Register:   executor.NewRegister(),
	}

	if err := runCtx.buildGraph(); err != nil {
		return nil, err
	}

	return runCtx, nil
}

func (r *RunContext) buildGraph() error {
	// fetch the node
	view, err := r.Store.GetNodeView(r.PlaybookID)
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

func (r *RunContext) getNodeView(id string) (*graph.NodeView, error) {
	// fetch the node
	view, err := r.Store.GetNodeView(id)
	if err != nil {
		return nil, fmt.Errorf("getNodeView: id = %s err = %v", id, err)
	}
	// check to see if this node is already found (not a tree, but can be a DAG)
	if err := r.Index.Add(view); err != nil {
		return nil, err
	}
	if view.Children == nil {
		view.Children = []*graph.NodeView{}
	}
	return view, nil
}

func (r RunContext) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("RootID = %s\n", r.PlaybookID))
	sb.WriteString(fmt.Sprintf("RootView = %v\n", r.Root))
	return sb.String()
}

func (r *RunContext) Execute(n *graph.Node, stdout, stderr io.Writer) (*graph.NodeExecutionResult, error) {
	contentType := executor.ContentType(n.ContentLanguage)
	if contentType == executor.Empty {
		// Let's default to shell
		contentType = executor.Shell
	}

	exec, err := r.Register.Get(contentType)
	if err != nil {
		return nil, err
	}

	tools.Log.Info().Msgf("execute node %v content-type %v", n.ID, contentType)
	req := &executor.ExecRequest{
		Hdr: &executor.RequestHdr{
			ID: uuid.NewString(),
		},
		Content: []byte(n.Content),
		Stdout:  stdout,
		Stderr:  stderr,
	}
	res := exec.Execute(context.Background(), req)
	return graph.NewNodeExecutionResult(res.Hdr.RequestID, n.ID, "..."), nil
}

func (ni NodeIndex) Add(view *graph.NodeView) error {
	id := view.Node.ID
	if ni.ContainsID(id) {
		return fmt.Errorf("an entry with id=%s exists", id)
	}
	ni[id] = view
	return nil
}

func (ni NodeIndex) ContainsID(id string) bool {
	_, found := ni[id]
	return found
}

func (ni NodeIndex) Get(id string) (*graph.NodeView, error) {
	e, found := ni[id]
	if !found {
		return nil, fmt.Errorf("nodeIndex.Get(id=%s) not found", id)
	}
	return e, nil
}
