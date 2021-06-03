package kmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	executor "github.com/aardlabs/terminal-poc/executors"
	"github.com/spf13/cobra"
)

func NewCmdExecutor() *cobra.Command {
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute",
		Args:  MinimumArgs(2, "You need to specify something to execute"),
		RunE: func(cmd *cobra.Command, args []string) error {
			register, err := executor.NewRegister()
			if err != nil {
				return err
			}

			contentType := executor.ContentType(args[0])

			for count, content := range args[1:] {
				req := executor.DefaultRequest()

				req.Content = []byte(content)
				req.ContentType = contentType

				req.Stdout = &prefixWriter{
					writer: os.Stdout,
					prefix: []byte(fmt.Sprint(count, "-out> ")),
				}

				req.Stderr = &prefixWriter{
					writer: os.Stderr,
					prefix: []byte(fmt.Sprint(count, "-err> ")),
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
					cmd.PrintErrf("Execution failed: %v\n", res.Err)
				} else {
					cmd.Printf("Exit status: %d\n", res.ExitStatus)
				}
			}

			register.Cleanup()

			return nil
		},
	}
	cmd.Flags().DurationVarP(&timeout, "timeout", "t",
		0, "Wait some amount of time before giving up on a command to return")

	return cmd
}

type prefixWriter struct {
	writer io.Writer
	prefix []byte
}

func (pw *prefixWriter) Write(data []byte) (int, error) {
	pw.writer.Write(pw.prefix)
	return pw.writer.Write(data)
}

func (pw *prefixWriter) Close() error {
	return nil
}
