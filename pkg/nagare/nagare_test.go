package nagare

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
)

func newTestMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			NewNagareExtension(),
		),
	)
}

func TestNagareExtensionRendersModernDiagramDSL(t *testing.T) {
	md := newTestMarkdown()
	markdown := `# Test Page

Modern diagram:

` + "```nagare" + `
@layout(w:500,h:300)
client:Browser(url: "https://example.com", text: "Web App", x:50,y:100,w:180,h:120)
server:Server(title: "API Server", icon: "server", port: 8080, x:300,y:100,w:150,h:50)
client.e --> server.w
` + "```" + `
`

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	result := buf.String()
	if !strings.Contains(result, "<svg") {
		t.Fatalf("Expected SVG output in result, got: %s", result)
	}
	if !strings.Contains(result, "<polyline") {
		t.Errorf("Expected rendered connector in modern diagram output, got: %s", result)
	}
	if !strings.Contains(result, "<g transform=\"translate(50.000000,100.000000)\"") {
		t.Errorf("Expected translated component group in modern diagram output, got: %s", result)
	}
}

func TestNagareExtensionRendersCharts(t *testing.T) {
	md := newTestMarkdown()
	markdown := `# Chart Page

` + "```nagare" + `
chart
title: Test Chart
xaxis: number

series: test
color: #ff0000
data:
  0: 10
  1: 20
  2: 15
` + "```" + `
`

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	result := buf.String()
	if !strings.Contains(result, "<svg") {
		t.Fatalf("Expected SVG output in result, got: %s", result)
	}
	if !strings.Contains(result, "Test Chart") {
		t.Errorf("Expected chart title in rendered output, got: %s", result)
	}
	if !strings.Contains(result, "<path") {
		t.Errorf("Expected chart line path in rendered output, got: %s", result)
	}
}

func TestNagareExtensionMultipleBlocks(t *testing.T) {
	md := newTestMarkdown()
	markdown := `# Test Page

First nagare block:

` + "```nagare" + `
@layout(w:500,h:300)
client:Browser(url: "https://example.com", text: "Web App", x:50,y:100,w:180,h:120)
server:Server(title: "API Server", icon: "server", port: 8080, x:300,y:100,w:150,h:50)
client.e --> server.w
` + "```" + `

Some text in between.

Second nagare block:

` + "```nagare" + `
chart
title: Secondary Chart
xaxis: number

series: load
color: #22c55e
data:
  0: 1
  1: 3
  2: 2
` + "```" + `
`

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	result := buf.String()
	svgCount := strings.Count(result, "<svg")
	if svgCount != 2 {
		t.Errorf("Expected 2 SVG elements, got %d", svgCount)
	}
	if !strings.Contains(result, "Secondary Chart") {
		t.Errorf("Expected rendered chart title in result, got: %s", result)
	}
	if !strings.Contains(result, "<polyline") {
		t.Errorf("Expected rendered diagram connector in result, got: %s", result)
	}
}

func TestNagareExtensionInvalidCode(t *testing.T) {
	md := newTestMarkdown()
	markdown := "# Test Page\n\nHere's a bad nagare block:\n\n```nagare\nthis is invalid nagare syntax\n```\n\nRegular text."

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	result := buf.String()
	if !strings.Contains(result, "Error processing nagare block:") {
		t.Errorf("Expected error div in result, got: %s", result)
	}
	if !strings.Contains(result, "this is invalid nagare syntax") {
		t.Errorf("Expected original invalid block content in result, got: %s", result)
	}
	if !strings.Contains(result, "Regular text") {
		t.Errorf("Expected regular text in result")
	}
}
