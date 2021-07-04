package markdown

import (
	"github.com/muesli/termenv"
)

type LinenumRenderer struct {
	TermRenderer
}

func NewLinenumRenderer(styleName string) (*LinenumRenderer, error) {
	var err error
	lr := &LinenumRenderer{}
	lr.ansiRenderer, err = NewAnsiRenderer(styleName)
	return lr, err
}

func (lr *LinenumRenderer) Render(content string, cursor *Cursor) (string, error) {
	cp := termenv.ColorProfile()
	arrow := "\U000025AC\U000025B6 " // BLACK RECTANGLE followed by BLACK RIGHT-POINTING TRIANGLE
	indicator := termenv.String(arrow).Foreground(cp.Color("#ff0000")).Bold()
	writer := NewLinenumWriter(1, 10, indicator.String(), cursor)
	writer.SetIndicatorWidth(indicator.Width())
	content = writer.MarkCursorLocation(content)

	md := NewMarkdown(lr)
	err := md.Convert([]byte(content), writer)
	ferr := writer.Flush()
	if err == nil {
		err = ferr
	}

	return writer.String(), err
}
