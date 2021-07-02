package markdown

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
)

var (
	cursorStartMark = string([]byte{2}) // STX - "Start of Text" ASCII
	cursorStopMark  = string([]byte{3}) // ETX - "End of Text" ASCII
)

type LinenumWriter struct {
	linenum int

	cursor *Cursor

	indicator      string
	indicatorWidth int

	prefixFmt   string
	prefixWidth int
	termWidth   int

	contentMarked bool
	inCursor      bool

	input  bytes.Buffer
	output bytes.Buffer

	addIndicatorToNextLine bool
}

func NewLinenumWriter(start, prefixWidth int, indicator string, cursor *Cursor) *LinenumWriter {
	width, _, _ := terminal.GetSize(0)

	return &LinenumWriter{
		linenum:        start,
		cursor:         cursor,
		indicator:      indicator,
		indicatorWidth: len(indicator),
		prefixFmt:      fmt.Sprint("%", prefixWidth-2, "d: "), // two less to account for colon/space following
		prefixWidth:    prefixWidth,
		termWidth:      width,
	}
}

func (lw *LinenumWriter) SetIndicatorWidth(width int) {
	lw.indicatorWidth = width
}

func (lw *LinenumWriter) AddIndicatorToNextLine() {
	lw.addIndicatorToNextLine = true
}

func (lw *LinenumWriter) MarkCursorLocation(content string) string {
	if lw.cursor == nil {
		return content
	}

	lw.contentMarked = true

	stop := lw.cursor.Stop
	if stop > 0 && content[stop-1] == '\n' {
		stop-- // place end marker just in front of the last newline
	}

	return content[0:lw.cursor.Start] +
		cursorStartMark + content[lw.cursor.Start:stop] + cursorStopMark +
		content[stop:]
}

func (lw *LinenumWriter) Write(data []byte) (n int, err error) {
	n, err = lw.input.Write(data)
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

func (lw *LinenumWriter) Flush() (err error) {
	if lw.input.Len() > 0 {
		// flush remaining into the output
		err = lw.write(lw.input.String())
		lw.input.Reset()
	}
	return
}

func (lw *LinenumWriter) String() string {
	return lw.output.String()
}

//--------------------------------------------------------------------------------

func (lw *LinenumWriter) getLine() (*string, error) {
	i := strings.IndexRune(lw.input.String(), '\n')
	if i < 0 {
		return nil, nil
	}

	i++ // length is offset plus one

	buf := make([]byte, i)
	n, err := lw.input.Read(buf)
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
	addIndicator := false

	if lw.contentMarked {
		if strings.Contains(line, cursorStartMark) {
			line = strings.Replace(line, cursorStartMark, "", -1)
			lw.inCursor = true
		}
		addIndicator = lw.inCursor // add even if this is the stopping point
		if strings.Contains(line, cursorStopMark) {
			line = strings.Replace(line, cursorStopMark, "", -1)
			lw.inCursor = false
		}
	}

	var prefix string
	if lw.addIndicatorToNextLine || lw.cursor.Contains(lw.linenum) || addIndicator {
		prefix = lw.getIndicatorPrefix()
		lw.addIndicatorToNextLine = false
	} else {
		prefix = fmt.Sprintf(lw.prefixFmt, lw.linenum)
	}

	wrapped := lw.wrap(prefix + line)
	_, err := lw.output.Write(wrapped)
	lw.linenum++

	return err
}

func (lw *LinenumWriter) getIndicatorPrefix() string {
	space := " "
	if lw.linenum < 10 {
		space += " "
	}

	linenumPrefix := fmt.Sprintf("%s%d: ", space, lw.linenum)
	remainLen := lw.prefixWidth - (lw.indicatorWidth + len(linenumPrefix))

	spacePrefix := ""
	if remainLen > 0 {
		spacePrefix = strings.Repeat(" ", remainLen)
	}

	return spacePrefix + lw.indicator + linenumPrefix
}

func (lw *LinenumWriter) wrap(line string) []byte {
	lineBytes := []byte(line)

	if lw.termWidth <= lw.prefixWidth {
		return lineBytes
	}

	wrapBytes := wordwrap.Bytes(lineBytes, lw.termWidth-lw.prefixWidth)

	// here we ignore the first line since it's already indented with our prefix...
	// but then we have the remaining lines pushed in
	ignoreCnt := lw.prefixWidth
	indentWriter := indent.NewWriter(uint(lw.prefixWidth), func(w io.Writer) {
		if ignoreCnt > 0 {
			ignoreCnt--
		} else {
			w.Write([]byte(" "))
		}
	})

	indentWriter.Write(wrapBytes)

	return indentWriter.Bytes()
}
