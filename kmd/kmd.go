package kmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/1xyz/pryrite/app"
	"github.com/1xyz/pryrite/config"
	"github.com/1xyz/pryrite/inspector"
	"github.com/1xyz/pryrite/shells"
	"github.com/1xyz/pryrite/snippet"
	"github.com/1xyz/pryrite/tools"
)

func NewCmdRoot(cfg *config.Config) *cobra.Command {
	var rootCmd = &cobra.Command{
		Version:      app.Version,
		Use:          app.UsageName,
		Short:        "Work seamlessly with the aardy service from the command line",
		SilenceUsage: true,
	}
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display the program version",
		RunE: func(cmd *cobra.Command, args []string) error {
			tools.LogStdout(fmt.Sprintf("%s version %s (%s at %s)",
				rootCmd.Name(), rootCmd.Version, app.CommitHash, app.BuildTime))
			return nil
		},
	}

	gCtx := NewGraphContext(cfg)
	rootCmd.AddCommand(NewCmdSnippetList(gCtx))
	rootCmd.AddCommand(NewCmdSnippetSearch(gCtx))
	rootCmd.AddCommand(NewCmdSnippetDesc(gCtx))
	rootCmd.AddCommand(NewCmdSnippetSave(gCtx))
	rootCmd.AddCommand(NewCmdSnippetSlurp(gCtx))
	rootCmd.AddCommand(NewCmdSnippetEdit(gCtx))
	rootCmd.AddCommand(NewCmdSnippetRun(gCtx))
	rootCmd.AddCommand(NewCmdAuth(cfg))
	rootCmd.AddCommand(NewCmdConfig(cfg))
	rootCmd.AddCommand(NewCmdCompletion())
	rootCmd.AddCommand(NewCmdRawExecutor())
	rootCmd.AddCommand(inspector.NewCmdExecutor(nil))
	rootCmd.AddCommand(shells.NewCmdInit())
	rootCmd.AddCommand(shells.NewCmdHistory())
	rootCmd.AddCommand(versionCmd)

	return rootCmd
}

func Execute(cfg *config.Config) error {
	rootCmd := NewCmdRoot(cfg)
	return rootCmd.Execute()
}

func NewGraphContext(cfg *config.Config) *snippet.Context {
	return snippet.NewContext(cfg, fmt.Sprintf("TermConsole:%s", app.Version))
}
