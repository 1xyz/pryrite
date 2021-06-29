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

	writer io.WriteCloser
	wlock  sync.Mutex

	lastWrite time.Time
}

func (proxy *readWriterProxy) Monitor(output io.Reader) {
	go func() {
		buf := make([]byte, 65536)
		for {
			n, err := output.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				tools.Log.Err(err).Str("name", proxy.name).Msg("Unable to read monitored output")
				break
			}
			if n > 0 {
				_, err = proxy.Write(buf[0:n])
				if err != nil {
					tools.Log.Err(err).Str("name", proxy.name).Str("data", string(buf[0:n])).
						Msg("Unable to write monitored output")
				}
			}
		}
	}()
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

func (proxy *readWriterProxy) Write(data []byte) (int, error) {
	proxy.wlock.Lock()
	defer proxy.wlock.Unlock()

	origLen := len(data) // need to respond with the original length for success

	// always look for markers, even if no writer was assigned
	if proxy.markerRE != nil {
		var found []byte

		data = proxy.markerRE.ReplaceAllFunc(data, func(match []byte) []byte {
			if found != nil {
				// only replace the first one found, but this is weird, so warn
				tools.Log.Warn().Str("found", string(found)).Str("match", string(match)).
					Str("regexp", proxy.markerRE.String()).Msg("Found more than one match")
				return match
			}
			found = match
			return nil
		})

		if found != nil {
			// give the caller time to finish before we record "done"
			go func() {
				time.Sleep(10 * time.Millisecond)
				proxy.markerFound(string(found))
			}()
		}
	}

	if proxy.writer == nil {
		tools.Log.Error().Str("proxy", proxy.name).Str("data", string(data)).
			Msg("asked to write without a writer assigned")
		// lie to caller about the success of the write to avoid killing our repl
		return origLen, nil
	}

	_, err := proxy.writer.Write(data)
	proxy.lastWrite = time.Now()

	return origLen, err
}
