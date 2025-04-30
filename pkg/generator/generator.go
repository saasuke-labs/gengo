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
	Watch bool
}

type PostPageData struct {
	Title     string
	HeroImage HeroImage
	Tags      []string
	Article   template.HTML
}

type PostListData struct {
	Posts []PostMeta
	Watch bool
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

func generatePage(title string, content template.HTML, tmpl *template.Template, outputPath string, watchMode bool) {
	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
		return
	}
	defer f.Close()

	layoutTemplate.Execute(f, PageData{
		Title: title,
		HTML:  content,
		Watch: watchMode,
	})
}

func generatePostList(posts []PostMeta, tmpl *template.Template, outputPath string, watchMode bool) {
	html := bytes.NewBufferString("")

	tmpl.Execute(html, PostListData{
		Posts: posts,
		Watch: watchMode,
	})

	generatePage("Posts", template.HTML(html.String()), tmpl, outputPath, watchMode)
}

func generatePostPage(post PostMeta, tmpl *template.Template, outputPath string, watchMode bool) {

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

	generatePage(title, template.HTML(html.String()), tmpl, outFile, watchMode)
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

func GenerateSite(postsPath, outputPath string, watchMode bool) []FileProgress {
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
		generatePostPage(post, postTemplate, outputPath, watchMode)

		filesProgress = append(filesProgress, FileProgress{
			Filename: post.Markdown,
			Status:   Completed,
		})

	}

	// Generate post list
	indexPath := filepath.Join(outputPath, "index.html")
	generatePostList(posts, postListTemplate, indexPath, watchMode)

	filesProgress = append(filesProgress, FileProgress{
		Filename: indexPath,
		Status:   Completed,
	})

	return filesProgress
}

func GenerateSiteAsync(postsPath, outputPath string, watchMode bool) (<-chan FileProgress, []FileProgress) {
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
		defer close(progressCh)

		for _, post := range posts {
			progressCh <- FileProgress{Filename: post.Markdown, Status: Started}

			generatePostPage(post, postTemplate, outputPath, watchMode)

			progressCh <- FileProgress{Filename: post.Markdown, Status: Completed}

		}
		progressCh <- FileProgress{Filename: indexPath, Status: Started}

		generatePostList(posts, postListTemplate, indexPath, watchMode)

		progressCh <- FileProgress{Filename: indexPath, Status: Completed}

		i := 0
		for tag, posts := range tags {
			tagPath := filepath.Join(tagsPath, tag+".html")
			progressCh <- FileProgress{Filename: tagPath, Status: Started}

			generatePostList(posts, postListTemplate, tagPath, watchMode)

			progressCh <- FileProgress{Filename: tagPath, Status: Completed}

			i++
		}

	}()

	fmt.Println("returning progress channel")
	return progressCh, initialProgress
}

// func printAST(node ast.Node, source []byte) {
// 	for n := node.FirstChild(); n != nil; n = n.NextSibling() {
// 		switch n := n.(type) {
// 		case *ast.Heading:
// 			fmt.Printf("\nHeading: %d %s\n", n.Level, n.Text(source))
// 		case *ast.FencedCodeBlock:
// 			fmt.Printf("\nCode Block: %s \n%s\n\n", (*ast.FencedCodeBlock)(n).Language(source), n.Lines().Value(source))
// 		case *ast.Paragraph:
// 			fmt.Printf("\nParagraph: %s \n", (*ast.Paragraph)(n).Lines().Value(source))
// 		case *ast.Text:
// 			fmt.Printf("\nText: %s \n", (*ast.Text)(n).Value(source))
// 		case *ast.HTMLBlock:
// 			fmt.Printf("\nHTML Block: %s\n", n.Text(source))
// 		default:
// 			fmt.Printf("\nNode: %T\n", n)
// 		}
// 		printAST(n, source)
// 	}
// }

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "/", "")
	return s
}
