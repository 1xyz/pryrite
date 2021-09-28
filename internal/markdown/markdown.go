package markdown

import (
	"io/ioutil"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

const (
	DescriptionChunk ChunkType = iota
	CodeChunk
)

func Split(content string, handler ChunkFunc) (title string, err error) {
	return render(content, handler)
}

// only for debugging
func Dump(content string) error {
	_, err := render(content, nil)
	return err
}

func render(content string, handler ChunkFunc) (title string, err error) {
	cr := newChunkRenderer(handler, handler == nil)
	r := renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(cr, 1)))
	md := goldmark.New(goldmark.WithRenderer(r))
	err = md.Convert([]byte(content), ioutil.Discard)
	if err == nil {
		err = cr.finish()
	}
	title = cr.title
	return
}
