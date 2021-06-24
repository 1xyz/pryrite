package executor

import (
	"context"
	"io"
	"os"

	"github.com/google/uuid"
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
	ContentType *ContentType

	// the executor publishes all side effects (Stdout, Stderr, debugging events etc.)
	Stdout io.WriteCloser
	Stderr io.WriteCloser
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
	Name() string

	// ContentType returns the content type supported by the Executor
	ContentType() *ContentType

	// Execute requests the Executor to execute the provided request
	Execute(context.Context, *ExecRequest) *ExecResponse

	// Internal cleanup function invoked when the Register is torn down
	Cleanup()
}

func DefaultRequest() *ExecRequest {
	return &ExecRequest{
		Hdr: &RequestHdr{
			ID: uuid.NewString(),
		},
		Stdout: &IgnoreCloseWriter{os.Stdout},
		Stderr: &IgnoreCloseWriter{os.Stderr},
	}
}

type IgnoreCloseWriter struct {
	Writer io.Writer
}

func (icw *IgnoreCloseWriter) Write(data []byte) (int, error) {
	return icw.Writer.Write(data)
}

func (icw *IgnoreCloseWriter) Close() error {
	return nil
}
