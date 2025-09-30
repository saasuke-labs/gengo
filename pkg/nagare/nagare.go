package nagare

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

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

// NagareRenderer renders nagare code blocks by calling the HTTP service
type NagareRenderer struct {
	ServiceURL string
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

	// Call the HTTP service
	htmlContent, err := r.callNagareService(nagareBlock.Content)
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

	// Write the HTML response directly
	w.WriteString(htmlContent)

	return ast.WalkContinue, nil
}

func (r *NagareRenderer) callNagareService(content []byte) (string, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second, // 5 second timeout
	}

	// Create request with context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", r.ServiceURL, bytes.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("failed to create nagare request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call nagare service: %w", err)
	}
	defer resp.Body.Close()

	// Read response body (for both success and error cases)
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Include the error message from the response body
		errorMsg := string(responseBytes)
		if errorMsg == "" {
			errorMsg = "no error details provided"
		}
		return "", fmt.Errorf("nagare service returned status %d: %s", resp.StatusCode, errorMsg)
	}

	return string(responseBytes), nil
}

// Extension represents the nagare extension
type Extension struct {
	ServiceURL string
}

// Extend implements goldmark.Extender.Extend
func (e *Extension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&NagareTransformer{}, 999),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&NagareRenderer{ServiceURL: e.ServiceURL}, 999),
	))
}

// NewNagareExtension creates a new nagare extension
func NewNagareExtension(serviceURL string) *Extension {
	return &Extension{ServiceURL: serviceURL}
}
