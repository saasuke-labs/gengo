package gsx

import (
	"fmt"
	"strings"
)

// Node represents a GSX element or raw content.
type Node interface{}

type Element struct {
	Name       string
	Attributes []Attribute // in source order, so output is deterministic
	Children   []Node
}

// Attribute is one name="value" pair on an element.
type Attribute struct {
	Name  string
	Value string
}

type Text struct {
	Content string
}

// ParseGSX parses a simple GSX string and returns a root node.
func ParseGSX(input string) (Node, error) {
	p := &parser{input: input}
	return p.parse()
}

type parser struct {
	input string
	pos   int
}

func (p *parser) parse() (Node, error) {
	p.skipWhitespace()
	if p.peek() != '<' {
		return p.parseText()
	}
	return p.parseElement()
}

func (p *parser) parseElement() (Node, error) {
	if !p.consume("<") {
		return nil, fmt.Errorf("expected '<'")
	}
	name := p.readIdentifier()
	if name == "" {
		return nil, fmt.Errorf("missing tag name")
	}
	attrs := p.readAttributes()

	selfClosing := p.consume("/>")
	if selfClosing {
		return &Element{Name: name, Attributes: attrs}, nil
	}

	if !p.consume(">") {
		return nil, fmt.Errorf("expected '>'")
	}

	var children []Node
	for !p.startsWith("</" + name + ">") {
		child, err := p.parse()
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	p.consume("</" + name + ">")
	return &Element{Name: name, Attributes: attrs, Children: children}, nil
}

func (p *parser) parseText() (Node, error) {
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] != '<' {
		p.pos++
	}
	return &Text{Content: p.input[start:p.pos]}, nil
}

func (p *parser) readIdentifier() string {
	start := p.pos
	for p.pos < len(p.input) && (isAlphaNum(p.input[p.pos]) || p.input[p.pos] == '-') {
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *parser) readAttributes() []Attribute {
	var attrs []Attribute
	for {
		p.skipWhitespace()
		if p.peek() == '/' || p.peek() == '>' {
			break
		}
		key := p.readIdentifier()
		p.skipWhitespace()
		if !p.consume("=") {
			break
		}
		p.skipWhitespace()
		attrs = setAttribute(attrs, key, p.readQuotedString())
	}
	return attrs
}

// setAttribute adds an attribute, or replaces the value of one already
// present so that, as before, the last occurrence of a name wins.
func setAttribute(attrs []Attribute, name, value string) []Attribute {
	for i := range attrs {
		if attrs[i].Name == name {
			attrs[i].Value = value
			return attrs
		}
	}
	return append(attrs, Attribute{Name: name, Value: value})
}

func (p *parser) readQuotedString() string {
	if !p.consume(`"`) {
		return ""
	}
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] != '"' {
		p.pos++
	}
	val := p.input[start:p.pos]
	p.consume(`"`)
	return val
}

func (p *parser) consume(s string) bool {
	if strings.HasPrefix(p.input[p.pos:], s) {
		p.pos += len(s)
		return true
	}
	return false
}

func (p *parser) startsWith(s string) bool {
	return strings.HasPrefix(p.input[p.pos:], s)
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\n') {
		p.pos++
	}
}

func (p *parser) peek() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func isAlphaNum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
