package generator

import (
	"bytes"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func savePage(content template.HTML, outputPath string) {
	// Create the output directory if it doesn't exist
	// TODO - Optimize and create the directory only once for each section
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
		return
	}
	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
		return
	}
	defer f.Close()

	_, err = f.WriteString(string(content))
	if err != nil {
		log.Fatalf("failed to write to output file: %v", err)
		return
	}
}

func applyTemplate(templatePath string, data PageData) template.HTML {

	tmpl := template.Must(template.ParseFiles(templatePath))

	html := bytes.NewBufferString("")

	err := tmpl.Execute(html, data)

	if err != nil {
		log.Fatalf("failed to execute template: %v", err)
		return ""
	}

	return template.HTML(html.String())
}

func convertExtension(path, newExt string) string {
	base := filepath.Base(path)                         // e.g. "graphql-schema-stitching.mdx"
	ext := filepath.Ext(base)                           // e.g. ".mdx"
	name := strings.TrimSuffix(base, ext)               // e.g. "graphql-schema-stitching"
	return name + "." + strings.TrimPrefix(newExt, ".") // e.g. "graphql-schema-stitching.html"
}
