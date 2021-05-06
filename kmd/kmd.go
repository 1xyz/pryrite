package kmd

import (
	"github.com/aardlabs/terminal-poc/config"
	"github.com/spf13/cobra"
)

func NewCmdRoot() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "aard",
		Short: "Work seamlessly with the aard service from the command line",
	}
	rootCmd.AddCommand(NewCmdSnippetList())
	rootCmd.AddCommand(NewCmdAuth())
	rootCmd.AddCommand(NewCmdConfig())
	rootCmd.AddCommand(completionCmd)
	return rootCmd
}

func Execute(cfg *config.Config) error {
	rootCmd := NewCmdRoot()
	return rootCmd.Execute()
}
