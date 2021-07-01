package repl

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/app"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/internal/completer"
	"github.com/aardlabs/terminal-poc/internal/history"
	"github.com/aardlabs/terminal-poc/kmd"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type Context struct {
	Cfg      *config.Config
	GraphCtx *snippet.Context
	history  history.History
}

func (ctx *Context) NewRootCmd() *cobra.Command {
	return kmd.NewCmdRoot(ctx.Cfg)
}

type repl struct {
	ctx       *Context
	completer *completer.CobraCommandCompleter
	runner    *Runner
	prompt    *prompt.Prompt
}

func newRepl(cfg *config.Config) (*repl, error) {
	h, err := history.New(fmt.Sprintf("%s/repl.json", history.HistoryDir))
	if err != nil {
		return nil, err
	}
	histEntries, err := h.GetAll()
	if err != nil {
		return nil, err
	}

	ctx := &Context{
		Cfg:      cfg,
		GraphCtx: snippet.NewContext(cfg, app.Version),
		history:  h,
	}

	rootCmd := ctx.NewRootCmd()
	c := completer.NewCobraCommandCompleter(rootCmd)
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
		// Note: this option does not work well with history (bug) -
		//prompt.OptionCompletionOnDown(),
		prompt.OptionHistory(histEntries),
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
