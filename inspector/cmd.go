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
	rootCmd.AddCommand(newNilCmd("next", "Navigate to the next code block"))
	rootCmd.AddCommand(newNilCmd("prev", "Navigate to the previous code block"))
	rootCmd.AddCommand(newRunCmd(b))
	rootCmd.AddCommand(newWhereAmICmd(b))
	rootCmd.AddCommand(newNilCmd("quit", "Quit this session"))
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

func newNilCmd(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
}
