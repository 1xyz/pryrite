package markdown

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

const codeCopierControllerName = "code-copier"

type CodeBlockRenderer struct{}

func (cr *CodeBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindCodeBlock, cr.renderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, cr.renderCodeBlock)
}

func (cr *CodeBlockRenderer) renderCodeBlock(w util.BufWriter, source []byte, mdNode ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		var language []byte
		if mdNode.Kind() == ast.KindFencedCodeBlock {
			n := mdNode.(*ast.FencedCodeBlock)
			language = n.Language(source)
		}

		w.WriteString("<pre><code")

		if language != nil {
			w.WriteString(" class=\"language-")
			w.Write(language)
			w.WriteString("\"")
		}

		content := collectContent(source, mdNode)
		pc := extractPromptCommand(content)

		w.WriteString(" data-controller=\"")
		w.WriteString(codeCopierControllerName)
		w.WriteString("\"")

		if pc != nil {
			w.WriteString(" data-")
			w.WriteString(codeCopierControllerName)
			w.WriteString("-subargs-value=\"[")
			w.WriteString(pc.command.Join(","))
			w.WriteString("]\"")
		}

		w.WriteByte('>')
		w.WriteString(content)
	} else {
		w.WriteString("</code></pre>\n")
	}
	return ast.WalkContinue, nil
}

func collectContent(source []byte, n ast.Node) string {
	content := ""
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		content += string(line.Value(source))
	}
	return content
}
