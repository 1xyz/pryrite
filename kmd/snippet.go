package kmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/spf13/cobra"
)

type SnippetListOpts struct {
	Limit  int
	Kind   int
	User   string
	IsMine bool
}

func NewCmdSnippetList(gCtx *snippet.Context) *cobra.Command {
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
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.Config)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			limit := opts.Limit
			tools.Log.Info().Msgf("list Limit=%d", limit)
			nodes, err := snippet.GetSnippetNodes(gCtx, limit)
			if err != nil {
				return err
			}
			if err := snippet.RenderSnippetNodes(nodes); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&opts.Limit, "limit", "n",
		10, "Limit the number of results to display")
	return cmd
}

func NewCmdSnippetDesc(gCtx *snippet.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe the specified Snippet",
		Long: heredoc.Doc(`
              aard describe <name>, describes all data associated with the specified snippet. 

              Here, <name> can be the identifier or the URL of the snippet.
        `),
		Aliases: []string{"view", "show", "desc"},
		Example: heredoc.Doc(`
            To describe a specific snippet by URL, run:
              $ aard describe https://aard.app/edy6819l

            To describe a specific snippet by ID, run:
              $ aard describe edy6819l
		`),
		Args: MinimumArgs(1, "no name specified"),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.Config)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			tools.Log.Info().Msgf("describe name=%s", name)
			n, err := snippet.GetSnippetNode(gCtx, name)
			if err != nil {
				return err
			}
			if err := snippet.RenderSnippetNode(n, true /*show content*/); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}
