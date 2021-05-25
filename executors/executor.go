package executor

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

	// Content refers to the payload provided by the requester
	Content []byte

	// ContentType refers to the MIME type of the content
	ContentType ContentType

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

	// Exit status of the command
	ExitStatus int

	// Err is set to non-nil on error/cancellation
	Err error
}

type Executor interface {
	// Name returns the name of the executor
	Name() string

	// ContentTypes returns the content types supported by the Executor
	ContentTypes() []ContentType

	// Execute requests the Executor to execute the provided request
	Execute(context.Context, *ExecRequest) *ExecResponse
}
