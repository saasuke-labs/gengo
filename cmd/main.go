package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func main() {
	fmt.Println("Hello, World!")

	// Read the file passed as first argument
	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	// Read the file content
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	// Print the file content
	//fmt.Println(string(content))

	// Get AST from goldmark
	md := goldmark.New()
	reader := text.NewReader(content)

	doc := md.Parser().Parse(reader)
	if doc == nil {
		fmt.Println("Error parsing document")
		return
	}
	// Print the AST
	fmt.Println("AST:")
	printAST(doc, content, 0)

	html := toHTML(doc, content)

	fmt.Println("HTML:")
	fmt.Println(html)

	// Write html to the file passed as second argument
	if len(os.Args) > 2 {
		htmlFile, err := os.Create(os.Args[2])
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer htmlFile.Close()
		// Write the HTML content to the file
		_, err = htmlFile.WriteString(html)
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}
		fmt.Println("HTML written to", os.Args[2])
	} else {
		fmt.Println("No output file specified")
	}

}

func toHTML(n ast.Node, content []byte) string {
	html := ""
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {

		language := ""
		level := 0

		if node, ok := c.(*ast.FencedCodeBlock); ok {
			language = string(node.Language(content))
			fmt.Printf("%s (%s)", c.Kind().String(), language)
			html += fmt.Sprintf("<pre><code class=\"language-%s\">\n", language)
			html += string(node.Text(content))
			html += "</code></pre>\n"

		} else if node, ok := c.(*ast.Heading); ok {
			level = node.Level
			html += fmt.Sprintf("<h%d>%s</h%d>\n", level, c.Text(content), level)
		}

		// if !c.HasChildren() {
		// 	segment := c.Text(content)
		// 	if len(segment) > 0 {
		// 		fmt.Printf(": %q", segment)
		// 	}
		// }
		// fmt.Println()
		// fmt.Println()

		html += toHTML(c, content)
	}

	return html

}

func printAST(n ast.Node, content []byte, indent int) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {

		language := ""
		level := 0

		if node, ok := c.(*ast.FencedCodeBlock); ok {
			language = string(node.Language(content))
			fmt.Printf("%s%s (%s)", spaces(indent), c.Kind().String(), language)

		} else if node, ok := c.(*ast.Heading); ok {
			level = node.Level
			fmt.Printf("%s%s (%d)", spaces(indent), c.Kind().String(), level)
		} else {

			fmt.Printf("%s%s", spaces(indent), c.Kind().String())
		}

		if !c.HasChildren() {
			segment := c.Text(content)
			if len(segment) > 0 {
				fmt.Printf(": %q", segment)
			}
		}
		fmt.Println()
		fmt.Println()

		printAST(c, content, indent+2)
	}
}

func spaces(n int) string {
	return string(bytes.Repeat([]byte(" "), n))
}
