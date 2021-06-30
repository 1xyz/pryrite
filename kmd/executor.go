package kmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	executor "github.com/aardlabs/terminal-poc/executors"

	"github.com/spf13/cobra"
)

func NewCmdRawExecutor() *cobra.Command {
	var timeout time.Duration
	var disablePTY bool
	cmd := &cobra.Command{
		Hidden: true,
		Use:    "exec-raw",
		Short:  "Execute",
		Args:   MinimumArgs(2, "You need to specify something to execute"),
		RunE: func(cmd *cobra.Command, args []string) error {
			register, err := executor.NewRegister()
			if err != nil {
				return err
			}

			if disablePTY {
				executor.DisablePTY()
			}

			var contentType *executor.ContentType

			count := 0
			for _, content := range args {
				if contentType == nil {
					contentType, err = executor.Parse(content)
					if err != nil {
						return err
					}

					continue
				}

				count++

				req := executor.DefaultRequest()

				req.Content = []byte(content)
				req.ContentType = contentType
				contentType = nil

				req.Stdout = &prefixWriter{
					writer: os.Stdout,
					prefix: fmt.Sprint(count, "-out> "),
				}

				req.Stderr = &prefixWriter{
					writer: os.Stderr,
					prefix: fmt.Sprint(count, "-err> "),
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
					cmd.PrintErrf("%d-exit> execution failed: %v\n", count, res.Err)
				} else {
					cmd.Printf("%d-exit> status: %d\n", count, res.ExitStatus)
				}
			}

			register.Cleanup()

			return nil
		},
	}
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 0,
		"Wait some amount of time before giving up on a command to return")
	cmd.Flags().BoolVarP(&disablePTY, "disable-pty", "T", false,
		"Disable psuedo-terminal allocation")

	return cmd
}

type prefixWriter struct {
	writer    io.Writer
	prefix    string
	skipWrite bool
}

func (pw *prefixWriter) Write(data []byte) (int, error) {
	origLen := len(data)
	out := strings.ReplaceAll(string(data), "\n", "\n"+pw.prefix)
	if !pw.skipWrite {
		pw.skipWrite = true
		pw.writer.Write([]byte(pw.prefix))
	}
	_, err := pw.writer.Write([]byte(out))
	return origLen, err
}

func (pw *prefixWriter) Close() error {
	return nil
}
