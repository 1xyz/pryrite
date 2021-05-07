package kmd

import (
	"errors"
	"fmt"
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

func IsUserLoggedIn(cfg *config.Config) error {
	entry, found := cfg.GetDefaultEntry()
	if !found {
		return fmt.Errorf("no named entry found in configuration")
	}
	_, found = auth.GetLoggedInUser(entry)
	if !found {
		return fmt.Errorf("no user is logged in. See: aard auth login --help")
	}
	return nil
}
