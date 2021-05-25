package kernel

import (
	"context"
	"io"
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

type ExecRequest struct {
	Hdr *RequestHdr

	// Action refers to the intended action, this can be empty "" (default)
	// the action can be used to support, where there are multiple request actions
	// supported by the Kernel
	Action string

	// Content refers to the payload provided by the requester
	Content []byte

	// ContentType refers to the type content
	ContentType ContentType

	// Stdin messages are unique in that the request comes from the kernel, and the reply from the caller.
	// The caller is not required to support this, but if it does not, it must set 'allow_stdin' : False
	// in its execute requests. In this case, the kernel may not send Stdin requests. If that field is true,
	// the kernel may send Stdin requests and block waiting for a reply, so the frontend must answer.
	Stdin io.Reader

	// the kernel publishes all side effects (Stdout, Stderr, debugging events etc.)
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

type Kernel interface {
	// Name returns the name of the kernel
	Name() string

	// ContentTypes returns the content types supported by the Kernel
	ContentTypes() []ContentType

	// Execute requests the Kernel to execute the provided request
	Execute(context.Context, *ExecRequest) *ExecResponse
}
