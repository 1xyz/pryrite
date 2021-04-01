package history

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/events"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func Cmd(entry *config.Entry, argv []string, version string) error {
	usage := `
usage: pruney history show [-n=<count>]
       pruney history add-event <index> [-m=<message>]
       pruney history append-in
       pruney history append <command>

Options:
  -n=<count>     Limit the number of history shown. zero is unlimited [default: 0].
  -m=<message>   Message to be added with a new event [default: None].
  -h --help      Show this screen.
`
	opts, err := docopt.ParseArgs(usage, argv, version)
	if err != nil {
		tools.Log.Fatal().Msgf("error parsing arguments. err=%v", err)
	}

	tools.Log.Printf("history raw-opts = %v", opts)
	if tools.OptsBool(opts, "show") {
		limit := tools.OptsInt(opts, "-n")
		h, err := New()
		if err != nil {
			return err
		}
		items, err := h.GetAll()
		if err != nil {
			return err
		}
		hr := &historyRender{
			H:     items,
			Limit: limit,
		}
		hr.Render()
	} else if tools.OptsBool(opts, "append") {
		return appendHistory(tools.OptsStr(opts, "<command>"))
	} else if tools.OptsBool(opts, "append-in") {
		command, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		return appendHistory(string(command))
	} else if tools.OptsBool(opts, "add-event") {
		index := tools.OptsInt(opts, "<index>")
		message := tools.OptsStr(opts, "-m")
		h, err := New()
		if err != nil {
			return err
		}
		item, err := h.GetByIndex(index)
		if err != nil {
			return err
		}
		if _, err := events.AddConsoleEvent(entry, item.Command, message, true); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no command found")
	}
	return nil
}

func appendHistory(cmd string) error {
	cmd = strings.TrimSpace(cmd)
	if len(cmd) == 0 {
		return fmt.Errorf("empty command provided")
	}
	h, err := New()
	if err != nil {
		return err
	}
	item := &Item{
		CreatedAt: time.Now().UTC(),
		Command:   cmd,
	}
	if err := h.Append(item); err != nil {
		return err
	}
	return err
}
