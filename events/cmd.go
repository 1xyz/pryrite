package events

import (
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
)

func Cmd(entry *config.Entry, argv []string, version string) error {
	usage := `
usage: pruney events list [-n=<count>] 
       pruney events show <id>

Options:
  -n=<count>  Limit the number of events shown [default: 10].
  -h --help   Show this screen.

Examples:
  List the most recent 5 events
  $ pruney events list -n 5

  Show a specific event with id 25
  $ pruney events show 25
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
	}
	return nil
}
