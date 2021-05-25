package kmd

import (
	"context"
	"os"

	executor "github.com/aardlabs/terminal-poc/executors"
	"github.com/google/uuid"
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

			for _, content := range args[1:] {
				req := &executor.ExecRequest{
					Hdr: &executor.RequestHdr{
						ID: uuid.NewString(),
					},
					Content:     []byte(content),
					ContentType: contentType,
					Stdout:      os.Stdout,
					Stderr:      os.Stderr,
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
