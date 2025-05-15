package generator

import (
	"html/template"
)

type PageData struct {
	Title    string
	Tags     []string
	HTML     template.HTML
	Metadata map[string]string
	Section  string
	Sections []string
}

type PageTask struct {
	InputFile      string
	OutputFile     string
	Url            string
	Template       string
	LayoutTemplate string
	Metadata       map[string]string
	Tags           []string
	Section        string
	Sections       []string
}

func (t PageTask) Execute() error {

	html := generateMarkdownPage(t.InputFile)

	html = applyTemplate(t.Template, PageData{
		// See how to get the title from the HTML
		Title:    "",
		Tags:     t.Tags,
		Metadata: t.Metadata,
		HTML:     html,
	})
	html = applyTemplate(t.LayoutTemplate, PageData{
		// See how to get the title from the HTML
		Title:    "",
		Tags:     t.Tags,
		Metadata: t.Metadata,
		HTML:     html,
		Section:  t.Section,
		Sections: t.Sections,
	})

	savePage(html, t.OutputFile)
	return nil
}

func (t PageTask) Name() string {
	return t.OutputFile
}
