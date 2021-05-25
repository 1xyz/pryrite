package kmd

import (
	"context"
	"os"

	executor "github.com/aardlabs/terminal-poc/executors"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var register = executor.NewRegister()

func NewCmdExecutor() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute",
		Args:  MinimumArgs(1, "You need to specify something to execute"),
		RunE: func(cmd *cobra.Command, args []string) error {
			exec, _ := register.Get(executor.ContentType(args[0]))
			for _, content := range args[1:] {
				req := &executor.ExecRequest{
					Hdr: &executor.RequestHdr{
						ID: uuid.NewString(),
					},
					Content: []byte(content),
					Stdout:  os.Stdout,
					Stderr:  os.Stderr,
				}
				res := exec.Execute(context.Background(), req)
				if res.Err != nil {
					return res.Err
				}
				cmd.Printf("Exit status: %d\n", res.ExitStatus)
			}

			return nil
		},
	}
	return cmd
}
