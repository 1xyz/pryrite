package kmd

import (
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type ConfigAddOptions struct {
	ServiceURL string // Reference to the Service's URL
}

func NewCmdConfigAdd() *cobra.Command {
	opts := &ConfigAddOptions{}
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new named configuration",
		Example: heredoc.Doc(`
			To add a new named configuration, run:

			   $ aard config add my-config --service-url https://aard.app
		`),
		Args: MinimumArgs(1, "could not add configuration: no name provided"),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(args[0])
			fmt.Println(opts.ServiceURL)
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.ServiceURL,
		"service-url", "https://aard.app",
		"URL for the aard service")
	return cmd
}

func NewCmdConfigList() *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,

		Use:     "list",
		Short:   "Lists existing named configurations.",
		Aliases: []string{"ls"},
		Example: heredoc.Doc(`
			To list all available configurations, run::

			   $ aard config list
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("listing")
			return nil
		},
	}
	return cmd
}

func NewCmdConfigDelete() *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,

		Use:     "delete <name>",
		Short:   "Deletes a named configuration.",
		Aliases: []string{"remove", "rm"},
		Long: heredoc.Doc(`
            Deletes a named configuration. You cannot delete a configuration that is
            active,  To delete the current active configuration, first run 
            aard config activate another one.
        `),
		Example: heredoc.Doc(`
			To delete a named configuration, run:

               $ aard config delete my_config
		`),
		Args: MinimumArgs(1, "could not delete configuration: no name provided"),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("listing")
			return nil
		},
	}
	return cmd
}

func NewCmdConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage the set of aard named configurations.",
	}
	cmd.AddCommand(NewCmdConfigAdd())
	cmd.AddCommand(NewCmdConfigList())
	cmd.AddCommand(NewCmdConfigDelete())
	return cmd
}
