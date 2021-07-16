package executor

import (
	"bufio"
	"errors"
	"github.com/aardlabs/terminal-poc/tools"
	"io"
	"os"
	"strconv"
	"strings"
)

type BashExecutor struct {
	BaseExecutor

	// i/o for sending commands to the bash session
	cmdWriter *os.File

	// i/o for receiving exit status from executed commands in the bash session
	resultReader *os.File
}

// This is a bash REPL that:
//   * ignores backslashes (-r)
//   * differentiates commands by null-terminated input (-d) provided from a "commands-to-read" descriptor (-u)
//   * reports back exit status via a different descriptor (>&12)
const repl = `while IFS= read -u 11 -r -d $'\0' cmd; do eval "$cmd"; echo $? >&12; done`

var (
	Bash  = &ContentType{"text", "bash", map[string]string{}}
	Shell = &ContentType{"text", "shell", map[string]string{}}
)

// This is a "debug" version of the REPL to help track down issues:
// const repl = `while IFS= read -u 11 -r -d $'\0' cmd; do
//   echo "> $cmd" >&2;
//   eval "$cmd";
//   echo "DONE" >&2;
//   echo $? >&12;
// done`

func NewBashExecutor(content []byte, contentType *ContentType) (Executor, error) {
	if _, ok := contentType.Params["prompt-assign"]; ok {
		// The Bash Executor won't work remotely, so we need to fall back to one that does...
		return NewRemoteShellExecutor(content, contentType)
	}

	b := &BashExecutor{}
	b.setDefaults()

	b.prepareCmd = b.prepareBashCmd
	b.prepareIO = b.prepareBashIO
	b.cleanup = b.cleanupBash

	err := b.processContentType(content, Bash, contentType)
	if err != nil {
		if errors.Is(err, ErrUnsupportedContentType) {
			err := b.processContentType(content, Shell, contentType)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	b.command = "bash"
	b.commandArgs = append(b.commandArgs, "-c", repl)

	return b, nil
}

//--------------------------------------------------------------------------------

func (b *BashExecutor) prepareBashCmd(stdout, stderr io.WriteCloser, usePty bool) (execReadyCh, error) {
	execReady, err := b.BaseExecutor.defaultPrepareCmd(stdout, stderr, usePty)
	if err != nil {
		return nil, err
	}

	// Refer WAR-295
	// b.execCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// these are passed off to the bash session
	var cmdReader, resultWriter *os.File

	// prepare a pipe to let us inject commands (i.e. avoid their stdin)
	cmdReader, b.cmdWriter, err = os.Pipe()
	if err != nil {
		return nil, err
	}

	// prepare a pipe to let our injected commands communicate only to us (i.e. avoid their stdout)
	b.resultReader, resultWriter, err = os.Pipe()
	if err != nil {
		return nil, err
	}

	// offset our descriptors in case the user wants to get fancy with their own scripts
	// (i.e. they can still safely use FDs 3 thru 10)
	b.execCmd.ExtraFiles = make([]*os.File, 10)
	b.execCmd.ExtraFiles[8] = cmdReader    // this becomes file descriptor 11 in bash (in,out,err + 8)
	b.execCmd.ExtraFiles[9] = resultWriter // and this is 12

	return execReady, nil
}

func (b *BashExecutor) prepareBashIO(req *ExecRequest, isExecCmd bool) (resultReadyCh, error) {
	if isExecCmd {
		// this should _NEVER_ happen since we pass these over to the RemoteShellExecutor
		return nil, errors.New("unexpected call with execute command set")
	}

	// temporarily redirect inFile into caller's specified input file (Fd)
	b.inFile = req.In
	// temporarily redirect out/err into caller's writers
	b.stdout.SetWriter(req.Stdout)
	b.stderr.SetWriter(req.Stderr)

	resultReady := make(resultReadyCh, 1)
	go b.collectStatus(resultReady)

	command, err := b.getCommand(req.Content, req.ContentType)
	if err != nil {
		return nil, err
	}

	tools.Log.Debug().
		Str("command", string(command)).
		Msgf("prepareBashIO: writing command")
	b.cmdWriter.Write(command)
	b.cmdWriter.Write([]byte{0}) // null terminated for the repl's read to handle multiline snippets

	return resultReady, nil
}

func (b *BashExecutor) collectStatus(ready resultReadyCh) {
	result := collectorResult{exitStatus: -1}
	reader := bufio.NewReader(b.resultReader)

	var status string
	status, result.err = reader.ReadString('\n')
	if result.err != nil {
		return
	}

	result.exitStatus, result.err = strconv.Atoi(strings.TrimSpace(status))
	if result.err != nil {
		result.exitStatus = -1
	}

	ready <- result
	close(ready)
}

func (b *BashExecutor) cleanupBash(alreadyDone bool) {
	tools.Log.Info().Msgf("cleanupBash alreadyDone=%v", alreadyDone)

	if b.isRunning {
		b.cmdWriter.Close()
		b.resultReader.Close()
	}

	b.BaseExecutor.defaultCleanup(alreadyDone)
}
