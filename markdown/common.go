package markdown

import (
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

func NewAnsiRenderer(styleName string) (*ansi.ANSIRenderer, error) {
	var style *ansi.StyleConfig

	if styleName == "" {
		style = &glamour.DarkStyleConfig
	} else {
		style = glamour.DefaultStyles[styleName]
		if style == nil {
			return nil, fmt.Errorf("unknown style name: %s", styleName)
		}
	}

	return ansi.NewRenderer(ansi.Options{
		ColorProfile: termenv.TrueColor,
		Styles:       *style,
	}), nil
}

func NewMarkdown(r interface{}) goldmark.Markdown {
	rend := renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(r, 1)))
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRenderer(rend),
	)
}
