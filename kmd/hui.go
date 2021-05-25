package kmd

import (
	"github.com/aardlabs/terminal-poc/snippet"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/aardlabs/terminal-poc/tui"
	"github.com/spf13/cobra"
)

func NewCmdHui(gCtx *snippet.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Launch a Human friendly UI to execute",
		Long: examplef(`
              {AppName} run <name>.
        `),
		Example: examplef(`
            To run a snippet, run:
              $ {AppName} run <name>`),
		Args: cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return IsUserLoggedIn(gCtx.Config)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			tools.Log.Info().Msgf("run name=%s", name)
			return tui.LaunchUI(gCtx, name)
		},
	}

	return cmd
}
