package kmd

import (
	"errors"
	"github.com/MakeNowJust/heredoc"
	"github.com/aardlabs/terminal-poc/tools"
	"os"

	"github.com/spf13/cobra"
)

func NewCmdCompletion() *cobra.Command {
	var shellType string

	// completionCmd represents the completion command
	completionCmd := &cobra.Command{
		Hidden: true,
		Use:    "completion [bash|zsh|fish|powershell]",
		Short:  "Generate completion script",
		Long: heredoc.Docf(`
Generate shell completion scripts for the CLI commands.

When installing the CLI through a package manager, it's possible that
no additional shell configuration is necessary to gain completion support.
For Homebrew, see https://docs.brew.sh/Shell-Completion

If you need to set up completions manually, follow the instructions below. The exact
config file locations might vary based on your system. Make sure to restart your
shell before testing whether completions are working.:
Bash:

  $ source <(aard completion -s bash)

  # To load completions for each session, execute once:
  # Linux:
  $ aard completion -s bash > /etc/bash_completion.d/aard
  # macOS:
  $ aard completion -s bash > /usr/local/etc/bash_completion.d/aard

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ aard completion -s zsh > "${fpath[1]}/_aard"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ aard completion -s fish | source

  # To load completions for each session, execute once:
  $ aard completion -s fish > ~/.config/fish/completions/aard.fish

PowerShell:

  PS> aard completion -s powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> aard completion -s powershell > aard.ps1
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
					tools.LogStderrExit("GenBashCompletion", err)
				}
			case "zsh":
				if err := cmd.Root().GenZshCompletion(os.Stdout); err != nil {
					tools.LogStderrExit("GenZshCompletion", err)
				}
			case "fish":
				if err := cmd.Root().GenFishCompletion(os.Stdout, true); err != nil {
					tools.LogStderrExit("GenFishCompletion", err)
				}
			case "powershell":
				if err := cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout); err != nil {
					tools.LogStderrExit("GenPowerShellCompletionWithDesc", err)
				}
			}
			return nil
		},
	}

	completionCmd.Flags().StringVarP(&shellType, "shell",
		"s", "", "Shell type: {bash|zsh|fish|powershell}")

	return completionCmd
}
