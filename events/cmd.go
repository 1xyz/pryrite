package events

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
	"io/ioutil"
	"os"
	"strings"
)

func Cmd(entry *config.Entry, argv []string, version string) error {
	usage := `The "log" command allows you to retrieve events from the remote pruney log

usage: pruney log [-n=<count>]
       pruney log add [-m message] (<content>|--file=<filename>|--stdin)
       pruney log pbcopy <id>
       pruney log pbpaste [-m message]
       pruney log show <id> 

Options:
  -m=<message>   Message to be added with a new event [default: None].
  -n=<count>     Limit the number of events shown [default: 10].
  -h --help      Show this screen.

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
`
	opts, err := docopt.ParseArgs(usage, argv, version)
	if err != nil {
		tools.Log.Fatal().Msgf("error parsing arguments. err=%v", err)
	}
	tools.Log.Debug().Msgf("events.Cmd Opts = %v", opts)
	store := NewStore(entry.ServiceUrl)
	if tools.OptsBool(opts, "show") {
		id := tools.OptsStr(opts, "<id>")
		event, err := store.GetEvent(id)
		if err != nil {
			return err
		}
		evtRender := &eventRender{E: event}
		evtRender.Render()
	} else if tools.OptsBool(opts, "add") {
		//fmt.Printf("Opts = %v\n", opts)
		content := ""
		if tools.OptsContains(opts, "<content>") {
			content = tools.OptsStr(opts, "<content>")
		} else if tools.OptsContains(opts, "--file") {
			b, err := os.ReadFile(tools.OptsStr(opts, "--file"))
			if err != nil {
				return err
			}
			content = string(b)
		} else if tools.OptsContains(opts, "--stdin") {
			b, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			content = string(b)
		} else {
			return fmt.Errorf("unrecognized option")
		}

		message := tools.OptsStr(opts, "-m")
		if _, err := AddConsoleEvent(entry, content, message, true); err != nil {
			return err
		}
	} else if tools.OptsBool(opts, "pbcopy") {
		id := tools.OptsStr(opts, "<id>")
		event, err := store.GetEvent(id)
		if err != nil {
			return err
		}

		d, err := event.DecodeDetails()
		if err != nil {
			return err
		}
		if err := clipTo(d.Body()); err != nil {
			return fmt.Errorf("clipTo err = %v", err)
		}
		fmt.Printf("copied to clipboard!\n")
	} else if tools.OptsBool(opts, "pbpaste") {
		content, err := getClip()
		if err != nil {
			return fmt.Errorf("getClip err = %v", err)
		}
		message := tools.OptsStr(opts, "-m")
		if _, err := AddConsoleEvent(entry, content, message, true); err != nil {
			return err
		}
	} else {
		n := tools.OptsInt(opts, "-n")
		events, err := store.GetEvents(n)
		if err != nil {
			return err
		}
		evtRender := &eventsRender{E: events}
		evtRender.Render()
	}
	return nil
}

func AddConsoleEvent(entry *config.Entry, content, message string, doRender bool) (*Event, error) {
	store := NewStore(entry.ServiceUrl)
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return nil, fmt.Errorf("content cannot be empty")
	}
	if message == "None" {
		message = tools.TrimLength(content, maxColumnLen)
	}

	event, err := New("Console", message, "", &RawDetails{Raw: content})
	if err != nil {
		return nil, err
	}
	event, err = store.AddEvent(event)
	if err != nil {
		return nil, err
	}

	if doRender {
		evtRender := &eventRender{E: event}
		evtRender.Render()
	}
	return event, nil
}
