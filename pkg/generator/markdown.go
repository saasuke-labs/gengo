package generator

import (
	"html/template"

	"github.com/saasuke-labs/gengo/pkg/parser"
)

func generateMarkdownPage(markdownPath string) template.HTML {

	htmlPage := parser.MarkdownToHtml(markdownPath)

	return htmlPage.HTML
}
