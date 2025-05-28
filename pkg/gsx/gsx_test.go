package gsx

import (
	"testing"
)

func TestParseSimpleComponent(t *testing.T) {
	input := `<Tag name="Go" />`
	expected := `{{ template "Tag" (dict "name" "Go") }}`

	output, err := ParseString(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, output)
	}
}

func TestParseMixedHTMLAndComponent(t *testing.T) {
	input := `<div><Tag name="Go" /></div>`
	expected := `<div>{{ template "Tag" (dict "name" "Go") }}</div>`

	output, err := ParseString(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, output)
	}
}

func TestParseMultipleAttributes(t *testing.T) {
	input := `<PostCard title="Hello" author="Ana" />`
	expected := `{{ template "PostCard" (dict "title" "Hello" "author" "Ana") }}`

	output, err := ParseString(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, output)
	}
}

func TestParseComponentWithChildren(t *testing.T) {
	input := `<Card><p>Hello</p></Card>`
	expected := `{{ template "Card" (dict "inner" (html "<p>Hello</p>")) }}`

	output, err := ParseString(input, nil)

	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if output != expected {
		t.Errorf("expected:\\n%q\\ngot:\\n%q", expected, output)
	}
}
