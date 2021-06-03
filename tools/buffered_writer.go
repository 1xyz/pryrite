package tools

import (
	"bytes"
	"io"
	"sync"
)

type BufferedWriteCloser struct {
	Writer io.Writer

	buf  bytes.Buffer
	cond sync.Cond

	readWriterFlush bool
	readWriterErr   error
	readWriterDone  chan bool
}

// This WriteCloser is guaranteed to never let a Write block. However, it's at
// the expense of more memory usage (especially if the writer passed in is slow).
func NewBufferedWriteCloser(writer io.Writer) *BufferedWriteCloser {
	bw := &BufferedWriteCloser{
		Writer:         writer,
		buf:            bytes.Buffer{},
		cond:           sync.Cond{L: &sync.Mutex{}},
		readWriterDone: make(chan bool, 1),
	}

	go bw.readWriter()

	return bw
}

func (bw *BufferedWriteCloser) Write(data []byte) (n int, err error) {
	if bw.readWriterErr != nil {
		return 0, bw.readWriterErr
	}

	if len(data) == 0 {
		return 0, nil
	}

	bw.cond.L.Lock()
	n, err = bw.buf.Write(data)
	bw.cond.Signal()
	bw.cond.L.Unlock()
	return
}

func (bw *BufferedWriteCloser) Close() error {
	// ask reader to read until no more bytes (io.EOF)
	bw.cond.L.Lock()
	bw.readWriterFlush = true
	bw.cond.Signal()
	bw.cond.L.Unlock()

	// wait for reader to exit
	<-bw.readWriterDone

	if bw.readWriterErr == io.EOF {
		return nil
	}

	return bw.readWriterErr
}

func (bw *BufferedWriteCloser) readWriter() {
	defer func() {
		bw.readWriterDone <- true
	}()

	data := make([]byte, 65536)

	// condition needs to be locked before waiting, which will be automatically
	// unlocked while waiting and locked when woken up by a signal
	bw.cond.L.Lock()

	for {
		bw.cond.Wait()
		moreToRead := true
		for moreToRead {
			n, err := bw.buf.Read(data)
			if err != nil {
				// do not return if this is EOF unless we were asked to flush and this is our "done" cycle
				if err != io.EOF || bw.readWriterFlush {
					bw.readWriterErr = err
					bw.cond.L.Unlock()
					return
				}

				moreToRead = false
			}

			if n > 0 {
				_, err = bw.Writer.Write(data[0:n])
				if err != nil {
					bw.readWriterErr = err
					bw.cond.L.Unlock()
					return
				}

				moreToRead = bw.readWriterFlush || n == len(data)
			}
		}
	}
}
