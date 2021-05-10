package kmd

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/spf13/cobra"
)

func NewCmdRoot(cfg *config.Config, version string) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:          "aard",
		Short:        "Work seamlessly with the aard service from the command line",
		SilenceUsage: true,
	}
	gCtx := NewGraphContext(cfg, version)
	rootCmd.AddCommand(NewCmdSnippetList(gCtx))
	rootCmd.AddCommand(NewCmdSnippetDesc(gCtx))
	rootCmd.AddCommand(NewCmdSnippetAdd(gCtx))
	rootCmd.AddCommand(NewCmdAuth(cfg))
	rootCmd.AddCommand(NewCmdConfig(cfg))
	rootCmd.AddCommand(completionCmd)

	return rootCmd
}

func Execute(cfg *config.Config, version string) error {
	rootCmd := NewCmdRoot(cfg, version)
	return rootCmd.Execute()
}

func NewGraphContext(cfg *config.Config, version string) *snippet.Context {
	return snippet.NewContext(cfg, fmt.Sprintf("TermConsole:%s", version))
}
