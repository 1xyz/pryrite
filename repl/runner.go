package repl

import (
	"github.com/aardlabs/terminal-poc/app"
	"os"
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

func (r *Runner) Execute(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if len(cmd) == 0 {
		return
	}

	if cmd == "quit" {
		os.Exit(0)
		return
	}

	args := strings.Split(cmd, " ")
	if args[0] == app.Name {
		args = args[1:]
	}
	if len(args) == 0 {
		return
	}

	c := r.ctx.NewRootCmd()
	c.SetArgs(args)
	// Don't handle the error (kobra does it for you)
	c.Execute()
}
