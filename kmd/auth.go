package kmd

import (
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/aardlabs/terminal-poc/auth"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/spf13/cobra"
)

func NewCmdAuthLogin(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,

		Use:   "login <email>",
		Short: "Login an authorized user to the aard service",
		Long:  heredoc.Doc(`Authorizes the user to access the service via credentials`),
		Example: heredoc.Doc(`
			In order to login to this service, run:

              $ aard auth login alan@turing.me

            In order to view the current logged in user, run:

              $ aard config list 
		`),
		Args: MinimumArgs(1, "could not login: no email provided"),
		RunE: func(cmd *cobra.Command, args []string) error {
			email := args[0]
			tools.Log.Info().Msgf("auth login email=%s", email)
			entry, found := cfg.GetDefaultEntry()
			if !found {
				return fmt.Errorf("a active configuration is not found")
			}
			if err := auth.AuthUser(entry, email); err != nil {
				return err
			}
			tools.LogStdout(fmt.Sprintf("user %s logged in", email))
			return nil
		},
	}
	return cmd
}

func NewCmdAuthLogout(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout an authorized user",
		Example: heredoc.Doc(`
			To logout the current logged in user, run: 

              $ aard auth logout

            In order to view the current logged in user, run:

              $ aard config list 
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			entry, found := cfg.GetDefaultEntry()
			if !found {
				return fmt.Errorf("a active configuration is not found")
			}
			if err := auth.LogoutUser(entry); err != nil {
				return err
			}
			tools.LogStdout(fmt.Sprintf("user logged out"))
			return nil
		},
	}
	return cmd
}

func NewCmdAuth(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage credentials for a user",
	}
	cmd.AddCommand(NewCmdAuthLogin(cfg))
	cmd.AddCommand(NewCmdAuthLogout(cfg))
	return cmd
}
