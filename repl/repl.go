package repl

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/c-bata/go-prompt"
)

type repl struct {
	c *completor
	r *runner
	p *prompt.Prompt
}

func newRepl(gCtx *snippet.Context) (*repl, error) {
	c, err := newCompletor(gCtx)
	if err != nil {
		return nil, err
	}

	r, err := newRunner()
	if err != nil {
		return nil, err
	}

	fmt.Printf("kube-prompt %s (service-%s)\n", gCtx.Metadata.Agent, gCtx.ConfigEntry.ServiceUrl)
	fmt.Println("Please use `exit` or `Ctrl-D` to exit this program.")
	pt := prompt.New(
		r.Run,
		c.Complete,
		prompt.OptionTitle("kube-prompt: interactive kubernetes client"),
		prompt.OptionPrefix(">>> "),
		prompt.OptionInputTextColor(prompt.Yellow),
		//prompt.OptionCompletionWordSeparator(completer.FilePathCompletionSeparator),
	)

	return &repl{
		c: c, r: r, p: pt,
	}, nil
}

func Repl(gCtx *snippet.Context) error {
	defer fmt.Println("Bye!")
	r, err := newRepl(gCtx)
	if err != nil {
		return err
	}
	r.p.Run()
	return nil
}
