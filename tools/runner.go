package tools

import (
	"os"
	"strings"

	"github.com/aardlabs/terminal-poc/app"

	"github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"
)

type RunnerContext interface {
	NewRootCmd() *cobra.Command
	HistoryAppend(string) error
}

type Runner struct {
	ctx RunnerContext
}

func NewRunner(ctx RunnerContext) *Runner {
	return &Runner{
		ctx: ctx,
	}
}

func (r *Runner) Execute(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if len(cmd) == 0 {
		return
	}

	if cmd == "q" || cmd == "quit" || cmd == "exit" {
		os.Exit(0)
		return
	}

	args, _ := shellwords.Parse(cmd)
	if args[0] == app.Name {
		args = args[1:]
	}
	if len(args) == 0 {
		return
	}

	c := r.ctx.NewRootCmd()
	c.SetArgs(args)
	// Don't handle the error (kobra does it for you)
	if err := c.Execute(); err == nil {
		// append to command history if there is no error
		if err := r.ctx.HistoryAppend(cmd); err != nil {
			Log.Err(err).Msgf("history.Append %s", cmd)
		}
	}
}
