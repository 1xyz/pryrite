package auth

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/cmd"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
	"regexp"
)

var (
	emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

func Cmd(entry *config.Entry, params *cmd.Params) error {
	usage := `The "auth" command provides options to authenticate with the remote service

usage: pruney auth login <email>

Options:
   -h --help            Show this screen.

Examples:
   # Login into the user via the specified e-mail
   $ pruney auth login alan@turing.me
`
	opts, err := docopt.ParseArgs(usage, params.Argv, params.Version)
	if err != nil {
		tools.Log.Fatal().Msgf("error parsing arguments. err=%v", err)
	}

	if tools.OptsBool(opts, "login") {
		email := tools.OptsStr(opts, "<email>")
		// Ref: validate email -- https://golangcode.com/validate-an-email-address/
		if !isEmailValid(email) {
			return fmt.Errorf("email %v is not valid", email)
		}

		entry.User = email
		entry.AuthScheme = "Silly"
		if err := config.SetEntry(entry); err != nil {
			return nil
		}
		fmt.Printf("Login successful!\n")
	}
	return nil
}

// isEmailValid checks if the email provided passes the required structure and length.
func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}
