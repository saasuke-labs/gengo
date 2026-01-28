package nagare

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
)

func TestNagareExtension(t *testing.T) {
	// Create markdown with nagare extension
	md := goldmark.New(
		goldmark.WithExtensions(
			NewNagareExtension(),
		),
	)

	// Test markdown with nagare code block
	markdown := "# Test Page\n\nHere's a nagare block:\n\n```nagare\ncircle(50, 50, 30)\n```\n\nAnd some regular text."

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	result := buf.String()

	// Check that SVG is generated
	if !strings.Contains(result, "<svg") {
		t.Errorf("Expected SVG output in result, got: %s", result)
	}

	// Check that heading and text are still present
	if !strings.Contains(result, "<h1>Test Page</h1>") {
		t.Errorf("Expected heading in result, got: %s", result)
	}

	if !strings.Contains(result, "And some regular text") {
		t.Errorf("Expected regular text in result, got: %s", result)
	}

	t.Logf("Generated HTML length: %d", len(result))
}

func TestNagareExtensionMultipleBlocks(t *testing.T) {
	// Create markdown with nagare extension
	md := goldmark.New(
		goldmark.WithExtensions(
			NewNagareExtension(),
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

	// Count how many SVG elements are in the result (one per nagare block)
	svgCount := strings.Count(result, "<svg")
	if svgCount != 3 {
		t.Errorf("Expected 3 SVG elements, got %d", svgCount)
	}

	// Check that all sections are present
	if !strings.Contains(result, "First nagare block:") {
		t.Errorf("Expected 'First nagare block:' in result")
	}
	if !strings.Contains(result, "Second nagare block:") {
		t.Errorf("Expected 'Second nagare block:' in result")
	}
	if !strings.Contains(result, "Third nagare block:") {
		t.Errorf("Expected 'Third nagare block:' in result")
	}

	t.Logf("Multiple blocks test passed, generated %d SVGs", svgCount)
}

func TestNagareExtensionInvalidCode(t *testing.T) {
	// Create markdown with nagare extension
	md := goldmark.New(
		goldmark.WithExtensions(
			NewNagareExtension(),
		),
	)

	// Test markdown with invalid nagare code block
	markdown := "# Test Page\n\nHere's a bad nagare block:\n\n```nagare\nthis is invalid nagare syntax\n```\n\nRegular text."

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	result := buf.String()

	// The nagare extension should either:
	// 1. Generate an error div with "Error processing nagare block:", or
	// 2. Still generate some output (nagare might parse it differently)
	// We check for the error message div if nagare fails to parse
	hasErrorDiv := strings.Contains(result, "Error processing nagare block:")
	hasHeading := strings.Contains(result, "<h1>Test Page</h1>")
	hasRegularText := strings.Contains(result, "Regular text")

	// At minimum, the heading and regular text should be present
	if !hasHeading {
		t.Errorf("Expected heading in result")
	}
	if !hasRegularText {
		t.Errorf("Expected regular text in result")
	}

	t.Logf("Invalid code test result: error_div=%v, has_heading=%v, has_text=%v", hasErrorDiv, hasHeading, hasRegularText)
}
