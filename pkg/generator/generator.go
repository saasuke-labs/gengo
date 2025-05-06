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

// type HeroImage struct {
// 	Url         string `yaml:"url"`
// 	Attribution struct {
// 		Author string `yaml:"author"`
// 		Url    string `yaml:"url"`
// 	} `yaml:"attribution"`
// }

// type PostMeta struct {
// 	Slug        string    `yaml:"slug"`
// 	Title       string    `yaml:"title"`
// 	Description string    `yaml:"description"`
// 	Date        string    `yaml:"date"`
// 	Tags        []string  `yaml:"tags"`
// 	Markdown    string    `yaml:"markdown"`
// 	HeroImage   HeroImage `yaml:"hero-image"`
// }

type PageData struct {
	Title    string
	Tags     []string
	HTML     template.HTML
	Metadata map[string]string
}

// type PostListData struct {
// 	Posts []PostMeta
// }

type PageTask struct {
	InputFile      string
	OutputPath     string
	Template       string
	LayoutTemplate string
	Metadata       map[string]string
	Tags           []string
}

func (t PageTask) Generate() template.HTML {

	html := generateMarkdownPage(t.InputFile)

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

func generatePage(title string, content template.HTML, layoutTemplate *template.Template, outputPath string) {
	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
		return
	}
	defer f.Close()

	layoutTemplate.Execute(f, PageData{
		Title: title,
		HTML:  content,
	})
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

// func getPosts(postsPath string) []PostMeta {
// 	data, err := os.ReadFile(postsPath)
// 	if err != nil {
// 		log.Fatalf("failed to read posts file: %v", err)
// 	}
// 	var posts []PostMeta
// 	if err := yaml.Unmarshal(data, &posts); err != nil {
// 		log.Fatalf("failed to parse YAML: %v", err)
// 	}

// 	for i, post := range posts {
// 		if post.Slug == "" {
// 			posts[i].Slug = slugify(post.Title)
// 		}
// 	}

// 	return posts
// }

// func GenerateSite(manifestPath, outputPath string) []FileProgress {
// 	// posts := getPosts(postsPath)

// 	// tags := make(map[string][]PostMeta)
// 	// for _, post := range posts {
// 	// 	for _, tag := range post.Tags {
// 	// 		tags[tag] = append(tags[tag], post)
// 	// 	}
// 	// }
// 	// fmt.Println("TAGS:", tags)

// 	// //fmt.Println("POSTS:", posts)
// 	filesProgress := []FileProgress{}

// 	// os.MkdirAll(outputPath, 0755)

// 	// for _, post := range posts {
// 	// 	generatePostPage(post, postTemplate, outputPath)

// 	// 	filesProgress = append(filesProgress, FileProgress{
// 	// 		Filename: post.Markdown,
// 	// 		Status:   Completed,
// 	// 	})

// 	// }

// 	// // Generate post list
// 	// indexPath := filepath.Join(outputPath, "index.html")
// 	// generatePostList(posts, postListTemplate, indexPath)

// 	// filesProgress = append(filesProgress, FileProgress{
// 	// 	Filename: indexPath,
// 	// 	Status:   Completed,
// 	// })

// 	return filesProgress
// }

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
				Template:       manifest.DefaultPageTemplate,
				LayoutTemplate: manifest.DefaultLayoutTemplate,
				Metadata:       page.Metadata,
				Tags:           page.Tags,
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

	//fmt.Println("Sections:", manifest.Setions)
	// go func() {
	// 	var wg sync.WaitGroup

	// 	for sectionName, section := range manifest.Setions {
	// 		for _, page := range section.Pages {
	// 			go func(sectionName string, page Page) {
	// 				wg.Add(1)
	// 				defer wg.Done()

	// 				sectionBasePath := filepath.Join(outputDir, sectionName)
	// 				os.MkdirAll(sectionBasePath, 0755)

	// 				progressCh <- FileProgress{Filename: page.MarkdownPath, Status: Started, Section: sectionName}

	// 				outPath, html := generateMarkdownPage(MarkdownPage{
	// 					Title:        page.Title,
	// 					Description:  page.Description,
	// 					MarkdownPath: page.MarkdownPath,
	// 					PublishedAt:  page.PublishedAt,
	// 					Tags:         page.Tags,
	// 					Metadata:     page.Metadata,
	// 				}, sectionBasePath)

	// 				html = applyTemplate(manifest.DefaultPageTemplate, PageData{
	// 					Title:    page.Title,
	// 					Tags:     page.Tags,
	// 					Metadata: page.Metadata,
	// 					HTML:     html,
	// 				})

	// 				html = applyTemplate(manifest.DefaultLayoutTemplate, PageData{
	// 					Title:    page.Title,
	// 					Tags:     page.Tags,
	// 					Metadata: page.Metadata,
	// 					HTML:     html,
	// 				})

	// 				savePage(html, outPath)

	// 				progressCh <- FileProgress{Filename: page.MarkdownPath, Status: Completed, Section: sectionName}
	// 			}(sectionName, page)
	// 		}
	// 	}
	// 	time.Sleep(100 * time.Millisecond)
	// 	wg.Wait()
	// 	close(progressCh)
	// }()

	// tagsPath := filepath.Join(outputPath, "tags")

	// posts := getPosts(postsPath)
	//indexPath := filepath.Join(outputPath, "index.html")

	// tags := make(map[string][]PostMeta)
	// for _, post := range posts {
	// 	for _, tag := range post.Tags {
	// 		tags[tag] = append(tags[tag], post)
	// 	}
	// }

	//fmt.Println("POSTS:", posts)
	//initialProgress := make([]FileProgress, len(posts)+1+len(tags))
	// initialProgress := make([]FileProgress, 1)

	// initialProgress[0] = FileProgress{
	// 	Filename: indexPath,
	// 	Status:   Pending,
	// }

	// for i, post := range posts {
	// 	initialProgress[i+1] = FileProgress{
	// 		Filename: post.Markdown,
	// 		Status:   Pending,
	// 	}
	// }

	// i := 0
	// for tag, posts := range tags {
	// 	tagPath := filepath.Join(tagsPath, tag+".html")
	// 	initialProgress[i+1+len(posts)] = FileProgress{
	// 		Filename: tagPath,
	// 		Status:   Pending,
	// 	}
	// 	i++
	// }

	// os.MkdirAll(outputPath, 0755)
	// os.MkdirAll(outputPath, 0755)
	// os.MkdirAll(tagsPath, 0755)

	// go func() {
	// 	var wg sync.WaitGroup

	// 	for _, post := range posts {
	// 		go func(p PostMeta) {
	// 			wg.Add(1)
	// 			defer wg.Done()

	// 			progressCh <- FileProgress{Filename: p.Markdown, Status: Started}

	// 			generatePostPage(p, postTemplate, outputPath)

	// 			progressCh <- FileProgress{Filename: p.Markdown, Status: Completed}
	// 		}(post)

	// 	}
	// 	progressCh <- FileProgress{Filename: indexPath, Status: Started}

	// 	generatePostList(posts, postListTemplate, indexPath)

	// 	progressCh <- FileProgress{Filename: indexPath, Status: Completed}

	// 	for tag, posts := range tags {
	// 		go func(tag string, posts []PostMeta) {
	// 			wg.Add(1)
	// 			defer wg.Done()

	// 			tagPath := filepath.Join(tagsPath, tag+".html")
	// 			progressCh <- FileProgress{Filename: tagPath, Status: Started}

	// 			generatePostList(posts, postListTemplate, tagPath)

	// 			progressCh <- FileProgress{Filename: tagPath, Status: Completed}
	// 		}(tag, posts)

	// 	}

	// 	wg.Wait()
	// 	close(progressCh)
	// }()

	// files := make([]FileProgress, 0)
	// for sectionName, section := range manifest.Setions {
	// 	files = append(files, FileProgress{
	// 		Filename: path.Join(sectionName, "index.html"),
	// 		Status:   Pending,
	// 		Section:  sectionName,
	// 	})
	// 	for _, page := range section.Pages {
	// 		files = append(files, FileProgress{
	// 			Filename: page.MarkdownPath,
	// 			Status:   Pending,
	// 			Section:  sectionName,
	// 		})
	// 	}
	// }

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
