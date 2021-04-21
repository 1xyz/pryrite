package main

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/auth"
	"github.com/aardlabs/terminal-poc/capture"
	"github.com/aardlabs/terminal-poc/cmd"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/history"
	"github.com/aardlabs/terminal-poc/log"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
	"github.com/rs/zerolog"
	"os"
)

func setupLogfile() *os.File {
	var fp *os.File
	logFile := os.ExpandEnv("$HOME/.aardvark/aard.log")
	fp, err := tools.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND)
	if err != nil {
		tools.Log.Fatal().Err(err).Msgf("setupLogFile")
	}

	return fp
}

func main() {
	usage := `usage: aard [--version] [--verbose] [--help] <command> [<args>...]
options:
   -h, --help    Show this message.
   --verbose     Enable verbose logging (logfile: $HOME/.aard/aard.log).

The commands are:
   auth          authenticate with the remote service.
   config        provides options to configure aard.
   history       work with your local shell history.
   log           add & view snippets from the aard service.
   termcast      record/play an asciicast from your terminal.

See 'aard <command> --help' for more information on a specific command.

These are common aard commands used in various situations:

Configure and get started:
   # Setup up aard to talk to service 
   $ aard config add remote --service-url https://flaming-fishtoot.herokuapp.com/
   
   # Login into the user via the specified e-mail
   $ aard auth login alan@turing.me
   
Examine the snippet log:
   # List the most recent snippets from the log.
   aard log

   # Search the log for snippets involving cerbot
   aard log search cerbot

See 'aard <command> --help' for more information on a specific command.
`
	parser := &docopt.Parser{OptionsFirst: true}
	args, err := parser.ParseArgs(usage, nil, version)
	if err != nil {
		tools.Log.Fatal().Err(err).Msgf("parser.ParseArgs")
	}

	// setup logging
	logFp := setupLogfile()
	defer tools.CloseFile(logFp)
	level := zerolog.InfoLevel
	if tools.OptsBool(args, "--verbose") {
		level = zerolog.DebugLevel
	}
	tools.InitLogger(logFp, level)

	c := args["<command>"].(string)
	cArgs := args["<args>"].([]string)
	tools.Log.Debug().Msgf("global arguments: %v", args)
	tools.Log.Debug().Msgf("c arguments: %v %v", c, cArgs)
	p := cmd.Params{
		Agent:   "TermConsole",
		Version: version,
		Argv:    append([]string{c}, cArgs...),
		Command: c,
	}
	ec := RunCommand(&p)
	if ec > 0 {
		tools.Log.Warn().Msgf("c exited with %d", ec)
		os.Exit(ec)
	}
	tools.Log.Debug().Msgf("done")
}

type cmdFunc func(*config.Entry, *cmd.Params) error

var cmdFunctions = map[string]cmdFunc{
	"auth":     auth.Cmd,
	"termcast": capture.Cmd,
	"log":      log.Cmd,
	"history":  history.Cmd,
}

// RunCommand runs a specific command and the provided arguments
// Returns back an integer which is used as the exit code
// Ref: https://tldp.org/LDP/abs/html/exitcodes.html
func RunCommand(params *cmd.Params) int {
	if params.Command == "config" {
		if err := config.Cmd(params); err != nil {
			return cmd.LogErr(params.Command, err)
		}
	} else if params.Command == "history" {
		// history should continue to work if there is not config
		entry, _ := config.GetEntry("")
		if err := history.Cmd(entry, params); err != nil {
			return cmd.LogErr(params.Command, err)
		}
	} else {
		// Try creating a default config, if not found
		if err := config.CreateDefaultConfigIfEmpty(); err != nil {
			return cmd.LogErr(params.Command, err)
		}
		fn, found := cmdFunctions[params.Command]
		if !found {
			fmt.Printf("command %s not found!\n", params.Command)
			tools.Log.Warn().Msgf("%s is an unsupported command", params.Command)
			return 127 // command not found
		}
		entry, err := config.GetEntry("")
		if err != nil {
			return cmd.LogErr(params.Command, err)
		}
		if err := fn(entry, params); err != nil {
			return cmd.LogErr(params.Command, err)
		}
	}
	return 0
}
