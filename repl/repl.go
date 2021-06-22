package repl

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/app"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/kmd"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type Context struct {
	Cfg      *config.Config
	GraphCtx *snippet.Context
}

func (ctx *Context) NewRootCmd() *cobra.Command {
	return kmd.NewCmdRoot(ctx.Cfg)
}

type repl struct {
	ctx       *Context
	completer *Completer
	runner    *Runner
	prompt    *prompt.Prompt
}

func newRepl(cfg *config.Config) (*repl, error) {
	ctx := &Context{
		Cfg:      cfg,
		GraphCtx: snippet.NewContext(cfg, app.Version),
	}

	c, err := NewCompleter(ctx)
	if err != nil {
		return nil, err
	}

	r := NewRunner(ctx)
	fmt.Printf("%s %s (service: %s)\n", app.Name,
		ctx.GraphCtx.Metadata.Agent,
		ctx.GraphCtx.ConfigEntry.ServiceUrl)
	fmt.Println("Type `quit` or `Ctrl-D` to exit.")
	fmt.Println("Type `help <command>` to get help on a specific command")
	pt := prompt.New(
		r.Execute,
		c.Complete,
		prompt.OptionTitle(fmt.Sprintf("interactive %s", app.Name)),
		prompt.OptionPrefix(fmt.Sprintf("%s >>> ", app.Name)),
		prompt.OptionPrefixTextColor(prompt.White),
		prompt.OptionInputTextColor(prompt.Yellow),
		//prompt.OptionCompletionWordSeparator(completer.FilePathCompletionSeparator),
	)

	return &repl{
		ctx:       ctx,
		completer: c, runner: r, prompt: pt,
	}, nil
}

func Repl(cfg *config.Config) error {
	defer fmt.Println("Bye!")
	r, err := newRepl(cfg)
	if err != nil {
		return err
	}
	r.prompt.Run()
	return nil
}
