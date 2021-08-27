package inspector

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/aardlabs/terminal-poc/app"
	executor "github.com/aardlabs/terminal-poc/executors"
	"github.com/aardlabs/terminal-poc/historian"
	"github.com/aardlabs/terminal-poc/shells"

	"github.com/spf13/cobra"
)

func newRootCmd(n *NodeInspector) *cobra.Command {
	var rootCmd = &cobra.Command{
		Version: app.Version,
		Use:     "",
	}
	// NOTE: these are actually handled/processed over in codeBlock.handleExitCmd()
	rootCmd.AddCommand(newActionCmd(n, "next", []string{"n"}, "Navigate to the next code block"))
	rootCmd.AddCommand(newActionCmd(n, "prev", []string{"p"}, "Navigate to the previous code block"))
	rootCmd.AddCommand(newActionCmd(n, "jump", []string{"j"}, "Switch to another code block"))
	rootCmd.AddCommand(newRunCmd(n))
	rootCmd.AddCommand(NewCmdExecutor(n.runner.Register))
	rootCmd.AddCommand(newWhereAmICmd(n))
	rootCmd.AddCommand(newActionCmd(n, "quit", []string{"q", "exit"}, "Quit this session"))
	return rootCmd
}

func newRunCmd(n *NodeInspector) *cobra.Command {
	return &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		Short:   "Run this code block",
		RunE: func(cmd *cobra.Command, args []string) error {
			if n.currentBlock().RunBlock() {
				// Move to the next step if RunBlock signals a success
				n.processAction(&BlockAction{
					Action: BlockActionExecutionDone,
					Args:   []string{},
				})
			}
			return nil
		},
	}
}

func newWhereAmICmd(n *NodeInspector) *cobra.Command {
	return &cobra.Command{
		Use:   "whereami",
		Short: "Show content surrounding the current context",
		RunE: func(cmd *cobra.Command, args []string) error {
			n.currentBlock().WhereAmI()
			return nil
		},
	}
}

func newActionCmd(n *NodeInspector, use string, aliases []string, short string) *cobra.Command {
	return &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   short,
		RunE: func(cmd *cobra.Command, args []string) error {
			n.processAction(NewBlockAction(use))
			return nil
		},
	}
}

var localRegister *executor.Register

func NewCmdExecutor(register *executor.Register) *cobra.Command {
	var timeout time.Duration
	var disablePTY bool
	var ignoreWorkingDir bool
	var startPrompt string
	var usePrompt string
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute an ad-hoc command",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if disablePTY {
				executor.DisablePTY()
			}

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

			var content string
			addToHistory := false

			if !ignoreWorkingDir && len(args) == 1 && args[0][0] == shells.ExpandChar {
				item, err := shells.GetHistoryEntry(args[0])
				if err != nil {
					return err
				}

				if item.WorkingDir != "" {
					err = os.Chdir(item.WorkingDir)
					if err != nil {
						return err
					}
				}

				content = item.CommandLine
				addToHistory = true
			} else {
				content = strings.Join(args, " ")
			}

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

			startedAt := time.Now()
			res := register.Execute(ctx, req)
			duration := time.Since(startedAt)
			exitStatus := res.ExitStatus

			if res.Err != nil {
				exitStatus = -1
				cmd.PrintErrf("exit> execution failed: %v\n", res.Err)
			} else {
				cmd.Printf("exit> status: %d\n", exitStatus)
			}

			if addToHistory {
				wd, _ := os.Getwd()
				return shells.PutHistoryEntry(&historian.Item{
					RecordedAt:  startedAt,
					WorkingDir:  wd,
					CommandLine: content,
					ExitStatus:  &res.ExitStatus,
					Duration:    duration,
				})
			}

			return nil
		},
	}
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 0,
		"Wait some amount of time before giving up on a command to return")
	cmd.Flags().BoolVarP(&disablePTY, "disable-pty", "T", false,
		"Disable psuedo-terminal allocation")
	cmd.Flags().BoolVarP(&disablePTY, "ignore-workdir", "i", false,
		"Do not change to an item's saved working directory")
	cmd.Flags().StringVarP(&startPrompt, "start", "s", "",
		"Start this command as the named prompt")
	cmd.Flags().StringVarP(&usePrompt, "use", "u", "",
		"Use a running prompt to execute this command")

	return cmd
}
