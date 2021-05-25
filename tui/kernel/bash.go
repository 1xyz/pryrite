package kernel

import (
	"context"
	"os/exec"
)

type BashKernel struct{}

func (b *BashKernel) Name() string                { return "bash-kernel" }
func (b *BashKernel) ContentTypes() []ContentType { return []ContentType{Bash, Shell} }
func (b *BashKernel) Execute(ctx context.Context, req *ExecRequest) *ExecResponse {
	c := make(chan error, 1)
	cmd := exec.CommandContext(ctx, "bash", "-c", string(req.Content))
	if req.Stdin != nil {
		cmd.Stdin = req.Stdin
	}
	if req.Stdout != nil {
		cmd.Stdout = req.Stdout
	}
	if req.Stderr != nil {
		cmd.Stderr = req.Stderr
	}

	go func() {
		c <- cmd.Run()
		close(c)
	}()

	responseHdr := &ResponseHdr{req.Hdr.ID}
	select {
	case <-ctx.Done():
		<-c // wait for cmd.Run to complete
		return &ExecResponse{Hdr: responseHdr, Content: nil, Err: ctx.Err()}
	case err := <-c:
		return &ExecResponse{Hdr: responseHdr, Content: []byte{}, Err: err}
	}
}
