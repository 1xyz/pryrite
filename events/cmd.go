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

		renderDetail := true
		if tools.OptsContains(opts, "--file") {
			renderDetail = false
			filename := tools.OptsStr(opts, "--file")
			if err := WriteEventDetailsToFile(event, filename, false); err != nil {
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
			_, err := AddEventFromFile(entry, "Console", filename, message, true)
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
	return AddEvent(entry, "Console", content, message, doRender)
}

func AddEventFromFile(entry *config.Entry, kind, filename, message string, doRender bool) (*Event, error) {
	// ToDo: *maybe* a better way
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	content := string(b)
	return AddEvent(entry, kind, content, message, doRender)
}

func AddEvent(entry *config.Entry, kind, content, message string, doRender bool) (*Event, error) {
	store := NewStore(entry.ServiceUrl)
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return nil, fmt.Errorf("content cannot be empty")
	}
	if message == "None" {
		message = tools.TrimLength(content, maxColumnLen)
	}

	event, err := New(kind, message, "", &RawDetails{Raw: content})
	if err != nil {
		return nil, err
	}
	event, err = store.AddEvent(event)
	if err != nil {
		return nil, err
	}

	if doRender {
		evtRender := &eventRender{E: event, renderDetail: false}
		evtRender.Render()
	}
	return event, nil
}

func GetEvent(entry *config.Entry, eventID string) (*Event, error) {
	store := NewStore(entry.ServiceUrl)
	return store.GetEvent(eventID)
}

func WriteEventDetailsToFile(event *Event, filename string, overwrite bool) error {
	exists, err := tools.StatExists(filename)
	if err != nil {
		return err
	}
	if exists && !overwrite {
		return fmt.Errorf("cannot overwrite file = %v", filename)
	}
	fw, err := tools.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return err
	}
	defer tools.CloseFile(fw)
	if _, err := event.WriteBody(fw); err != nil {
		return err
	}
	return nil
}
