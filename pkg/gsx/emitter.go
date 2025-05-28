package gsx

import (
	"bytes"
	"fmt"
	"html"
)

// RenderToGoTemplate renders a parsed GSX node to a Go template-compatible string.
func RenderToGoTemplate(node Node) (string, error) {
	var buf bytes.Buffer
	err := renderNode(&buf, node)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderNode(buf *bytes.Buffer, node Node) error {
	switch n := node.(type) {
	case *Text:
		buf.WriteString(n.Content)
	case *Element:
		if isComponent(n.Name) {
			// Render as Go template invocation
			buf.WriteString("{{ template \"" + n.Name + "\" (dict")
			for k, v := range n.Attributes {
				buf.WriteString(fmt.Sprintf(` "%s" "%s"`, k, v))
			}
			if len(n.Children) > 0 {
				var innerBuf bytes.Buffer
				for _, child := range n.Children {
					renderNode(&innerBuf, child)
				}
				buf.WriteString(fmt.Sprintf(` "inner" (html "%s")`, innerBuf.String()))
			}
			buf.WriteString(") }}")
		} else {
			// Render as normal HTML tag
			buf.WriteString("<" + n.Name)
			for k, v := range n.Attributes {
				buf.WriteString(fmt.Sprintf(` %s="%s"`, k, html.EscapeString(v)))
			}
			buf.WriteString(">")
			for _, child := range n.Children {
				renderNode(buf, child)
			}
			buf.WriteString("</" + n.Name + ">")
		}
	default:
		return fmt.Errorf("unknown node type: %T", n)
	}
	return nil
}

func isComponent(name string) bool {
	return name != "" && name[0] >= 'A' && name[0] <= 'Z'
}
