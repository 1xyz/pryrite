package kmd

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/inspector"
	"github.com/aardlabs/terminal-poc/internal/ui/components"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
)

type SnippetListOpts struct {
	Limit       int
	ShowAllKind bool
	User        string
	IsMine      bool
	interactive bool
}

func NewCmdSnippetSearch(gCtx *snippet.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search available snippets",
		Long: examplef(`
              {AppName} search, searches all snippets which are visible to the current
              logged-in user. This includes both a user's own snippets as well
              as snippets shared.

			  The results of search are presented in a simple UI interface
        `),
		Example: examplef(`
            To search snippets for the term certutil, run:
              $ {AppName} search certutil
		`),
		Args: MinimumArgs(1, "You need to specify a search query"),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.ConfigEntry)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "-h" || args[0] == "--help" {
				return cmd.Help()
			}

			const limit = 25
			query := strings.Join(args, " ")
			kind := graph.Unknown
			tools.Log.Info().Msgf("search query=%s Limit=%d Kind=%v", query, limit, kind)
			nodes, err := snippet.SearchSnippetNodes(gCtx, query, limit, kind)
			if err != nil {
				return err
			}

			selNode, err := components.RenderNodesPicker(gCtx.ConfigEntry, nodes,
				"result for query: "+query+"; pick a node", 10, -1)
			if err != nil {
				if err == components.ErrNoEntryPicked {
					return nil
				}
				return err
			}

			return inspector.InspectNode(gCtx, selNode.Node.ID)
		},
	}
	return cmd
}

func NewCmdSnippetList(gCtx *snippet.Context) *cobra.Command {
	opts := &SnippetListOpts{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available snippets",
		Long: examplef(`
              {AppName} list, lists all snippets which are visible to the current
              logged-in user. This includes both a user's own snippets as well
              as snippet shared.

              By default, only "command" snippets are listed. This can be changed
              by using the --kind=all flag.
        `),
		Example: examplef(`
            To list the most recently created snippets, run:
              $ {AppName} list

            To list in an interactive UI, run:
              $ {AppName} list -i

            To list the most recent "n=100" snippets, run:
              $ {AppName} list -n 100

            To list all kinds of snippets that include non-command snippets, run:
              $ {AppName} list --all-kinds
		`),
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.ConfigEntry)
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
			if opts.interactive {
				selNode, err := components.RenderNodesPicker(gCtx.ConfigEntry, nodes,
					"select a node to inspect", 10, -1)
				if err != nil {
					if err == components.ErrNoEntryPicked {
						return nil
					}
					return err
				}
				return inspector.InspectNode(gCtx, selNode.Node.ID)
			} else {
				if err := snippet.RenderSnippetNodes(gCtx.ConfigEntry, nodes, kind); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&opts.Limit, "limit", "n",
		15, "Limit the number of results to display")
	cmd.Flags().BoolVarP(&opts.ShowAllKind, "all-kinds", "a",
		false, "include all kinds of snippets")
	cmd.Flags().BoolVarP(&opts.interactive, "interactive", "i",
		false, "show an interactive UI for listing")
	return cmd
}

func NewCmdSnippetDesc(gCtx *snippet.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe the specified Snippet",
		Long: examplef(`
              {AppName} describe <name>, describes all data associated with the specified snippet.

              Here, <name> can be the identifier or the URL of the snippet.
        `),

		Aliases: []string{"view", "show", "desc"},
		Example: examplef(`
            To describe a specific snippet by URL, run:
              $ {AppName} describe https://aardy.app/edy6819l

            To describe a specific snippet by ID, run:
              $ {AppName} describe edy6819l
		`),
		Args: MinimumArgs(1, "no name specified"),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.ConfigEntry)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			tools.Log.Info().Msgf("describe name=%s", name)
			view, err := snippet.GetSnippetNodeViewWithChildren(gCtx, name)
			if err != nil {
				return err
			}

			return snippet.RenderSnippetNodeView(gCtx.ConfigEntry, view)
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
		Long: examplef(`
              {AppName} save <content>, save content to the remote service.

              Here, <content> can be any content (typically a shell commend) that you want to be saved.
        `),
		Aliases: []string{"add", "stash"},
		Example: examplef(`
            To save a specified docker command, run:

              $ {AppName} save docker-compose run --rm --service-ports development bash
		`),
		Args: MinimumArgs(1, "no content specified"),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.ConfigEntry)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "-h" || args[0] == "--help" {
				return cmd.Help()
			}

			content := strings.Join(args, " ")
			tools.Log.Info().Msgf("add content=%s", content)
			shell := os.Getenv("SHELL")
			contentType := "text/"
			if strings.HasSuffix(shell, "/bash") {
				contentType += "bash"
			} else {
				contentType += "shell"
			}
			n, err := snippet.AddSnippetNode(gCtx, content, contentType)
			if err != nil {
				return err
			}

			tools.Log.Info().Msgf("AddSnippetNode n.ID = %s", n.ID)
			fmt.Printf("Added a new node with id = %s\n", snippet.GetNodeViewURL(gCtx, n))
			return nil
		},
	}
	return cmd
}

func NewCmdSnippetEdit(gCtx *snippet.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit and update content of the specified snippet",
		Long: examplef(`
              {AppName} edit <name>, edits the content of the specified snippet.

              Here, <name> can be the identifier or the URL of the snippet.

              The edit command allows you to directly edit the content of a command. It will open
              the editor defined by the EDITOR environment variable, or fall back to 'nano' for Linux/OSX
              or 'notepad' for Windows. Upon exiting the editor, the content will be updated on the remote
              service.
        `),
		Example: examplef(`
            To edit a specific snippet by URL, run:
              $ {AppName} edit https://aardy.app/edy6819l

            To edit a specific snippet by ID, run:
              $ {AppName} edit edy6819l
		`),
		Args: MinimumArgs(1, "no name specified"),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.ConfigEntry)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "-h" || args[0] == "--help" {
				return cmd.Help()
			}

			name := args[0]
			tools.Log.Info().Msgf("edit name=%s", name)
			_, err := snippet.EditSnippetNode(gCtx, name, true /*save*/)
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

func NewCmdSnippetRun(gCtx *snippet.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <name>",
		Short: "Interactively run the content of the specified snippet",
		Long: examplef(`
              {AppName} inspect <name>, inspects the content of the specified snippet.

              Here, <name> can be the identifier or the URL of the snippet.

              The run command allows you to interactively view & run the content of a snippet.
        `),
		Example: examplef(`
            To run a specific snippet by URL, run:
              $ {AppName} inspect https://aardy.app/edy6819l

            To run a specific snippet by ID, run:
              $ {AppName} inspect edy6819l
		`),
		Args: MinimumArgs(1, "no name specified"),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.ConfigEntry)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "-h" || args[0] == "--help" {
				return cmd.Help()
			}

			name := args[0]
			tools.Log.Info().Msgf("edit name=%s", name)
			return inspector.InspectNode(gCtx, name)
		},
	}
	return cmd
}
