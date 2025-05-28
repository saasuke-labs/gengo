package gsx

import "os"

type Options struct {
	Indent bool   // whether to pretty-print output
	Target string // e.g. "gotemplate" (future)
}

func ParseString(input string, opts *Options) (string, error) {
	node, err := ParseGSX(input)

	if err != nil {
		return "", err
	}
	tmpl, err := RenderToGoTemplate(node)

	if err != nil {
		return "", err
	}

	return tmpl, nil

}

func ParseFile(path string, opts *Options) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return ParseString(string(raw), opts)
}
