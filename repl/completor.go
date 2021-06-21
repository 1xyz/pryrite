package repl

import (
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/c-bata/go-prompt"
)

func newCompletor(gCtx *snippet.Context) (*completor, error) {
	return &completor{
		gCtx: gCtx,
	}, nil
}

type completor struct {
	gCtx *snippet.Context
}

func (c *completor) Complete(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "users", Description: "Store the username and age"},
		{Text: "articles", Description: "Store the article text posted by user"},
		{Text: "comments", Description: "Store the text commented to articles"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
