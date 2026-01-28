package nagare

import (
	"bytes"

	nagarediagram "github.com/saasuke-labs/nagare/pkg/diagram"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// NagareCodeBlock represents a nagare code block in the AST
type NagareCodeBlock struct {
	ast.BaseBlock
	Content []byte
}

// Dump implements ast.Node.Dump
func (n *NagareCodeBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// Kind implements ast.Node.Kind
func (n *NagareCodeBlock) Kind() ast.NodeKind {
	return nagareCodeBlockKind
}

var nagareCodeBlockKind = ast.NewNodeKind("NagareCodeBlock")

// NagareTransformer transforms fenced code blocks with language "nagare" into NagareCodeBlock
type NagareTransformer struct{}

// Transform implements ast.Transformer.Transform
func (t *NagareTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	var nagareBlocks []struct {
		parent ast.Node
		fcb    *ast.FencedCodeBlock
		block  *NagareCodeBlock
	}

	// First pass: collect all nagare blocks
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if fcb, ok := n.(*ast.FencedCodeBlock); ok {
			// Check if this is a nagare code block
			if fcb.Info != nil && bytes.Equal(fcb.Info.Value(reader.Source()), []byte("nagare")) {
				// Create a new NagareCodeBlock
				nagareBlock := &NagareCodeBlock{}

				// Copy content from the fenced code block
				var content bytes.Buffer
				for i := 0; i < fcb.Lines().Len(); i++ {
					line := fcb.Lines().At(i)
					content.Write(line.Value(reader.Source()))
				}
				nagareBlock.Content = content.Bytes()

				// Store for replacement
				nagareBlocks = append(nagareBlocks, struct {
					parent ast.Node
					fcb    *ast.FencedCodeBlock
					block  *NagareCodeBlock
				}{fcb.Parent(), fcb, nagareBlock})
			}
		}

		return ast.WalkContinue, nil
	})

	// Second pass: replace all collected blocks
	for _, nb := range nagareBlocks {
		nb.parent.ReplaceChild(nb.parent, nb.fcb, nb.block)
	}
}

// NagareRenderer renders nagare code blocks using the nagare library
type NagareRenderer struct {
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs
func (r *NagareRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(nagareCodeBlockKind, r.renderNagareCodeBlock)
}

func (r *NagareRenderer) renderNagareCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	nagareBlock := node.(*NagareCodeBlock)

	// Use the nagare library directly to render the diagram
	svg, err := nagarediagram.CreateDiagram(string(nagareBlock.Content))
	if err != nil {
		// Fallback: render as regular code block with error message
		w.WriteString("<div class=\"nagare-error\">")
		w.WriteString("<p><strong>Error processing nagare block:</strong> ")
		w.Write(util.EscapeHTML([]byte(err.Error())))
		w.WriteString("</p>")
		w.WriteString("</div>")
		w.WriteString("<pre><code class=\"language-nagare\">")
		w.Write(util.EscapeHTML(nagareBlock.Content))
		w.WriteString("</code></pre>")
		return ast.WalkContinue, nil
	}

	// Write the SVG directly
	w.WriteString(svg)

	return ast.WalkContinue, nil
}

// Extension represents the nagare extension
type Extension struct {
}

// Extend implements goldmark.Extender.Extend
func (e *Extension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&NagareTransformer{}, 999),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&NagareRenderer{}, 999),
	))
}

// NewNagareExtension creates a new nagare extension
func NewNagareExtension() *Extension {
	return &Extension{}
}
