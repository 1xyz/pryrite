package main

import (
	"fmt"
	"github.com/aardlabs/terminal-poc/capture"
	"github.com/docopt/docopt-go"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"time"
)

func main() {
	usage := `Terminal poc.

Usage:
  term_poc capture [<command>...]
  term_poc replay --input=<file>
  term_poc -h | --help
  term_poc --version

Options:
  --input=<file>  Input file to replay content.
  -h --help       Show this screen.

Examples:
  $ term_poc capture
  start capturing output from the current shell until exit or ctrl+D is called 

  $ term_poc capture cat go.mod
  Capture the output of command: cat go.mod

  $ term_poc replay --input /tmp/frameset-755e7e9c-9436-4e2c-8417-1548bba315b2.json
  Replay a capture of events on this console
`
	tStart := time.Now()
	defer func() {
		log.Infof("completed at %v", time.Since(tStart))
	}()

	arguments, err := docopt.ParseDoc(usage)
	if err != nil {
		log.Fatalf("doctopt.ParseDoc err = %v", err)
	}

	log.Debugf("Args: ", arguments)
	doCapture, err := arguments.Bool("capture")
	if err != nil {
		log.Fatalf("arguments.Bool err = %v", err)
	}
	if doCapture {
		cmd, cmdArgs, err := parseCommand(arguments)
		if err != nil {
			log.Fatalf("parseCommand err = %v", err)
		}

		sessionID := uuid.New().String()
		tmpDir, err := ioutil.TempDir("", "frames")
		if err != nil {
			log.Fatalf("ioutil.TempDir err = %v", err)
		}
		outJsonFile := path.Join(tmpDir, fmt.Sprintf("frameset-%s.json", sessionID))
		log.Printf("cmd = %v, cmdArgs = %v sessionID = %s output = %s",
			cmd, cmdArgs, sessionID, outJsonFile)

		if err := capture.Capture(sessionID, outJsonFile, cmd, cmdArgs...); err != nil {
			log.Errorf("capture error = %v", err)
		}
		return
	}

	doReplay, err := arguments.Bool("replay")
	if err != nil {
		log.Fatalf("arguments.Bool err = %v", err)
	}
	if doReplay {
		inpFile, err := arguments.String("--input")
		if err != nil {
			log.Fatalf("argments.String = %v", err)
		}

		log.Printf("Replaying from %s", inpFile)
		if err := capture.Replay(inpFile); err != nil {
			log.Fatalf("capture.Replay err = %v", err)
		}
	}
}

func parseCommand(arguments docopt.Opts) (string, []string, error) {
	iface, ok := arguments["<command>"]
	if !ok {
		return "", nil, fmt.Errorf("cannot find a command in arguments")
	}
	args, ok := iface.([]string)
	if !ok {
		return "", nil, fmt.Errorf("cannot cast to type []string")
	}
	if len(args) == 0 {
		return os.Getenv("SHELL"), []string{}, nil
	}
	if len(args) == 1 {
		return args[0], []string{}, nil
	}
	return args[0], args[1:], nil
}
