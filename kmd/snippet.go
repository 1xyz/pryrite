package kmd

import (
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type SnippetListOpts struct {
	Limit  int
	Kind   int
	User   string
	IsMine bool
}

func NewCmdSnippetList() *cobra.Command {
	opts := &SnippetListOpts{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List snippets",
		Long: heredoc.Doc(`
              aard list, lists all snippets which are visible to the current
              logged-in user. This includes both a user's own snippets as well
              as snippet shared. 

              By default, only "command" snippets are listed. This can be changed
              by using the --kind=all flag.
        `),
		Example: heredoc.Doc(`
            To list the most recently created snippets, run: 
              $ aard list

            To list the most recent "n=100" snippets, run:
              $ aard list -n 100

            To list all kind of snippets that include non-command snippets, run:
              $ aard list --kind=all

            To filter the snippets listed to those created by specific user, run:
              $ aard list --user=alanturing

            To filter the snippets listed to those created by me, run:
              $ aard list --mine
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(args[0])
			return nil
		},
	}
	cmd.Flags().IntVarP(&opts.Limit, "limit", "n",
		8, "Limit the number of results to display")
	return cmd
}
