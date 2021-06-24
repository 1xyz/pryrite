package executor

import (
	"io"
	"regexp"
	"sync"
	"time"

	"github.com/aardlabs/terminal-poc/tools"
)

type readWriterProxy struct {
	name string

	markerRE    *regexp.Regexp
	markerFound func(string)

	reader io.Reader
	writer io.WriteCloser
	rlock  sync.Mutex
	wlock  sync.Mutex

	lastWrite time.Time
}

func (proxy *readWriterProxy) SetReader(reader io.Reader) {
	proxy.rlock.Lock()
	proxy.reader = reader
	proxy.rlock.Unlock()
}

func (proxy *readWriterProxy) SetWriter(writer io.WriteCloser) {
	proxy.SetWriterMarker(writer, nil, nil)
}

func (proxy *readWriterProxy) SetWriterMarker(writer io.WriteCloser, markerRE *regexp.Regexp, markerFound func(string)) {
	proxy.wlock.Lock()
	wtr := proxy.writer
	lastWrite := proxy.lastWrite
	proxy.wlock.Unlock()

	if writer == nil && wtr != nil {
		// block (UNLOCKED!) to make sure the upstream writer has all the bytes before we
		// let the execute complete
		if time.Since(lastWrite) > time.Second {
			time.Sleep(100 * time.Millisecond)
		}

		err := wtr.Close()
		if err != nil {
			tools.Log.Err(err).Str("proxy", proxy.name).Msg("writer close failed")
		}

		proxy.wlock.Lock()
		proxy.lastWrite = time.Unix(0, 0)
	} else {
		proxy.wlock.Lock()
	}

	proxy.markerRE = markerRE
	proxy.markerFound = markerFound
	proxy.writer = writer
	proxy.wlock.Unlock()
}

func (proxy *readWriterProxy) Read(buf []byte) (int, error) {
	proxy.rlock.Lock()
	defer proxy.rlock.Unlock()

	if proxy.reader == nil {
		return 0, io.EOF
	}

	return proxy.reader.Read(buf)
}

func (proxy *readWriterProxy) Write(data []byte) (int, error) {
	proxy.wlock.Lock()
	defer proxy.wlock.Unlock()

	if proxy.writer == nil {
		tools.Log.Error().Str("proxy", proxy.name).Str("data", string(data)).
			Msg("asked to write without a writer assigned")
		// lie to bash about the success of the write to avoid killing our repl
		return len(data), nil
	}

	if proxy.markerRE == nil {
		return proxy.writer.Write(data)
	}

	origLen := len(data) // need to respond with the original length for success

	var found []byte
	data = proxy.markerRE.ReplaceAllFunc(data, func(match []byte) []byte {
		found = match
		return nil
	})

	_, err := proxy.writer.Write(data)
	proxy.lastWrite = time.Now()

	if found != nil {
		// give the caller time to finish before we record "done"
		go func() {
			time.Sleep(10 * time.Millisecond)
			proxy.markerFound(string(found))
		}()
	}

	return origLen, err
}
