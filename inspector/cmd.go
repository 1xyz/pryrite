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
	rootCmd.AddCommand(newNilCmd("next", " navigate to the next code block"))
	rootCmd.AddCommand(newNilCmd("prev", " navigate to the previous code block"))
	rootCmd.AddCommand(newNilCmd("quit", "quit this session"))
	return rootCmd
}

func newNilCmd(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
}
