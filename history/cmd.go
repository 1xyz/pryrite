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
	usage := `The "history" command allows pruney to work with local shell command history

usage: pruney history [-n=<count>]
       pruney history log <index> -m=<message>
       pruney history append (<command> | --stdin)

Options:
  -c=<command>   Specify the command to be appended to the local history.
  -n=<count>     Limit the number of history shown. zero is unlimited [default: 25].
  -m=<message>   Message to be added with a new event.
  --stdin        Read from the standard input.
  -h --help      Show this screen.

Examples(s):
  # Show the last three entries from the local history.
  $ pruney history -n 3
  ┏━━━━━━━┳━━━━━━━━━━━━━━━━┳━━━━━━━━━┓
  ┃ INDEX ┃ DATE           ┃ COMMAND ┃
  ┃    90 ┃ Apr  1 2:11PM  ┃ ls -l   ┃                                                                                                                                                                      ┃
  ┃    91 ┃ Apr  1 2:11PM  ┃ cd /    ┃                                                                                                                                                                      ┃
  ┃    92 ┃ Apr  1 2:11PM  ┃ ls -la  ┃
  ┗━━━━━━━┻━━━━━━━━━━━━━━━━┻━━━━━━━━━┛
  # Note: the above result show last three history entries   

  # Log a specific history entry, by index to the pruney remote log.
  $ pruney history log 102 -m "list longform with hidden attrs"
 
  # Append to the local history. Note: Intended to be call by shell hooks. 
  $ pruney history append "ls -l " 
`
	opts, err := docopt.ParseArgs(usage, argv, version)
	if err != nil {
		tools.Log.Fatal().Msgf("error parsing arguments. err=%v", err)
	}

	//fmt.Printf("raw opts = %v", opts)
	tools.Log.Printf("history raw-opts = %v", opts)
	if tools.OptsBool(opts, "append") {
		useStdin := tools.OptsBool(opts, "--stdin")
		if useStdin {
			command, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			return appendHistory(string(command))
		} else {
			return appendHistory(tools.OptsStr(opts, "<command>"))
		}
	} else if tools.OptsBool(opts, "log") {
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
