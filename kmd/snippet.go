package kmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/1xyz/pryrite/historian"
	"github.com/1xyz/pryrite/inspector"
	"github.com/1xyz/pryrite/internal/common"
	"github.com/1xyz/pryrite/internal/slurp"
	"github.com/1xyz/pryrite/internal/ui/components"
	"github.com/1xyz/pryrite/shells"

	"github.com/spf13/cobra"

	"github.com/1xyz/pryrite/graph"
	"github.com/1xyz/pryrite/snippet"
	"github.com/1xyz/pryrite/tools"
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
		Long: tools.Examplef(`
              {AppName} search, searches all snippets which are visible to the current
              logged-in user. This includes both a user's own snippets as well
              as snippets shared.

			  The results of search are presented in a simple UI interface
        `),
		Example: tools.Examplef(`
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
		Long: tools.Examplef(`
              {AppName} list, lists all snippets which are visible to the current
              logged-in user. This includes both a user's own snippets as well
              as snippet shared.

              By default, only "command" snippets are listed. This can be changed
              by using the --kind=all flag.
        `),
		Example: tools.Examplef(`
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
		Long: tools.Examplef(`
              {AppName} describe <name>, describes all data associated with the specified snippet.

              Here, <name> can be the identifier or the URL of the snippet.
        `),

		Aliases: []string{"view", "show", "desc"},
		Example: tools.Examplef(`
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

			view, err := snippet.GetSnippetNodeWithChildren(gCtx, name)
			if err != nil {
				return err
			}

			return snippet.RenderSnippetNodeView(gCtx.ConfigEntry, view)
		},
	}
	return cmd
}

func NewCmdSnippetSave(gCtx *snippet.Context) *cobra.Command {
	showHelp := func(cmd *cobra.Command) {
		fmt.Println(cmd.Long)
		fmt.Println("Examples:")
		fmt.Println(cmd.Example)
	}

	cmd := &cobra.Command{
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,

		Use:   "save [flags] <content>...",
		Short: "Save a new snippet with the specified content",
		Long: tools.Examplef(`
              {AppName} save <content>, save content to the remote service.

              Here, <content> can be any content (typically a shell commend) that you want to be saved.
              Or, use caret notation to save an item from the history log.

              Flags:

                -h, --help     help for save
                -m, --message  Include an optional message
        `),
		Aliases: []string{"add", "stash"},
		Example: tools.Examplef(`
            To save a specified docker command, run:

              $ {AppName} save docker-compose run --rm --service-ports development bash

            To save a specified command with a message, run:

              $ {AppName} save -m "list files" ls -l
		`),
		Args: MinimumArgs(1, "no content specified"),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.ConfigEntry)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			saveArgs, err := ParseSaveArgs(args)
			if err != nil {
				switch err {
				case ErrShowHelp:
					showHelp(cmd)
					return nil
				default:
					showHelp(cmd)
					return err
				}
			}

			shell := os.Getenv("SHELL")
			contentType := "text/"
			if strings.HasSuffix(shell, "/bash") {
				contentType += "bash"
			} else {
				contentType += "shell"
			}

			tools.Log.Info().Msgf("AddSnippetNode: message = %s, command = %s, content-type = %s",
				saveArgs.Message, saveArgs.Command, contentType)
			n, err := snippet.AddSnippetNode(gCtx, saveArgs.Message, saveArgs.Command, contentType)
			if err != nil {
				return err
			}

			tools.Log.Info().Msgf("AddSnippetNode n.ID = %s", n.ID)
			fmt.Printf("Added a new node with id = %s\n", common.GetNodeURL(gCtx.ConfigEntry, n.ID))
			return nil
		},
	}
	return cmd
}

func NewCmdSnippetSlurp(gCtx *snippet.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slurp",
		Short: "Reports a batch of commands to curate into snippets",
		Args:  cobra.RangeArgs(0, 1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.ConfigEntry)
		},
		// TODO: add --shell argument to handle cases like ssh XXX | slurp --shell zsh
		RunE: func(cmd *cobra.Command, args []string) error {
			slurpAll := false

			if len(args) == 0 {
				len := shells.GetHistoryLen()
				if len > 10 {
					fmt.Printf("\nSlurp all %d entries? (y|N)  ", len)
					var resp string
					fmt.Scanln(&resp)
					resp = strings.ToLower(resp)
					if resp == "" || resp[0] != "y"[0] {
						return nil
					}
				}

				slurpAll = true
			}

			var generator func(slurp.Digester) error

			if !slurpAll && args[0] == "-" {
				var slurper slurp.Slurper = &slurp.HistorySlurper{}
				generator = func(cb slurp.Digester) error {
					return slurper.Slurp("", os.Stdin, cb)
				}
			}

			var duration time.Duration
			var err error
			if !slurpAll {
				duration, err = time.ParseDuration(args[0])
			}
			if err == nil {
				generator = func(cb slurp.Digester) error {
					return shells.EachHistoryEntry(false, duration, true, func(item *historian.Item) error {
						// FIXME: get proper location (expose some of newSlurpFactory)
						slurp := &slurp.Slurp{
							ExecutedAt: &item.RecordedAt,
							Location: &url.URL{
								Scheme: "bash",
								Host:   "localhost",
							},
							Language:    "bash",
							Commandline: item.CommandLine,
						}

						if item.ExitStatus != nil {
							slurp.ExitStatus = fmt.Sprint(*item.ExitStatus)
						}

						return cb(slurp)
					})
				}
			}

			if generator == nil {
				return fmt.Errorf("unknown slurp argument: %s", args[0])
			}

			cnt := 0
			err = generator(func(slurp *slurp.Slurp) error {
				tools.Log.Info().Interface("slurp", slurp).Msg("AddSnippetNodeFromSlurp")
				n, err := snippet.AddSnippetNodeFromSlurp(gCtx, slurp)
				if err != nil {
					return err
				}

				tools.Log.Info().Msgf("AddSnippetNodeFromSlurp n.ID = %s", n.ID)
				cnt++
				return nil
			})

			if err != nil {
				return err
			}

			u, _ := url.Parse(gCtx.ConfigEntry.DashboardUrl)
			u.Path += "slurps"
			cmd.Printf("\nSaved %d snippets available at %s\n\n", cnt, u)

			return nil
		},
	}
	return cmd
}

func NewCmdSnippetEdit(gCtx *snippet.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit and update content of the specified snippet",
		Long: tools.Examplef(`
              {AppName} edit <name>, edits the content of the specified snippet.

              Here, <name> can be the identifier or the URL of the snippet.

              The edit command allows you to directly edit the content of a command. It will open
              the editor defined by the EDITOR environment variable, or fall back to 'nano' for Linux/OSX
              or 'notepad' for Windows. Upon exiting the editor, the content will be updated on the remote
              service.
        `),
		Example: tools.Examplef(`
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
		Long: tools.Examplef(`
              {AppName} inspect <name>, inspects the content of the specified snippet.

              Here, <name> can be the identifier or the URL of the snippet.

              The run command allows you to interactively view & run the content of a snippet.
        `),
		Example: tools.Examplef(`
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
			tools.Log.Info().Msgf("run name=%s", name)
			store, err := gCtx.GetStore()
			if err != nil {
				return err
			}

			nodeId, err := store.ExtractID(name)
			if err != nil {
				return err
			}

			return inspector.InspectNode(gCtx, nodeId)
		},
	}
	return cmd
}

type SaveArgs struct {
	Message string
	Command string
}

var (
	ErrEOL        = errors.New("parse incomplete: encountered end of line")
	ErrCmdMissing = errors.New("command missing")
	ErrShowHelp   = errors.New("show help")
)

func ParseSaveArgs(args []string) (*SaveArgs, error) {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		return nil, ErrShowHelp
	}

	var s SaveArgs
	index := 0
	if args[0] == "-m" || args[0] == "--message" {
		message, i, err := getMessage(args[1:])
		if err != nil {
			return nil, err
		}

		s.Message = message
		index = i + 1
	}

	if len(args) <= index {
		return nil, ErrCmdMissing
	}

	remainArgs := args[index:]
	cmd := remainArgs[0]

	if len(remainArgs) == 1 && cmd[0] == shells.ExpandChar {
		item, err := shells.GetHistoryEntry(cmd)
		if err != nil {
			return nil, err
		}

		s.Command = item.CommandLine
	} else {
		s.Command = strings.Join(remainArgs, " ")
	}

	return &s, nil
}

func getMessage(args []string) (string, int, error) {
	if len(args) == 0 {
		return "", 0, fmt.Errorf("EOF")
	}
	if args[0][0] != '"' {
		message := args[0]
		return message, 1, nil
	}

	var messages []string
	for i, arg := range args {
		if i == 0 {
			// strip the prefix quote
			arg = arg[1:]
		}
		if strings.HasSuffix(arg, "\"") && !strings.HasSuffix(arg, "\\\"") {
			// strip the suffix quote
			messages = append(messages, arg[0:len(arg)-1])
			return strings.Join(messages, " "), i + 1, nil
		}
		messages = append(messages, arg)
	}
	return "", 0, ErrEOL
}
