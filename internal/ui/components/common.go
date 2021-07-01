package components

import (
	"errors"
	"github.com/aardlabs/terminal-poc/tools"
	"io"
	"os"
)

var ErrNoEntryPicked = errors.New("no entry picked")

func colLen() int {
	_, c, err := tools.GetTermWindowSize()
	if err != nil {
		return 40 // default col len
	}
	return c
}

// bellSkipper implements an io.WriteCloser that skips the terminal bell
// character (ASCII code 7), and writes the rest to os.Stdout. It is used to
// replace readline.Stdout's, that is the package used by promptui to display the
// prompts.
//
// This is a workaround for the bell issue documented in
// https://github.com/manifoldco/promptui/issues/49.
type bellSkipper struct{}

func newBellSkipper() io.WriteCloser {
	return &bellSkipper{}
}

// Write implements an io.WriterCloser over os.Stderr, but it skips the terminal
// bell character.
func (bs *bellSkipper) Write(b []byte) (int, error) {
	const charBell = 7 // c.f. readline.CharBell
	if len(b) == 1 && b[0] == charBell {
		return len(b), nil
	}
	return os.Stdout.Write(b)
}

// Close implements an io.WriterCloser over os.Stderr.
func (bs *bellSkipper) Close() error {
	return os.Stdout.Close()
}
