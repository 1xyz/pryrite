package main

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/events"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
	"github.com/rs/zerolog"
	"os"
)

const version = "0.1.alpha"

func setupLogfile() *os.File {
	var fp *os.File
	logFile := os.ExpandEnv("$HOME/.pruney/pruney.log")
	fp, err := tools.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND)
	if err != nil {
		tools.Log.Fatal().Err(err).Msgf("setupLogFile")
	}

	return fp
}

func main() {
	usage := `usage: pruney [--version] [(--verbose|--quiet)] [--help]
           <command> [<args>...]
options:
   -h, --help
   --verbose      Change the logging level verbosity
The commands are:
   config   Explore configuration commands
See 'pruney <command> --help' for more information on a specific command.
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
	verbose := tools.OptsBool(args, "--verbose")
	quiet := tools.OptsBool(args, "--quiet")
	if verbose == true {
		level = zerolog.DebugLevel
	} else if quiet == true {
		level = zerolog.WarnLevel
	}

	tools.InitLogger(logFp, level)

	cmd := args["<command>"].(string)
	cmdArgs := args["<args>"].([]string)

	tools.Log.Debug().Msgf("global arguments: %v", args)
	tools.Log.Debug().Msgf("command arguments: %v %v", cmd, cmdArgs)
	RunCommand(cmd, cmdArgs, version)
	tools.Log.Debug().Msgf("done")
}

// RunCommand runs a specific command and the provided arguments
func RunCommand(c string, args []string, version string) {
	argv := append([]string{c}, args...)
	switch c {
	case "config":
		if err := config.Cmd(argv, version); err != nil {
			fmt.Printf("command failed:. err: %v\n", err)
		}
	case "events":
		entry, err := config.GetEntry("")
		if err != nil {
			fmt.Printf("config.getEntry err: %v\n", err)
			return
		}
		if err := events.Cmd(entry, argv, version); err != nil {
			fmt.Printf("command failed err = %v\n", err)
		}
	default:
		tools.Log.Debug().Msgf("%s is an unsupported command", c)
	}
}
