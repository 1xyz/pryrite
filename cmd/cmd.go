package cmd

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/tools"
)

type Params struct {
	Agent   string
	Version string
	Argv    []string
	Command string
}

func LogErr(cmd string, err error) int {
	fmt.Printf("command failed:. err: %v\n", err)
	tools.Log.Err(err).Msgf("command %v failed", cmd)
	return 1
}
