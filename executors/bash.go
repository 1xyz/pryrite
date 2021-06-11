package executor

import (
	"bufio"
	"context"
	"errors"
	"go.uber.org/atomic"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/aardlabs/terminal-poc/tools"
)

type BashExecutor struct {
	bash        *exec.Cmd
	bashDone    chan error
	isRunning   atomic.Bool
	isExecuting bool

	// i/o for sending commands to the bash session
	cmdWriter *os.File

	// i/o for receiving exit status from executed commands in the bash session
	resultReader *os.File
}

type collectorResult struct {
	err        error
	exitStatus int
}

// This is a bash REPL that:
//   * ignores backslashes (-r)
//   * differentiates commands by null-terminated input (-d) provided from a "commands-to-read" descriptor (-u)
//   * reports back exit status via a different descriptor (>&12)
const repl = `while IFS= read -u 11 -r -d $'\0' cmd; do eval "$cmd"; echo $? >&12; done`

// This is a "debug" version of the REPL to help track down issues:
// const repl = `while IFS= read -u 11 -r -d $'\0' cmd; do
//   echo "> $cmd";
//   eval "$cmd";
//   echo "DONE";
//   echo $? >&12;
// done`

func NewBashExecutor() (Executor, error) {
	b := &BashExecutor{
		bashDone: make(chan error, 1),
	}
	err := b.init()
	return b, err
}

func (b *BashExecutor) Name() string { return "bash-executor" }

func (b *BashExecutor) ContentTypes() []ContentType { return []ContentType{Bash, Shell} }

func (b *BashExecutor) Execute(ctx context.Context, req *ExecRequest) *ExecResponse {
	if b.isExecuting {
		return &ExecResponse{
			ExitStatus: -1,
			Err:        errors.New("an execution is already in progress"),
		}
	}

	b.isExecuting = true
	defer func() { b.isExecuting = false }()

	b.ensureRunning()

	// update proxies with the requested i/o
	inProxy := b.bash.Stdin.(*readWriterProxy)
	outProxy := b.bash.Stdout.(*readWriterProxy)
	errProxy := b.bash.Stderr.(*readWriterProxy)
	inProxy.SetReader(req.Stdin)
	outProxy.SetWriter(req.Stdout)
	errProxy.SetWriter(req.Stderr)

	resultReady := make(chan *collectorResult, 1)
	go b.collectStatus(resultReady)

	b.cmdWriter.Write(req.Content)
	b.cmdWriter.Write([]byte{0}) // null terminated for the repl's read to handle multiline snippets

	responseHdr := &ResponseHdr{req.Hdr.ID}

	var err error
	exitStatus := -1

	select {
	case result := <-resultReady: // got a result
		err = result.err
		exitStatus = result.exitStatus
	case <-ctx.Done(): // interrupted by the context
		err = ctx.Err()
		// it's not possible to kill the current command so we have to error out and reset ourselves
		b.Reset()
	case err = <-b.bashDone: // bash exited!
	}

	// update proxies to avoid confusing caller if more junk comes in
	inProxy.SetReader(nil)
	outProxy.SetWriter(nil)
	errProxy.SetWriter(nil)

	return &ExecResponse{Hdr: responseHdr, ExitStatus: exitStatus, Err: err}
}

func (b *BashExecutor) Cleanup() {
	if b.isRunning.Load() {
		stopKill(b.bash.Process)
		b.bash.Process.Wait() // the previous kill won't let the `Run()` call clean up children
		b.cmdWriter.Close()
		b.resultReader.Close()
		b.isRunning.Store(false)
	}
}

func (b *BashExecutor) Reset() error {
	b.Cleanup()
	return b.init()
}

//--------------------------------------------------------------------------------

func (b *BashExecutor) init() error {
	b.isRunning.Store(false)
	b.bash = exec.Command("bash", "-c", repl)
	//b.bashDone = make(chan error, 1)

	// make sure bash is in its own process group so we can terminate itself _and_ any children
	b.bash.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// proxy in/out/err to allow for dynamic reassignment for each execution
	b.bash.Stdin = &readWriterProxy{name: "stdin"}
	b.bash.Stdout = &readWriterProxy{name: "stdout"}
	b.bash.Stderr = &readWriterProxy{name: "stderr"}

	var err error

	// these are passed off to the bash session
	var cmdReader, resultWriter *os.File

	// prepare a pipe to let us inject commands (i.e. avoid their stdin)
	cmdReader, b.cmdWriter, err = os.Pipe()
	if err != nil {
		return err
	}

	// prepare a pipe to let our injected commands communicate only to us (i.e. avoid their stdout)
	b.resultReader, resultWriter, err = os.Pipe()
	if err != nil {
		return err
	}

	// offset our descriptors in case the user wants to get fancy with their own scripts
	// (i.e. they can still safely use FDs 3 thru 10)
	b.bash.ExtraFiles = make([]*os.File, 10)
	b.bash.ExtraFiles[8] = cmdReader    // this becomes file descriptor 11 in bash (in,out,err + 8)
	b.bash.ExtraFiles[9] = resultWriter // and this is 12

	return nil
}

func (b *BashExecutor) ensureRunning() {
	if b.isRunning.Load() {
		return
	}

	go func() {
		b.bashDone <- b.bash.Run() // this calls `Wait()` for us
		close(b.bashDone)
		b.isRunning.Store(false)
	}()

	b.isRunning.Store(true)
}

func (b *BashExecutor) collectStatus(ready chan *collectorResult) {
	result := &collectorResult{exitStatus: -1}
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

// attempt an interrupt immediately and then in the background try a full kill if still running
// FIXME: this won't work on windows!!
func stopKill(proc *os.Process) {
	syscall.Kill(-proc.Pid, syscall.SIGINT)
	go func() {
		time.Sleep(500 * time.Millisecond)
		syscall.Kill(-proc.Pid, syscall.SIGKILL)
	}()
}

//--------------------------------------------------------------------------------

type readWriterProxy struct {
	name   string
	reader io.Reader
	writer io.WriteCloser
	rlock  sync.Mutex
	wlock  sync.Mutex
}

func (proxy *readWriterProxy) SetReader(reader io.Reader) {
	proxy.rlock.Lock()
	proxy.reader = reader
	proxy.rlock.Unlock()
}

func (proxy *readWriterProxy) SetWriter(writer io.WriteCloser) {
	proxy.wlock.Lock()
	wtr := proxy.writer
	proxy.wlock.Unlock()

	if writer == nil && wtr != nil {
		// block to make sure the upstream writer has all the bytes before we
		// let the execute complete
		err := wtr.Close()
		if err != nil {
			tools.Log.Err(err).Str("proxy", proxy.name).Msg("writer close failed")
		}
	}

	proxy.wlock.Lock()
	proxy.writer = writer
	proxy.wlock.Unlock()
}

func (proxy *readWriterProxy) Read(buf []byte) (int, error) {
	proxy.rlock.Lock()
	rdr := proxy.reader
	proxy.rlock.Unlock()

	if rdr == nil {
		// this is "common" for a nil stdin
		tools.Log.Debug().Str("proxy", proxy.name).Msg("asked to read without a reader assigned")
		return 0, io.EOF
	}

	return rdr.Read(buf)
}

func (proxy *readWriterProxy) Write(data []byte) (int, error) {
	proxy.wlock.Lock()
	wtr := proxy.writer
	proxy.wlock.Unlock()

	if wtr == nil {
		tools.Log.Error().Str("proxy", proxy.name).Msg("asked to write without a writer assigned")
		// lie to bash about the success of the write to avoid killing our repl
		return len(data), nil
	}

	return wtr.Write(data)
}
