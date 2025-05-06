package generator

import (
	"blog-down/pkg/parser"
	_ "embed"
	"html/template"
	"path/filepath"
)

type MarkdownPage struct {
	Title        string            `yaml:"title"`
	Description  string            `yaml:"description"`
	MarkdownPath string            `yaml:"markdown-path"`
	PublishedAt  string            `yaml:"published-at"`
	Tags         []string          `yaml:"tags"`
	Metadata     map[string]string `yaml:"metadata"`
}

func generateMarkdownPage(page MarkdownPage, outputPath string) (string, template.HTML) {

	htmlPage := parser.MarkdownToHtml(page.MarkdownPath)

	title := page.Title

	if title == "" {
		title = htmlPage.Title
	}

	slug := slugify(title)
	outFile := filepath.Join(outputPath, slug+".html")

	return outFile, htmlPage.HTML
}
