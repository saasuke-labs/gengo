package generator

import (
	"blog-down/pkg/parser"
	_ "embed"
	"html/template"
)

func generateMarkdownPage(markdownPath string) template.HTML {

	htmlPage := parser.MarkdownToHtml(markdownPath)

	return htmlPage.HTML
}
