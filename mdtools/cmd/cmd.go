package cmd

import (
	"fmt"
	"github.com/1xyz/pryrite/app"
	"github.com/1xyz/pryrite/mdtools/markdown"
	"github.com/1xyz/pryrite/tools"
	"github.com/spf13/cobra"
)

func NewCmdRoot() *cobra.Command {
	var rootCmd = &cobra.Command{
		Version:      app.Version,
		Use:          app.Name,
		Short:        fmt.Sprintf("%s is a markdown executor", app.Name),
		SilenceUsage: true,
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display the program version",
		RunE: func(cmd *cobra.Command, args []string) error {
			tools.LogStdout(app.Version)
			return nil
		},
	}

	var execCmd = &cobra.Command{
		Use:   "open",
		Short: "open a markdown file to inspect",
		Args:  minArgs(1, "You need to specify a local or http(s) URL to a markdown file"),
		Example: fmt.Sprintf(" %s open _examples/hello_world.md\n %s open https://raw.githubusercontent.com/1xyz/pryrite/main/_examples/hello-world.md\n",
			app.Name, app.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			tools.LogStdout("execute filename=%s\n", args[0])
			filename := args[0]
			return markdown.MDFileInspect(filename)
		},
	}

	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(versionCmd)
	return rootCmd
}

func Execute() error {
	rootCmd := NewCmdRoot()
	return rootCmd.Execute()
}

func minArgs(n int, msg string) cobra.PositionalArgs {
	if msg == "" {
		return cobra.MinimumNArgs(1)
	}
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return fmt.Errorf("missing arguments\n%s", msg)
		}
		return nil
	}
}
