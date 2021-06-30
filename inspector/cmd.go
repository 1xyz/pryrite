package inspector

import (
	"context"
	"strings"
	"time"

	"github.com/aardlabs/terminal-poc/app"
	executor "github.com/aardlabs/terminal-poc/executors"

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
	rootCmd.AddCommand(NewCmdExecutor(b.runner.Register))
	rootCmd.AddCommand(newWhereAmICmd(b))
	rootCmd.AddCommand(newActionCmd(b, "quit", []string{"q", "exit"}, "Quit this session"))
	return rootCmd
}

func newRunCmd(b *codeBlock) *cobra.Command {
	return &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "Run this code block",
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

var localRegister *executor.Register

func NewCmdExecutor(register *executor.Register) *cobra.Command {
	var timeout time.Duration
	var disablePTY bool
	var startPrompt string
	var usePrompt string
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute an ad-hoc command",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Try to use the current register being run, otherwise, use our own local one.
			// This is important for finding prompts.
			if register == nil {
				register = localRegister
				if register == nil {
					var err error
					register, err = executor.NewRegister()
					if err != nil {
						return err
					}

					localRegister = register
				}
			}

			// FIXME: use a shared/common register between these and runs
			req := executor.DefaultRequest()

			content := strings.Join(args, " ")
			req.Content = []byte(content)
			var err error
			req.ContentType, err = executor.Parse("text/bash")
			if err != nil {
				return err
			}

			if startPrompt != "" {
				req.ContentType.Params["prompt-assign"] = startPrompt
			} else if usePrompt != "" {
				req.ContentType.Params["prompt"] = usePrompt
			}

			var ctx context.Context
			if timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), timeout)
				defer cancel()
			} else {
				ctx = context.Background()
			}

			res := register.Execute(ctx, req)
			if res.Err != nil {
				cmd.PrintErrf("exit> execution failed: %v\n", res.Err)
			} else {
				cmd.Printf("exit> status: %d\n", res.ExitStatus)
			}

			return nil
		},
	}
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 0,
		"Wait some amount of time before giving up on a command to return")
	cmd.Flags().BoolVarP(&disablePTY, "disable-pty", "T", false,
		"Disable psuedo-terminal allocation")
	cmd.Flags().StringVarP(&startPrompt, "start", "s", "",
		"Start this command as the named prompt")
	cmd.Flags().StringVarP(&usePrompt, "use", "u", "",
		"Use a running prompt to execute this command")

	return cmd
}
