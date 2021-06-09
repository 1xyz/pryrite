package tools

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func BashExec(cmd string) error {
	binaryStr := strings.TrimSpace("bash")
	binary, err := exec.LookPath(binaryStr)
	if err != nil {
		return err
	}

	env := os.Environ()
	return syscall.Exec(binary, []string{"bash", "-c", cmd}, env)
}
