package executor

import (
	"os"
	"os/signal"
	"syscall"
)

type Canceler struct {
	stop chan bool
}

func (clr *Canceler) OnCancel(cb func()) {
	clr.stop = make(chan bool, 1)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT)
	go func() {
		defer signal.Stop(sigCh)
		for {
			select {
			case <-clr.stop:
				return
			case <-sigCh:
				cb()
			}
		}
	}()
}

func (clr *Canceler) Stop() {
	clr.stop <- true
}
