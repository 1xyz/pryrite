package inspector

import (
	"github.com/aardlabs/terminal-poc/app"
	"github.com/spf13/cobra"
)

func newRootCmd(b *codeBlock) *cobra.Command {
	var rootCmd = &cobra.Command{
		Version: app.Version,
		Use:     "",
	}
	// NOTE: these are actually handled/processed over in codeBlock.handleExitCmd()
	rootCmd.AddCommand(newActionCmd(b, "next", []string{"n"}, "Navigate to the next code block"))
	rootCmd.AddCommand(newActionCmd(b, "prev", []string{"p"}, "Navigate to the previous code block"))
	rootCmd.AddCommand(newActionCmd(b, "jump", []string{"j"}, "Switch to another code block"))
	rootCmd.AddCommand(newRunCmd(b))
	rootCmd.AddCommand(newWhereAmICmd(b))
	rootCmd.AddCommand(newActionCmd(b, "quit", []string{"q", "exit"}, "Quit this session"))
	return rootCmd
}

func newRunCmd(b *codeBlock) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run this code block",
		RunE: func(cmd *cobra.Command, args []string) error {
			b.RunBlock()
			return nil
		},
	}
}

func newWhereAmICmd(b *codeBlock) *cobra.Command {
	return &cobra.Command{
		Use:   "whereami",
		Short: "Show content surrounding the current context",
		RunE: func(cmd *cobra.Command, args []string) error {
			b.WhereAmI()
			return nil
		},
	}
}

func newActionCmd(b *codeBlock, use string, aliases []string, short string) *cobra.Command {
	return &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   short,
		RunE: func(cmd *cobra.Command, args []string) error {
			b.SetExitAction(NewBlockAction(use))
			return nil
		},
	}
}
