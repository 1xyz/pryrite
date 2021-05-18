package kmd

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/spf13/cobra"
)

type SnippetListOpts struct {
	Limit       int
	ShowAllKind bool
	User        string
	IsMine      bool
}

func NewCmdSnippetSearch(gCtx *snippet.Context) *cobra.Command {
	opts := &SnippetListOpts{}
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search available snippets",
		Long: heredoc.Doc(`
              aard search, searches all snippets which are visible to the current
              logged-in user. This includes both a user's own snippets as well
              as snippet shared.

              By default, only "command" snippets are searched. This can be changed
              by using the --kind=all flag.
        `),
		Example: heredoc.Doc(`
            To search snippets for the term certutil, run:
              $ aard search "certutil"

            To limit the search result to 10 entries, run:
              $ aard search "certutil" -n 10

            To include all kinds of snippets that include non-command snippets, run:
              $ aard search certutil --all-kinds
		`),
		Args: MinimumArgs(1, "You need to specify a search query"),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.Config)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			limit := opts.Limit
			kind := graph.Command
			if opts.ShowAllKind {
				kind = graph.Unknown
			}
			tools.Log.Info().Msgf("search query=%s Limit=%d Kind=%v", query, limit, kind)
			nodes, err := snippet.SearchSnippetNodes(gCtx, query, limit, kind)
			if err != nil {
				return err
			}
			if err := snippet.RenderSnippetNodes(gCtx.Config, nodes, kind); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&opts.Limit, "limit", "n",
		100, "Limit the number of results to display")
	cmd.Flags().BoolVarP(&opts.ShowAllKind, "all-kinds", "a",
		false, "include all kinds of snippets")
	return cmd
}

func NewCmdSnippetList(gCtx *snippet.Context) *cobra.Command {
	opts := &SnippetListOpts{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available snippets",
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

            To list all kinds of snippets that include non-command snippets, run:
              $ aard list --all-kinds
		`),
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.Config)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			limit := opts.Limit
			tools.Log.Info().Msgf("list Limit=%d", limit)
			kind := graph.Command
			if opts.ShowAllKind {
				kind = graph.Unknown
			}
			nodes, err := snippet.GetSnippetNodes(gCtx, limit, kind)
			if err != nil {
				return err
			}
			if err := snippet.RenderSnippetNodes(gCtx.Config, nodes, kind); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&opts.Limit, "limit", "n",
		100, "Limit the number of results to display")
	cmd.Flags().BoolVarP(&opts.ShowAllKind, "all-kinds", "a",
		false, "include all kinds of snippets")
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
              $ aard describe https://aardy.app/edy6819l

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
			view, err := snippet.GetSnippetNodeView(gCtx, name)
			if err != nil {
				return err
			}

			if err := snippet.RenderSnippetNodeView(gCtx.Config, view); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func NewCmdSnippetSave(gCtx *snippet.Context) *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,

		Use:   "save <content>...",
		Short: "Save a new snippet with the specified content",
		Long: heredoc.Doc(`
              aard save <content>, save content to the remote service.

              Here, <content> can be any content (typically a shell commend) that you want to be saved.
        `),
		Aliases: []string{"add", "stash"},
		Example: heredoc.Doc(`
            To save a specified docker command, run:

              $ aard save docker-compose run --rm --service-ports development bash
		`),
		Args: MinimumArgs(1, "no content specified"),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.Config)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "-h" || args[0] == "--help" {
				return cmd.Help()
			}

			content := strings.Join(args, " ")
			tools.Log.Info().Msgf("add content=%s", content)
			n, err := snippet.AddSnippetNode(gCtx, content)
			if err != nil {
				return err
			}

			tools.Log.Info().Msgf("AddSnippetNode n.ID = %s", n.ID)
			fmt.Printf("Added a new node with id = %s\n", n.ID)
			return nil
		},
	}
	return cmd
}

func NewCmdSnippetEdit(gCtx *snippet.Context) *cobra.Command {
	cmd := &cobra.Command{
		Short: "Edit and update content of the specified snippet",
		Long: heredoc.Doc(`
              aard edit <name>, edits the content of the specified snippet.

              Here, <name> can be the identifier or the URL of the snippet.

              The edit command allows you to directly edit the content of a command. It will open
              the editor defined by the EDITOR environment variable, or fall back to 'nano' for Linux/OSX
              or 'notepad' for Windows. Upon exiting the editor, the content will be updated on the remote
              service.
        `),
		Example: heredoc.Doc(`
            To edit a specific snippet by URL, run:
              $ aard edit https://aardy.app/edy6819l

            To edit a specific snippet by ID, run:
              $ aard edit edy6819l
		`),
		Args: MinimumArgs(1, "no name specified"),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.Config)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "-h" || args[0] == "--help" {
				return cmd.Help()
			}

			name := args[0]
			tools.Log.Info().Msgf("edit name=%s", name)
			_, err := snippet.EditSnippetNode(gCtx, name)
			if err != nil {
				return err
			}
			//if err := snippet.RenderSnippetNodeView(gCtx.Config, n, true /*show content*/); err != nil {
			//	return err
			//}
			return nil
		},
	}
	return cmd
}
