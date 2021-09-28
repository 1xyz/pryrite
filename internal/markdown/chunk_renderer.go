package markdown

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

const descriptionContentType = "text/markdown"
const codeContentType = "text/code"

//var Detector = language.NewDetector()

type ChunkType uint
type ChunkFunc func(chunk string, chunkType ChunkType, contentType string) error

type chunkRenderer struct {
	handler   ChunkFunc
	title     string
	level     int
	nextStart int
	lastSrc   []byte
	dumping   bool
}

func newChunkRenderer(chunkHandler ChunkFunc, dump bool) *chunkRenderer {
	return &chunkRenderer{
		handler: chunkHandler,
		level:   0x7fffffff, // any number way bigger than any level we might encounter will suffice
		dumping: dump,
	}
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs
func (cr *chunkRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	if cr.dumping {
		// block elements
		reg.Register(ast.KindDocument, cr.dump)
		reg.Register(ast.KindHeading, cr.dump)
		reg.Register(ast.KindBlockquote, cr.dump)
		reg.Register(ast.KindCodeBlock, cr.dump)
		reg.Register(ast.KindFencedCodeBlock, cr.dump)
		reg.Register(ast.KindHTMLBlock, cr.dump)
		reg.Register(ast.KindList, cr.dump)
		reg.Register(ast.KindListItem, cr.dump)
		reg.Register(ast.KindParagraph, cr.dump)
		reg.Register(ast.KindTextBlock, cr.dump)
		reg.Register(ast.KindThematicBreak, cr.dump)
		// inline elements
		reg.Register(ast.KindAutoLink, cr.dump)
		reg.Register(ast.KindCodeSpan, cr.dump)
		reg.Register(ast.KindEmphasis, cr.dump)
		reg.Register(ast.KindImage, cr.dump)
		reg.Register(ast.KindLink, cr.dump)
		reg.Register(ast.KindRawHTML, cr.dump)
		reg.Register(ast.KindText, cr.dump)
		reg.Register(ast.KindString, cr.dump)
	} else {
		// we only want to slurp out headings (for title) and code snippets
		reg.Register(ast.KindHeading, cr.getTitle)
		reg.Register(ast.KindCodeBlock, cr.split)
		reg.Register(ast.KindFencedCodeBlock, cr.split)

		// unfortunately, some of goldmark's logic is dependant on executing a
		// callback, otherwise it'll hit a strange "index out of range" error,
		// so we register a no-op for those we're not using
		reg.Register(ast.KindDocument, cr.noop)
		reg.Register(ast.KindBlockquote, cr.noop)
		reg.Register(ast.KindHTMLBlock, cr.noop)
		reg.Register(ast.KindList, cr.noop)
		reg.Register(ast.KindListItem, cr.noop)
		reg.Register(ast.KindParagraph, cr.noop)
		reg.Register(ast.KindTextBlock, cr.noop)
		reg.Register(ast.KindThematicBreak, cr.noop)
		reg.Register(ast.KindAutoLink, cr.noop)
		reg.Register(ast.KindCodeSpan, cr.noop)
		reg.Register(ast.KindEmphasis, cr.noop)
		reg.Register(ast.KindImage, cr.noop)
		reg.Register(ast.KindLink, cr.noop)
		reg.Register(ast.KindRawHTML, cr.noop)
		reg.Register(ast.KindText, cr.noop)
		reg.Register(ast.KindString, cr.noop)
	}
}

func (cr *chunkRenderer) getTitle(w util.BufWriter, source []byte, mdNode ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	heading := mdNode.(*ast.Heading)
	if heading.Level < cr.level {
		cr.title = string(heading.Text(source))
		cr.level = heading.Level
	}

	return ast.WalkContinue, nil
}

func (cr *chunkRenderer) split(w util.BufWriter, source []byte, mdNode ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	var start, stop int
	var language string
	var descChunk *string

	if mdNode.Type() == ast.TypeBlock {
		lines := mdNode.Lines()
		if lines.Len() > 0 {
			start = eatSpace(lines.At(0).Start-1, source)
			stop = lines.At(lines.Len() - 1).Stop
		} else {
			// ignore empty code blocks since there's nothing to execute
			return ast.WalkContinue, nil
		}
		if mdNode.Kind() == ast.KindFencedCodeBlock {
			language = string(mdNode.(*ast.FencedCodeBlock).Language(source))
		}
	} else {
		textNode := mdNode.FirstChild().(*ast.Text)
		start = textNode.Segment.Start
		stop = textNode.Segment.Stop
	}

	if start > cr.nextStart {
		// copy the content found up to the start of this code chunk
		desc := string(source[cr.nextStart:start])
		descChunk = &desc
	}

	chunk := string(source[start:stop])
	cr.nextStart = stop

	if language == "" {
		//language, _ = Detector.LanguageOf(chunk)
		if descChunk != nil && mdNode.Kind() == ast.KindFencedCodeBlock && language != "" {
			// update the previous chunk with the language value for the "top" part of the fence
			*descChunk = strings.TrimRight(*descChunk, "\r\n") + language + "\n"
		}
	}

	if descChunk != nil {
		err := cr.handler(*descChunk, DescriptionChunk, descriptionContentType)
		if err != nil {
			return ast.WalkStop, err
		}
	}

	if language == "" {
		language = codeContentType
	}

	contentType := extractContentType(chunk, language)

	err := cr.handler(chunk, CodeChunk, contentType)
	if err != nil {
		return ast.WalkStop, err
	}

	cr.lastSrc = source

	return ast.WalkContinue, nil
}

func (cr *chunkRenderer) finish() error {
	// may be more remaining content if last chunk is a fenced code block
	stop := len(cr.lastSrc)
	if cr.nextStart < stop {
		chunk := cr.lastSrc[cr.nextStart:stop]
		return cr.handler(string(chunk), DescriptionChunk, descriptionContentType)
	}

	return nil
}

func eatSpace(pos int, source []byte) int {
	for ; pos > 0; pos-- {
		if source[pos] == ' ' || source[pos] == '\t' {
			continue
		}
		break
	}

	return pos + 1
}

func (cr *chunkRenderer) dump(w util.BufWriter, source []byte, mdNode ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		fmt.Printf("--- %s (%d) %s\n", mdNode.Kind().String(), mdNode.Type(), strings.Repeat("-", 40))
		mdNode.Dump(source, 0)
		if mdNode.Type() == ast.TypeBlock || mdNode.Type() == ast.TypeDocument {
			fmt.Println("lines:")
			lines := mdNode.Lines()
			for i := 0; i < lines.Len(); i++ {
				s := lines.At(i)
				fmt.Printf("  %d [%d:%d:%d]: %s", i, s.Start, s.Stop, s.Padding, s.Value(source))
			}
		} else if mdNode.FirstChild() != nil {
			textNode := mdNode.FirstChild().(*ast.Text)
			s := textNode.Segment
			fmt.Printf("line [%d:%d:%d]: %s", s.Start, s.Stop, s.Padding, s.Value(source))
		}
		fmt.Print("\n" + strings.Repeat("-", 80) + "\n\n")
	}
	return ast.WalkContinue, nil
}

func (cr *chunkRenderer) noop(_ util.BufWriter, source []byte, _ ast.Node, _ bool) (ast.WalkStatus, error) {
	// track in case we have no code snippets so the finish call can dump the content to the handler
	cr.lastSrc = source
	return ast.WalkContinue, nil
}
