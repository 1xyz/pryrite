package main

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/events"
	"github.com/aardlabs/terminal-poc/history"
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
	usage := `usage: pruney [--version] [--verbose] [--help] <command> [<args>...]
options:
   -h, --help    Show this message.
   --verbose     Enable verbose logging (logfile: $HOME/.pruney/pruney.log).

The commands are:
   config        provides options to configure pruney .
   history       work with local shell command history .
   log           work with events from the remote pruney log service.

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
	if tools.OptsBool(args, "--verbose") {
		level = zerolog.DebugLevel
	}
	tools.InitLogger(logFp, level)

	cmd := args["<command>"].(string)
	cmdArgs := args["<args>"].([]string)
	tools.Log.Debug().Msgf("global arguments: %v", args)
	tools.Log.Debug().Msgf("command arguments: %v %v", cmd, cmdArgs)
	ec := RunCommand(cmd, cmdArgs, version)
	if ec > 0 {
		tools.Log.Warn().Msgf("command exited with %d", ec)
		os.Exit(ec)
	}
	tools.Log.Debug().Msgf("done")
}

// RunCommand runs a specific command and the provided arguments
// Returns back an integer which is used as the exit code
// Ref: https://tldp.org/LDP/abs/html/exitcodes.html
func RunCommand(cmd string, args []string, version string) int {
	argv := append([]string{cmd}, args...)
	switch cmd {
	case "config":
		if err := config.Cmd(argv, version); err != nil {
			return logErr(cmd, err)
		}
	case "log":
		entry, err := config.GetEntry("")
		if err != nil {
			return logErr(cmd, err)
		}
		if err := events.Cmd(entry, argv, version); err != nil {
			return logErr(cmd, err)
		}
	case "history":
		entry, _ := config.GetEntry("")
		if err := history.Cmd(entry, argv, version); err != nil {
			return logErr(cmd, err)
		}
	default:
		tools.Log.Warn().Msgf("%s is an unsupported command", cmd)
		return 127 // command not found
	}
	return 0
}

func logErr(cmd string, err error) int {
	fmt.Printf("command failed:. err: %v\n", err)
	tools.Log.Err(err).Msgf("command %v failed", cmd)
	return 1
}
