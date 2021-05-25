package executor

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type BashExecutor struct {
	bash      *exec.Cmd
	bashDone  chan error
	isRunning bool

	// i/o for sending commands to the bash session
	cmdWriter *os.File

	// i/o for receiving exit status from executed commands in the bash session
	resultReader *os.File
}

type collectorResult struct {
	err        error
	exitStatus int
}

const repl = `while IFS= read -u 11 -r line; do
    eval "$line";
    echo $? >&12;
done`

func NewBashExecutor() (Executor, error) {
	b := &BashExecutor{}

	b.bash = exec.Command("bash", "-c", repl)
	b.bashDone = make(chan error, 1)

	// FIXME: proxy!!
	b.bash.Stdout = os.Stdout
	b.bash.Stderr = os.Stderr

	var err error

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
	b.bash.ExtraFiles = make([]*os.File, 10)
	b.bash.ExtraFiles[8] = cmdReader    // this becomes file descriptor 11 in bash (in,out,err + 8)
	b.bash.ExtraFiles[9] = resultWriter // and this is 12

	return b, nil
}

func (b *BashExecutor) Name() string { return "bash-executor" }

func (b *BashExecutor) ContentTypes() []ContentType { return []ContentType{Bash, Shell} }

func (b *BashExecutor) Execute(ctx context.Context, req *ExecRequest) *ExecResponse {
	b.ensureRunning()

	resultReady := make(chan *collectorResult, 1)
	go b.collectStatus(resultReady)

	b.cmdWriter.Write(req.Content)
	b.cmdWriter.WriteString("\n")

	responseHdr := &ResponseHdr{req.Hdr.ID}

	var err error
	exitStatus := -1

	select {
	case result := <-resultReady: // got a result
		err = result.err
		exitStatus = result.exitStatus
	case <-ctx.Done(): // interrupted by the context
		err = ctx.Err()
		// FIXME: how to kill on timeout?
	case err = <-b.bashDone: // bash exited!
	}

	return &ExecResponse{Hdr: responseHdr, ExitStatus: exitStatus, Err: err}
}

func (b *BashExecutor) Cleanup() {
	if b.isRunning {
		b.bash.Process.Kill()
		b.bash.Process.Wait() // the previous kill won't let the `Run()` call clean up children
		b.cmdWriter.Close()
		b.resultReader.Close()
		b.isRunning = false
	}
}

func (b *BashExecutor) ensureRunning() {
	if b.isRunning {
		return
	}

	go func() {
		b.bashDone <- b.bash.Run() // this calls `Wait()` for us
		close(b.bashDone)
		b.isRunning = false
	}()

	b.isRunning = true
}

func (b *BashExecutor) collectStatus(ready chan *collectorResult) {
	result := &collectorResult{exitStatus: -1}
	reader := bufio.NewReader(b.resultReader)

	var status string
	status, result.err = reader.ReadString('\n')
	if result.err != nil {
		return
	}

	result.exitStatus, result.err = strconv.Atoi(strings.Trim(status, " \t\r\n"))
	if result.err != nil {
		result.exitStatus = -1
	}

	ready <- result
	close(ready)
}
