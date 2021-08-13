package tools

import (
	"os"
	"path/filepath"
)

var myDirectory string

func MyPathTo(filename string) string {
	if myDirectory == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		myDir := filepath.Join(homeDir, ".aardvark")
		err = os.Mkdir(myDir, 0700)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}

		myDirectory = myDir
	}

	return filepath.Join(myDirectory, filename)
}
