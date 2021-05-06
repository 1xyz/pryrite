package kmd

import (
	"github.com/aardlabs/terminal-poc/config"
	"github.com/spf13/cobra"
)

func NewCmdRoot(cfg *config.Config) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "aard",
		Short: "Work seamlessly with the aard service from the command line",
	}
	rootCmd.AddCommand(NewCmdSnippetList())
	rootCmd.AddCommand(NewCmdAuth())
	rootCmd.AddCommand(NewCmdConfig(cfg))
	rootCmd.AddCommand(completionCmd)
	return rootCmd
}

func Execute(cfg *config.Config) error {
	rootCmd := NewCmdRoot(cfg)
	return rootCmd.Execute()
}
