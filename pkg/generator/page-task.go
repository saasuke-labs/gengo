package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	ttemplate "text/template"
	"time"
)

type DataTask struct {
	Key     string
	UrlTmpl string
	Headers map[string]string
}

type PageData struct {
	Title    string
	Tags     []string
	HTML     template.HTML
	Metadata map[string]string
	Section  string
	Sections []string
	Data     map[string]interface{}
}

type PageTask struct {
	Title          string
	InputFile      string
	OutputFile     string
	Url            string
	Template       string
	LayoutTemplate string
	Metadata       map[string]string
	Tags           []string
	Section        string
	Sections       []string
	DataTasks      []DataTask
}

// renderURLTemplate renders a source URL template using the page's metadata context.
func renderURLTemplate(urlTmpl string, metadata map[string]string) (string, error) {
	funcMap := ttemplate.FuncMap{
		"now": time.Now,
	}
	tmpl, err := ttemplate.New("url").Funcs(funcMap).Parse(urlTmpl)
	if err != nil {
		return "", fmt.Errorf("invalid URL template %q: %w", urlTmpl, err)
	}
	ctx := struct{ Metadata map[string]string }{Metadata: metadata}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("failed to render URL template %q: %w", urlTmpl, err)
	}
	return buf.String(), nil
}

// fetchSource fetches a single source and returns parsed data based on Content-Type.
func fetchSource(task DataTask, metadata map[string]string) (interface{}, error) {
	url, err := renderURLTemplate(task.UrlTmpl, metadata)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %w", url, err)
	}

	for name, value := range task.Headers {
		req.Header.Set(name, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response from %s: %w", url, err)
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var parsed interface{}
		if err := json.Unmarshal(body, &parsed); err != nil {
			fmt.Printf("Warning: failed to parse JSON from %s: %v; returning raw string\n", url, err)
			return string(body), nil
		}
		return parsed, nil
	}
	if strings.Contains(contentType, "text/html") {
		return template.HTML(body), nil
	}
	return string(body), nil
}

// fetchAllData fetches all DataTasks in parallel and returns a map of key → data.
func fetchAllData(tasks []DataTask, metadata map[string]string) map[string]interface{} {
	result := make(map[string]interface{}, len(tasks))
	if len(tasks) == 0 {
		return result
	}

	type entry struct {
		key  string
		data interface{}
	}
	ch := make(chan entry, len(tasks))
	var wg sync.WaitGroup

	for _, task := range tasks {
		wg.Add(1)
		go func(t DataTask) {
			defer wg.Done()
			data, err := fetchSource(t, metadata)
			if err != nil {
				fmt.Printf("Warning: failed to fetch source '%s': %v\n", t.Key, err)
				ch <- entry{t.Key, nil}
			} else {
				ch <- entry{t.Key, data}
			}
		}(task)
	}

	wg.Wait()
	close(ch)
	for e := range ch {
		result[e.key] = e.data
	}
	return result
}

// renderMarkdownTemplate pre-processes a markdown file as a Go text template using
// the provided data and metadata. Returns the raw markdown bytes.
func renderMarkdownTemplate(filePath string, data map[string]interface{}, metadata map[string]string) ([]byte, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	funcMap := ttemplate.FuncMap{
		"now": time.Now,
	}
	tmpl, err := ttemplate.New(filepath.Base(filePath)).Funcs(funcMap).Parse(string(raw))
	if err != nil {
		return raw, nil // not a template, return as-is
	}

	ctx := struct {
		Data     map[string]interface{}
		Metadata map[string]string
	}{
		Data:     data,
		Metadata: metadata,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		fmt.Printf("Warning: failed to render markdown template %s: %v; using raw content\n", filePath, err)
		return raw, nil
	}
	return buf.Bytes(), nil
}

func getHtmlFromFile(filePath string) template.HTML {
	return getHtmlFromFileWithData(filePath, nil, nil)
}

func getHtmlFromFileWithData(filePath string, data map[string]interface{}, metadata map[string]string) template.HTML {
	if filePath == "" {
		return template.HTML("")
	}

	extension := filepath.Ext(filePath)

	if extension == ".html" {
		content, err := os.ReadFile(filePath)
		if err != nil {
			panic(fmt.Sprintf("Error reading file %s: %v", filePath, err))
		}
		return template.HTML(content)
	}

	if extension == ".md" {
		if len(data) > 0 {
			processed, err := renderMarkdownTemplate(filePath, data, metadata)
			if err != nil {
				fmt.Printf("Warning: %v\n", err)
				return generateMarkdownPage(filePath)
			}
			return generateMarkdownPageFromBytes(processed)
		}
		return generateMarkdownPage(filePath)
	}

	panic(fmt.Sprintf("Unsupported file type %s for file %s", extension, filePath))
}

func (t PageTask) Execute() error {
	data := fetchAllData(t.DataTasks, t.Metadata)

	html := getHtmlFromFileWithData(t.InputFile, data, t.Metadata)

	if t.Template != "" {
		html = applyTemplate(t.Template, PageData{
			Title:    "",
			Tags:     t.Tags,
			Metadata: t.Metadata,
			HTML:     html,
			Data:     data,
		})
	}
	html = applyTemplate(t.LayoutTemplate, PageData{
		Title:    t.Title,
		Tags:     t.Tags,
		Metadata: t.Metadata,
		HTML:     html,
		Section:  t.Section,
		Sections: t.Sections,
		Data:     data,
	})

	savePage(html, t.OutputFile)
	return nil
}

func (t PageTask) Name() string {
	return t.OutputFile
}
