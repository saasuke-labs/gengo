package testutils

import (
	"fmt"
	"os"

	"golang.org/x/net/html"
)

func CompareHtmlFiles(file1, file2 string) bool {
	n1, err := ParseHTMLFile(file1)
	if err != nil {
		return false
	}
	n2, err := ParseHTMLFile(file2)
	if err != nil {
		return false
	}

	fmt.Println("File1:", n1)
	fmt.Println("File2:", n2)

	return CompareNodes(n1, n2)
}

// ParseHTMLFile reads and parses an HTML file into a DOM node
func ParseHTMLFile(path string) (*html.Node, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return html.Parse(f)
}

func CompareNodes(n1, n2 *html.Node) bool {
	// Skip whitespace-only text nodes
	for n1 != nil && isIgnorableText(n1) {
		n1 = n1.NextSibling
	}
	for n2 != nil && isIgnorableText(n2) {
		n2 = n2.NextSibling
	}

	if n1 == nil || n2 == nil {
		return n1 == nil && n2 == nil
	}

	if n1.Type != n2.Type || n1.Data != n2.Data {
		return false
	}

	if len(n1.Attr) != len(n2.Attr) {
		return false
	}
	attrMap := map[string]string{}
	for _, attr := range n1.Attr {
		attrMap[attr.Key] = attr.Val
	}
	for _, attr := range n2.Attr {
		if val, ok := attrMap[attr.Key]; !ok || val != attr.Val {
			return false
		}
	}

	// Compare children
	c1 := n1.FirstChild
	c2 := n2.FirstChild
	for {
		// Skip ignorable text nodes
		for c1 != nil && isIgnorableText(c1) {
			c1 = c1.NextSibling
		}
		for c2 != nil && isIgnorableText(c2) {
			c2 = c2.NextSibling
		}

		if c1 == nil || c2 == nil {
			return c1 == nil && c2 == nil
		}

		if !CompareNodes(c1, c2) {
			return false
		}

		c1 = c1.NextSibling
		c2 = c2.NextSibling
	}

	return true
}

// Helper: ignore whitespace-only text nodes
func isIgnorableText(n *html.Node) bool {
	return n.Type == html.TextNode && isAllWhitespace(n.Data)
}

func isAllWhitespace(s string) bool {
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}
