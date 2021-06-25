package inspector

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/graph"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/c-bata/go-prompt"
	"strings"
)

func newCodeBlock(index int, prefix string, b *graph.Block, n *graph.Node) *codeBlock {
	return &codeBlock{
		index:      index,
		block:      b,
		node:       n,
		exitAction: nil,
		prefix:     prefix,
	}
}

type codeBlock struct {
	index      int
	block      *graph.Block
	node       *graph.Node
	exitAction *BlockAction
	prefix     string
}

func (c *codeBlock) OpenRepl() {
	c.SetExitAction(nil)
	completer := NewCobraCommandCompleter(newRootCmd(c))
	prefix := fmt.Sprintf("[%d] %s :: %s", c.index, c.prefix, c.block.ID)
	pt := prompt.New(
		c.handleCommand,
		completer.Complete,
		prompt.OptionTitle(fmt.Sprintf("interactive")),
		prompt.OptionPrefix(fmt.Sprintf("%s >>> ", prefix)),
		prompt.OptionPrefixTextColor(prompt.Green),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionSetExitCheckerOnInput(func(in string, breakline bool) bool {
			if in == "quit" {
				return true
			}
			c.handleExitCmd(in)
			return c.exitAction != nil
		}),
	)

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

func (c *codeBlock) handleExitCmd(in string) {
	cmdToActions := map[string]BlockActionType{
		"quit": BlockActionQuit,
		"next": BlockActionNext,
		"prev": BlockActionPrev,
	}
	if action, found := cmdToActions[in]; found {
		c.exitAction = &BlockAction{Action: action}
	}
}

func (c *codeBlock) SetExitAction(action *BlockAction) { c.exitAction = action }
func (c *codeBlock) GetExitAction() (*BlockAction, bool) {
	return c.exitAction, c.exitAction != nil
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
	return fmt.Sprintf("[%d] %s/%s - %s", c.index, c.node.ID, c.block.ID, tools.TrimLength(c.block.Content, 15))
}
