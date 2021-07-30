package executor

import (
	"errors"
	"io"
	"reflect"

	"github.com/aardlabs/terminal-poc/tools"
)

type CommandFeeder struct {
	writer io.WriteCloser
	input  chan []byte
}

func NewCommandFeeder(inputWriter io.WriteCloser) *CommandFeeder {
	// yeah, apparently Go doesn't let you check nil for interfaces: https://play.golang.org/p/I8GNCS6sBUb
	if reflect.ValueOf(inputWriter).IsNil() {
		return &CommandFeeder{
			input: make(chan []byte, 100),
		}
	}

	return &CommandFeeder{writer: inputWriter}
}

func (cf *CommandFeeder) Write(p []byte) (int, error) {
	if cf.writer == nil {
		cf.input <- p
		return len(p), nil
	}

	return cf.writer.Write(p)
}

func (cf *CommandFeeder) Put(content []byte) {
	if content != nil && content[len(content)-1] != '\n' {
		content = append(content, '\n')
	}
	tools.Trace("exec", "put", len(content), string(content))
	if cf.writer == nil {
		cf.input <- content
	} else {
		if len(content) < 1 {
			content = []byte{4} // EOT
		}
		_, err := cf.writer.Write(content)
		if err != nil {
			tools.Log.Err(err).Str("content", string(content)).Msg("PTY write failed")
		}
	}
}

func (cf *CommandFeeder) Read(buf []byte) (int, error) {
	if cf.input == nil {
		return 0, errors.New("read was invoked with PTYs enabled")
	}

	b := <-cf.input
	if b == nil {
		return 0, io.EOF
	}

	return copy(buf, b), nil
}

func (cf *CommandFeeder) Close() error {
	if cf.input == nil {
		cf.Put(nil) // send EOT and let things shutdown smoothly
		return nil
	}

	cf.input <- nil
	return nil
}
