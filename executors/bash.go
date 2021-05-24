package executor

import (
	"context"
	"errors"
	"os/exec"
)

type BashExecutor struct{}

func (b *BashExecutor) Name() string { return "bash-executor" }

func (b *BashExecutor) ContentTypes() []ContentType { return []ContentType{Bash, Shell} }

func (b *BashExecutor) Execute(ctx context.Context, req *ExecRequest) *ExecResponse {
	c := make(chan error, 1)
	cmd := exec.CommandContext(ctx, "bash", "-c", string(req.Content))
	cmd.Stdin = req.Stdin
	cmd.Stdout = req.Stdout
	cmd.Stderr = req.Stderr

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

func (b *BashExecutor) ExecuteStart(ctx context.Context, req *ExecRequest) *ExecAsync {
	async := &ExecAsync{}

	if req.Stdin != nil {
		async.Err = errors.New("stdin may not be set when working with async execution")
		return async
	}

	async.cmd = exec.CommandContext(ctx, "bash", "-i")
	async.cmd.Stdout = req.Stdout
	async.cmd.Stderr = req.Stderr

	var err error
	async.stdin, err = async.cmd.StdinPipe()
	if err != nil {
		async.Err = err
		return async
	}

	err = async.cmd.Start()
	if err != nil {
		async.Err = err
		return (*ExecAsync)(async)
	}

	return async
}

func (b *BashExecutor) ExecuteFeed(async *ExecAsync, content []byte) {
	_, err := async.stdin.Write(content)
	if err != nil {
		async.Err = err
	}
}

func (b *BashExecutor) ExecuteStop(async *ExecAsync) error {
	return nil
}
