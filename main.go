package main

import (
	_ "embed"
	"fmt"
	"github.com/1xyz/pryrite/app"
	"github.com/1xyz/pryrite/mdtools/cmd"
	"github.com/1xyz/pryrite/tools"
	"os"
)

var (
	//go:embed version.txt
	version string
)

func main() {
	app.Name = "pryrite"

	wr, err := tools.OpenLogger(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tools.OpenLogger err = %v", err)
		os.Exit(1)
	}
	defer func() {
		if err := wr.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "wr.close %v", err)
		}
	}()

	// Don't handle error will be handled by cobra
	cmd.Execute()
}
