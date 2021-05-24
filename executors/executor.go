package executor

import (
	"context"
	"io"
	"os/exec"
)

type RequestHdr struct {
	// A unique ID, identifying request
	ID string
	// ExecutionID uniquely identifies a playbook execution
	ExecutionID string
	// NodeID identifies the node being execution
	NodeID string
	// UserID identifies the user executing the run
	UserID string
}

type ExecAction uint

const (
	Run     ExecAction = iota // Run a command and wait for more requests
	RunOnce                   // Prepends the content within the template tag to the container designated by the target dom id.
)

func (action ExecAction) String() string {
	return [...]string{"run", "run-once"}[action]
}

type ExecRequest struct {
	Hdr *RequestHdr

	// Action refers to the intended action;
	// the action can be used to support, where there are multiple request actions
	// supported by the Executor
	Action ExecAction

	// Content refers to the payload provided by the requester
	Content []byte

	// ContentType refers to the MIME type of the content
	ContentType ContentType

	// Stdin messages are unique in that the request comes from the executor, and the reply from the caller.
	// The caller is not required to support this, but if it does not, it must set 'allow_stdin' : False
	// in its execute requests. In this case, the executor may not send Stdin requests. If that field is true,
	// the executor may send Stdin requests and block waiting for a reply, so the frontend must answer.
	Stdin io.Reader

	// the executor publishes all side effects (Stdout, Stderr, debugging events etc.)
	Stdout io.Writer
	Stderr io.Writer
}

type ResponseHdr struct {
	// RequestID refers to request's unique ID
	RequestID string
}

type ExecResponse struct {
	Hdr *ResponseHdr

	// Content is the response content, if any.
	Content []byte

	// Err is set to non-nil on error/cancellation
	Err error
}

type ExecAsync struct {
	Responses chan *ExecResponse

	// Err is set to non-nil on error/cancellation
	Err error

	// The internal command for processing requests
	cmd *exec.Cmd

	stdin io.WriteCloser
}

type Executor interface {
	// Name returns the name of the executor
	Name() string

	// ContentTypes returns the content types supported by the Executor
	ContentTypes() []ContentType

	// Execute requests the Executor to execute the provided request
	Execute(context.Context, *ExecRequest) *ExecResponse

	ExecuteStart(context.Context, *ExecRequest) *ExecAsync

	ExecuteFeed(*ExecAsync, []byte)

	ExecuteStop(*ExecAsync) error
}
