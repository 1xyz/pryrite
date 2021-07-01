package executor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/aardlabs/terminal-poc/tools"

	pseudoTY "github.com/creack/pty"
	"github.com/mattn/go-shellwords"
	"github.com/rs/zerolog"
)

type BaseExecutor struct {
	name        string
	contentType *ContentType

	command     string
	commandArgs []string
	isRunning   bool

	execCmd     *exec.Cmd
	execDone    chan error
	execErr     error
	isExecuting bool
	skipCount   uint

	stdin  *CommandFeeder
	stdout *readWriterProxy
	stderr *readWriterProxy

	// callbacks for replacing parts of the base logic
	prepareCmd func(stdout, stderr io.WriteCloser, usePty bool) (execReadyCh, error)
	prepareIO  func(*ExecRequest, bool) (resultReadyCh, error)
	cancel     func()
	clearIO    func(bool)
	cleanup    func(bool)
}

type execReadyCh chan error
type resultReadyCh chan collectorResult

var expectExitResultReady = make(resultReadyCh, 1)

type collectorResult struct {
	err        error
	exitStatus int
}

func (be *BaseExecutor) Name() string { return be.name }

func (be *BaseExecutor) ContentType() *ContentType { return be.contentType }

func (be *BaseExecutor) Execute(ctx context.Context, req *ExecRequest) *ExecResponse {
	resp := &ExecResponse{
		Hdr:        &ResponseHdr{req.Hdr.ID},
		ExitStatus: -1,
	}

	if be.execErr != nil {
		resp.Err = be.execErr
		return resp
	}

	// on by default but may be globally disabled or per commend through the content-type
	usePty := true
	if disablePTY {
		usePty = false
	} else {
		if disPTY, ok := be.contentType.Params["disable-pty"]; ok {
			usePty = strings.ToLower(disPTY) == "true"
		}
	}

	resp.Err = be.ensureRunning(req.Stdout, req.Stderr, usePty)
	if resp.Err != nil {
		be.execErr = resp.Err
		return resp
	}

	if be.isExecuting {
		resp.Err = ErrExecInProgress
		return resp
	}

	be.isExecuting = true
	defer func() {
		be.isExecuting = false
		if be.skipCount > 0 {
			be.skipCount--
		}
	}()

	canceler := &Canceler{}
	canceler.OnCancel(be.cancel)
	defer canceler.Stop()

	tools.Log.Debug().Str("name", be.Name()).Str("id", req.Hdr.ID).Str("command", string(req.Content)).
		Str("contentType", req.ContentType.String()).Int("skip", int(be.skipCount)).Bool("pty", usePty).
		Msg("Executing")

	var resultReady resultReadyCh
	resultReady, resp.Err = be.prepareIO(req, be.skipCount > 0)
	if resp.Err != nil {
		return resp
	}

	select {
	case result := <-resultReady: // got a result
		resp.ExitStatus = result.exitStatus
		resp.Err = result.err
	case <-ctx.Done(): // interrupted by the context
		resp.Err = ctx.Err()
		// it's difficult to kill the current command and its children so we have to error out and reset ourselves
		// TODO: have our REPL above run every command in its own process group and report that back to us
		alreadyDone := false
		be.cleanup(alreadyDone)
	case resp.Err = <-be.execDone: // the underlying process exited!
		resp.ExitStatus = be.execCmd.ProcessState.ExitCode()
		alreadyDone := true
		be.cleanup(alreadyDone)
		if resp.Err == nil && resultReady != expectExitResultReady {
			resp.Err = fmt.Errorf("%s terminated unexpectedly (status:%d)",
				be.execCmd.Path, resp.ExitStatus)
		}
	}

	be.clearIO(be.skipCount > 0)

	tools.Log.Debug().Str("name", be.Name()).Interface("response", resp).Msg("Execution complete")

	return resp
}

func (be *BaseExecutor) Cleanup() {
	alreadyDone := false
	be.cleanup(alreadyDone)
}

//--------------------------------------------------------------------------------

func (be *BaseExecutor) setDefaults() {
	be.prepareCmd = be.defaultPrepareCmd
	be.prepareIO = be.defaultPrepareIO
	be.cancel = be.defaultCancel
	be.clearIO = be.defaultClearIO
	be.cleanup = be.defaultCleanup
}

func (be *BaseExecutor) processContentType(content []byte, myContentType, wantContentType *ContentType) error {
	if !myContentType.ParentOf(wantContentType, nil) {
		return ErrUnsupportedContentType
	}

	if _, ok := wantContentType.Params["prompt"]; ok {
		// this happens if there wasn't a prompt-assign found associated with the same content-type
		return fmt.Errorf("prompt requested but none found running: %s", wantContentType)
	}

	be.execDone = make(chan error, 1)

	be.name = myContentType.Subtype + "-executor"
	be.contentType = myContentType.Clone()

	prompt := wantContentType.Params["prompt-assign"]
	if prompt != "" {
		// restrict matching this executor to only prompted request content-types
		be.contentType.Params["prompt"] = prompt

		commandLine := wantContentType.Params["command"]
		if commandLine == "" {
			commandLine = string(content)
		}

		args, err := shellwords.Parse(commandLine)
		if err != nil {
			return err
		}

		be.command = args[0]
		be.commandArgs = args[1:]

		// since we use this command for the executor itself, we need to skip it when requested
		be.skipCount = 1
	}

	return nil
}

func (be *BaseExecutor) getCommand(content []byte, contentType *ContentType) ([]byte, error) {
	if _, ok := contentType.Params["prompt"]; ok {
		var start, stop int
		n, _ := fmt.Sscanf(contentType.Params["command"], "%d:%d", &start, &stop)
		if n != 1 && n != 2 {
			return content, nil
		}

		if n == 1 {
			return content[start:], nil
		}

		return content[start:stop], nil
	}

	return content, nil
}

func (be *BaseExecutor) ensureRunning(stdout, stderr io.WriteCloser, usePty bool) error {
	if be.isRunning {
		return nil
	}

	execReady, err := be.prepareCmd(stdout, stderr, usePty)
	if err != nil {
		return err
	}

	tools.Trace("exec", "command ready to run")

	go func() {
		err := be.execCmd.Run() // this calls `Wait()` for us
		tools.Trace("exec", "command stopped", err)
		be.execDone <- err
	}()

	select {
	case err = <-execReady:
		if err != nil {
			return err
		}
	case err = <-be.execDone:
		if err != nil {
			return err
		}
	case <-time.After(10 * time.Second):
		be.stdin.Put(nil)
		stopKill(be.execCmd.Process)
		return errors.New("gave up waiting for executor to be ready")
	}

	be.isRunning = true

	if tools.Log.GetLevel() <= zerolog.DebugLevel {
		args := zerolog.Arr()
		for _, arg := range be.commandArgs {
			args = args.Str(arg)
		}
		var pid int
		if be.execCmd.Process == nil {
			pid = -1
		} else {
			pid = be.execCmd.Process.Pid
		}
		tools.Log.Debug().Str("execType", be.contentType.String()).Str("name", be.Name()).
			Str("command", be.command).Array("args", args).Int("pid", pid).
			Msg("Executor is running")
	}

	return nil
}

func (be *BaseExecutor) defaultPrepareCmd(stdout, stderr io.WriteCloser, usePty bool) (execReadyCh, error) {
	be.execCmd = exec.Command(be.command, be.commandArgs...)

	var outPTY, errPTY *os.File
	if usePty {
		var err error
		var outTTY, errTTY *os.File
		outPTY, outTTY, err = pseudoTY.Open()
		if err != nil {
			return nil, err
		}

		errPTY, errTTY, err = pseudoTY.Open()
		if err != nil {
			return nil, err
		}

		be.execCmd.SysProcAttr = &syscall.SysProcAttr{}
		be.execCmd.SysProcAttr.Setsid = true
		be.execCmd.SysProcAttr.Setctty = true

		be.execCmd.Stdin = outTTY
		be.execCmd.Stdout = outTTY
		be.execCmd.Stderr = errTTY
	}

	be.stdin = NewCommandFeeder(outPTY)
	// proxy out/err to allow for dynamic reassignment for each execution
	be.stdout = &readWriterProxy{name: "stdout", writer: stdout}
	be.stderr = &readWriterProxy{name: "stderr", writer: stderr}

	if outPTY != nil {
		be.stdout.Monitor(outPTY)
		be.stderr.Monitor(errPTY)
	}

	// default is to be ready immediately (but executors can/will override this)
	execReady := make(execReadyCh, 1)
	execReady <- nil

	return execReady, nil
}

func (be *BaseExecutor) defaultCancel() {
	if be.isRunning {
		stopKill(be.execCmd.Process)
	}
}

func (be *BaseExecutor) defaultPrepareIO(req *ExecRequest, isExecCmd bool) (resultReadyCh, error) {
	if isExecCmd {
		// just a test connection so provide EOF to the input to have the command exit
		be.stdin.Put(nil)
	} else {
		command, err := be.getCommand(req.Content, req.ContentType)
		if err != nil {
			return nil, err
		}

		be.stdin.Put(command)

		// for now we can only support one command per execution for the basic input feeder executors
		be.stdin.Close()
	}

	// this will never "fire"--instead, we expect the command to exit once done processing
	ready := expectExitResultReady
	return ready, nil
}

func (be *BaseExecutor) defaultClearIO(isExecCmd bool) {
	// update proxies to avoid confusing caller if more junk comes in
	be.stdout.SetWriter(nil)
	be.stderr.SetWriter(nil)
}

func (be *BaseExecutor) defaultCleanup(alreadyDone bool) {
	if be.isRunning {
		if !alreadyDone {
			be.stdin.Close()
			stopKill(be.execCmd.Process)
			<-be.execDone // wait for our goroutine to exit
		}
		be.isRunning = false
	}
}

// attempt an interrupt immediately and then in the background try a full kill if still running
// FIXME: this won't work on windows!!
func stopKill(proc *os.Process) {
	if proc == nil {
		return
	}
	syscall.Kill(-proc.Pid, syscall.SIGINT)
	go func() {
		time.Sleep(500 * time.Millisecond)
		syscall.Kill(-proc.Pid, syscall.SIGKILL)
	}()
}
