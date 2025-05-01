package generator

import (
	"blog-down/pkg/parser"
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type HeroImage struct {
	Url         string `yaml:"url"`
	Attribution struct {
		Author string `yaml:"author"`
		Url    string `yaml:"url"`
	} `yaml:"attribution"`
}

type PostMeta struct {
	Slug        string    `yaml:"slug"`
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Date        string    `yaml:"date"`
	Tags        []string  `yaml:"tags"`
	Markdown    string    `yaml:"markdown"`
	HeroImage   HeroImage `yaml:"hero-image"`
}

type PageData struct {
	Title string
	HTML  template.HTML
}

type PostPageData struct {
	Title     string
	HeroImage HeroImage
	Tags      []string
	Article   template.HTML
}

type PostListData struct {
	Posts []PostMeta
}

type FileStatus string

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

//go:embed templates/post.html
var postTemplateHTML string

//go:embed templates/postList.html
var postListTemplateHTML string

//go:embed templates/layout.html
var layoutTemplateHTML string

// This should not be Global. Depends on the command to be executed.
var postTemplate *template.Template
var layoutTemplate *template.Template
var postListTemplate *template.Template

func init() {
	postTemplate = template.Must(template.New("post").Parse(postTemplateHTML))
	postListTemplate = template.Must(template.New("postList").Parse(postListTemplateHTML))
	layoutTemplate = template.Must(template.New("layout").Parse(layoutTemplateHTML))
}

func generatePage(title string, content template.HTML, tmpl *template.Template, outputPath string) {
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

func generatePostList(posts []PostMeta, tmpl *template.Template, outputPath string) {
	html := bytes.NewBufferString("")

	tmpl.Execute(html, PostListData{
		Posts: posts,
	})

	generatePage("Posts", template.HTML(html.String()), tmpl, outputPath)
}

func generatePostPage(post PostMeta, tmpl *template.Template, outputPath string) {

	htmlPage := parser.MarkdownToHtml(post.Markdown)

	title := post.Title
	if title == "" {
		title = htmlPage.Title
	}

	if post.Slug == "" {
		post.Slug = slugify(title)
	}

	outFile := filepath.Join(outputPath, post.Slug+".html")

	html := bytes.NewBufferString("")
	tmpl.Execute(html, PostPageData{
		Title:     title,
		HeroImage: post.HeroImage,
		Tags:      post.Tags,
		Article:   htmlPage.HTML,
	})

	generatePage(title, template.HTML(html.String()), tmpl, outFile)
}

func getPosts(postsPath string) []PostMeta {
	data, err := os.ReadFile(postsPath)
	if err != nil {
		log.Fatalf("failed to read posts file: %v", err)
	}
	var posts []PostMeta
	if err := yaml.Unmarshal(data, &posts); err != nil {
		log.Fatalf("failed to parse YAML: %v", err)
	}

	for i, post := range posts {
		if post.Slug == "" {
			posts[i].Slug = slugify(post.Title)
		}
	}

	return posts
}

func GenerateSite(postsPath, outputPath string) []FileProgress {
	posts := getPosts(postsPath)

	tags := make(map[string][]PostMeta)
	for _, post := range posts {
		for _, tag := range post.Tags {
			tags[tag] = append(tags[tag], post)
		}
	}
	fmt.Println("TAGS:", tags)

	//fmt.Println("POSTS:", posts)
	filesProgress := []FileProgress{}

	os.MkdirAll(outputPath, 0755)

	for _, post := range posts {
		generatePostPage(post, postTemplate, outputPath)

		filesProgress = append(filesProgress, FileProgress{
			Filename: post.Markdown,
			Status:   Completed,
		})

	}

	// Generate post list
	indexPath := filepath.Join(outputPath, "index.html")
	generatePostList(posts, postListTemplate, indexPath)

	filesProgress = append(filesProgress, FileProgress{
		Filename: indexPath,
		Status:   Completed,
	})

	return filesProgress
}

func GenerateSiteAsync(postsPath, outputPath string) (<-chan FileProgress, []FileProgress) {

	progressCh := make(chan FileProgress)

	tagsPath := filepath.Join(outputPath, "tags")

	posts := getPosts(postsPath)
	indexPath := filepath.Join(outputPath, "index.html")

	tags := make(map[string][]PostMeta)
	for _, post := range posts {
		for _, tag := range post.Tags {
			tags[tag] = append(tags[tag], post)
		}
	}
	fmt.Println("TAGS:", tags)

	//fmt.Println("POSTS:", posts)
	initialProgress := make([]FileProgress, len(posts)+1+len(tags))

	initialProgress[0] = FileProgress{
		Filename: indexPath,
		Status:   Pending,
	}

	for i, post := range posts {
		initialProgress[i+1] = FileProgress{
			Filename: post.Markdown,
			Status:   Pending,
		}
	}

	i := 0
	for tag, posts := range tags {
		tagPath := filepath.Join(tagsPath, tag+".html")
		initialProgress[i+1+len(posts)] = FileProgress{
			Filename: tagPath,
			Status:   Pending,
		}
		i++
	}

	os.MkdirAll(outputPath, 0755)
	os.MkdirAll(outputPath, 0755)
	os.MkdirAll(tagsPath, 0755)

	go func() {
		var wg sync.WaitGroup

		for _, post := range posts {
			go func(p PostMeta) {
				wg.Add(1)
				defer wg.Done()

				progressCh <- FileProgress{Filename: p.Markdown, Status: Started}

				generatePostPage(p, postTemplate, outputPath)

				progressCh <- FileProgress{Filename: p.Markdown, Status: Completed}
			}(post)

		}
		progressCh <- FileProgress{Filename: indexPath, Status: Started}

		generatePostList(posts, postListTemplate, indexPath)

		progressCh <- FileProgress{Filename: indexPath, Status: Completed}

		for tag, posts := range tags {
			go func(tag string, posts []PostMeta) {
				wg.Add(1)
				defer wg.Done()

				tagPath := filepath.Join(tagsPath, tag+".html")
				progressCh <- FileProgress{Filename: tagPath, Status: Started}

				generatePostList(posts, postListTemplate, tagPath)

				progressCh <- FileProgress{Filename: tagPath, Status: Completed}
			}(tag, posts)

		}

		wg.Wait()
		close(progressCh)
	}()

	fmt.Println("returning progress channel")
	return progressCh, initialProgress
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "/", "")
	return s
}
