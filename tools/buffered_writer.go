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

	for {
		// this must be locked before waiting, which will be automatically
		// unlocked while waiting and locked when woken up by a signal
		bw.cond.L.Lock()
		bw.cond.Wait()

		needLock := false
		moreToRead := true

		for moreToRead {
			if needLock {
				// only occurs when we're looping to drain the reader buf
				bw.cond.L.Lock()
			}

			n, err := bw.buf.Read(data)
			// unlock ASAP to make sure the writes to the buffer can continue
			// since our writer below might be slow
			bw.cond.L.Unlock()
			needLock = true
			if err != nil {
				// do not return if this is EOF unless we were asked to flush
				// and this is our "done" cycle
				if err != io.EOF || bw.readWriterFlush {
					bw.readWriterErr = err
					return
				}

				moreToRead = false
			}

			if n > 0 {
				_, err = bw.Writer.Write(data[0:n])
				if err != nil {
					bw.readWriterErr = err
					return
				}

				moreToRead = bw.readWriterFlush || n == len(data)
			}
		}
	}
}
