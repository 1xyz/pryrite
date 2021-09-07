package shells

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aardlabs/terminal-poc/app"
	"github.com/aardlabs/terminal-poc/historian"
	"github.com/aardlabs/terminal-poc/tools"

	"github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"
)

func NewCmdInit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Provides an eval string to integrate history logging from your shell",
		RunE: func(cmd *cobra.Command, args []string) error {
			parent, err := GetParent()
			if err != nil {
				return err
			}

			shell := parent.Executable()

			err = syncThisDir(shell)
			if err != nil {
				return err
			}

			initPath := fmt.Sprintf("%s/%s-init.sh", shell, shell)
			fmt.Printf("source %s\n", tools.MyPathTo(initPath))

			return nil
		},
	}

	return cmd
}

func NewCmdHistory() *cobra.Command {
	reverse := false
	currentSession := false
	showWorkingDir := false

	cmd := &cobra.Command{
		Use:     "history [<duration_limit>]",
		Aliases: []string{"h", "hist"},
		Short:   "Display the list of entries captured in the history log",
		Long: tools.Examplef(`
            Displays the list of entries captured in the history log

            {AppName} history will accept an optional agument to limit entries shown.
            This value is a number followed by one of the following unit abbreviations:

              s - seconds
              m - minutes
              h - hours
        `),
		Example: tools.Examplef(`
            To list all entries recorded in the previous 10 minutes:
              $ {AppName} history 10m

            To list the entries with the oldest items last:
              $ {AppName} history -r
		`),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var duration time.Duration

			if len(args) > 0 {
				var err error
				duration, err = time.ParseDuration(args[0])
				if err != nil {
					return err
				}
			}

			return EachHistoryEntry(reverse, duration, currentSession, func(item *historian.Item) error {
				fmt.Println(item.StringWithOpts(showWorkingDir))
				return nil
			})
		},
	}

	cmd.Flags().BoolVarP(&reverse, "reverse", "r", reverse,
		"Show the entries in reverse")
	cmd.Flags().BoolVarP(&currentSession, "session", "s", currentSession,
		"Limit results to only those from the current session")
	cmd.Flags().BoolVarP(&showWorkingDir, "workdir", "w", showWorkingDir,
		"Show the working directory of each entry")

	cmd.AddCommand(newCmdStart())
	cmd.AddCommand(newCmdStop())
	cmd.AddCommand(newCmdGet())

	return cmd
}

//--------------------------------------------------------------------------------

func newCmdStart() *cobra.Command {
	cmd := &cobra.Command{
		Hidden:      true,
		Annotations: map[string]string{"SkipUpdateCheck": "true"},
		Use:         "start <command-line>",
		Short:       "Start a history entry for the provided command-line",
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parent, err := GetParent()
			if err != nil {
				return err
			}

			cmdArgs, err := shellwords.Parse(args[0])
			if err == nil {
				exe := cmdArgs[0]
				if filepath.Base(exe) == app.Name && len(cmdArgs) > 1 && cmdArgs[1][0] == "h"[0] {
					// ignore our own history command lines...
					// NOTE: this will _not_ catch aliased aardy history calls
					return nil
				}
			}

			parentPID := parent.Pid()
			hist := openHistorian(parent.Executable(), false)
			defer hist.Close()

			workingDir, _ := os.Getwd()

			return hist.Put(&historian.Item{
				WorkingDir:  workingDir,
				CommandLine: args[0],
				ParentPID:   &parentPID,
			})
		},
	}

	return cmd
}

func newCmdStop() *cobra.Command {
	cmd := &cobra.Command{
		Hidden:      true,
		Annotations: map[string]string{"SkipUpdateCheck": "true"},
		Use:         "stop <exit-status>",
		Short:       "Stop and record a history log for the most recently started entry in this session",
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			now := time.Now()

			exitStatus, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			parent, err := GetParent()
			if err != nil {
				return err
			}

			hist := openHistorian(parent.Executable(), false)
			defer hist.Close()

			item, err := getMostRecentEntry(parent, hist, true, true)
			if err != nil {
				return err
			}

			if item == nil {
				tools.Log.Warn().Msg("Not saving exit status--no recent report found")
				return nil
			}

			if item.ExitStatus != nil {
				// this happens often, esp when they use aardy history since we intentially ignore that report...
				return nil
			}

			item.ExitStatus = &exitStatus
			item.Duration = now.Sub(item.RecordedAt)

			return hist.Put(item)
		},
	}

	return cmd
}

func newCmdGet() *cobra.Command {
	asOffset := false

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a commandline from the history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0][0] == ExpandChar {
				item, err := GetHistoryEntry(args[0])
				if err != nil {
					return err
				}

				fmt.Print(item.CommandLine)
				return nil
			}

			id, err := strconv.ParseInt(args[0], 10, 0)
			if err != nil {
				return err
			}

			parent, err := GetParent()
			if err != nil {
				return err
			}

			hist := openHistorian(parent.Executable(), true)
			defer hist.Close()

			var item *historian.Item
			if asOffset {
				item, err = getRelativeItem(hist, nil, id)
			} else {
				item, err = getItem(hist, uint64(id))
			}

			if err != nil {
				return err
			}

			fmt.Print(item.CommandLine)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&asOffset, "offset", "o", asOffset,
		"Interpret the <id> as an offset (negative values will count from most recent to oldest)")

	return cmd
}
