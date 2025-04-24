package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	_ "github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"
)

var (
	watchMode     bool
	sseClients    = make(map[chan string]bool)
	sseRegister   = make(chan chan string)
	sseUnregister = make(chan chan string)
)

type PostMeta struct {
	Slug     string   `yaml:"slug"`
	Title    string   `yaml:"title"`
	Date     string   `yaml:"date"`
	Tags     []string `yaml:"tags"`
	Markdown string   `yaml:"markdown"`
}

type SiteData struct {
	Title string
	HTML  template.HTML
	Watch bool
}

//go:embed templates/layout.html
var layoutHTML string

func main() {
	var postsPath string
	var outputPath string
	var rootCmd = &cobra.Command{
		Use:   "mdsite",
		Short: "Static site generator from Markdown",
		Run: func(cmd *cobra.Command, args []string) {
			generateSite(postsPath, outputPath)
			if watchMode {
				go startSSEServer()
				go serveOutput(outputPath)
				watchAndRebuild(postsPath, outputPath)
				select {}
			}
		},
	}

	rootCmd.Flags().StringVar(&postsPath, "posts", "posts.yaml", "Path to posts.yaml")
	rootCmd.Flags().StringVar(&outputPath, "output", "output", "Output directory")
	rootCmd.Flags().BoolVar(&watchMode, "watch", false, "Enable watch mode with hot reload")
	rootCmd.Execute()
}

func generateSite(postsPath, outputPath string) {
	data, err := os.ReadFile(postsPath)
	if err != nil {
		log.Fatalf("failed to read posts file: %v", err)
	}
	var posts []PostMeta
	if err := yaml.Unmarshal(data, &posts); err != nil {
		log.Fatalf("failed to parse YAML: %v", err)
	}

	fmt.Println("POSTS:", posts)

	md := goldmark.New(
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
	tmpl := template.Must(template.New("layout").Parse(layoutHTML))

	os.MkdirAll(outputPath, 0755)
	for _, post := range posts {
		content, err := os.ReadFile(post.Markdown)
		if err != nil {
			log.Printf("failed to read %s: %v", post.Markdown, err)
			continue
		}
		var buf bytes.Buffer
		context := parser.NewContext()
		doc := md.Parser().Parse(text.NewReader(content), parser.WithContext(context))
		printAST(doc, content)

		title := post.Title
		if title == "" {
			if h1 := findFirstH1(doc, content); h1 != "" {
				title = h1
			}
		}
		if post.Slug == "" {
			post.Slug = slugify(title)
		}

		md.Renderer().Render(&buf, content, doc)

		outFile := filepath.Join(outputPath, post.Slug+".html")
		f, err := os.Create(outFile)
		if err != nil {
			log.Printf("failed to create output file: %v", err)
			continue
		}

		fmt.Println("Generated html:", buf.String())
		tmpl.Execute(f, SiteData{
			Title: title,
			HTML:  template.HTML(buf.String()),
			Watch: watchMode,
		})
		f.Close()
	}
}

func printAST(node ast.Node, source []byte) {
	for n := node.FirstChild(); n != nil; n = n.NextSibling() {
		switch n := n.(type) {
		case *ast.Heading:
			fmt.Printf("\nHeading: %d %s\n", n.Level, n.Text(source))
		case *ast.FencedCodeBlock:
			fmt.Printf("\nCode Block: %s \n%s\n\n", (*ast.FencedCodeBlock)(n).Language(source), n.Lines().Value(source))
		case *ast.Paragraph:
			fmt.Printf("\nParagraph: %s \n", (*ast.Paragraph)(n).Lines().Value(source))
		case *ast.Text:
			fmt.Printf("\nText: %s \n", (*ast.Text)(n).Value(source))
		case *ast.HTMLBlock:
			fmt.Printf("\nHTML Block: %s\n", n.Text(source))
		default:
			fmt.Printf("\nNode: %T\n", n)
		}
		printAST(n, source)
	}
}

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
	s = strings.ReplaceAll(s, "/", "")
	return s
}

func watchAndRebuild(postsPath, outputPath string) {
	w, _ := fsnotify.NewWatcher()
	defer w.Close()
	//contentDir := filepath.Dir(postsPath)
	w.Add(postsPath)
	w.Add("templates")
	filepath.WalkDir("content", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			w.Add(path)
		}
		return nil
	})
	for {
		select {
		case <-w.Events:
			generateSite(postsPath, outputPath)
			broadcastReload()
		case err := <-w.Errors:
			log.Println("watch error:", err)
		}
	}
}

func serveOutput(outputPath string) {
	http.Handle("/", http.FileServer(http.Dir(outputPath)))
	http.HandleFunc("/reload", handleSSE)
	log.Println("Serving on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	client := make(chan string)
	sseRegister <- client
	defer func() { sseUnregister <- client }()
	for {
		select {
		case <-client:
			fmt.Fprintf(w, "data: reload\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func startSSEServer() {
	for {
		select {
		case c := <-sseRegister:
			sseClients[c] = true
		case c := <-sseUnregister:
			delete(sseClients, c)
			close(c)
		}
	}
}

func broadcastReload() {
	for c := range sseClients {
		c <- "reload"
	}
}
