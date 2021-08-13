package executor

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/aardlabs/terminal-poc/tools"
)

type RemoteShellExecutor struct {
	BaseExecutor
}

const rshellReadyMarker = "__AARDY_READY"

var rshellReadyMarkerRE = regexp.MustCompile(`(?m)^` + rshellReadyMarker + `\s*`)

// Unlike the Bash Executor, we can't use file descriptors for communicating the exit status.
// Instead, this will issue each command followed by an echo and watch for the marker in output.

const rshellExitMarker = "__AARDY_EXIT="

var rshellExitMarkerRE = regexp.MustCompile(`(?m)^` + rshellExitMarker + `\d+\s*`)

func NewRemoteShellExecutor(content []byte, contentType *ContentType) (Executor, error) {
	se := &RemoteShellExecutor{}
	se.setDefaults()

	se.prepareCmd = se.prepareShellCmd
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

	se.name = "remote-" + se.name

	if se.command == "" {
		return nil, fmt.Errorf("remote shell requires a command (%s)", contentType)
	}

	return se, nil
}

func (se *RemoteShellExecutor) prepareShellCmd(stdout, stderr io.WriteCloser, usePty bool) (execReadyCh, error) {
	execReady, err := se.defaultPrepareCmd(stdout, stderr, usePty)
	if err != nil {
		return nil, err
	}

	<-execReady // drop immediate ready signal from the default prep

	go func() {
		se.stdin.Put([]byte("echo " + rshellReadyMarker))
		se.stdout.SetWriterMarker(stdout, rshellReadyMarkerRE, func(marker string) {
			execReady <- nil
		})
	}()

	return execReady, nil
}

func (se *RemoteShellExecutor) prepareShellIO(req *ExecRequest, isExecCmd bool) (resultReadyCh, error) {
	if isExecCmd {
		resultReady := make(resultReadyCh, 1)
		resultReady <- collectorResult{} // already waited for execReady as part of ensureRunning
		return resultReady, nil
	}

	command, err := se.getCommandFrom(req.Content, req.ContentType)
	if err != nil {
		return nil, err
	}

	// feed in the command and follow it up with a marker indicating the exit status
	se.stdin.Put(command)
	se.stdin.Put([]byte("echo " + rshellExitMarker + "$?"))

	ready := make(resultReadyCh, 1)
	se.stdout.SetWriterMarker(req.Stdout, rshellExitMarkerRE, func(marker string) {
		var err error
		var status int
		vals := strings.Split(strings.TrimSpace(marker), "=")
		if len(vals) < 2 {
			tools.Log.Error().Str("marker", marker).Msg("Unexpected exit marker found")
			status = -1
		} else {
			status, err = strconv.Atoi(vals[1])
		}
		ready <- collectorResult{err: err, exitStatus: status}
	})

	se.stderr.SetWriter(req.Stderr)

	return ready, nil
}
