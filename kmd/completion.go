package kmd

import (
	"errors"
	"os"

	"github.com/aardlabs/terminal-poc/tools"

	"github.com/spf13/cobra"
)

func NewCmdCompletion() *cobra.Command {
	var shellType string

	// completionCmd represents the completion command
	completionCmd := &cobra.Command{
		Hidden: true,
		Use:    "completion [bash|zsh|fish|powershell]",
		Short:  "Generate completion script",
		Long: examplef(`
Generate shell completion scripts for the CLI commands.

When installing the CLI through a package manager, it's possible that
no additional shell configuration is necessary to gain completion support.
For Homebrew, see https://docs.brew.sh/Shell-Completion

If you need to set up completions manually, follow the instructions below. The exact
config file locations might vary based on your system. Make sure to restart your
shell before testing whether completions are working.:
Bash:

  $ source <({AppName} completion -s bash)

  # To load completions for each session, execute once:
  # Linux:
  $ {AppName} completion -s bash > /etc/bash_completion.d/{AppName}
  # macOS:
  $ {AppName} completion -s bash > /usr/local/etc/bash_completion.d/{AppName}

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ {AppName} completion -s zsh > "${fpath[1]}/_{AppName}"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ {AppName} completion -s fish | source

  # To load completions for each session, execute once:
  $ {AppName} completion -s fish > ~/.config/fish/completions/{AppName}.fish

PowerShell:

  PS> {AppName} completion -s powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> {AppName} completion -s powershell > {AppName}.ps1
  # and source this file from your PowerShell profile.
`),
		DisableFlagsInUseLine: false,
		Args:                  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if shellType == "" {
				return &FlagError{Err: errors.New("error: the value for `--shell` is required")}
			}

			switch shellType {
			case "bash":
				if err := cmd.Root().GenBashCompletion(os.Stdout); err != nil {
					tools.LogStderrExit(err, "GenBashCompletion")
				}
			case "zsh":
				if err := cmd.Root().GenZshCompletion(os.Stdout); err != nil {
					tools.LogStderrExit(err, "GenZshCompletion")
				}
			case "fish":
				if err := cmd.Root().GenFishCompletion(os.Stdout, true); err != nil {
					tools.LogStderrExit(err, "GenFishCompletion")
				}
			case "powershell":
				if err := cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout); err != nil {
					tools.LogStderrExit(err, "GenPowerShellCompletionWithDesc")
				}
			}
			return nil
		},
	}

	completionCmd.Flags().StringVarP(&shellType, "shell",
		"s", "", "Shell type: {bash|zsh|fish|powershell}")

	return completionCmd
}
