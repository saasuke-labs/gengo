package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWherePages_EmptyConditions(t *testing.T) {
	pages := []Page{
		{Title: "Page 1", Flags: []string{"pinned"}},
		{Title: "Page 2", Flags: []string{"archived"}},
	}

	result := wherePages("", pages)
	assert.Equal(t, pages, result)
}

func TestWherePages_SinglePositiveCondition(t *testing.T) {
	pages := []Page{
		{Title: "Page 1", Flags: []string{"pinned"}},
		{Title: "Page 2", Flags: []string{"archived"}},
		{Title: "Page 3", Flags: []string{"pinned", "featured"}},
	}

	result := wherePages("pinned", pages)
	assert.Len(t, result, 2)
	assert.Equal(t, "Page 1", result[0].Title)
	assert.Equal(t, "Page 3", result[1].Title)
}

func TestWherePages_SingleNegativeCondition(t *testing.T) {
	pages := []Page{
		{Title: "Page 1", Flags: []string{"pinned"}},
		{Title: "Page 2", Flags: []string{"archived"}},
		{Title: "Page 3", Flags: []string{}},
	}

	result := wherePages("!archived", pages)
	assert.Len(t, result, 2)
	assert.Equal(t, "Page 1", result[0].Title)
	assert.Equal(t, "Page 3", result[1].Title)
}

func TestWherePages_MultipleConditions(t *testing.T) {
	pages := []Page{
		{Title: "Page 1", Flags: []string{"pinned"}},
		{Title: "Page 2", Flags: []string{"pinned", "archived"}},
		{Title: "Page 3", Flags: []string{"pinned", "featured"}},
		{Title: "Page 4", Flags: []string{"archived"}},
	}

	// Pages with pinned AND NOT archived
	result := wherePages("pinned,!archived", pages)
	assert.Len(t, result, 2)
	assert.Equal(t, "Page 1", result[0].Title)
	assert.Equal(t, "Page 3", result[1].Title)
}

func TestWherePages_ComplexConditions(t *testing.T) {
	pages := []Page{
		{Title: "Page 1", Flags: []string{"pinned", "featured"}},
		{Title: "Page 2", Flags: []string{"pinned", "archived"}},
		{Title: "Page 3", Flags: []string{"featured"}},
		{Title: "Page 4", Flags: []string{}},
	}

	// Pages with pinned AND featured AND NOT archived
	result := wherePages("pinned,featured,!archived", pages)
	assert.Len(t, result, 1)
	assert.Equal(t, "Page 1", result[0].Title)
}

func TestWherePages_NoMatches(t *testing.T) {
	pages := []Page{
		{Title: "Page 1", Flags: []string{"pinned"}},
		{Title: "Page 2", Flags: []string{"archived"}},
	}

	result := wherePages("featured", pages)
	assert.Len(t, result, 0)
}

func TestWherePages_NoPagesWithoutFlags(t *testing.T) {
	pages := []Page{
		{Title: "Page 1", Flags: []string{}},
		{Title: "Page 2", Flags: []string{}},
	}

	result := wherePages("!pinned,!archived", pages)
	assert.Len(t, result, 2)
}

func TestMatchesConditions(t *testing.T) {
	tests := []struct {
		name       string
		pageFlags  []string
		conditions []string
		expected   bool
	}{
		{
			name:       "Empty conditions",
			pageFlags:  []string{"pinned"},
			conditions: []string{},
			expected:   true,
		},
		{
			name:       "Single positive match",
			pageFlags:  []string{"pinned"},
			conditions: []string{"pinned"},
			expected:   true,
		},
		{
			name:       "Single positive no match",
			pageFlags:  []string{"archived"},
			conditions: []string{"pinned"},
			expected:   false,
		},
		{
			name:       "Single negative match",
			pageFlags:  []string{"pinned"},
			conditions: []string{"!archived"},
			expected:   true,
		},
		{
			name:       "Single negative no match",
			pageFlags:  []string{"archived"},
			conditions: []string{"!archived"},
			expected:   false,
		},
		{
			name:       "Multiple conditions all match",
			pageFlags:  []string{"pinned", "featured"},
			conditions: []string{"pinned", "featured"},
			expected:   true,
		},
		{
			name:       "Multiple conditions one fails",
			pageFlags:  []string{"pinned"},
			conditions: []string{"pinned", "featured"},
			expected:   false,
		},
		{
			name:       "Mixed conditions match",
			pageFlags:  []string{"pinned", "featured"},
			conditions: []string{"pinned", "!archived"},
			expected:   true,
		},
		{
			name:       "Mixed conditions fail",
			pageFlags:  []string{"pinned", "archived"},
			conditions: []string{"pinned", "!archived"},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesConditions(tt.pageFlags, tt.conditions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"pinned", "featured", "archived"}

	assert.True(t, contains(slice, "pinned"))
	assert.True(t, contains(slice, "featured"))
	assert.True(t, contains(slice, "archived"))
	assert.False(t, contains(slice, "notfound"))
	assert.False(t, contains([]string{}, "pinned"))
}
