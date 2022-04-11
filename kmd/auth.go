package kmd

import (
	"fmt"

	"github.com/1xyz/pryrite/auth"
	"github.com/1xyz/pryrite/config"
	"github.com/1xyz/pryrite/tools"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func NewCmdAuthLogin(cfg *config.Config) *cobra.Command {
	serviceURL := ""
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login an authorized user to the aard service.",
		Long:  heredoc.Doc(`Authorizes the user to access the service.`),
		Example: tools.Examplef(`
			In order to login to this service, run:

              $ {AppName} auth login

            In order to view the current logged in user, run:

              $ {AppName} config list
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tools.Log.Info().Msgf("auth login serviceURL=%s", serviceURL)
			entry, found := cfg.GetDefaultEntry()
			if !found {
				return fmt.Errorf("a active configuration is not found")
			}
			if err := auth.AuthUser(entry); err != nil {
				return err
			}
			tools.LogStdout(fmt.Sprintf("User logged in as %s", entry.Email))
			return nil
		},
	}
	return cmd
}

func NewCmdAuthLogout(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout an authorized user",
		Example: tools.Examplef(`
			To logout the current logged in user, run:

              $ {AppName} auth logout

            In order to view the current logged in user, run:

              $ {AppName} config list
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
			tools.LogStdout("user logged out")
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
