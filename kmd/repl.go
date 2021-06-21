package kmd

import (
	"github.com/aardlabs/terminal-poc/repl"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/spf13/cobra"
)

func NewCmdRepl(gCtx *snippet.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repl",
		Short: "Launch a REPL",
		Long: examplef(`
              {AppName} repl.
        `),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return repl.Repl(gCtx)
		},
	}

	return cmd
}
