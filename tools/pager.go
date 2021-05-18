package tools

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

type pagerWriter struct {
	buf                    []byte
	cmd                    *exec.Cmd
	supportRawControlChars bool
}

const defaultPager = "less"

// OutputWriteCloser is the interface that groups the basic Write and Close methods.
type OutputWriteCloser interface {
	io.Writer
	io.Closer

	// SupportRawControlChars indicates if the writer allows
	// writing raw control characters
	SupportRawControlChars() bool
}

func NewPager() (*pagerWriter, error) {
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = defaultPager
	}
	execPath, err := exec.LookPath(pager)
	if err != nil {
		return nil, fmt.Errorf("could not locater %s %v", pager, err)
	}

	var cmd *exec.Cmd = nil
	if pager == defaultPager {
		// less supports -r (use that arg)
		cmd = exec.Command(execPath, "-r")
	} else {
		cmd = exec.Command(execPath)
	}

	return &pagerWriter{
		buf:                    []byte{},
		cmd:                    cmd,
		supportRawControlChars: pager == defaultPager,
	}, nil
}

func (p *pagerWriter) Write(b []byte) (int, error) {
	p.buf = append(p.buf, b...)
	return len(b), nil
}

func (p *pagerWriter) Close() error {
	p.cmd.Stdin = bytes.NewReader(p.buf)
	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr
	if err := p.cmd.Start(); err != nil {
		return err
	}
	Log.Info().Msgf("waiting for %v to quit. Process=%v", p.cmd, p.cmd.Process)
	if err := p.cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func (p *pagerWriter) SupportRawControlChars() bool { return p.supportRawControlChars }

type outputWriter struct{}

func (w *outputWriter) SupportRawControlChars() bool { return false }
func (w *outputWriter) Write(p []byte) (int, error)  { return os.Stdout.Write(p) }
func (w *outputWriter) Close() error                 { return nil }

func OpenOutputWriter() (OutputWriteCloser, error) {
	termEnabled := IsTermEnabled(syscall.Stdout)
	Log.Info().Msgf("IsTermEnabled(stdout) = %b", termEnabled)
	if !termEnabled {
		return &outputWriter{}, nil
	}
	pw, err := NewPager()
	if err != nil {
		return &outputWriter{}, nil
	}
	return pw, nil
}
