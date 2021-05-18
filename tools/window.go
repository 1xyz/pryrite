package tools

import (
	"fmt"
	"github.com/creack/pty"
	"golang.org/x/term"
	"os"
)

// GetTermWindowSize returns the terminal window size in rows, columns
func GetTermWindowSize() (int, int, error) {
	rows, cols, err := pty.Getsize(os.Stdin)
	if err != nil {
		return 0, 0, fmt.Errorf("pty.Getsize err = %v", err)
	}
	return rows, cols, nil
}

// IsTermEnabled returns true if the given file descriptor is a terminal.
func IsTermEnabled(fd int) bool {
	return term.IsTerminal(fd)
}
