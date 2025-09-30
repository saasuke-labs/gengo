package main

import (
	"fmt"

	"github.com/saasuke-labs/gengo/pkg/parser"
)

func main() {
	fmt.Println("Testing gengo parser with nagare content...")

	result := parser.MarkdownToHtml("test-nagare.md")

	fmt.Printf("Title: %s\n", result.Title)
	fmt.Printf("HTML: %s\n", result.HTML)
}
