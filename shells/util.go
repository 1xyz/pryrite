package shells

import (
	"os"

	"github.com/mitchellh/go-ps"
)

func GetParent() (ps.Process, error) {
	parentPID := os.Getppid()
	parent, err := ps.FindProcess(parentPID)
	if err != nil {
		return nil, err
	}

	if parent.Executable() == "go" {
		// this is only during development when letting go compile-and-run...
		parentPID = parent.PPid()
		parent, err = ps.FindProcess(parentPID)
		if err != nil {
			return nil, err
		}
	}

	return parent, nil
}
