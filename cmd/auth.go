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

usage: aard auth login [<service_url>]

Options:
   -h --help            Show this screen.

Examples:
   # Login the user with the default service:
   $ aard auth login
`
	opts, err := docopt.ParseArgs(usage, params.Argv, params.Version)
	if err != nil {
		tools.Log.Fatal().Msgf("error parsing arguments. err=%v", err)
	}

	if tools.OptsBool(opts, "login") {
		serviceUrl := tools.OptsStr(opts, "<service_url>")
		if err := auth.AuthUser(entry, serviceUrl); err != nil {
			return err
		}
		fmt.Printf("Login successful!\n")
	}
	return nil
}
