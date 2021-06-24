package executor

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type RemoteShellExecutor struct {
	BaseExecutor
}

// Unlike the Bash Executor, we can't use file descriptors for communicating the exit status.
// Instead, this will issue each command follwed by an echo and watch for the marker in output.

const rshellExitMarker = "__AARDY_EXIT="

var rshellExitMarkerRE = regexp.MustCompile(`(?m)^` + rshellExitMarker + `\d+\s*`)

func NewRemoteShellExecutor(content []byte, contentType *ContentType) (Executor, error) {
	se := &RemoteShellExecutor{}
	se.setDefaults()

	se.prepareIO = se.prepareShellIO

	err := se.processContentType(content, Bash, contentType)
	if err != nil {
		if errors.Is(err, ErrUnsupportedContentType) {
			err := se.processContentType(content, Shell, contentType)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if se.command == "" {
		return nil, fmt.Errorf("remote shell requires a command (%s)", contentType)
	}

	return se, nil
}

func (se *RemoteShellExecutor) prepareShellIO(req *ExecRequest, isExecCmd bool) (resultReadyCh, error) {
	if isExecCmd {
		return se.defaultPrepareIO(req, isExecCmd)
	}

	cf := se.execCmd.Stdin.(*CommandFeeder)

	command, err := se.getCommand(req.Content, req.ContentType)
	if err != nil {
		return nil, err
	}

	// feed in the command and follow it up with a marker indicating the exit status
	cf.Put(command)
	cf.Put([]byte("echo " + rshellExitMarker + "$?"))

	ready := make(resultReadyCh)
	se.execCmd.Stdout.(*readWriterProxy).SetWriterMarker(req.Stdout, rshellExitMarkerRE, func(marker string) {
		vals := strings.Split(strings.TrimSpace(marker), "=")
		status, err := strconv.Atoi(vals[1])
		ready <- collectorResult{err: err, exitStatus: status}
	})

	se.execCmd.Stderr.(*readWriterProxy).SetWriter(req.Stderr)

	return ready, nil
}
