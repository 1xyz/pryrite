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
	ContentType ContentType

	// Stdin messages are unique in that the request comes from the kernel, and the reply from the caller.
	// The caller is not required to support this, but if it does not, it must set 'allow_stdin' : False
	// in its execute requests. In this case, the kernel may not send Stdin requests. If that field is true,
	// the kernel may send Stdin requests and block waiting for a reply, so the frontend must answer.
	Stdin io.Reader

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
	// Name returns the name of the executor
	Name() string

	// ContentTypes returns the content types supported by the Executor
	ContentTypes() []ContentType

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
