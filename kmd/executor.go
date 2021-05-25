package kmd

import (
	"context"
	"fmt"
	"io"
	"os"

	executor "github.com/aardlabs/terminal-poc/executors"
	"github.com/spf13/cobra"
)

func NewCmdExecutor() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute",
		Args:  MinimumArgs(1, "You need to specify something to execute"),
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
					writer: os.Stdout,
					prefix: []byte(fmt.Sprint(count, "-err> ")),
				}

				res := register.Execute(context.Background(), req)
				if res.Err != nil {
					return res.Err
				}

				cmd.Printf("Exit status: %d\n", res.ExitStatus)
			}

			register.Cleanup()

			return nil
		},
	}

	return cmd
}

type prefixWriter struct {
	writer io.Writer
	prefix []byte
}

func (pw *prefixWriter) Write(data []byte) (int, error) {
	// println("barfwrite", string(data))
	pw.writer.Write(pw.prefix)
	return pw.writer.Write(data)
}
