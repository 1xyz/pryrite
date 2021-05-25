package executor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
)

type BashExecutor struct{}

type collectorResult struct {
	err        error
	exitStatus int
	context    []string
}

func (b *BashExecutor) Name() string { return "bash-executor" }

func (b *BashExecutor) ContentTypes() []ContentType { return []ContentType{Bash, Shell} }

func (b *BashExecutor) Execute(ctx context.Context, req *ExecRequest) *ExecResponse {
	c := make(chan error, 1)
	cmd := exec.CommandContext(ctx, "bash")

	// prepare the input to allow for injection of our own follow-on commands
	cmdInput, err := cmd.StdinPipe()
	if err != nil {
		return &ExecResponse{Err: err}
	}

	cmd.Stdout = req.Stdout
	cmd.Stderr = req.Stderr

	// prepare a pipe to let our injected commands communicate only to us (i.e. avoid their stdout)
	readFile, writeFile, err := os.Pipe()
	if err != nil {
		return &ExecResponse{Err: err}
	}

	// offset our descriptor in case the user wants to get fancy with their own scripts
	cmd.ExtraFiles = make([]*os.File, 10)
	cmd.ExtraFiles[9] = writeFile // this becomes file descriptor 12 in bash (in,out,err + 9)

	result := &collectorResult{}
	go b.collectStatusAndContext(cmdInput, readFile, result)

	go func() {
		c <- cmd.Run()
		close(c)
	}()

	cmdInput.Write(req.Content)
	cmdInput.Write([]byte("\n"))

	responseHdr := &ResponseHdr{req.Hdr.ID}

	select {
	case <-ctx.Done():
		err = ctx.Err()
		if err == nil {
			err = <-c // wait for cmd.Run to complete
		}
	case err = <-c:
	}

	if err == nil {
		// don't stomp on an error from the cmd or ctx
		err = result.err
	}

	return &ExecResponse{Hdr: responseHdr, Err: err}
}

func (b *BashExecutor) collectStatusAndContext(cmdInput io.WriteCloser, readFile *os.File, result *collectorResult) {
	var err error
	state := 0
	reader := bufio.NewReader(readFile)

	cmdInput.Write([]byte("set -o posix\necho $? >&12\n"))

	for {
		var data []byte

		data, _, err = reader.ReadLine()
		if err != nil {
			break
		}

		line := string(data)

		if state == 0 {
			state++

			result.exitStatus, err = strconv.Atoi(line)
			if err != nil {
				break
			}

			// list out variables
			cmdInput.Write([]byte("set >&12; echo AARDY_COLLECTION_DONE >&12\n"))

		} else if state == 1 {
			if line == "AARDY_COLLECTION_DONE" {
				break
			}

			result.context = append(result.context, line)

		} else {
			err = fmt.Errorf("unknown state value: %d", state)
			break
		}
	}

	closeErr := cmdInput.Close()

	if err == nil {
		// don't overwrite an err from above!
		result.err = closeErr
	}
}
