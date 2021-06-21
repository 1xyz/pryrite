package kmd

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/aardlabs/terminal-poc/app"
	"github.com/aardlabs/terminal-poc/config"
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/aardlabs/terminal-poc/update"
)

func NewCmdRoot(cfg *config.Config) *cobra.Command {
	var rootCmd = &cobra.Command{
		Version:      app.Version,
		Use:          app.Name,
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
	updateCheck := false
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: examplef("Install the latest version of {AppName}"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if updateCheck {
				if !update.Check(cfg, true) {
					tools.LogStdout("The latest version is installed")
				}
			} else {
				result, err := update.GetLatest(cfg)
				if err != nil {
					return err
				}
				tools.LogStdout(fmt.Sprintf("Done! The latest version is now installed: %s", result))
			}
			return nil
		},
	}
	updateCmd.Flags().BoolVarP(&updateCheck, "check", "",
		updateCheck, "check for updates")
	gCtx := NewGraphContext(cfg)
	rootCmd.AddCommand(NewCmdSnippetList(gCtx))
	rootCmd.AddCommand(NewCmdSnippetSearch(gCtx))
	rootCmd.AddCommand(NewCmdSnippetDesc(gCtx))
	rootCmd.AddCommand(NewCmdSnippetSave(gCtx))
	rootCmd.AddCommand(NewCmdSnippetEdit(gCtx))
	rootCmd.AddCommand(NewCmdHui(gCtx))
	rootCmd.AddCommand(NewCmdRepl(gCtx))
	rootCmd.AddCommand(NewCmdAuth(cfg))
	rootCmd.AddCommand(NewCmdConfig(cfg))
	rootCmd.AddCommand(NewCmdCompletion())
	rootCmd.AddCommand(NewCmdExecutor())
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)

	return rootCmd
}

func Execute(cfg *config.Config) error {
	rootCmd := NewCmdRoot(cfg)
	return rootCmd.Execute()
}

func NewGraphContext(cfg *config.Config) *snippet.Context {
	return snippet.NewContext(cfg, fmt.Sprintf("TermConsole:%s", app.Version))
}

func examplef(format string, args ...string) string {
	args = append(args, "{AppName}", app.Name)
	r := strings.NewReplacer(args...)
	return heredoc.Doc(r.Replace(format))
}
