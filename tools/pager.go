package tools

import (
	"fmt"
	"golang.org/x/term"
	"io"
	"os"
	"os/exec"
	"syscall"
)

type Pager struct {
	buf    []byte
	w      io.WriteCloser
	stdout *os.File
	cmd    *exec.Cmd
}

func NewPager() (*Pager, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	pager := os.Getenv("PAGER")
	if pager == "" {
		return nil, fmt.Errorf("could not find $PAGER to be set")
	}
	execPath, err := exec.LookPath(pager)
	if err != nil {
		return nil, fmt.Errorf("could not locater %s %v", pager, err)
	}

	stdout := os.Stdout
	os.Stdout = w

	cmd := exec.Command(execPath)
	cmd.Stdin = r
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	return &Pager{
		buf:    []byte{},
		cmd:    cmd,
		w:      w,
		stdout: stdout,
	}, nil
}

func (p *Pager) Write(b []byte) (int, error) {
	p.buf = append(p.buf, b...)
	return len(b), nil
}

func (p *Pager) Close() error {
	if err := p.cmd.Start(); err != nil {
		return err
	}
	if _, err := p.w.Write(p.buf); err != nil {
		return err
	}
	if err := p.w.Close(); err != nil {
		return err
	}
	if err := p.cmd.Wait(); err != nil {
		return err
	}
	// restore stdout
	os.Stdout = p.stdout
	return nil
}

func OpenOutputWriter() (io.WriteCloser, error) {
	if !term.IsTerminal(syscall.Stdout) {
		return os.Stdout, nil
	}
	pw, err := NewPager()
	if err != nil {
		return nil, err
	}
	return pw, nil
}
