package generator

import (
	"bytes"
	"html/template"
)

type HomeData struct {
}

type HomeTask struct {
	Title          string
	Sections       []string
	OutputFile     string
	Template       string
	LayoutTemplate string
	Metadata       map[string]string
}

func (t HomeTask) Execute() error {
	// TODO - See error from save page or generating the html
	html := bytes.NewBufferString("")
	tmpl := template.Must(template.ParseFiles(t.Template))

	tmpl.Execute(html, HomeData{})

	html2 := applyTemplate(t.LayoutTemplate, PageData{
		Title:    t.Title,
		HTML:     template.HTML(html.String()),
		Sections: t.Sections,
		Section:  "",
		Metadata: t.Metadata,
	})

	savePage(html2, t.OutputFile)

	return nil

}

func (t HomeTask) Name() string {
	return t.OutputFile
}
