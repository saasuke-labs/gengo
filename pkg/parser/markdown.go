package parser

import (
	"bytes"
	"html/template"
	"log"
	"os"

	"github.com/saasuke-labs/gengo/pkg/nagare"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

type HtmlPage struct {
	Title string
	HTML  template.HTML
}

var md goldmark.Markdown

func init() {
	// Configure extensions
	extensions := []goldmark.Extender{
		extension.GFM,
		highlighting.NewHighlighting(
			// highlighting.WithFormatOptions(
			// 	htmlchroma.WithLineNumbers(true),
			// ),
			highlighting.WithStyle("github"), // choose a theme
			highlighting.WithGuessLanguage(false),
		),
		nagare.NewNagareExtension(),
	}

	md = goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithHardWraps(), html.WithXHTML()),
	)
}

func MarkdownToHtml(markdownPath string) HtmlPage {
	content, err := os.ReadFile(markdownPath)
	if err != nil {
		log.Fatalf("failed to read %s: %v", markdownPath, err)
	}
	context := parser.NewContext()
	doc := md.Parser().Parse(text.NewReader(content), parser.WithContext(context))

	title := ""
	h1 := findFirstH1(doc, content)

	if h1 != "" {
		title = h1

	}

	var article bytes.Buffer

	md.Renderer().Render(&article, content, doc)

	return HtmlPage{
		Title: title,
		HTML:  template.HTML(article.String()),
	}
}

func findFirstH1(doc ast.Node, source []byte) string {
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		if h, ok := n.(*ast.Heading); ok && h.Level == 1 {
			return string(h.Text(source))
		}
	}
	return ""
}
