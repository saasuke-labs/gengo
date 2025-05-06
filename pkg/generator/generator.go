package generator

import (
	"bytes"
	_ "embed"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type PageData struct {
	Title    string
	Tags     []string
	HTML     template.HTML
	Metadata map[string]string
}

type PageTask struct {
	InputFile      string
	OutputPath     string
	Url            string
	Template       string
	LayoutTemplate string
	Metadata       map[string]string
	Tags           []string
}

func (t PageTask) Generate() template.HTML {

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
	})

	return html
}

type SectionTask struct {
	OutputPath     string
	Template       string
	LayoutTemplate string
	Pages          []Page
}

type SectionData struct {
	Pages []Page
}

func (t SectionTask) Generate() template.HTML {
	html := bytes.NewBufferString("")
	tmpl := template.Must(template.ParseFiles(t.Template))

	tmpl.Execute(html, SectionData{
		Pages: t.Pages,
	})

	html2 := applyTemplate(t.LayoutTemplate, PageData{
		// See how to get the title from the HTML
		Title: "",
		HTML:  template.HTML(html.String()),
	})

	return html2
}

func (t PageTask) GetOutputPath() string {
	return t.OutputPath
}

func (t SectionTask) GetOutputPath() string {
	return t.OutputPath
}

type FileStatus string

type Page struct {
	Title        string            `yaml:"title"`
	Description  string            `yaml:"description"`
	MarkdownPath string            `yaml:"markdown-path"`
	PublishedAt  string            `yaml:"published-at"`
	LastEditedAt string            `yaml:"last-edited-at"`
	Tags         []string          `yaml:"tags"`
	Metadata     map[string]string `yaml:"metadata"`
}

type Section struct {
	Template     string `yaml:"template"`
	PageTemplate string `yaml:"page-template"`
	Pages        []Page `yaml:"pages"`
}

type ManifestFile struct {
	Title                  string             `yaml:"title"`
	DefaultLayoutTemplate  string             `yaml:"default-layout-template"`
	DefaultPageTemplate    string             `yaml:"default-page-template"`
	DefaultSectionTemplate string             `yaml:"default-section-template"`
	Setions                map[string]Section `yaml:"sections"`
}

type Task interface {
	Generate() template.HTML
	GetOutputPath() string
}

const (
	Pending   FileStatus = "pending"
	Started   FileStatus = "started"
	Completed FileStatus = "completed"
	Failed    FileStatus = "failed"
)

// FileProgress represents a progress update for a file
type FileProgress struct {
	Filename string
	Status   FileStatus
}

func savePage(content template.HTML, outputPath string) {
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

func getManifest(manifestPath string) ManifestFile {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		log.Fatalf("failed to read manifest file: %v", err)
	}
	var manifest ManifestFile
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		log.Fatalf("failed to parse YAML: %v", err)
	}

	return manifest
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

func calculateFilesToGenerate(manifest ManifestFile, outDir string) []Task {
	tasks := make([]Task, 0)

	for sectionName, section := range manifest.Setions {
		tags := make(map[string][]Page)

		sectionBasePath := filepath.Join(outDir, sectionName)
		os.MkdirAll(sectionBasePath, 0755)

		outPath := filepath.Join(sectionBasePath, "index.html")

		tasks = append(tasks, &SectionTask{
			OutputPath:     outPath,
			Template:       manifest.DefaultSectionTemplate,
			LayoutTemplate: manifest.DefaultLayoutTemplate,
			Pages:          section.Pages,
		})

		for _, page := range section.Pages {
			outputFilename := convertExtension(page.MarkdownPath, ".html")
			outPath := filepath.Join(sectionBasePath, outputFilename)
			tasks = append(tasks, PageTask{
				InputFile:      page.MarkdownPath,
				OutputPath:     outPath,
				Url:            filepath.Join("/", sectionName, outputFilename),
				Template:       manifest.DefaultPageTemplate,
				LayoutTemplate: manifest.DefaultLayoutTemplate,
				Metadata:       page.Metadata,
				Tags:           page.Tags,
			})

			for _, tag := range page.Tags {
				if _, ok := tags[tag]; !ok {
					tags[tag] = make([]Page, 0)
				}
				tags[tag] = append(tags[tag], page)
			}
		}

		tagsBasePath := filepath.Join(sectionBasePath, "tags")

		for tag, pages := range tags {
			tagPath := filepath.Join(tagsBasePath, slugify(tag)+".html")
			os.MkdirAll(tagsBasePath, 0755)

			// TODO - Create specific task for tags
			tasks = append(tasks, &SectionTask{
				OutputPath:     tagPath,
				Template:       manifest.DefaultSectionTemplate,
				LayoutTemplate: manifest.DefaultLayoutTemplate,
				Pages:          pages,
			})

		}
	}

	return tasks
}

func GenerateSiteAsync(manifestPath, outputDir string) ([]FileProgress, <-chan FileProgress) {

	manifest := getManifest(manifestPath)
	progressCh := make(chan FileProgress)

	filesToGenerate := calculateFilesToGenerate(manifest, outputDir)

	files := make([]FileProgress, len(filesToGenerate))
	for idx, fileToGenerate := range filesToGenerate {

		files[idx] = FileProgress{
			Filename: fileToGenerate.GetOutputPath(),
			Status:   Pending,
		}
	}

	go func() {
		var wg sync.WaitGroup

		for _, task := range filesToGenerate {
			wg.Add(1)
			go func(task Task) {
				defer wg.Done()

				progressCh <- FileProgress{Filename: task.GetOutputPath(), Status: Started}

				html := task.Generate()

				savePage(html, task.GetOutputPath())

				progressCh <- FileProgress{Filename: task.GetOutputPath(), Status: Completed}
			}(task)
		}

		wg.Wait()
		close(progressCh)
	}()

	return files, progressCh
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "/", "")
	return s
}
