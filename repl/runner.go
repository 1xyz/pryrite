package repl

import (
	"strings"
)

type Runner struct {
	ctx *Context
}

func NewRunner(ctx *Context) *Runner {
	return &Runner{
		ctx: ctx,
	}
}

func (r *Runner) Run(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if len(cmd) == 0 {
		return
	}
	args := strings.Split(cmd, " ")
	c := r.ctx.NewRootCmd()
	c.SetArgs(args)
	// Don't handle the error (kobra does it for you)
	c.Execute()
}
