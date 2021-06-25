package inspector

import (
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
)

func NewCobraCommandCompleter(cmd *cobra.Command) *CobraCommandCompleter {
	return &CobraCommandCompleter{
		rootCmd: cmd,
	}
}

type CobraCommandCompleter struct {
	rootCmd *cobra.Command
}

func (c *CobraCommandCompleter) Complete(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}
	args := strings.Split(d.TextBeforeCursor(), " ")
	w := d.GetWordBeforeCursor()
	s := c.commandSuggestions(c.rootCmd, args, 0, w)
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func (c *CobraCommandCompleter) commandSuggestions(cmd *cobra.Command, args []string, index int, w string) []prompt.Suggest {
	if len(args) == 0 || index >= len(args) {
		return cmdSuggestions(cmd)
	}

	currentArg := args[index]
	childCmd, found := findCmd(cmd, currentArg)
	if !found {
		if strings.HasPrefix(w, "-") {
			return optSuggestions(cmd)
		}
		return cmdSuggestions(cmd)
	}
	return c.commandSuggestions(childCmd, args, index+1, w)
}

func optSuggestions(cmd *cobra.Command) []prompt.Suggest {
	flags := cmd.Flags()
	s := make([]prompt.Suggest, 0)
	flags.VisitAll(func(f *pflag.Flag) {
		s = append(s, prompt.Suggest{
			Text:        "-" + f.Name,
			Description: f.Usage,
		})
	})
	s = append(s, prompt.Suggest{Text: "--help", Description: "Help on command " + cmd.Name()})
	return s
}

func findCmd(cmd *cobra.Command, arg string) (*cobra.Command, bool) {
	commands := cmd.Commands()
	for i, c := range cmd.Commands() {
		if c.Hidden {
			continue
		}
		if c.Name() == arg {
			return commands[i], true
		}
	}
	return nil, false
}

func cmdSuggestions(cmd *cobra.Command) []prompt.Suggest {
	childCmds := cmd.Commands()
	var s []prompt.Suggest
	for _, childCmd := range childCmds {
		if childCmd.Hidden {
			continue
		}
		s = append(s, prompt.Suggest{
			Text:        childCmd.Name(),
			Description: childCmd.Short,
		})
	}
	return s
}
