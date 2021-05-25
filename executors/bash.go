package executor

import (
	"bufio"
	"context"
	"errors"
	"io"
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

	// proxy in/out/err to allow for dynamic reassignment for each execution
	b.bash.Stdin = &readWriterProxy{}
	b.bash.Stdout = &readWriterProxy{}
	b.bash.Stderr = &readWriterProxy{}

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

	// update proxies with the requested i/o
	inProxy := b.bash.Stdin.(*readWriterProxy)
	inProxy.SetReader(req.Stdin)
	outProxy := b.bash.Stdout.(*readWriterProxy)
	outProxy.SetWriter(req.Stdout)
	errProxy := b.bash.Stderr.(*readWriterProxy)
	errProxy.SetWriter(req.Stderr)

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

	// update proxies to avoid confusing caller if more junk comes in
	inProxy.SetReader(nil)
	outProxy.SetWriter(nil)
	errProxy.SetWriter(nil)

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

type readWriterProxy struct {
	reader io.Reader
	writer io.Writer
}

func (proxy *readWriterProxy) SetReader(reader io.Reader) {
	proxy.reader = reader
}

func (proxy *readWriterProxy) SetWriter(writer io.Writer) {
	proxy.writer = writer
}

func (proxy *readWriterProxy) Read(buf []byte) (int, error) {
	if proxy.reader == nil {
		return 0, errors.New("proxy was asked to read without a reader assigned")
	}

	return proxy.reader.Read(buf)
}

func (proxy *readWriterProxy) Write(data []byte) (int, error) {
	if proxy.writer == nil {
		return 0, errors.New("proxy was asked to write without a writer assigned")
	}

	return proxy.writer.Write(data)
}
