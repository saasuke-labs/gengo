package generator

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
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
var md goldmark.Markdown
var postTemplate *template.Template
var layourTemplate *template.Template
var postListTemplate *template.Template

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(extension.GFM, highlighting.NewHighlighting(
			// highlighting.WithFormatOptions(
			// 	htmlchroma.WithLineNumbers(true),
			// ),
			highlighting.WithStyle("github"), // choose a theme
			highlighting.WithGuessLanguage(false),
		)),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithHardWraps(), html.WithXHTML()),
	)
	postTemplate = template.Must(template.New("post").Parse(postTemplateHTML))
	postListTemplate = template.Must(template.New("postList").Parse(postListTemplateHTML))
	layourTemplate = template.Must(template.New("layour").Parse(layoutTemplateHTML))
}

func generatePage(title string, content template.HTML, tmpl *template.Template, outputPath string, watchMode bool) {
	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
		return
	}
	defer f.Close()

	layourTemplate.Execute(f, PageData{
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

func generatePostPage(post PostMeta, md goldmark.Markdown, tmpl *template.Template, outputPath string, watchMode bool) {

	content, err := os.ReadFile(post.Markdown)
	if err != nil {
		log.Fatalf("failed to read %s: %v", post.Markdown, err)
		return
	}
	context := parser.NewContext()
	doc := md.Parser().Parse(text.NewReader(content), parser.WithContext(context))
	//printAST(doc, content)

	title := post.Title
	if title == "" {
		if h1 := findFirstH1(doc, content); h1 != "" {
			title = h1
		}
	}
	if post.Slug == "" {
		post.Slug = slugify(title)
	}
	var article bytes.Buffer

	md.Renderer().Render(&article, content, doc)

	outFile := filepath.Join(outputPath, post.Slug+".html")

	html := bytes.NewBufferString("")
	tmpl.Execute(html, PostPageData{
		Title:     title,
		HeroImage: post.HeroImage,
		Tags:      post.Tags,
		Article:   template.HTML(article.String()),
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
		generatePostPage(post, md, postTemplate, outputPath, watchMode)

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

			generatePostPage(post, md, postTemplate, outputPath, watchMode)

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

func findFirstH1(doc ast.Node, source []byte) string {
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		if h, ok := n.(*ast.Heading); ok && h.Level == 1 {
			return string(h.Text(source))
		}
	}
	return ""
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "/", "")
	return s
}
