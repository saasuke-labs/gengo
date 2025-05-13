package generator

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
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
		Section:  t.Section,
		Sections: t.Sections,
	})

	return html
}

type HomeData struct {
}

type HomeTask struct {
	Sections       []string
	OutputFile     string
	Template       string
	LayoutTemplate string
}

type SectionTask struct {
	Section        string
	Sections       []string
	OutputFile     string
	Template       string
	LayoutTemplate string
	Pages          []Page
}

type SectionData struct {
	Section string

	Pages []Page
}

func (t SectionTask) Generate() template.HTML {
	html := bytes.NewBufferString("")
	tmpl := template.Must(template.ParseFiles(t.Template))

	tmpl.Execute(html, SectionData{
		Section: t.Section,
		Pages:   t.Pages,
	})

	html2 := applyTemplate(t.LayoutTemplate, PageData{
		// See how to get the title from the HTML
		Title:    "",
		HTML:     template.HTML(html.String()),
		Section:  t.Section,
		Sections: t.Sections,
	})

	return html2
}

func (t HomeTask) Generate() template.HTML {
	html := bytes.NewBufferString("")
	tmpl := template.Must(template.ParseFiles(t.Template))

	tmpl.Execute(html, HomeData{})

	html2 := applyTemplate(t.LayoutTemplate, PageData{
		// See how to get the title from the HTML
		Title:    "",
		HTML:     template.HTML(html.String()),
		Sections: t.Sections,
		Section:  "",
	})

	return html2
}

func (t PageTask) GetOutputPath() string {
	return t.OutputFile
}

func (t SectionTask) GetOutputPath() string {
	return t.OutputFile
}

func (t HomeTask) GetOutputPath() string {
	return t.OutputFile
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
	Section      string
}

func (p *Page) OutFileName() string {
	return convertExtension(p.MarkdownPath, ".html")
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
	HomeTemplate           string             `yaml:"home-template"`
	Sections               map[string]Section `yaml:"sections"`
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

func getManifest(manifestPath string) ManifestFile {
	log.Printf("Reading manifest file: %s", manifestPath)
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

func calculateFilesToGenerate(manifest ManifestFile, baseDir, outDir string) []Task {
	tasks := make([]Task, 0)

	sections := make([]string, 0)
	for section, _ := range manifest.Sections {
		sections = append(sections, section)
	}

	if manifest.HomeTemplate != "" {
		homePath := filepath.Join(outDir, "index.html")

		tasks = append(tasks, &HomeTask{
			Sections:       sections,
			OutputFile:     homePath,
			Template:       path.Join(baseDir, manifest.HomeTemplate),
			LayoutTemplate: path.Join(baseDir, manifest.DefaultLayoutTemplate),
		})
	}

	for sectionName, section := range manifest.Sections {
		tags := make(map[string][]Page)

		sectionBasePath := filepath.Join(outDir, sectionName)

		// Do not generate the section page if there is no template configured
		if manifest.DefaultSectionTemplate != "" {
			os.MkdirAll(sectionBasePath, 0755)

			outFile := filepath.Join(sectionBasePath, "index.html")

			tasks = append(tasks, &SectionTask{
				Section:        sectionName,
				Sections:       sections,
				OutputFile:     outFile,
				Template:       path.Join(baseDir, manifest.DefaultSectionTemplate),
				LayoutTemplate: path.Join(baseDir, manifest.DefaultLayoutTemplate),
				Pages:          section.Pages,
			})
		}

		for _, page := range section.Pages {
			outputFilename := convertExtension(page.MarkdownPath, ".html")
			outPath := filepath.Join(sectionBasePath, outputFilename)
			tasks = append(tasks, PageTask{
				InputFile:      path.Join(baseDir, page.MarkdownPath),
				OutputFile:     outPath,
				Url:            filepath.Join("/", sectionName, outputFilename),
				Template:       path.Join(baseDir, manifest.DefaultPageTemplate),
				LayoutTemplate: path.Join(baseDir, manifest.DefaultLayoutTemplate),
				Metadata:       page.Metadata,
				Tags:           page.Tags,
				Section:        sectionName,
				Sections:       sections,
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
			tagOutputFile := filepath.Join(tagsBasePath, slugify(tag)+".html")
			os.MkdirAll(tagsBasePath, 0755)

			// TODO - Create specific task for tags
			tasks = append(tasks, &SectionTask{
				OutputFile:     tagOutputFile,
				Template:       path.Join(baseDir, manifest.DefaultSectionTemplate),
				LayoutTemplate: path.Join(baseDir, manifest.DefaultLayoutTemplate),
				Pages:          pages,
				Section:        sectionName,
				Sections:       sections,
			})

		}
	}

	return tasks
}

func GenerateSiteAsync(manifestPath, outputDir string) ([]FileProgress, <-chan FileProgress) {

	manifest := getManifest(manifestPath)
	baseDir := filepath.Dir(manifestPath)

	fmt.Println("Generating site...", manifest)
	progressCh := make(chan FileProgress)

	filesToGenerate := calculateFilesToGenerate(manifest, baseDir, outputDir)

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
