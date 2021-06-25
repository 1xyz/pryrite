package executor

import (
	"io"
)

type CommandFeeder struct {
	input chan []byte
}

func NewCommandFeeder() *CommandFeeder {
	return &CommandFeeder{
		input: make(chan []byte, 100),
	}
}

func (cf *CommandFeeder) Put(content []byte) {
	if content != nil && content[len(content)-1] != '\n' {
		content = append(content, '\n')
	}
	cf.input <- content
}

func (cf *CommandFeeder) Read(buf []byte) (int, error) {
	b := <-cf.input
	if b == nil {
		return 0, io.EOF
	}

	return copy(buf, b), nil
}

func (cf *CommandFeeder) Close() error {
	cf.input <- nil
	return nil
}
