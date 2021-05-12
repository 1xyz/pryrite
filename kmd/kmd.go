package kmd

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/spf13/cobra"
)

func NewCmdRoot(cfg *config.Config, version string) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:          "aard",
		Short:        "Work seamlessly with the aard service from the command line",
		SilenceUsage: true,
	}
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display the program version",
		RunE: func(cmd *cobra.Command, args []string) error {
			tools.LogStdout(version)
			return nil
		},
	}
	gCtx := NewGraphContext(cfg, version)
	rootCmd.AddCommand(NewCmdSnippetList(gCtx))
	rootCmd.AddCommand(NewCmdSnippetSearch(gCtx))
	rootCmd.AddCommand(NewCmdSnippetDesc(gCtx))
	rootCmd.AddCommand(NewCmdSnippetSave(gCtx))
	rootCmd.AddCommand(NewCmdSnippetEdit(gCtx))
	rootCmd.AddCommand(NewCmdAuth(cfg))
	rootCmd.AddCommand(NewCmdConfig(cfg))
	rootCmd.AddCommand(NewCmdCompletion())
	rootCmd.AddCommand(versionCmd)
	return rootCmd
}

func Execute(cfg *config.Config, version string) error {
	rootCmd := NewCmdRoot(cfg, version)
	return rootCmd.Execute()
}

func NewGraphContext(cfg *config.Config, version string) *snippet.Context {
	return snippet.NewContext(cfg, fmt.Sprintf("TermConsole:%s", version))
}
