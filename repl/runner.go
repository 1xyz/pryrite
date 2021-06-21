package repl

import "github.com/c-bata/go-prompt"

type runner struct {
	e *prompt.Executor
}

func (r *runner) Run(cmd string) {

}

func newRunner() (*runner, error) {
	return &runner{}, nil
}
