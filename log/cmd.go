package log

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/cmd"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
	"io/ioutil"
	"os"
	"strings"
)

func Cmd(entry *config.Entry, params *cmd.Params) error {
	usage := `The "log" command allows you to work with snippet content from the remote pruney service

usage: pruney log [-n=<count>]
       pruney log add [-m message] (<content>|--file=<filename>|--stdin)
       pruney log pbcopy <id>
       pruney log pbpaste [-m message]
       pruney log search <query>...
       pruney log show <id> [--file=<filename>]

Options:
  -m=<message>       Message to be added when creating a new snippet [default: ].
  -n=<count>         Limit the number of snippets shown [default: 10].
  --file=<filename>  Filename to read/write snippet content from/to.
  -h --help          Show this screen.

Examples:
  List the most recent 5 snippets from the log.
  $ pruney log -n 5

  Log a snippet with content and message.
  $ pruney log add  -m "certbot manual nginx plugin" "certbot run -a manual -i nginx -d example.com"

  Log a new snippet from the specified file.
  $ pruney log add -m "main gist file" --file main.go

  Log a new snippet from stdin. In this example pipe the content of /tmp/example.log.
  $ cat /tmp/example.log | pruney log add -m "Example log snippet" --stdin

  Copy the content of snippet with id 25 to your local clipboard.
  $ pruney log pbcopy 25

  Log the content of the clipboard as a new snippet.
  $ pruney log pbpaste

  Search for snippets that match the provided query.
  $ pruney log search cerbot 

  Show a specific snippet logged with id 25 in detail.
  $ pruney log show 25

  Show a specific snippet with id 25 and write the snippet content to a file
  $ pruney log show 25 --file=/tmp/event-25.txt
`
	opts, err := docopt.ParseArgs(usage, params.Argv, params.Version)
	if err != nil {
		tools.Log.Fatal().Msgf("error parsing arguments. err=%v", err)
	}
	tools.Log.Debug().Msgf("events.Cmd Opts = %v", opts)
	store := graph.NewStore(entry)
	if tools.OptsBool(opts, "show") {
		id := tools.OptsStr(opts, "<id>")
		event, err := store.GetNode(id)
		if err != nil {
			return err
		}

		renderDetail := true
		if tools.OptsContains(opts, "--file") {
			renderDetail = false
			filename := tools.OptsStr(opts, "--file")
			if err := graph.WriteSnippetDetails(event, filename, false); err != nil {
				return err
			}
		}
		evtRender := &nodeRender{Node: event, renderDetail: renderDetail}
		evtRender.Render()
	} else if tools.OptsBool(opts, "add") {
		//fmt.Printf("Opts = %v\n", opts)
		message := tools.OptsStr(opts, "-m")
		content := ""
		if tools.OptsContains(opts, "<content>") {
			content = tools.OptsStr(opts, "<content>")
		} else if tools.OptsContains(opts, "--file") {
			filename := tools.OptsStr(opts, "--file")
			_, err := graph.AddSnippetFromFile(entry, graph.Command, entry.ClientID, params.Agent, params.Version, filename, message)
			return err
		} else if tools.OptsContains(opts, "--stdin") {
			b, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			content = string(b)
		} else {
			return fmt.Errorf("unrecognized option")
		}

		if _, err := graph.AddCommandSnippet(entry, entry.ClientID, params.Agent, params.Version, content, message); err != nil {
			return err
		}
	} else if tools.OptsBool(opts, "pbcopy") {
		id := tools.OptsStr(opts, "<id>")
		event, err := store.GetNode(id)
		if err != nil {
			return err
		}

		d, err := event.DecodeDetails()
		if err != nil {
			return err
		}
		if err := clipTo(d); err != nil {
			return fmt.Errorf("clipTo err = %v", err)
		}
		fmt.Printf("copied to clipboard!\n")
	} else if tools.OptsBool(opts, "pbpaste") {
		content, err := getClip()
		if err != nil {
			return fmt.Errorf("getClip err = %v", err)
		}
		message := tools.OptsStr(opts, "-m")
		node, err := graph.AddCommandSnippet(entry, entry.ClientID, params.Agent, params.Version, content, message)
		if err != nil {
			return err
		}
		fmt.Printf("logged command with snippet id = %v\n", node.ID)
	} else if tools.OptsBool(opts, "search") {
		queryRaw := tools.OptsStrSlice(opts, "<query>")
		query := strings.Join(queryRaw, " ")
		nodes, err := store.SearchNodes(query)
		if err != nil {
			return err
		}
		evtRender := &nodesRender{Nodes: nodes}
		evtRender.Render()
	} else {
		n := tools.OptsInt(opts, "-n")
		nodes, err := store.GetNodes(n)
		if err != nil {
			return err
		}
		evtRender := &nodesRender{Nodes: nodes}
		evtRender.Render()
	}
	return nil
}
