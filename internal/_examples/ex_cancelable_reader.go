package main

import (
	"context"
	"fmt"
	pio "github.com/aardlabs/terminal-poc/internal/io"
	"io"
	"log"
	"os"
	"time"
)

// Example - demonstrating using the cancelable reader
// Note(s): run w/ race flag: to isolate race errors with cancelable rdr
// -- go run -race internal/_examples/ex_cancelable_reader.go
func main() {
	fd := os.Stdin
	for {
		rdr := pio.NewCancelableReadCloser(context.Background())
		fmt.Printf("start...\n")
		if err := ioLoop(rdr, fd); err != nil {
			log.Fatalf("error io.Loop = %v", err)
		}
		fmt.Printf("os.stdin fd = %v\n\n", int(fd.Fd()))
	}
}

func ioLoop(rdr *pio.CancelableReadCloser, fd *os.File) error {
	if err := rdr.Start(fd); err != nil {
		return err
	}

	go func(r io.Reader) {
		p := make([]byte, 32*1024)
		for {
			n, err := r.Read(p)
			if err != nil {
				log.Printf("r.Read err = %v", err)
				break
			}
			log.Printf("bytes read = %v\n", n)
			_, ew := os.Stdout.Write(p[0:n])
			if ew != nil {
				log.Printf("write err = %v", ew)
				break
			}
		}
	}(rdr)

	time.Sleep(10 * time.Second)
	if err := rdr.Close(); err != nil {
		log.Printf("rdr/close err = %v", err)
	}
	return nil
}
