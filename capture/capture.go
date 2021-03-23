package capture

import (
	"fmt"
	"github.com/creack/pty"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func Capture(sessionID, filename, cmdName string, args ...string) error {
	cmd := exec.Command(cmdName, args...)

	// Refer "github.com/creack/pty" for example(s)
	// Start the command with a pty.
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("pty.Start %v", err)
	}

	// Make sure to close the pty at the end.
	defer func() {
		if err = ptmx.Close(); err != nil {
			log.Warnf("ptmx.Close err = %v", err)
		}
	}() // Best effort.

	// Handle pty size
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Warnf("pty.InheritSize err = %v", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.

	// Set stdin in raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("term.MakeRaw err = %v", err)
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			log.Warnf("term.Restore err = %v", err)
		}
	}() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	go func() {
		// Copy stdin -> pty
		// blocks until EOF is reached
		if _, err := io.Copy(ptmx, os.Stdin); err != nil {
			log.Warnf("io.copy(ptmx, os.Stdin) err = %v", err)
		}
	}()

	outfw, err := NewFrameSetWriter(sessionID, filename)
	if err != nil {
		return fmt.Errorf("NewFrameSetWriter error = %v", err)
	}
	defer func() {
		if err := outfw.Close(); err != nil {
			log.Warnf("outfw.close err = %v", err)
		}
	}()

	omw := io.MultiWriter(os.Stdout, outfw)
	// Copy pty -> stdout; blocks until EOF is reached
	go func() {
		if _, err := io.Copy(omw, ptmx); err != nil {
			log.Warnf("io.copy ptmx to os.Stdout %v", err)
		}
	}()
	if err := cmd.Wait(); err != nil {
		log.Printf("cmd.wait err = %v", err)
	}
	return nil
}
