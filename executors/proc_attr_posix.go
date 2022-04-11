//go:build !windows
// +build !windows

package executor

import "syscall"

func (be *BaseExecutor) setProcAttr() {
	be.execCmd.SysProcAttr = &syscall.SysProcAttr{}
	be.execCmd.SysProcAttr.Setsid = true
	be.execCmd.SysProcAttr.Setctty = true
}
