package capture

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/cmd"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
	"os"
)

func Cmd(entry *config.Entry, params *cmd.Params) error {
	usage := `The "termcast" command provides capture/play commands terminal stdout as an asciicast

usage: pruney termcast record --title=<title> [--file=<filename>] 
       pruney termcast play (<id> | --file=<filename>)

Options:
  --file=<filename>    The name of the cast file.
  --title=<title>      The title of the cast to be recorded.
  -h --help            Show this screen.

Examples:
   # Start a new a new ascii cast with title "foobar"
   # The cast would be saved as an event
   $ pruney termcast record --title "foobar"

   # Play an event with id 25, (Assuming it it of type asciicast)
   $ pruney termcast play 25
`
	opts, err := docopt.ParseArgs(usage, params.Argv, params.Version)
	if err != nil {
		tools.Log.Fatal().Msgf("error parsing arguments. err=%v", err)
	}

	if tools.OptsBool(opts, "record") {
		filename := ""
		if tools.OptsContains(opts, "--file") {
			filename = tools.OptsStr(opts, "--file")
		} else {
			f, err := tools.CreateTempFile("", "t_*.cast")
			if err != nil {
				return err
			}
			filename = f
		}

		title := tools.OptsStr(opts, "--title")
		shell := os.Getenv("SHELL")
		if len(shell) == 0 {
			return fmt.Errorf("unknown shell")
		}
		fmt.Printf("Type exit to end the cast")
		if err := Capture(title, filename, shell); err != nil {
			return err
		}
		if _, err := graph.AddSnippetFromFile(entry, graph.AsciiCast, entry.ClientID, params.Agent, params.Version, filename, title); err != nil {
			return err
		}
		fmt.Printf("written cast to file %v\n", filename)
	} else if tools.OptsBool(opts, "play") {
		filename := ""
		if tools.OptsContains(opts, "--file") {
			filename = tools.OptsStr(opts, "--file")
		} else {
			f, err := tools.CreateTempFile("", "t_*.cast")
			if err != nil {
				return err
			}
			filename = f

			eventID := tools.OptsStr(opts, "<id>")
			event, err := graph.GetSnippet(entry, eventID)
			if err != nil {
				return err
			}
			if event.Kind != graph.AsciiCast {
				return fmt.Errorf("the event %s is not of type %s", event.Kind, graph.AsciiCast)
			}
			if err := graph.WriteSnippetDetails(event, filename, true); err != nil {
				return err
			}
		}

		if err := Play(filename); err != nil {
			return err
		}
		fmt.Printf("Done!\n")
	}

	return nil
}
