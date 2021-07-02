package tools

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type LinenumWriter struct {
	linenum    int
	cursorLine int

	numPrefixFmt    string
	cursorPrefixFmt string

	inputBuf  bytes.Buffer
	outputBuf bytes.Buffer
}

func NewLinenumWriter(start, cursorLine, prefixWidth int, cursor string) *LinenumWriter {
	return &LinenumWriter{
		linenum:         start,
		cursorLine:      cursorLine,
		numPrefixFmt:    fmt.Sprint("%", prefixWidth, "d: %s"),
		cursorPrefixFmt: fmt.Sprint("%", prefixWidth, "s: %s"),
	}
}

func (lw *LinenumWriter) Write(data []byte) (n int, err error) {
	n, err = lw.inputBuf.Write(data)
	if err != nil {
		return
	}

	for {
		var line *string

		line, err = lw.getLine()
		if err != nil {
			return
		}

		if line == nil {
			// no line found in the buffer yet...
			return
		}

		err = lw.write(*line)
		if err != nil {
			return
		}
	}
}

func (lw *LinenumWriter) Close() error {
	// flush remaining into the output
	err := lw.write(lw.inputBuf.String())
	lw.inputBuf.Truncate(0)
	return err
}

func (lw *LinenumWriter) String() string {
	return lw.outputBuf.String()
}

func (lw *LinenumWriter) getLine() (*string, error) {
	i := strings.IndexRune(lw.inputBuf.String(), '\n')
	if i < 0 {
		return nil, nil
	}

	i++ // length is offset plus one

	buf := make([]byte, i)
	n, err := lw.inputBuf.Read(buf)
	if err != nil {
		return nil, err
	}

	if n != i {
		return nil, io.ErrShortBuffer
	}

	str := string(buf)
	return &str, nil
}

func (lw *LinenumWriter) write(line string) error {
	formatted := fmt.Sprintf(lw.numPrefixFmt, lw.linenum, line)
	lw.linenum++

	_, err := lw.outputBuf.WriteString(formatted)
	return err
}
