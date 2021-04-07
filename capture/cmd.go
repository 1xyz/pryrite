package capture

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/events"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
	"os"
)

func Cmd(entry *config.Entry, argv []string, version string) error {
	usage := `The "capture" command provides capture/play commands terminal stdout as an asciicast

usage: pruney capture record --title=<title> [--file=<filename>] 
       pruney capture play (<id> | --file=<filename>)

Options:
  --file=<filename>    The name of the cast file.
  --title=<title>      The title of the cast to be recorded.
  -h --help            Show this screen.

Examples:
`
	opts, err := docopt.ParseArgs(usage, argv, version)
	if err != nil {
		tools.Log.Fatal().Msgf("error parsing arguments. err=%v", err)
	}

	fmt.Printf("Opts = %v\n", opts)
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
		if _, err := events.AddEventFromFile(entry, "AsciiCast", filename, title, false); err != nil {
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
			event, err := events.GetEvent(entry, eventID)
			if err != nil {
				return err
			}
			if err := events.WriteEventDetailsToFile(event, filename, true); err != nil {
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
