package cmd

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/history"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func HistoryCmd(entry *config.Entry, params *Params) error {
	usage := `The "history" command allows one to work with local shell command history

usage: aard history [-n=<count>]
       aard history log <index> [-m=<message>]
       aard history append (<command> | --stdin)

Options:
  -c=<command>   Specify the command to be appended to the local history.
  -n=<count>     Limit the number of history shown. zero is unlimited [default: 25].
  -m=<message>   Message to be added with a new event [default: ].
  --stdin        Read from the standard input.
  -h --help      Show this screen.

Examples(s):
  # Show the last three entries from the local history.
  $ aard history -n 3
  ┏━━━━━━━┳━━━━━━━━━━━━━━━━┳━━━━━━━━━┓
  ┃ INDEX ┃ DATE           ┃ COMMAND ┃
  ┃    90 ┃ Apr  1 2:11PM  ┃ ls -l   ┃
  ┃    91 ┃ Apr  1 2:11PM  ┃ cd /    ┃
  ┃    92 ┃ Apr  1 2:11PM  ┃ ls -la  ┃
  ┗━━━━━━━┻━━━━━━━━━━━━━━━━┻━━━━━━━━━┛
  # Note: the above result show last three history entries   

  # Log a specific history entry, by index to the aard remote log.
  $ aard history log 102 -m "list longform with hidden attrs"
 
  # Append to the local history. Note: Intended to be call by shell hooks. 
  $ aard history append "ls -l " 
`
	opts, err := docopt.ParseArgs(usage, params.Argv, params.Version)
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
		if entry == nil {
			return fmt.Errorf("no remote configuration found")
		}
		index := tools.OptsInt(opts, "<index>")
		message := tools.OptsStr(opts, "-m")
		h, err := history.New()
		if err != nil {
			return err
		}
		item, err := h.GetByIndex(index)
		if err != nil {
			return err
		}
		node, err := graph.AddCommandSnippet(entry, entry.ClientID, params.Agent, params.Version, item.Command, message)
		if err != nil {
			return err
		}
		fmt.Printf("logged a new snippet with id=%v\n", node.ID)
	} else {
		limit := tools.OptsInt(opts, "-n")
		h, err := history.New()
		if err != nil {
			return err
		}
		items, err := h.GetAll()
		if err != nil {
			return err
		}
		hr := &history.HistoryRender{
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
		// ignore this. zsh sends an empty string
		return nil
	}
	if cmd == "aard history" {
		// skip inserting the history command
		return nil
	}
	h, err := history.New()
	if err != nil {
		return err
	}
	item := &history.Item{
		CreatedAt: time.Now().UTC(),
		Command:   cmd,
	}
	if err := h.Append(item); err != nil {
		return err
	}
	return err
}
