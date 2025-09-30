package nagare

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yuin/goldmark"
)

func TestNagareExtension(t *testing.T) {
	// Create a mock HTTP server that returns some HTML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the request body (the nagare code)
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)

		// Return some mock HTML canvas
		w.Header().Set("Content-Type", "text/html")
		htmlResponse := `<canvas id="nagare-canvas" width="400" height="300">
			<p>Your browser doesn't support canvas. Code was: ` + string(body) + `</p>
		</canvas>`
		w.Write([]byte(htmlResponse))
	}))
	defer server.Close()

	// Create markdown with nagare extension
	md := goldmark.New(
		goldmark.WithExtensions(
			NewNagareExtension(server.URL),
		),
	)

	// Test markdown with nagare code block
	markdown := "# Test Page\n\nHere's a nagare block:\n\n```nagare\nsome nagare code\nline 2\n```\n\nAnd some regular text."

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	result := buf.String()

	// Check that the canvas HTML is present
	if !bytes.Contains([]byte(result), []byte("<canvas id=\"nagare-canvas\"")) {
		t.Errorf("Expected canvas HTML in result, got: %s", result)
	}

	// Check that the original nagare code is referenced in the canvas fallback
	if !bytes.Contains([]byte(result), []byte("some nagare code")) {
		t.Errorf("Expected nagare code in result, got: %s", result)
	}

	t.Logf("Generated HTML: %s", result)
}

func TestNagareExtensionMultipleBlocks(t *testing.T) {
	// Create a mock HTTP server that returns some HTML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the request body (the nagare code)
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)

		// Return some mock HTML canvas with the code content
		w.Header().Set("Content-Type", "text/html")
		htmlResponse := `<canvas id="nagare-canvas" data-code="` + string(body) + `">
			<p>Canvas for: ` + string(body) + `</p>
		</canvas>`
		w.Write([]byte(htmlResponse))
	}))
	defer server.Close()

	// Create markdown with nagare extension
	md := goldmark.New(
		goldmark.WithExtensions(
			NewNagareExtension(server.URL),
		),
	)

	// Test markdown with multiple nagare code blocks
	markdown := `# Test Page

First nagare block:

` + "```nagare" + `
circle(50, 50, 30)
` + "```" + `

Some text in between.

Second nagare block:

` + "```nagare" + `
rectangle(10, 10, 100, 50)
` + "```" + `

More text.

Third nagare block:

` + "```nagare" + `
line(0, 0, 200, 200)
` + "```" + `
`

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	result := buf.String()

	// Count how many canvas elements are in the result
	canvasCount := bytes.Count([]byte(result), []byte("<canvas id=\"nagare-canvas\""))
	if canvasCount != 3 {
		t.Errorf("Expected 3 canvas elements, got %d. Result: %s", canvasCount, result)
	}

	// Check that all three code snippets are present
	if !bytes.Contains([]byte(result), []byte("circle(50, 50, 30)")) {
		t.Errorf("Expected first nagare code in result, got: %s", result)
	}
	if !bytes.Contains([]byte(result), []byte("rectangle(10, 10, 100, 50)")) {
		t.Errorf("Expected second nagare code in result, got: %s", result)
	}
	if !bytes.Contains([]byte(result), []byte("line(0, 0, 200, 200)")) {
		t.Errorf("Expected third nagare code in result, got: %s", result)
	}

	t.Logf("Multiple blocks HTML: %s", result)
}

func TestNagareExtensionErrorHandling(t *testing.T) {
	// Create a mock HTTP server that returns a 400 error with error message
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid nagare syntax: expected 'circle' but got 'triangle'"))
	}))
	defer server.Close()

	// Create markdown with nagare extension
	md := goldmark.New(
		goldmark.WithExtensions(
			NewNagareExtension(server.URL),
		),
	)

	// Test markdown with nagare code block that will cause an error
	markdown := "# Test Page\n\nHere's a bad nagare block:\n\n```nagare\ntriangle(50, 50, 30)\n```\n\nRegular text."

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	result := buf.String()

	// Check that error message is displayed
	if !bytes.Contains([]byte(result), []byte("Error processing nagare block:")) {
		t.Errorf("Expected error message in result, got: %s", result)
	}

	// Check that the specific error message from server is included
	if !bytes.Contains([]byte(result), []byte("Invalid nagare syntax: expected 'circle' but got 'triangle'")) {
		t.Errorf("Expected server error message in result, got: %s", result)
	}

	// Check that the original nagare code is still shown in a code block
	if !bytes.Contains([]byte(result), []byte("triangle(50, 50, 30)")) {
		t.Errorf("Expected nagare code in fallback, got: %s", result)
	}

	// Check that it's wrapped in a code block
	if !bytes.Contains([]byte(result), []byte("<pre><code class=\"language-nagare\">")) {
		t.Errorf("Expected code block wrapper in result, got: %s", result)
	}

	t.Logf("Error handling HTML: %s", result)
}
