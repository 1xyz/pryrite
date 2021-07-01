package components

import (
	"errors"
	"github.com/aardlabs/terminal-poc/tools"
)

var ErrNoEntryPicked = errors.New("no entry picked")

func colLen() int {
	_, c, err := tools.GetTermWindowSize()
	if err != nil {
		return 40 // default col len
	}
	return c
}
