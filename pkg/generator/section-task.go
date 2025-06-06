package generator

import (
	"bytes"
	"html/template"
)

type SectionTask struct {
	Title          string
	Section        string
	Sections       []string
	OutputFile     string
	Template       string
	LayoutTemplate string
	Pages          []Page
	Metadata       map[string]string
}

type SectionData struct {
	Section  string
	Metadata map[string]string
	Pages    []Page
}

func (t SectionTask) Execute() error {
	html := bytes.NewBufferString("")
	tmpl := template.Must(template.ParseFiles(t.Template))

	tmpl.Execute(html, SectionData{
		Section: t.Section,
		Pages:   t.Pages,
	})

	html2 := applyTemplate(t.LayoutTemplate, PageData{
		Title:    t.Title,
		HTML:     template.HTML(html.String()),
		Section:  t.Section,
		Sections: t.Sections,
		Metadata: t.Metadata,
	})

	savePage(html2, t.OutputFile)
	return nil
}

func (t SectionTask) Name() string {
	return t.OutputFile
}
