package events

import (
	"encoding/json"
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
	"github.com/google/uuid"
	"strings"
	"time"
)

func Cmd(entry *config.Entry, argv []string, version string) error {
	usage := `The "log" command allows you to retrieve events from the remote pruney log

usage: pruney log [-n=<count>] 
       pruney log show <id>
       pruney log add <content> [-m=<message>]

Options:
  -m=<message>   Message to be added with a new event [default: None].
  -n=<count>     Limit the number of events shown [default: 10].
  -h --help      Show this screen.

Examples:
  List the most recent 5 events from the log
  $ pruney log  -n 5

  Show a specific event with id 25
  $ pruney log show 25

  Log an event with content and message
  $ pruney log add  "certbot run -a manual -i nginx -d example.com" -m "certbot manual nginx plugin" 
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
		content := tools.OptsStr(opts, "<content>")
		message := tools.OptsStr(opts, "-m")
		_, err := AddConsoleEvent(entry, content, message, true)
		return err
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

	rawJson, err := json.Marshal(&ConsoleMetadata{Raw: content})
	if err != nil {
		return nil, err
	}
	event, err := store.AddEvent(&Event{
		CreatedAt: time.Now().UTC(),
		Kind:      "Console",
		Details:   rawJson,
		Metadata: Metadata{
			SessionID: uuid.New().String(),
			Title:     message,
			URL:       "",
		},
	})
	if err != nil {
		return nil, err
	}

	if doRender {
		evtRender := &eventRender{E: event}
		evtRender.Render()
	}
	return event, nil
}
