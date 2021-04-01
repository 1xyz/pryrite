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
	usage := `
usage: pruney events list [-n=<count>] 
       pruney events show <id>
       pruney events add <content> [-m=<message>]

Options:
  -m=<message>   Message to be added with a new event [default: None].
  -n=<count>     Limit the number of events shown [default: 10].
  -h --help      Show this screen.

Examples:
  List the most recent 5 events
  $ pruney events list -n 5

  Show a specific event with id 25
  $ pruney events show 25

  Add an event message with content and message
  $ pruney events add  "certbot run -a manual -i nginx -d example.com" -m "certbot manual nginx plugin" 
`
	opts, err := docopt.ParseArgs(usage, argv, version)
	if err != nil {
		tools.Log.Fatal().Msgf("error parsing arguments. err=%v", err)
	}
	tools.Log.Debug().Msgf("events.Cmd Opts = %v", opts)
	store := NewStore(entry.ServiceUrl)
	if tools.OptsBool(opts, "list") {
		n := tools.OptsInt(opts, "-n")
		events, err := store.GetEvents(n)
		if err != nil {
			return err
		}
		evtRender := &eventsRender{E: events}
		evtRender.Render()
	} else if tools.OptsBool(opts, "show") {
		id := tools.OptsStr(opts, "<id>")
		event, err := store.GetEvent(id)
		if err != nil {
			return err
		}
		evtRender := &eventRender{E: event}
		evtRender.Render()
	} else if tools.OptsBool(opts, "add") {
		content := tools.OptsStr(opts, "<content>")
		content = strings.TrimSpace(content)
		if len(content) == 0 {
			return fmt.Errorf("content cannot be empty")
		}

		message := tools.OptsStr(opts, "-m")
		if message == "None" {
			message = tools.TrimLength(content, maxColumnLen)
		}

		rawJson, err := json.Marshal(&ConsoleMetadata{Raw: content})
		if err != nil {
			return err
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
			return err
		}
		evtRender := &eventRender{E: event}
		evtRender.Render()
	}
	return nil
}
