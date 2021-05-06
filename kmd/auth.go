package kmd

import (
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func NewCmdAuthLogin() *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,

		Use:   "login <email>",
		Short: "Login an authorized user to the aard service",
		Long:  heredoc.Doc(`Authorizes the user to access the service via credentials`),
		Example: heredoc.Doc(`
			In order to login to this service, run:

              $ aard auth login alan@turing.me
		`),
		Args: MinimumArgs(1, "could not login: no email provided"),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(args[0])
			return nil
		},
	}
	return cmd
}

func NewCmdAuthLogout() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout an authorized user",
		Example: heredoc.Doc(`
			To logout the current logged in user, run: 

              $ aard auth logout
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Logged out")
			return nil
		},
	}
	return cmd
}

func NewCmdAuth() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage credentials for a user",
	}
	cmd.AddCommand(NewCmdAuthLogin())
	cmd.AddCommand(NewCmdAuthLogout())
	return cmd
}
