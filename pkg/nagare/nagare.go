package nagare

import (
	"bytes"
	"fmt"
	"strings"

	nagarelib "github.com/saasuke-labs/nagare/pkg/nagare"
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

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if fcb, ok := n.(*ast.FencedCodeBlock); ok {
			if fcb.Info != nil && bytes.Equal(fcb.Info.Value(reader.Source()), []byte("nagare")) {
				nagareBlock := &NagareCodeBlock{}

				var content bytes.Buffer
				for i := 0; i < fcb.Lines().Len(); i++ {
					line := fcb.Lines().At(i)
					content.Write(line.Value(reader.Source()))
				}
				nagareBlock.Content = content.Bytes()

				nagareBlocks = append(nagareBlocks, struct {
					parent ast.Node
					fcb    *ast.FencedCodeBlock
					block  *NagareCodeBlock
				}{fcb.Parent(), fcb, nagareBlock})
			}
		}

		return ast.WalkContinue, nil
	})

	for _, nb := range nagareBlocks {
		nb.parent.ReplaceChild(nb.parent, nb.fcb, nb.block)
	}
}

// NagareRenderer renders nagare code blocks using the nagare library
type NagareRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs
func (r *NagareRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(nagareCodeBlockKind, r.renderNagareCodeBlock)
}

func (r *NagareRenderer) renderNagareCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	nagareBlock := node.(*NagareCodeBlock)

	svg, err := renderNagareSVG(string(nagareBlock.Content))
	if err != nil {
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

	w.WriteString(svg)
	return ast.WalkContinue, nil
}

func renderNagareSVG(code string) (string, error) {
	svg, err := nagarelib.RenderToSVG(code)
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(svg)
	if trimmed == "" {
		return "", fmt.Errorf("nagare returned an empty SVG")
	}
	if isBareBackgroundSVG(trimmed) {
		return "", fmt.Errorf("nagare rendered only a background; check the block for chart or diagram errors")
	}
	return svg, nil
}

func isBareBackgroundSVG(svg string) bool {
	if !strings.Contains(svg, "<svg") {
		return false
	}
	graphicTags := []string{"<g", "<circle", "<line", "<polyline", "<path", "<text", "<ellipse", "<polygon"}
	for _, tag := range graphicTags {
		if strings.Contains(svg, tag) {
			return false
		}
	}
	return strings.Count(svg, "<rect") == 1
}

// Extension represents the nagare extension
type Extension struct{}

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
func NewNagareExtension() *Extension { return &Extension{} }
