package kmd

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/spf13/cobra"
)

type ConfigAddOptions struct {
	ServiceURL string // Reference to the Service's URL
}

func NewCmdConfigAdd(cfg *config.Config) *cobra.Command {
	opts := &ConfigAddOptions{}
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new named configuration",
		Example: heredoc.Doc(`
			To add a new named configuration, run:

			   $ aard config add my-config --service-url https://aardy.app
		`),
		Args: MinimumArgs(1, "could not add configuration: no name provided"),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			tools.Log.Info().Msgf("Addconfig name=%s, serviceURL=%s",
				name, opts.ServiceURL)
			if err := cfg.Add(name, opts.ServiceURL); err != nil {
				return err
			}
			if err := cfg.SaveFile(config.DefaultConfigFile); err != nil {
				return err
			}
			tools.LogStdout(fmt.Sprintf("configuration added with name = %s", name))
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.ServiceURL,
		"service-url", "https://aardy.app",
		"URL for the aard service")
	return cmd
}

func NewCmdConfigList(cfg *config.Config) *cobra.Command {
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
			tr := &config.TableRender{Config: cfg}
			tr.Render()
			return nil
		},
	}
	return cmd
}

func NewCmdConfigActivate(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,

		Use:     "activate <name>",
		Short:   "Activates an existing named configuration.",
		Aliases: []string{"ls"},
		Example: heredoc.Doc(`
			To activate an existing named configuration, run::

			   $ aard config activate my-config

            To list all configurations, run::

			   $ aard config list
		`),
		Args: MinimumArgs(1, "could not activate configuration: no name provided"),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			tools.Log.Info().Msgf("config activate called name = %s", name)
			if err := cfg.SetDefault(name); err != nil {
				return err
			}
			if err := cfg.SaveFile(config.DefaultConfigFile); err != nil {
				return err
			}
			tools.LogStdout(fmt.Sprintf("Configuration %s is active", name))
			return nil
		},
	}
	return cmd
}

func NewCmdConfigDelete(cfg *config.Config) *cobra.Command {
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
			name := args[0]
			tools.Log.Info().Msgf("config delete name=%s", name)
			entry, found := cfg.Get(name)
			if !found {
				return fmt.Errorf("no configuration entry found with name = %s", name)
			}
			if entry.Name == cfg.DefaultEntry {
				return fmt.Errorf("the specified configuration entry is active and cannot be deleted")
			}
			if err := cfg.Del(name); err != nil {
				return err
			}
			if err := cfg.SaveFile(config.DefaultConfigFile); err != nil {
				return err
			}
			tools.LogStdout("configuration deleted!")
			return nil
		},
	}
	return cmd
}

func NewCmdConfig(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage the set of aard named configurations",
	}
	cmd.AddCommand(NewCmdConfigAdd(cfg))
	cmd.AddCommand(NewCmdConfigList(cfg))
	cmd.AddCommand(NewCmdConfigDelete(cfg))
	cmd.AddCommand(NewCmdConfigActivate(cfg))
	return cmd
}
