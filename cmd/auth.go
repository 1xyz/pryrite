package cmd

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/auth"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
)

func AuthCmd(entry *config.Entry, params *Params) error {
	usage := `The "auth" command provides options to authenticate with the remote service

usage: aard auth login <email>

Options:
   -h --help            Show this screen.

Examples:
   # Login into the user via the specified e-mail
   $ aard auth login alan@turing.me
`
	opts, err := docopt.ParseArgs(usage, params.Argv, params.Version)
	if err != nil {
		tools.Log.Fatal().Msgf("error parsing arguments. err=%v", err)
	}

	if tools.OptsBool(opts, "login") {
		email := tools.OptsStr(opts, "<email>")
		if err := auth.AuthUser(entry, email); err != nil {
			return err
		}
		fmt.Printf("Login successful!\n")
	}
	return nil
}
