package inspector

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/graph/log"
	"github.com/aardlabs/terminal-poc/internal/completer"
	"github.com/aardlabs/terminal-poc/run"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/c-bata/go-prompt"
)

func newCodeBlock(index int, prefix string, b *graph.Block, n *graph.Node, r *run.Run) *codeBlock {
	return &codeBlock{
		index:      index,
		block:      b,
		node:       n,
		exitAction: nil,
		prefix:     prefix,
		runner:     r,
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
}

func (c *codeBlock) OpenRepl() {
	c.SetExitAction(nil)
	cc := completer.NewCobraCommandCompleter(newRootCmd(c))
	prefix := fmt.Sprintf("[%d/%d] %s :: %s", c.index+1, c.nBlocks, c.prefix, c.block.ID)
	pt := prompt.New(
		c.handleCommand,
		cc.Complete,
		prompt.OptionTitle("interactive inspector"),
		prompt.OptionPrefix(fmt.Sprintf("%s >>> ", prefix)),
		prompt.OptionPrefixTextColor(prompt.Green),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionSetExitCheckerOnInput(func(in string, breakline bool) bool {
			return c.exitAction != nil
		}),
	)

	c.WhereAmI()
	pt.Run()
}

func (c *codeBlock) handleCommand(s string) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return
	}

	cmd := newRootCmd(c)
	cmd.SetArgs(strings.Split(s, " "))
	// allow cobra to handle this
	cmd.Execute()
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
		case log.ExecStateCompleted, log.ExecStateCanceled:
			doneCh <- true
			close(doneCh)
		case log.ExecStateFailed:
			tools.LogStdError("\U0000274C Execution failed; %s\n", entry.Err)
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
	sb := strings.Builder{}
	for _, block := range c.node.Blocks {
		if block.ID == c.block.ID {
			sb.WriteString("\U0001F449 \U0001F449 ")
		}
		sb.WriteString(block.Content)
	}
	fmt.Println(md(sb.String(), ""))
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
}
