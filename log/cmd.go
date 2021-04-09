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
)

func Cmd(entry *config.Entry, params *cmd.Params) error {
	usage := `The "log" command allows you to retrieve events from the remote pruney log

usage: pruney log [-n=<count>]
       pruney log add [-m message] (<content>|--file=<filename>|--stdin)
       pruney log pbcopy <id>
       pruney log pbpaste [-m message]
       pruney log show <id> [--file=<filename>]

Options:
  -m=<message>       Message to be added with a new event [default: None].
  -n=<count>         Limit the number of events shown [default: 10].
  --file=<filename>  Filename to read/write details content from/to.
  -h --help          Show this screen.

Examples:
  List the most recent 5 events from the log
  $ pruney log -n 5

  Log an event with content and message
  $ pruney log add  -m "certbot manual nginx plugin" "certbot run -a manual -i nginx -d example.com"

  Log an event with a specific file 
  $ pruney log add -m "main gist file" --file main.go

  Log an event from stdin
  $ cat /tmp/example.log | pruney log add -m "Example log snippet" --stdin

  Copy the content of event 25 to your local clipboard
  $ pruney log pbcopy 25

  Copy the content of the clipboard to a new event
  $ pruney log pbpaste

  Show a specific event with id 25
  $ pruney log show 25

  Show a specific event with id 25 and write the details content to a file
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
		evtRender := &eventRender{E: event, renderDetail: renderDetail}
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
		if err := clipTo(d.Summary()); err != nil {
			return fmt.Errorf("clipTo err = %v", err)
		}
		fmt.Printf("copied to clipboard!\n")
	} else if tools.OptsBool(opts, "pbpaste") {
		content, err := getClip()
		if err != nil {
			return fmt.Errorf("getClip err = %v", err)
		}
		message := tools.OptsStr(opts, "-m")
		if _, err := graph.AddCommandSnippet(entry, entry.ClientID, params.Agent, params.Version, content, message); err != nil {
			return err
		}
	} else {
		n := tools.OptsInt(opts, "-n")
		events, err := store.GetNodes(n)
		if err != nil {
			return err
		}
		evtRender := &eventsRender{E: events}
		evtRender.Render()
	}
	return nil
}
