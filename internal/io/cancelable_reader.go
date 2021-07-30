package io

import (
	"context"
	"github.com/aardlabs/terminal-poc/tools"
	"io"
	"os/exec"
)

// CancelableReadCloser essentially reads from the provided reader
// so that it can be read by the Read(..) method.
//
// The read allows the caller to cancel (i.e via close) the reader without
// waiting for the read to be stuck in a blocking mode.
//
// see: _examples/ex_cancelable_reader.go for example usage
type CancelableReadCloser struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
	done   chan error
}

func NewCancelableReadCloser(ctx context.Context) *CancelableReadCloser {
	return &CancelableReadCloser{
		cmd:  exec.CommandContext(ctx, "cat"),
		done: make(chan error, 1),
	}
}

func (crc *CancelableReadCloser) Start(input io.Reader) error {
	var err error
	crc.cmd.Stdin = input
	crc.stdout, err = crc.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := crc.cmd.Start(); err != nil {
		return err
	}

	go func(c *exec.Cmd) {
		crc.done <- c.Wait()
	}(crc.cmd)

	return nil
}

func (crc *CancelableReadCloser) Read(buf []byte) (int, error) {
	return crc.stdout.Read(buf)
}

func (crc *CancelableReadCloser) Close() error {
	if err := crc.cmd.Process.Kill(); err != nil {
		tools.Log.Warn().Err(err).Msgf("CancelableReaderCloser: closer err = %v", err)
		// do not return the error; Process.Kill always returns an error
	}
	return <-crc.done
}
