package kmd

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/spf13/cobra"
)

const AppName = "aardy"

type VersionInfo struct {
	Version    string
	CommitHash string
	BuildTime  string
}

func NewCmdRoot(cfg *config.Config, versionInfo *VersionInfo) *cobra.Command {
	var rootCmd = &cobra.Command{
		Version:      versionInfo.Version,
		Use:          AppName,
		Short:        "Work seamlessly with the aard service from the command line",
		SilenceUsage: true,
	}
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display the program version",
		RunE: func(cmd *cobra.Command, args []string) error {
			tools.LogStdout(fmt.Sprintf("%s version %s (%s at %s)",
				rootCmd.Name(), rootCmd.Version, versionInfo.CommitHash, versionInfo.BuildTime))
			return nil
		},
	}
	gCtx := NewGraphContext(cfg, versionInfo.Version)
	rootCmd.AddCommand(NewCmdSnippetList(gCtx))
	rootCmd.AddCommand(NewCmdSnippetSearch(gCtx))
	rootCmd.AddCommand(NewCmdSnippetDesc(gCtx))
	rootCmd.AddCommand(NewCmdSnippetSave(gCtx))
	rootCmd.AddCommand(NewCmdSnippetEdit(gCtx))
	rootCmd.AddCommand(NewCmdAuth(cfg))
	rootCmd.AddCommand(NewCmdConfig(cfg))
	rootCmd.AddCommand(NewCmdCompletion())
	rootCmd.AddCommand(NewCmdExecutor())
	rootCmd.AddCommand(versionCmd)
	return rootCmd
}

func Execute(cfg *config.Config, versionInfo *VersionInfo) error {
	rootCmd := NewCmdRoot(cfg, versionInfo)
	return rootCmd.Execute()
}

func NewGraphContext(cfg *config.Config, version string) *snippet.Context {
	return snippet.NewContext(cfg, fmt.Sprintf("TermConsole:%s", version))
}

func examplef(format string, args ...string) string {
	args = append(args, "{AppName}", AppName)
	r := strings.NewReplacer(args...)
	return heredoc.Doc(r.Replace(format))
}
