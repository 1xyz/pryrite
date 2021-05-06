/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package kmd

import (
	"github.com/aardlabs/terminal-poc/tools"
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

  $ source <(kobra completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ kobra completion bash > /etc/bash_completion.d/kobra
  # macOS:
  $ kobra completion bash > /usr/local/etc/bash_completion.d/kobra

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ kobra completion zsh > "${fpath[1]}/_kobra"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ kobra completion fish | source

  # To load completions for each session, execute once:
  $ kobra completion fish > ~/.config/fish/completions/kobra.fish

PowerShell:

  PS> kobra completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> kobra completion powershell > kobra.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
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
	},
}
