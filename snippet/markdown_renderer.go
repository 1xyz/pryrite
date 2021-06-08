package snippet

import (
	"bytes"
	"fmt"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/muesli/termenv"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type MarkdownRenderer struct {
	ar *ansi.ANSIRenderer
}

func NewMarkdownRenderer(styleName string) (*MarkdownRenderer, error) {
	var style *ansi.StyleConfig

	if styleName == "" {
		style = &glamour.DarkStyleConfig
	} else {
		style = glamour.DefaultStyles[styleName]
		if style == nil {
			return nil, fmt.Errorf("unknown style name: %s", styleName)
		}
	}

	return &MarkdownRenderer{
		ar: ansi.NewRenderer(ansi.Options{
			ColorProfile: termenv.TrueColor,
			Styles:       *style,
		}),
	}, nil
}

func (mr *MarkdownRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	mr.ar.RegisterFuncs(reg)
}

func (mr *MarkdownRenderer) Render(content string) (string, error) {
	rend := renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(mr, 1)))

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRenderer(rend),
	)

	buf := &bytes.Buffer{}
	err := md.Convert([]byte(content), buf)

	return buf.String(), err
}
