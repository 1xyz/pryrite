package main

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/kmd"
	"github.com/aardlabs/terminal-poc/tools"
	"os"
)

func main() {
	verbose := false
	wr, err := tools.OpenLogger(verbose)
	if err != nil {
		tools.Log.Fatal().Err(err).Msgf("tools.OpenLogger")
	}
	defer func() {
		if err := wr.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "wr.close %v", err)
		}
	}()

	if err := config.CreateDefaultConfigIfEmpty(); err != nil {
		tools.LogErrExit("CreateDefaultConfig", err)
	}
	cfg, err := config.Default()
	if err != nil {
		tools.LogErrExit("config.Default", err)
	}
	if err := kmd.Execute(cfg); err != nil {
		tools.LogErrExit("command failed", err)
	}
}

//
//func main() {
//	usage := `usage: aard [--version] [--verbose] [--help] <command> [<args>...]
//options:
//   -h, --help    Show this message.
//   --verbose     Enable verbose logging (logfile: $HOME/.aard/aard.log).
//
//The commands are:
//   auth          authenticate with the remote service.
//   config        provides options to configure aard.
//   history       work with your local shell history.
//   log           add & view snippets from the aard service.
//   termcast      record/play an asciicast from your terminal.
//
//See 'aard <command> --help' for more information on a specific command.
//
//These are common aard commands used in various situations:
//
//Configure and get started:
//   # Setup up aard to talk to service
//   $ aard config add remote --service-url https://flaming-fishtoot.herokuapp.com/
//
//   # Login into the user via the specified e-mail
//   $ aard auth login alan@turing.me
//
//Examine the snippet log:
//   # List the most recent snippets from the log.
//   aard log
//
//   # Search the log for snippets involving cerbot
//   aard log search cerbot
//
//See 'aard <command> --help' for more information on a specific command.
//`
//	parser := &docopt.Parser{OptionsFirst: true}
//	args, err := parser.ParseArgs(usage, nil, version)
//	if err != nil {
//		fmt.Printf("parse error = %v", err)
//	}
//
//	wr, err := tools.OpenLogger(tools.OptsBool(args, "--verbose"))
//	if err != nil {
//		tools.Log.Fatal().Err(err).Msgf("tools.OpenLogger")
//	}
//	defer func() {
//		if err := wr.Close(); err != nil {
//			fmt.Printf("wr.close %v", err)
//		}
//	}()
//
//	c := args["<command>"].(string)
//	cArgs := args["<args>"].([]string)
//	tools.Log.Debug().Msgf("global arguments: %v", args)
//	tools.Log.Debug().Msgf("c arguments: %v %v", c, cArgs)
//	p := cmd.Params{
//		Agent:   "TermConsole",
//		Version: version,
//		Argv:    append([]string{c}, cArgs...),
//		Command: c,
//	}
//	ec := RunCommand(&p)
//	if ec > 0 {
//		tools.Log.Warn().Msgf("c exited with %d", ec)
//		os.Exit(ec)
//	}
//	tools.Log.Debug().Msgf("done")
//}
//
//type cmdFunc func(*config.Entry, *cmd.Params) error
//
//var cmdFunctions = map[string]cmdFunc{
//	"auth":     cmd.AuthCmd,
//	"termcast": cmd.CaptureCmd,
//	"log":      cmd.LogCmd,
//}
//
//// RunCommand runs a specific command and the provided arguments
//// Returns back an integer which is used as the exit code
//// Ref: https://tldp.org/LDP/abs/html/exitcodes.html
//func RunCommand(params *cmd.Params) int {
//	if params.Command == "config" {
//		if err := cmd.ConfigCmd(params); err != nil {
//			return cmd.LogErr(params.Command, err)
//		}
//	} else if params.Command == "history" {
//		// history should continue to work if there is not config
//		entry, _ := config.GetEntry("")
//		if err := cmd.HistoryCmd(entry, params); err != nil {
//			return cmd.LogErr(params.Command, err)
//		}
//	} else {
//		// Try creating a default config, if not found
//		if err := config.CreateDefaultConfigIfEmpty(); err != nil {
//			return cmd.LogErr(params.Command, err)
//		}
//		fn, found := cmdFunctions[params.Command]
//		if !found {
//			fmt.Printf("command %s not found!\n", params.Command)
//			tools.Log.Warn().Msgf("%s is an unsupported command", params.Command)
//			return 127 // command not found
//		}
//		entry, err := config.GetEntry("")
//		if err != nil {
//			return cmd.LogErr(params.Command, err)
//		}
//		if err := fn(entry, params); err != nil {
//			return cmd.LogErr(params.Command, err)
//		}
//	}
//	return 0
//}
