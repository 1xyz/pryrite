package markdown

import (
	"bytes"

	"github.com/charmbracelet/glamour/ansi"
	"github.com/yuin/goldmark/renderer"
)

type TermRenderer struct {
	ansiRenderer *ansi.ANSIRenderer
}

func NewTermRenderer(styleName string) (*TermRenderer, error) {
	var err error
	tr := &TermRenderer{}
	tr.ansiRenderer, err = NewAnsiRenderer(styleName)
	return tr, err
}

func (tr *TermRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	tr.ansiRenderer.RegisterFuncs(reg)
}

func (tr *TermRenderer) Render(content string) (string, error) {
	md := NewMarkdown(tr)
	buf := &bytes.Buffer{}
	err := md.Convert([]byte(content), buf)
	return buf.String(), err
}
