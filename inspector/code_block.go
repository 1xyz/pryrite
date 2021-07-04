package inspector

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/graph/log"
	"github.com/aardlabs/terminal-poc/internal/history"
	"github.com/aardlabs/terminal-poc/markdown"
	"github.com/aardlabs/terminal-poc/run"
	"github.com/aardlabs/terminal-poc/tools"
)

func newCodeBlock(index int, prefix string, b *graph.Block, n *graph.Node, r *run.Run, h history.History) *codeBlock {
	return &codeBlock{
		index:      index,
		block:      b,
		node:       n,
		exitAction: nil,
		prefix:     prefix,
		runner:     r,
		hist:       h,
	}
}

type codeBlock struct {
	// index is this codeblock's entry in codeblock list
	index int

	// nBlocks is the total number of code blocks
	nBlocks int

	block  *graph.Block
	node   *graph.Node
	runner *run.Run

	// exitAction is actionable exit information
	// when the code block exits from the repl, an exitAction can be set
	// if an action needs to be taken by the caller
	exitAction *BlockAction

	// prefix is the prompt prefix string
	prefix string

	// history used when executing commands with this codeblock
	hist history.History
}

func (c *codeBlock) SetNBlocks(nblocks int)            { c.nBlocks = nblocks }
func (c *codeBlock) SetExitAction(action *BlockAction) { c.exitAction = action }
func (c *codeBlock) GetExitAction() (*BlockAction, bool) {
	return c.exitAction, c.exitAction != nil
}

func (c *codeBlock) RunBlock() {
	tools.LogStdout("Running (%s): %s\n", c.block.ContentType, c.block.Content)
	doneCh := make(chan bool)
	c.runner.SetExecutionUpdateFn(func(entry *log.ResultLogEntry) {
		logResultLogEntry(entry)
		switch entry.State {
		case log.ExecStateCompleted, log.ExecStateFailed:
			doneCh <- true
			close(doneCh)
		default:
			// do nothing
		}
	})
	c.runner.SetStatusUpdateFn(func(status *run.Status) {
		switch status.Level {
		case run.StatusError:
			tools.Log.Error().Msgf(status.Message)
		default:
			tools.Log.Info().Msgf(status.Message)
		}
	})
	defer func() {
		c.runner.SetExecutionUpdateFn(nil)
		c.runner.SetStatusUpdateFn(nil)
	}()

	reqID, err := c.runner.ExecuteBlock(c.node, c.block, os.Stdout, os.Stderr)
	if err != nil {
		tools.LogStdError("\U0000274C  Execution failed: %s", err)
		return
	}

	sigCh := make(chan os.Signal, 1)
	// ToDo: the Ctrl+C is not captured because the go-prompt's run
	// is on the main go-routine.
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigCh:
		tools.LogStdout("signal %v received", sig)
		c.runner.CancelBlock(c.node.ID, reqID)
	case <-doneCh:
		return
	}
}

func (c *codeBlock) WhereAmI() {
	cursor := &markdown.Cursor{}
	sb := strings.Builder{}
	for _, block := range c.node.Blocks {
		if block.ID == c.block.ID {
			cursor.Start = sb.Len()
			cursor.Stop = cursor.Start + len(block.Content)
		}
		sb.WriteString(block.Content)
	}
	fmt.Println(md(sb.String(), "", cursor))
}

func (c *codeBlock) String() string {
	return fmt.Sprintf("[%d/%d] %s/%s - %s",
		c.index, c.nBlocks, c.node.ID, c.block.ID, tools.TrimLength(c.block.Content, 15))
}

func (c *codeBlock) QualifiedID() string {
	return fmt.Sprintf("%s :: %s", c.prefix, c.block.ID)
}
func (c *codeBlock) Block() *graph.Block { return c.block }
func (c *codeBlock) Index() int          { return c.index }
func (c *codeBlock) Len() int            { return c.nBlocks }

func logResultLogEntry(entry *log.ResultLogEntry) {
	tools.Log.Info().Msgf(
		"executionID: %s | requestID: %s | nodeID: %s | blockID: %s | result: %v | Err: %s",
		entry.ExecutionID,
		entry.RequestID,
		entry.NodeID,
		entry.BlockID,
		entry.State,
		entry.Err)
	displayResultLogEntry(entry)
}

func displayResultLogEntry(entry *log.ResultLogEntry) {
	switch entry.State {
	case log.ExecStateCompleted:
		exitInfo := ""
		if len(entry.ExitStatus) > 0 {
			exitInfo = fmt.Sprintf("exit-status: [%s]", entry.ExitStatus)
		}
		tools.LogStdout("\U00002705 Execution completed; %s\n", exitInfo)
	case log.ExecStateCanceled, log.ExecStateFailed:
		exitInfo := ""
		if len(entry.ExitStatus) > 0 {
			exitInfo = fmt.Sprintf("exit-status: [%s]", entry.ExitStatus)
		}
		tools.LogStdError("\U0000274C Execution %s; %s %s\n", entry.State, entry.Err, exitInfo)
	default:
		// do nothing
	}
}
