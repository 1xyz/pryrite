package history

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/tools"
	"github.com/docopt/docopt-go"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func Cmd(argv []string, version string) error {
	usage := `
usage: pruney history append <command>
       pruney history append-in
       pruney history show

Options:
  -h --help            Show this screen.
`
	opts, err := docopt.ParseArgs(usage, argv, version)
	if err != nil {
		tools.Log.Fatal().Msgf("error parsing arguments. err=%v", err)
	}

	fmt.Printf("opts = %v", opts)
	if tools.OptsBool(opts, "show") {
		h, err := New()
		if err != nil {
			return err
		}
		if _, err := h.GetAll(); err != nil {
			return err
		}
	} else if tools.OptsBool(opts, "append") {
		return appendHistory(tools.OptsStr(opts, "<command>"))
	} else if tools.OptsBool(opts, "append-in") {
		command, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		return appendHistory(string(command))
	} else {
		return fmt.Errorf("no command found")
	}
	return nil
}

func appendHistory(cmd string) error {
	cmd = strings.TrimSpace(cmd)
	if len(cmd) == 0 {
		return fmt.Errorf("empty command provided")
	}
	h, err := New()
	if err != nil {
		return err
	}
	item := &Item{
		CreatedAt: time.Now().UTC(),
		Command:   cmd,
	}
	if err := h.Append(item); err != nil {
		return err
	}
	return err
}
