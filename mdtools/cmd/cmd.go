package cmd

import (
	"fmt"
	"github.com/1xyz/pryrite/mdtools/markdown"
	"github.com/1xyz/pryrite/tools"
	"github.com/spf13/cobra"
)

func NewCmdRoot() *cobra.Command {
	var rootCmd = &cobra.Command{
		Version:      "1.0",
		Use:          "aardy-mdtools",
		Short:        "aardy is a markdown executor",
		SilenceUsage: true,
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display the program version",
		RunE: func(cmd *cobra.Command, args []string) error {
			tools.LogStdout("version a")
			return nil
		},
	}

	var execCmd = &cobra.Command{
		Use:   "execute",
		Short: "Execute a markdown file",
		Args:  minArgs(1, "You need to specify a search query"),
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
			return fmt.Errorf("missing arguments, requires %d args got %d", n, len(args))
		}
		return nil
	}
}
