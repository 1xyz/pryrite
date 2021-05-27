package kmd

import (
	"errors"

	"github.com/aardlabs/terminal-poc/auth"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/spf13/cobra"
)

func MinimumArgs(n int, msg string) cobra.PositionalArgs {
	if msg == "" {
		return cobra.MinimumNArgs(1)
	}

	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return &FlagError{Err: errors.New(msg)}
		}
		return nil
	}
}

func IsUserLoggedIn(entry *config.Entry) error {
	_, found := auth.GetLoggedInUser(entry)
	if !found {
		return auth.AuthUser(entry, "")
	}
	return nil
}
