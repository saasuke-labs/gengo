package generator

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A page that declares `data:` in its front matter gets the referenced source
// fetched (URL rendered with the page metadata, headers sent) and exposed as
// .Data to the markdown pre-render pass and to the page and layout templates.
func TestSources_EndToEnd(t *testing.T) {
	var gotPath, gotHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotHeader = r.Header.Get("X-Site")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"comments":[{"author":"Ada","text":"First!"},{"author":"Linus","text":"Nice post"}]}`))
	}))
	defer server.Close()

	baseDir := t.TempDir()
	outDir := t.TempDir()

	write := func(name, content string) {
		path := filepath.Join(baseDir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	}

	write("layout.html", `<html><body>{{ .HTML }}</body></html>`)
	write("page.html", `<article>{{ .HTML }}</article><ul>{{ range .Data.comments.comments }}<li>{{ .author }}: {{ .text }}</li>{{ end }}</ul>`)
	write("blog/hello.md", "---\ntitle: Hello\nmetadata:\n  page-id: hello-world\ndata:\n  comments:\n    source: kotomi\n---\n# Hello\n\nComments: {{ len .Data.comments.comments }}\n")

	manifest := ManifestFile{
		Title:                 "Site",
		DefaultLayoutTemplate: "layout.html",
		DefaultPageTemplate:   "page.html",
		Sources: map[string]Source{
			"kotomi": {
				Url:     server.URL + `/pages/{{ index .Metadata "page-id" }}/comments`,
				Headers: map[string]string{"X-Site": "my-site"},
			},
		},
	}

	pages, err := pagesFromGlobs(baseDir, FileGlobs{"blog/*.md"})
	require.NoError(t, err)
	require.Len(t, pages, 1)
	manifest.Sections = map[string]Section{"blog": {Pages: pages}}

	for _, task := range scheduleTasks(manifest, baseDir, outDir) {
		require.NoError(t, task.Execute())
	}

	assert.Equal(t, "/pages/hello-world/comments", gotPath)
	assert.Equal(t, "my-site", gotHeader)

	out, err := os.ReadFile(filepath.Join(outDir, "blog", "hello.html"))
	require.NoError(t, err)
	html := string(out)
	assert.Contains(t, html, "Comments: 2")
	assert.Contains(t, html, "<li>Ada: First!</li>")
	assert.Contains(t, html, "<li>Linus: Nice post</li>")
	assert.NotContains(t, html, "page-id", "front matter must not leak into the page")
}
