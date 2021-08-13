package tools

import (
	"fmt"
	"os"
	"runtime"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// GetTermWindowSize returns the terminal window size in rows, columns
func GetTermWindowSize() (int, int, error) {
	var rows, cols int

	if runtime.GOOS == "windows" {
		// FIXME: windows either crashes with creak/pty or errors out w/ term.GetSize
		// these are the defaults for the new windows terminal...
		rows = 30
		cols = 120
	} else {
		var err error
		rows, cols, err = pty.Getsize(os.Stdin)
		if err != nil {
			return 0, 0, fmt.Errorf("pty.Getsize err = %v", err)
		}
	}

	return rows, cols, nil
}

// IsTermEnabled returns true if the given file descriptor is a terminal.
func IsTermEnabled(fd int) bool {
	return term.IsTerminal(fd)
}
